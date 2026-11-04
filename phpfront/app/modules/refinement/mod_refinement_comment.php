<?php

// @route POST /refinement/threads/{id}/comments

require "ext/autoload.php";

// mod_refinement_comment is a correction: one change to what the run before it
// produced, and the run that applies it over the artifacts as they stand.
//
// Every comment after a thread's first is a correction, which is what separates
// "design this" from "use ULID as primary keys": the row says which comment it
// corrects, and Apply all composes its request out of them.
class mod_refinement_comment implements Module {
	/** BOARD is the thread list, which a thread's own path hangs off. */
	const BOARD = "/refinement";

	/** MISSING is what the page of a thread that is not there says. */
	const MISSING = "No thread with that ID. Someone may have deleted its design, which deletes the thread.";

	/** WANTED is what the box asks for. */
	const WANTED = "A correction says one thing to change.";

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
		$to = self::BOARD . "/threads/" . Task\Route::pathEscape($id);
		$body = trim(Task\Route::input("comment"));

		if ($body === "") {
			Task\Route::fail($to, self::WANTED);

			return;
		}

		$comment = $store->correct($id, $chrome->name(), $body);

		Task\Refinement\Work::start($store, $id, $comment, $body);
		Task\Route::go($to);
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_refinement_comment;

	$mod->serve();
}
