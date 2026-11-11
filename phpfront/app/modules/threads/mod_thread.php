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
// A prompt belongs to the page it was asked from, so reading one takes
// prompt/view as well; a note takes note/view alone.
class mod_thread implements Module {
	/** MISSING is what the page of a thread that is not there says. */
	const MISSING = "No thread with that ID. Someone may have deleted it.";

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
		return array("note", "run", "user");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("note/view")) {
			$view->refused();

			return;
		}

		$store = new Task\Threads\Notes;
		$note = $store->get(Task\Route::param("id"));

		if ($note === null || $note->getParentId() !== "") {
			$view->missing(self::MISSING);

			return;
		}

		if (Task\Note::isPrompt($note) && !$chrome->allowed("prompt/view")) {
			$view->refused();

			return;
		}

		$replies = $store->replies(array($note->getId()));
		$under = isset($replies[$note->getId()]) ? $replies[$note->getId()] : array();
		$title = Task\Note::excerpt(Task\Threads\Talk::said($note));

		$chrome->heading($title, array(
			array("text" => "Threads", "href" => "/threads"),
			array("text" => $title, "href" => ""),
		));

		$head = new Task\Head($title);
		$seen = new Task\Seen($chrome);

		$head->buttons($this->menu($note, $chrome)->data());

		$view->page(array(
			"head" => $head->data(),
			"id" => $note->getId(),
			"messages" => Task\Threads\Talk::messages($note, $under, $chrome, $seen),
			"editor" => self::field(),
			"replies" => $chrome->allowed("note/edit"),
		), "thread.tpl");
	}

	/** menu is the thread's own actions: a thread is deleted, never rewritten. */
	function menu($note, $chrome) {
		$menu = new Task\Menu("Actions on this thread");

		$menu->whenDanger(Task\Note::deletes($note, $chrome), "trash", "Delete", array(
			"data-delete" => "/api/notes/" . $note->getId(),
			"data-confirm" => "Delete this thread and every message in it? It leaves every listing; nobody reads it again.",
			"data-confirm-ok" => "Delete",
			"data-next" => Task\Route::href("/threads"),
		));

		return $menu;
	}

	/**
	 * field is the box under the conversation, as components/editor.tpl reads
	 * it: two rows and no placeholder, with an `@` opening the list of users,
	 * because a reply is where an agent is named.
	 */
	static function field() {
		return array(
			"name" => "body",
			"rows" => 2,
			"label" => "A reply",
			"placeholder" => "",
			"mentions" => true,
			"toolbar" => "inline",
		);
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_thread;

	$mod->serve();
}
