<?php

// @route GET /orb

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_orb lists the rooms of the orb: the ones taking participants and the
// ones that ended with Claude's summary.
//
// The page is read only. New meeting and every room's own actions call
// /api/meetings from the browser, so the markup sets the data-* attributes
// assets/orb.js reads and nothing here writes.
class mod_orb implements Module {
	/** LEAD is the paragraph under the heading. */
	const LEAD = "Talk in a room that keeps each speaker apart: live voice relayed by task-ui, a transcript under the name of who said it, and Claude's summary once the room ends.";

	/** NONE is what the block says with no rooms. */
	const NONE = "No meetings yet. Start one and send the link to whoever should join.";

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
		$counts = $store->counts();
		$rows = [];

		foreach ($store->all(40) as $row) {
			$rows[] = $this->row($chrome, $row, $counts);
		}

		$head = new Task\Head("Meetings", self::LEAD);

		$view->page([
			"head" => $head->data(),
			"rows" => $rows,
			"create" => $chrome->allowed("meeting/edit"),
			"block" => [
				"title" => "Meetings",
				"note" => "",
				"empty" => count($rows) === 0 ? self::NONE : "",
				"buttons" => ["items" => []],
			],
		]);
	}

	/** row is one room as the table draws it. */
	function row($chrome, $meeting, $counts) {
		$id = $meeting->getId();
		$me = $chrome->user;
		$open = $meeting->getState() === "open";
		$here = isset($counts[$id]) ? $counts[$id] : 0;
		$mine = $me !== null && $meeting->getUserId() === $me->getId();

		return [
			"id" => $id,
			"title" => $meeting->getTitle(),
			"open" => $open,
			"state" => $open ? "live" : "ended",
			"ago" => Task\Clock::ago($meeting->getStartedAt()),
			"here" => $here,
			"user" => $chrome->who()->of($meeting->getUserName()),
			"action" => $open ? "Open" : "Summary",
			"delete" => !$open && $chrome->allowed("meeting/delete") && ($mine || Task\User::admin($me)),
		];
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_orb;

	$mod->serve();
}
