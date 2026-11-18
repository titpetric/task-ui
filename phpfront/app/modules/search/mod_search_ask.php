<?php

// @route POST /search/ask

require "ext/autoload.php";

// mod_search_ask starts the agent on a question and sends the reader to the
// thread that answers it. A reload of the thread starts no second run.
class mod_search_ask implements Module {
	/** SEARCH is where a question that asks nothing sends the reader back to. */
	const SEARCH = "/search";

	/** WANTED is what an empty box is told. */
	const WANTED = "Type what to look for.";

	function name() {
		return "search";
	}

	function title() {
		return "Search";
	}

	function icon() {
		return "magnifier";
	}

	function tables() {
		return [];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("prompt/edit") || !$chrome->allowed("prompt/view") || !$chrome->allowed("note/view") || !Task\Host::claude()) {
			$view->refused();

			return;
		}

		$query = trim(Task\Route::input("q"));

		if ($query === "") {
			Task\Route::fail(self::SEARCH, self::WANTED);

			return;
		}

		$note = Task\Host::dispatch($query, "prompts/search.md", "");

		Task\Route::go("/threads/" . Task\Route::pathEscape($note));
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_search_ask;

	$mod->serve();
}
