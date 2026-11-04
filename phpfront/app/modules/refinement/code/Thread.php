<?php

namespace Task\Refinement;

// Thread is what a thread is made of, as text: the lines that name it, the
// grouping the page draws, and the markdown the page renders.
//
// Every method is static and takes rows. The statements are in storage/, the
// gate and the redirect are in the mod_ files, and this is the part that is
// neither: what a title is, what counts as a rule, what is a picture and what
// is prose. A fixture pins all of it without a database.
class Thread {
	/** WIDTH is how long a line the thread list shows may be. */
	const WIDTH = 120;

	/**
	 * FILES is where a relative link in a thread's markdown resolves: the
	 * artifact files, which a run wrote and which the report names from inside
	 * the thread's own folder.
	 *
	 * At the root and not under the base. Serving a file is a byte stream and
	 * stays Go, so the prefix is the same string before and after the cutover,
	 * and the same one phpscript.ArtifactPrefix hands the plugin.
	 */
	const FILES = "/artifacts/files/";

	/** NAMED is what a thread with nothing to name it is called. */
	const NAMED = "a thread";

	/** cut is one line as the thread list shows it, WIDTH at most. */
	static function cut($line) {
		$line = trim($line);

		if (strlen($line) > self::WIDTH) {
			return substr($line, 0, self::WIDTH - 3) . "...";
		}

		return $line;
	}

	/** title is the first line of a comment, for the thread list. */
	static function title($body) {
		$lines = explode("\n", trim($body));
		$line = self::cut($lines[0]);

		return $line === "" ? self::NAMED : $line;
	}

	/**
	 * reportTitle is the first heading of the report, which is what the thread
	 * is called once a run has written one.
	 *
	 * A thread opens named after the line the person typed, because that is
	 * all there is at the time. The runs then go and find out what the thing
	 * actually is, and README.md leads with it: the report's own heading is a
	 * better name than the request that started it, and a correction that
	 * changes the heading renames the thread with it.
	 *
	 * ATX and setext both, and the first one anywhere in the document, so a
	 * report opening with a badge or a comment still answers.
	 *
	 * Empty when the report has no heading at all. That is not a title of "",
	 * it is the absence of one: the caller keeps what the thread had.
	 */
	static function reportTitle($markdown) {
		$lines = explode("\n", $markdown);
		$count = count($lines);

		for ($i = 0; $i < $count; $i++) {
			$line = trim($lines[$i]);

			if ($line === "") {
				continue;
			}

			// # The heading, with any hashes the writer balanced it with.
			if (substr($line, 0, 2) === "# ") {
				return self::cut(trim(substr($line, 2), " #"));
			}

			// The heading, with a line of = under it.
			if ($i + 1 < $count && preg_match('/^=+$/', trim($lines[$i + 1]))) {
				return self::cut($line);
			}
		}

		return "";
	}

	/**
	 * designID is the comment a thread opened on, out of the rows the page
	 * already has, or "" for a thread with none left.
	 *
	 * The same rule the store uses: the oldest comment that corrects nothing
	 * and records no Apply all. The page needs it because the delete button on
	 * the design has to say that it takes the thread.
	 */
	static function designID($comments) {
		foreach ($comments as $comment) {
			if ($comment["kind"] !== "apply" && !$comment["correction"]) {
				return $comment["id"];
			}
		}

		return "";
	}

	/**
	 * entries groups a thread's comments the way the page draws them: a
	 * comment that stands on its own, or an Apply all with the ones it folded
	 * under it.
	 *
	 * Folded comments are not dropped from the page. They are the record of
	 * what was asked and of the run that answered each one, and the page keeps
	 * them behind a disclosure rather than behind nothing.
	 */
	static function entries($comments) {
		$folded = array();

		foreach ($comments as $comment) {
			$into = $comment["collapsed_into"];

			if ($into !== "") {
				if (!isset($folded[$into])) {
					$folded[$into] = array();
				}

				$folded[$into][] = $comment;
			}
		}

		$entries = array();

		foreach ($comments as $comment) {
			if ($comment["collapsed_into"] !== "") {
				continue;
			}

			$under = isset($folded[$comment["id"]]) ? $folded[$comment["id"]] : array();

			$entries[] = array(
				"comment" => $comment,
				"folded" => $under,
				"folded_count" => count($under),
				"applied" => $comment["kind"] === "apply",
			);
		}

		return $entries;
	}

	/**
	 * request is what a person wrote, or what an Apply all composed, as the
	 * HTML the thread shows.
	 *
	 * Markdown, because a request is written as markdown: a flattened one is
	 * sections and a numbered list, and a typed one is as likely to carry a
	 * list or a path in backticks. Printing the source left the reader with
	 * the hashes and the asterisks.
	 *
	 * The renderer is the instance's, which skips raw HTML and keeps a link to
	 * a scheme it knows, so a request a reader typed is as safe to draw as a
	 * report a run wrote.
	 */
	static function request($body, $folder) {
		$text = trim($body);

		if ($text === "") {
			return "";
		}

		return \Task\Host::markdown(self::flatHeadings($text), self::FILES, $folder);
	}

	/**
	 * flatHeadings takes every heading of a request down to the smallest one,
	 * which the page draws as a bold line.
	 *
	 * A request is a paragraph inside a thread, not a document: a run of `##
	 * The design` and `## The rules` set in the page's own h2 reads as though
	 * the thread had been cut into chapters, and the thread's own headings
	 * lose against them.
	 *
	 * Inside a fence nothing is a heading. A shell snippet's `# comment` is a
	 * comment, and rewriting it would be rewriting the code the request is
	 * showing.
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

			$found = array();

			if (preg_match('/^#{1,6}\s+(.*)$/', $trimmed, $found)) {
				$lines[$i] = "###### " . trim($found[1], " #");
			}
		}

		return implode("\n", $lines);
	}

	/**
	 * report renders the README a run wrote as the HTML the page shows.
	 *
	 * It is markdown the instance's own agent wrote into the artifact tree,
	 * not anything a reader typed, and the page draws it rather than printing
	 * it. An image in it is embedded, which is how a schema diagram is read on
	 * the page instead of downloaded.
	 *
	 * The folder is the thread's own, and the report names its files from
	 * inside it: [schema/user.up.sql](schema/user.up.sql) is the thread's copy
	 * and the link has to carry the folder the run left out.
	 */
	static function report($markdown, $folder) {
		$text = trim($markdown);

		if ($text === "") {
			return "";
		}

		return \Task\Host::markdown($text, self::FILES, $folder);
	}

	/**
	 * split separates the report into what is looked at and what is read: the
	 * lines that are nothing but an embedded image, and everything else.
	 *
	 * The point of a refinement is the picture. What the run wrote around it is
	 * how it got there, which is worth keeping and not worth putting first:
	 * the page draws the images and folds the prose behind "show thinking".
	 *
	 * A report that embeds nothing is all prose, and then the prose is the
	 * output: there is nothing to look at and the page shows it rather than
	 * hiding the only thing it has.
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

			// An image inside a fence is a fragment being shown, not a picture
			// being embedded, and a line with prose around the image is a
			// sentence: only a line that is the image moves.
			if (!$fenced && preg_match('/^!\[[^\]]*\]\([^)]+\)$/', $trimmed)) {
				$images[] = $trimmed;

				continue;
			}

			$prose[] = $line;
		}

		// Lifting a line out leaves the blank lines that framed it, and two
		// gaps where there was one reads as a missing paragraph.
		$text = preg_replace('/\n{3,}/', "\n\n", implode("\n", $prose));

		return array("images" => implode("\n\n", $images), "prose" => trim($text));
	}
}
