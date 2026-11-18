<?php

// @route GET /search

require "ext/autoload.php";

// mod_search is the search page: one box over the documentation and job targets
// of every project the reader may see. The box starts a run at /search/ask,
// which answers in a thread.
class mod_search implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "Ask for what you need in the documentation and jobs of every project you can see.";

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
		return ["search_document", "search_section", "search_term"];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("prompt/edit") || !$chrome->allowed("prompt/view") || !$chrome->allowed("note/view") || !Task\Host::claude()) {
			$view->refused();

			return;
		}

		$chrome->heading("Search", [["text" => "Search", "href" => ""]]);

		$head = new Task\Head("Search", self::LEAD);

		$view->page(["head" => $head->data()]);
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_search;

	$mod->serve();
}
