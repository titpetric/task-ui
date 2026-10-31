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
		$folder = $thread["id"];
		$entries = RefinementManager::entries($comments);

		foreach ($entries as $i => $entry) {
			$entries[$i]["comment"]["html"] = RefinementPage::request($entry["comment"]["body"], $folder);

			foreach ($entry["folded"] as $j => $under) {
				$entries[$i]["folded"][$j]["html"] = RefinementPage::request($under["body"], $folder);
			}
		}

		$tpl->assign("base", $this->base);
		$tpl->assign("thread", $thread);
		$tpl->assign("entries", $entries);
		// Which comment is the design, so its delete button can say that
		// it takes the thread with it.
		$tpl->assign("design", RefinementManager::designID($comments));
		$tpl->assign("scratchpad", Task::artifact($folder . "/" . RefinementPrompt::SCRATCHPAD) !== "");

		$report = RefinementPage::split($thread["report_markdown"]);

		$tpl->assign("output", RefinementPage::report($report["images"], $folder));
		$tpl->assign("thinking", RefinementPage::report($report["prose"], $folder));
		$tpl->assign("visual", $report["images"] !== "");

		echo $tpl->get();
	}

	/**
	 * request is what a person wrote, or what an Apply all composed, as
	 * the HTML the thread shows.
	 *
	 * Markdown, because a request is written as markdown: a flattened one
	 * is sections and a numbered list, and a typed one is as likely to
	 * carry a list or a path in backticks. Printing the source left the
	 * reader with the hashes and the asterisks.
	 *
	 * The renderer is the instance's, which skips raw HTML and keeps a
	 * link to a scheme it knows, so a request a reader typed is as safe
	 * to draw as a report a run wrote.
	 */
	static function request($body, $folder) {
		$text = trim($body);

		if ($text === "") {
			return "";
		}

		return Task::markdown(RefinementPage::flatHeadings($text), $folder);
	}

	/**
	 * flatHeadings takes every heading of a request down to the smallest
	 * one, which the page draws as a bold line.
	 *
	 * A request is a paragraph inside a thread, not a document: a run of
	 * `## The design` and `## The rules` set in the page's own h2 reads as
	 * though the thread had been cut into chapters, and the thread's own
	 * headings lose against them.
	 *
	 * Inside a fence nothing is a heading. A shell snippet's `# comment`
	 * is a comment, and rewriting it would be rewriting the code the
	 * request is showing.
	 */
	static function flatHeadings($markdown) {
		$lines = explode("\n", $markdown);
		$fenced = false;

		foreach ($lines as $i => $line) {
			$trimmed = trim($line);

			if (substr($trimmed, 0, 3) === "```" || substr($trimmed, 0, 3) === "~~~") {
				$fenced = !$fenced;
				continue;
			}

			if ($fenced) {
				continue;
			}

			if (preg_match('/^#{1,6}\s+(.*)$/', $trimmed, $found)) {
				$lines[$i] = "###### " . trim($found[1], " #");
			}
		}

		return implode("\n", $lines);
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

	/**
	 * split separates the report into what is looked at and what is
	 * read: the lines that are nothing but an embedded image, and
	 * everything else.
	 *
	 * The point of a refinement is the picture. What the run wrote
	 * around it is how it got there, which is worth keeping and not
	 * worth putting first: the page draws the images and folds the prose
	 * behind "show thinking".
	 *
	 * A report that embeds nothing is all prose, and then the prose is
	 * the output: there is nothing to look at and the page shows it
	 * rather than hiding the only thing it has.
	 */
	static function split($markdown) {
		$images = array();
		$prose = array();
		$fenced = false;

		foreach (explode("\n", $markdown) as $line) {
			$trimmed = trim($line);

			if (substr($trimmed, 0, 3) === "```" || substr($trimmed, 0, 3) === "~~~") {
				$fenced = !$fenced;
			}

			// An image inside a fence is a fragment being shown, not a
			// picture being embedded, and a line with prose around the
			// image is a sentence: only a line that is the image moves.
			if (!$fenced && preg_match('/^!\[[^\]]*\]\([^)]+\)$/', $trimmed)) {
				$images[] = $trimmed;
				continue;
			}

			$prose[] = $line;
		}

		// Lifting a line out leaves the blank lines that framed it, and
		// two gaps where there was one reads as a missing paragraph.
		$text = preg_replace('/\n{3,}/', "\n\n", implode("\n", $prose));

		return array("images" => implode("\n\n", $images), "prose" => trim($text));
	}
}
