<?php

// @route GET /threads/{id}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_thread is one thread read as a conversation: the messages in the order
// they were said, the reader's own on one side and everyone else's on the
// other, each with the recording it was spoken into where there is one.
//
// The page is the same conversation the Threads pane draws, and `?pane=1`
// answers that pane alone: the markup and no shell, for the script to swap in.
// A prompt belongs to the page it was asked from, so reading one takes
// prompt/view as well; a note takes note/view alone, and a chat takes one of
// its participants or an admin.
class mod_thread implements Module {
	function name() {
		return "threads";
	}

	function title() {
		return "Threads";
	}

	function icon() {
		return "note";
	}

	function tables() {
		return ["note", "note_event", "note_participant", "run", "user"];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("note/view") && !$chrome->allowed("prompt/view")) {
			$view->refused();

			return;
		}

		$store = new Task\Threads\Notes;
		$note = $store->get(Task\Route::param("id"));
		$pane = Task\Route::input("pane") === "1";

		if ($note === null || $note->getParentId() !== "") {
			$this->problem($view, $pane, Task\Threads\Conversation::MISSING);

			return;
		}

		$conversation = Task\Threads\Conversation::view($note, $chrome, $store, $pane);

		if (isset($conversation["error"])) {
			$this->problem($view, $pane, $conversation["error"]);

			return;
		}

		if ($pane) {
			header("Content-Type: text/html; charset=utf-8");
			echo $view->partial("pane.tpl", ["data_pane" => $conversation]);

			return;
		}

		$chrome->heading($conversation["title"], [
			["text" => "Threads", "href" => "/threads"],
			["text" => $conversation["title"], "href" => ""],
		]);

		$view->page($conversation, "thread.tpl");
	}

	/** problem answers a pane with its message, or a page with the whole shell. */
	function problem($view, $pane, $why) {
		if ($pane) {
			header("Content-Type: text/html; charset=utf-8");
			echo '<p class="empty-state">' . htmlspecialchars($why, ENT_QUOTES) . "</p>";

			return;
		}

		if ($why === Task\Threads\Conversation::MISSING) {
			$view->missing($why);

			return;
		}

		$view->refused();
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_thread;

	$mod->serve();
}
