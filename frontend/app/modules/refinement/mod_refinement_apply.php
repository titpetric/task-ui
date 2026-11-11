<?php

// @route POST /refinement/threads/{id}/apply

require "ext/autoload.php";

// mod_refinement_apply folds every rule the thread still holds into one request
// and runs it over the artifacts that are already there.
//
// A thread runs as a sequence: each correction is a run over what the run
// before it left, which drifts. By the tenth, the first correction is a thing
// nobody has read in an hour. This turns the sequence into a statement: here is
// the design, here is every rule, make the artifacts satisfy all of it at once.
//
// What it records is a comment of its own, which is neither the design nor a
// rule, with every request standing before it folded into it. The rules
// survive, because a fold is not a deletion: the next Apply all composes out of
// the same comments this one did.
class mod_refinement_apply implements Module {
	/** BOARD is the thread list, which a thread's own path hangs off. */
	const BOARD = "/refinement";

	/** MISSING is what the page of a thread that is not there says. */
	const MISSING = "No thread with that ID. Someone may have deleted its design, which deletes the thread.";

	/** NOTHING is what a thread whose design was deleted is told. */
	const NOTHING = "This thread has no design left to apply its rules to.";

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
		$design = $store->design($id);

		// A thread whose design was deleted is a thread with no design, and a
		// rule corrects something: there is nothing here to apply it to.
		if (!is_array($design)) {
			Task\Route::fail($to, self::NOTHING);

			return;
		}

		$rules = Task\Refinement\Prompt::rules($store->rules($id));
		$body = Task\Refinement\Prompt::flatten(Task\Row::text($design, "body"), $rules);
		$comment = $store->apply($id, $chrome->name(), $body);

		Task\Refinement\Work::start($store, $id, $comment, $body);
		Task\Route::go($to);
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_refinement_apply;

	$mod->serve();
}
