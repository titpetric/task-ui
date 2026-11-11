<?php

// @route POST /refinement/threads/{id}/fork

require "ext/autoload.php";

// mod_refinement_fork opens a thread on a new subject, carrying the rules of
// one that already works.
//
// A thread's corrections generalise. "Every id is a ULID", "no foreign keys",
// "the diagram is generated and the generation is repeatable" are not about
// users and their logins; they are how this person wants the work done, and
// typing them again for the next subject is the thing worth not doing.
//
// So a fork is the subject and the rules: the new thread opens on what was
// typed, every rule of the old one is copied into it as a rule of its own, and
// the run it starts is the flattened request. It is the first run of a new
// thread and it starts on an empty folder, because the rules travel and the
// artifacts do not.
class mod_refinement_fork implements Module {
	/** BOARD is the thread list, which a thread's own path hangs off. */
	const BOARD = "/refinement";

	/** MISSING is what the page of a thread that is not there says. */
	const MISSING = "No thread with that ID. Someone may have deleted its design, which deletes the thread.";

	/** WANTED is what the box asks for. */
	const WANTED = "A fork starts from the subject of the new work.";

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

		$from = Task\Row::text($thread, "id");
		$want = trim(Task\Route::input("want"));

		if ($want === "") {
			$back = self::BOARD . "/threads/" . Task\Route::pathEscape($from);

			Task\Route::fail($back, self::WANTED);

			return;
		}

		$user = $chrome->name();
		$opened = $store->start($user, $want);
		$id = $opened["id"];

		$store->carry($from, $id, $opened["comment_id"]);

		// The first run of the fork is the flattened request, not the subject
		// alone: the rules arrived with it and a run that is told the subject
		// and not the rules produces the thing the rules were written against.
		$design = $store->design($id);
		$rules = Task\Refinement\Prompt::rules($store->rules($id));
		$body = Task\Refinement\Prompt::flatten(Task\Row::text($design, "body"), $rules);
		$comment = $store->apply($id, $user, $body);

		Task\Refinement\Work::start($store, $id, $comment, $body);
		Task\Route::go(self::BOARD . "/threads/" . Task\Route::pathEscape($id));
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_refinement_fork;

	$mod->serve();
}
