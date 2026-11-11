<?php

// @route POST /refinement/threads/{id}/comments/{comment}/delete

require "ext/autoload.php";

// mod_refinement_delete retires one rule, or the whole thread when the comment
// is the design the rules correct.
//
// The artifacts stay where they are either way. A rule is a thing the thread
// asks for and dropping the ask is what this does; the folder is the artifact
// tree's, which the instance's own artifacts view clears.
class mod_refinement_delete implements Module {
	/** BOARD is the thread list, where deleting a design sends the reader. */
	const BOARD = "/refinement";

	/** MISSING is what the page of a thread that is not there says. */
	const MISSING = "No thread with that ID. Someone may have deleted its design, which deletes the thread.";

	function name() {
		return "refinement";
	}

	function title() {
		return "Refinement";
	}

	function icon() {
		return "flow";
	}

	function tables() {
		return array("refinement_thread", "refinement_comment", "refinement_correction");
	}

	function serve() {
		$db = Task\DB::writer();
		$chrome = new Task\Chrome($db);
		$view = new Task\View($this, $chrome);

		// Changing a thread takes what starting its runs takes: a reader who
		// may not prompt may read a thread and may not delete what it asks for.
		if (!$chrome->allowed("module") || !$chrome->allowed("prompt/edit")) {
			$view->refused();

			return;
		}

		$store = new Task\Refinement\Threads($db);
		$thread = $store->find(Task\Route::param("id"));

		if (!is_array($thread)) {
			$view->missing(self::MISSING);

			return;
		}

		$id = Task\Row::text($thread, "id");
		$comment = Task\Route::param("comment");
		$design = $store->design($id);

		// The design is what the rules correct, so removing it removes the
		// thread: nothing is left for a rule to be a rule about.
		if (is_array($design) && Task\Row::text($design, "id") === $comment) {
			$store->drop($id);
			Task\Route::go(self::BOARD);

			return;
		}

		$store->remove($id, $comment);
		Task\Route::go(self::BOARD . "/threads/" . Task\Route::pathEscape($id));
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_refinement_delete;

	$mod->serve();
}
