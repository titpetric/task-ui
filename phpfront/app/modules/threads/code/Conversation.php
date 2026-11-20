<?php

namespace Task\Threads;

// Conversation is one thread as the pane reads it: its title, the messages,
// the agent picker, the reply box and the actions.
//
// Both the thread's own page and the right-hand pane of the Threads page read
// this one scope, so a swap of the pane and a reload of the page draw the same
// conversation and enforce the same permissions.
class Conversation {
	/** MISSING is what a pane says for a thread that is not there. */
	const MISSING = "No thread with that ID. Someone may have deleted it.";

	/** view is one thread as pane.tpl reads it, or an error the pane shows. */
	static function view($note, $chrome, $store, $split = false) {
		$prompt = \Task\Note::isPrompt($note);

		if ($prompt && (!$chrome->allowed("prompt/view") || !\Task\Note::seesPlace($chrome, $note->getDir()))) {
			return ["error" => \Task\View::REFUSED . \Task\View::GRANTED];
		}

		if (!$prompt && !\Task\Note::isChat($note) && !$chrome->allowed("note/view")) {
			return ["error" => \Task\View::REFUSED . \Task\View::GRANTED];
		}

		$chat = \Task\Note::isChat($note);
		$userID = $chrome->user === null ? "" : $chrome->user->getID();
		$member = $chat && $store->participant($note->getId(), $userID);

		if ($chat && !$member && !\Task\User::admin($chrome->user)) {
			return ["error" => self::MISSING];
		}

		$replies = $store->replies([$note->getId()]);
		$under = isset($replies[$note->getId()]) ? $replies[$note->getId()] : [];
		$title = \Task\Note::excerpt(Talk::opening($note, $under));
		$work = Work::byQuestion($store->events($note->getId()), $chrome);
		$users = new \Task\User;
		$agentNames = [];

		foreach ($users->agents() as $user) {
			$agentNames[] = $user->getName();
		}

		$agent = $note->getAgent();

		if ($agent === "" && $prompt) {
			$agent = count($agentNames) > 0 ? $agentNames[0] : "claude";
		}

		$menu = self::menu($note, $chrome);
		$pick = !$chat && count($agentNames) > 0 && $note->getState() !== \Task\Note::WORKING && ($chrome->name() === $note->getUserName() || \Task\User::admin($chrome->user));
		$head = new \Task\Head($title);

		$head->menu($menu->data());
		$seen = new \Task\Seen($chrome);

		$pane = [
			"id" => $note->getId(),
			"title" => $title,
			"head" => $head->data(),
			"head_view" => [
				"id" => $note->getId(),
				"title" => $title,
				"meta" => self::meta($note, $agent, $chat),
				"agents" => $agentNames,
				"agent" => $agent,
				"can_pick" => $pick,
				"menu" => $menu->data(),
			],
			"status" => self::status($note, $work, $chrome, $under),
			"split" => $split,
			"channel" => $note->getChannel(),
			"chat_thread" => $chat,
			"prompt" => $prompt,
			"last_event" => $store->lastEvent($note->getId()),
			"messages" => Talk::messages($note, $under, $chrome, $seen, $work, self::answers($under, $work)),
			"agent" => $agent,
			"agent_names" => $agentNames,
			"pick_agent" => $pick,
			"editor" => self::field(),
			"replies" => $chat ? $member && $chrome->allowed("note/edit") : ($prompt ? $chrome->allowed("prompt/edit") : $chrome->allowed("note/edit")),
			"mic" => $chrome->allowed("note/edit"),
		];
		$pane["composer"] = \Task\Composer::reply($pane);

		return $pane;
	}

	/** meta is the line under a title: what the thread is, where, and its state. */
	static function meta($note, $agent, $chat) {
		$parts = [];

		if ($chat) {
			$parts[] = "chat";
		} elseif (\Task\Note::isPrompt($note)) {
			$parts[] = "prompt";
		} else {
			$parts[] = "note";
		}

		if ($note->getDir() !== "") {
			$parts[] = $note->getDir();
		}

		if ($agent !== "" && !$chat) {
			$parts[] = "@" . $agent;
		}

		if ($note->getState() !== "") {
			$parts[] = $note->getState();
		}

		return implode(" · ", $parts);
	}

	/**
	 * answers is the reply each run wrote, by note ID: the message that carries
	 * the run, after the question that started it. A run that ended without a
	 * reply names none.
	 */
	static function answers($replies, $work) {
		$started = [];

		foreach ($work as $runs) {
			foreach ($runs as $run) {
				if ($run["id"] !== "" && $run["question"] !== "") {
					$started[$run["id"]] = $run["question"];
				}
			}
		}

		$out = [];

		foreach ($replies as $reply) {
			$run = $reply->getRunId();
			$question = isset($started[$run]) ? $started[$run] : "";

			if ($question !== "" && $reply->getId() !== $question) {
				$out[$reply->getId()] = true;
			}
		}

		return $out;
	}

	/** status is the pinned line of a working thread, with the run it may end. */
	static function status($note, $work, $chrome, $replies = []) {
		if ($note->getState() !== \Task\Note::WORKING) {
			return [
				"working" => false,
				"label" => "",
				"state" => "",
				"started" => "",
				"run" => "",
				"can_end" => false,
			];
		}

		$label = "Working";
		$state = "working";
		$id = "";

		foreach ($work as $runs) {
			foreach ($runs as $run) {
				if ($run["id"] !== "") {
					$id = $run["id"];
				}

				if (($run["status"]["body"] ?? "") !== "") {
					$label = $run["status"]["body"];
					$state = ($run["status"]["state"] ?? "") !== "" ? $run["status"]["state"] : "working";
				}
			}
		}

		if ($id === "") {
			$id = $note->getRunId();

			foreach ($replies as $reply) {
				if ($reply->getRunId() !== "") {
					$id = $reply->getRunId();
				}
			}
		}

		$runs = new \Task\Runs\Runs;
		$row = $runs->get($id);
		$started = $row === null ? "" : \Task\Run::started($row->getStartedAt());

		if ($started === "") {
			$at = \Task\Clock::parse($note->getUpdatedAt());
			$started = $at === null ? "" : (string)$at->unixMilli();
		}

		return [
			"working" => true,
			"label" => $label,
			"state" => $state,
			"started" => $started,
			"run" => $id,
			"can_end" => $row !== null && !\Task\Run::ended($row->getState()) && \Task\Run::canEnd($chrome, $row),
		];
	}

	/** menu is the thread's own actions: a thread is deleted, never rewritten. */
	static function menu($note, $chrome) {
		$menu = new \Task\Menu("Actions on this thread");

		$menu->whenDanger(\Task\Note::deletes($note, $chrome), "trash", "Delete", [
			"data-delete" => "/api/notes/" . $note->getId(),
			"data-confirm" => "Delete this thread and every message in it? It leaves every listing; nobody reads it again.",
			"data-confirm-ok" => "Delete",
			"data-next" => \Task\Route::href("/threads"),
		]);

		return $menu;
	}

	/** field is the box under the conversation, as components/editor.tpl reads it. */
	static function field() {
		return \Task\Note::field();
	}
}
