<?php

// @route GET /notes
// @route POST /notes

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_notes is the notes board: the notes newest first, each with its
// replies, and the agent a note names to start work.
class mod_notes implements Module {
	function name() {
		return "notes";
	}

	function title() {
		return "Notes";
	}

	function icon() {
		return "note";
	}

	function tables() {
		return array("note", "user");
	}

	function serve() {
		$chrome = new Task\Chrome;

		if (!$chrome->allowed("note/view")) {
			Task\Route::refused();
			$chrome->heading("Not allowed");

			(new Task\View($this, $chrome))->page(array("refused" => true), "forbidden.tpl");

			return;
		}

		if (Task\Route::posted()) {
			$this->post($chrome);

			return;
		}

		$chrome->heading("Notes", array(array("text" => "Notes", "href" => "")));

		$store = new Task\Notes\NoteStore;
		$replies = $store->replies(Task\Notes\NoteStore::NOTES);
		$agents = (new Task\User)->agents();
		$threads = array();

		foreach ($store->threads(Task\Notes\NoteStore::NOTES) as $note) {
			$threads[] = $this->thread($note, $replies, $chrome);
		}

		(new Task\View($this, $chrome))->page(array(
			"agent" => count($agents) > 0 ? $agents[0]->getName() : "",
			"threads" => $threads,
			"edits" => $chrome->allowed("note/edit"),
		));
	}

	/** post writes a note the board's form sent, and answers with a redirect. */
	function post($chrome) {
		if (!$chrome->allowed("note/edit")) {
			Task\Route::refused();

			return;
		}

		$body = Task\Route::input("body");

		if ($body === "") {
			Task\Route::fail("/notes", "A note says something.");

			return;
		}

		$note = (new Task\Notes\NoteStore(Task\DB::writer()))->post($chrome->name(), $body, $this->mentions($body), Task\Route::input("parent"));

		Task\Host::publish("notes", array("event" => "note", "id" => $note->getId()));
		Task\Route::go("/notes");
	}

	/** thread is one note with its replies, as the note component reads it. */
	function thread($note, $replies, $chrome) {
		$out = $this->one($note, $chrome);

		$out["replyable"] = $chrome->allowed("note/edit");
		$out["replies"] = array();

		$id = $note->getId();

		if (!isset($replies[$id])) {
			return $out;
		}

		foreach ($replies[$id] as $reply) {
			$out["replies"][] = $this->one($reply, $chrome);
		}

		return $out;
	}

	/** one is a note or a reply, with the menu of what the reader may do to it. */
	function one($note, $chrome) {
		$id = $note->getId();
		$author = $note->getUserName();
		$run = $note->getRunId();
		$menu = new Task\Menu("Actions on the note of " . $author);

		$menu->when($chrome->allowed("note/edit") && $chrome->name() === $author, "edit", "Edit", array("data-edit-open" => $id));
		$menu->when($run !== "", "terminal", "Open its run", array("href" => Task\Route::href("/runs/" . $run)));
		$menu->whenDanger($chrome->allowed("note/delete"), "trash", "Delete", array(
			"data-delete" => "/api/notes/" . $id,
			"data-confirm" => "Delete the note of " . $author . " and its replies?",
			"data-confirm-ok" => "Delete",
			"data-next" => "reload",
		));

		return array(
			"id" => $id,
			"body" => $chrome->message($note->getBody()),
			"author" => $chrome->who()->of($author),
			"state" => $note->getState(),
			"run" => $run,
			"short" => substr($run, 0, 8),
			"ago" => Task\Clock::ago($note->getCreatedAt()),
			"edited" => $note->getEditedAt() !== "",
			"menu" => $menu->data(),
		);
	}

	/**
	 * mentions is the names a note names, as the column holds them: each one
	 * spaced, so a LIKE for "% name %" matches a whole name alone.
	 */
	function mentions($body) {
		$found = array();
		// The match is its own statement. An out parameter written by a call
		// that is an operand of && is not visible to the operand after it, so
		// `preg_match_all(..., $names) && isset($names[1])` reads $names as
		// unset. See docs/todo-correction.md.
		$count = preg_match_all('/@([A-Za-z0-9._-]+)/', $body, $names);

		if ($count > 0 && isset($names[1])) {
			foreach ($names[1] as $name) {
				if (!in_array($name, $found, true)) {
					$found[] = $name;
				}
			}
		}

		if (count($found) === 0) {
			return "";
		}

		return " " . implode(" ", $found) . " ";
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	(new mod_notes)->serve();
}
