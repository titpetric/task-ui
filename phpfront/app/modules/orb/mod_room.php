<?php

// @route GET /orb/{id}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_room is one meeting: who is in it, the live transcript under the name of
// each speaker, and once it ends Claude's summary above the whole log.
//
// The page is read only. Joining, speaking, leaving and finishing call
// /api/meetings and its websocket from the browser, which is assets/orb.js.
class mod_room implements Module {
	/** GONE is what the page says for an ID no meeting holds. */
	const GONE = "No meeting with that ID. It may have been deleted.";

	function name() {
		return "orb";
	}

	function title() {
		return "Meetings";
	}

	function icon() {
		return "mic";
	}

	function tables() {
		return ["meeting", "meeting_line", "meeting_participant", "user"];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("meeting/view")) {
			$view->refused();

			return;
		}

		$store = new Task\Orb\Meetings($chrome->db);
		$meeting = $store->get(Task\Route::param("id"));

		if ($meeting === null) {
			$view->missing(self::GONE);

			return;
		}

		$title = $meeting->getTitle();
		$chrome->heading($title, [["text" => "Meetings", "href" => "/orb"], ["text" => $title, "href" => ""]]);

		$people = $this->people($chrome, $store->participants($meeting->getId()));
		$me = $chrome->user;
		$host = $me !== null && $meeting->getUserId() === $me->getId();
		$admin = Task\User::admin($me);
		$open = $meeting->getState() === "open";
		$summary = trim($meeting->getSummary());
		$users = new Task\User;
		$agentNames = [];

		foreach ($users->agents() as $user) {
			$agentNames[] = $user->getName();
		}

		$agent = $meeting->getAgent();

		if ($agent !== "") {
			$people = $this->withAgent($people, $meeting->getAgentId(), $agent);
		}

		$view->page([
			"head" => $this->head($chrome, $meeting, $people, $open, $host || $admin),
			"room" => [
				"id" => $meeting->getId(),
				"title" => $title,
				"open" => $open,
				"state" => $meeting->getState(),
				"self" => $me !== null ? $me->getId() : "",
				"self_name" => $chrome->name(),
				"host" => $host,
				"admin" => $admin,
				"can_edit" => $chrome->allowed("meeting/edit"),
				"can_delete" => $chrome->allowed("meeting/delete"),
				"started" => Task\Clock::seconds($meeting->getStartedAt()),
				"limit" => (int)$meeting->getLimitSeconds(),
				"summary" => $summary === "" ? "" : $chrome->markdown($summary),
				"summary_state" => $meeting->getSummaryState(),
				"summary_message" => $meeting->getSummaryMessage(),
				"summary_user" => $meeting->getSummaryUserName(),
				"agent" => $agent,
				"agent_names" => $agentNames,
				"pick_agent" => $open && $chrome->allowed("meeting/edit") && ($host || $admin),
				"ended" => Task\Clock::ago($meeting->getEndedAt()),
			],
			"people" => $people,
			"lines" => $this->lines($meeting, $store->lines($meeting->getId())),
		], "room.tpl");
	}

	/**
	 * head is the top of the page: what the room is, and what may be done to
	 * it by this reader.
	 */
	function head($chrome, $meeting, $people, $open, $host) {
		$head = new Task\Head($meeting->getTitle());
		$head->icon("mic");
		$head->badge($open ? "live" : "ended", $open ? "ok" : "");

		$here = count($people);
		$said = "Started " . Task\Clock::ago($meeting->getStartedAt()) . " by " . $meeting->getUserName();

		if (!$open) {
			$said .= ", " . ($here === 1 ? "one person joined" : $here . " people joined");
		}

		$head->faint($said);

		$buttons = new Task\Buttons;
		$buttons->item("copy", "Copy link", ["data-orb-copy" => $meeting->getId()]);

		if ($open && $chrome->allowed("meeting/edit")) {
			$buttons->item("stop", "Leave", ["data-orb-leave" => $meeting->getId(), "data-joined" => "0"]);
		}

		if ($open && $host) {
			$buttons->danger("stop", "Finish", [
				"data-orb-finish" => $meeting->getId(),
				"data-confirm" => "End the meeting and summarise the transcript?",
			]);
		}

		if (!$open && $host && $chrome->allowed("meeting/delete")) {
			$buttons->danger("trash", "Delete", [
				"data-orb-delete" => $meeting->getId(),
				"data-confirm" => "Delete this meeting, its transcript and its audio?",
			]);
		}

		$head->buttons($buttons->data());

		return $head->data();
	}

	/** people is the participants as the room draws them, the one reading marked. */
	function people($chrome, $participants) {
		$me = $chrome->user;
		$out = [];

		foreach ($participants as $person) {
			$here = $person->getLeftAt() === "";

			$self = $me !== null && $person->getUserId() === $me->getId();

			$out[] = [
				"id" => $person->getUserId(),
				"name" => $person->getUserName(),
				"className" => self::className($self, $here),
				"here" => $here,
				"self" => $self,
				"joins" => (int)$person->getJoins(),
			];
		}

		return $out;
	}

	/** withAgent puts the room's agent in the participant rail. */
	function withAgent($people, $id, $name) {
		array_unshift($people, [
			"id" => $id,
			"name" => $name,
			"className" => self::className(false, true),
			"here" => true,
			"self" => false,
			"joins" => 0,
		]);

		return $people;
	}

	/** className is the class of a participant: one reading, and one here or gone. */
	static function className($self, $here) {
		$classes = ["orb-person"];

		if ($self) {
			$classes[] = "self";
		}

		$classes[] = $here ? "here" : "gone";

		return implode(" ", $classes);
	}

	/** lines is a room's log as it draws it: what a speaker said, and what happened. */
	function lines($meeting, $lines) {
		$out = [];

		foreach ($lines as $line) {
			$out[] = [
				"id" => $line->getId(),
				"kind" => $line->getKind(),
				"said" => $line->getKind() === "said",
				"name" => $line->getUserName(),
				"user" => $line->getUserId(),
				"body" => $line->getBody(),
				"start" => (int)$line->getStartMs(),
				"end" => (int)$line->getEndMs(),
				"at" => Task\Clock::format($line->getCreatedAt(), "15:04"),
				"created" => $line->getCreatedAt(),
			];
		}

		return $out;
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_room;
	$mod->serve();
}
