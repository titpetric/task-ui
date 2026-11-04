<?php

// @route POST /refinement/threads

require "ext/autoload.php";

// mod_refinement_start opens a thread from the paragraph the board's box holds,
// and starts the run that answers it.
//
// The answer is a redirect to the thread, so a reload does not start a second
// run: the reader lands on the page that shows the one that is working.
//
// Starting a run takes prompt/edit, which is what the board enforces when the
// dispatch reaches it. The gate is here as well so a reader who may not prompt
// is refused the page rather than left with a thread whose run never started.
class mod_refinement_start implements Module {
	/** BOARD is where a failure sends the reader back to, and where the form posted from. */
	const BOARD = "/refinement";

	/** CORRECTION is what a correction that lost its thread is told. */
	const CORRECTION = "That is a correction and it needs its thread. Open the thread and send it from the box at the bottom.";

	/** WANTED is what the box asks for. */
	const WANTED = "A thread starts from one paragraph of what to design.";

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
		return array("refinement_thread", "refinement_comment");
	}

	function serve() {
		$db = Task\DB::writer();
		$chrome = new Task\Chrome($db);

		if (!$chrome->allowed("module") || !$chrome->allowed("prompt/edit")) {
			$view = new Task\View($this, $chrome);

			$view->refused();

			return;
		}

		// A correction that arrives here is a correction that lost its thread:
		// the comment form posts to a thread's own comments and only the form
		// that opens a thread posts here. Opening a thread from it would bury
		// the correction in a thread of its own, with none of the work it was
		// correcting, so this says what happened instead.
		if (Task\Route::input("comment") !== "") {
			Task\Route::fail(self::BOARD, self::CORRECTION);

			return;
		}

		$want = trim(Task\Route::input("want"));

		if ($want === "") {
			Task\Route::fail(self::BOARD, self::WANTED);

			return;
		}

		$store = new Task\Refinement\Threads($db);
		$opened = $store->start($chrome->name(), $want);

		Task\Refinement\Work::start($store, $opened["id"], $opened["comment_id"], $want);
		Task\Route::go(self::BOARD . "/threads/" . Task\Route::pathEscape($opened["id"]));
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_refinement_start;

	$mod->serve();
}
