<?php

// The routing a path needs and the pages the module draws.
//
// Every page renders through MiniTPL, from templates/, compiled into
// templates_c/ which the module ships. The module escapes what it
// renders: MiniTPL does that by default for a text node, an attribute
// and an HTML comment, and the one value that reaches the page
// unescaped is the report, which is markdown a run wrote and the module
// rendered.
//
// Made with the module's base, the prefix task-ui mounted it at, so a
// link the page builds is a link that resolves.
class RefinementPage {
	var $base;

	function __construct($base) {
		$this->base = $base;
	}

	/**
	 * path is what is left of a request for the module to answer: the
	 * path past the module's own prefix, without the query and without
	 * the slashes at either end.
	 */
	function path($uri) {
		$parts = explode("?", $uri);
		$path = $parts[0];

		if (strpos($path, $this->base) === 0) {
			$path = substr($path, strlen($this->base));
		}

		return trim($path, "/");
	}

	/** template is the module's engine, over the templates compiled ahead of time. */
	static function template($name) {
		$tpl = new MiniTPL\Template("templates/");

		$tpl->set_compile_location("templates_c/", true);
		$tpl->load($name);

		return $tpl;
	}

	/** go answers with a redirect, which a form that changed something should. */
	static function go($to) {
		header("Location: " . $to);
		http_response_code(303);
	}

	/** missing answers for a thread that is not there. */
	static function missing() {
		http_response_code(404);
		echo "<p>No thread there.</p>";
	}

	/** fail says what a form needs and sends the reader back to it. */
	static function fail($to, $why) {
		http_response_code(400);
		echo "<p class=\"bad\">", htmlspecialchars($why, ENT_QUOTES), "</p>";
		echo "<p><a href=\"", htmlspecialchars($to, ENT_QUOTES), "\">Back</a></p>";
	}

	/** threads draws the thread list and the form that starts one. */
	function threads($threads) {
		$tpl = RefinementPage::template("threads.tpl");

		$tpl->assign("base", $this->base);
		$tpl->assign("threads", $threads);

		echo $tpl->get();
	}

	/**
	 * thread draws one thread: the report its last run wrote, links to
	 * the folder that holds the files and to the context the runs keep,
	 * the requests, and the boxes that ask for the next one.
	 *
	 * The files are not listed. They are a folder of the artifact tree
	 * and the instance already has the page that browses one, with the
	 * preview, the versions and the download beside it.
	 */
	function thread($thread, $comments) {
		$tpl = RefinementPage::template("thread.tpl");

		$tpl->assign("base", $this->base);
		$tpl->assign("thread", $thread);
		$tpl->assign("entries", RefinementManager::entries($comments));
		// Which comment is the design, so its delete button can say that
		// it takes the thread with it.
		$tpl->assign("design", RefinementManager::designID($comments));
		$tpl->assign("context", Task::artifact($thread["id"] . "/CONTEXT.md") !== "");
		$tpl->assign("report", RefinementPage::report($thread["report_markdown"], $thread["id"]));

		echo $tpl->get();
	}

	/**
	 * report renders the README a run wrote as the HTML the page shows.
	 *
	 * It is markdown the instance's own agent wrote into the artifact
	 * tree, not anything a reader typed, and the page draws it rather
	 * than printing it. An image in it is embedded, which is how a schema
	 * diagram is read on the page instead of downloaded.
	 *
	 * The folder is the thread's own, and the report names its files from
	 * inside it: [schema/user.up.sql](schema/user.up.sql) is the thread's
	 * copy and the link has to carry the folder the run left out.
	 */
	static function report($markdown, $folder) {
		$text = trim($markdown);

		if ($text === "") {
			return "";
		}

		return Task::markdown($text, $folder);
	}
}
