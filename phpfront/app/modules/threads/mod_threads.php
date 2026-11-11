<?php

// @route GET /threads

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_threads is the timeline: every thread the reader may see, newest first,
// and the box that starts one.
//
// A thread is a note of the board, a prompt asked from a page, or a question
// put to the orb: one table, one list, one page per thread. The list is rows
// and not the threads themselves, because a thread is read on its own page,
// where the messages take sides and a recording plays.
//
// The page reads and does not write. The box holds data-note-form, which the
// script posts to /api/notes and then refreshes the list in place.
class mod_threads implements Module {
	/** LEAD is the sentence under the heading, as HTML: it names a tag. */
	const LEAD = 'Everything said on this instance: a note to someone with <code>@name</code>, a prompt asked from a page, and what the orb heard. If you can run jobs, a note is an instruction to the people it names, and naming several splits one request between them.';

	/** AGENT is what the sentence says further where an agent answers notes. */
	const AGENT = ' Name <span class="mention">@%s</span> to start the agent on a thread, and reply to follow up: a reply goes to it without naming it again.';

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

		$chrome->heading("Threads", array(array("text" => "Threads", "href" => "")));

		$store = new Task\Note;
		$users = new Task\User;
		$roots = $store->timeline();
		$replies = $store->replies(Task\Note::ids($roots));
		$agents = $users->agents();
		$seen = new Task\Seen($chrome);
		$rows = array();

		foreach ($roots as $note) {
			$under = isset($replies[$note->getId()]) ? $replies[$note->getId()] : array();
			$rows[] = Task\Threads\Timeline::row($note, $under, $chrome, $seen);
		}

		$agent = count($agents) > 0 ? $agents[0]->getName() : "";
		$head = new Task\Head("Threads", self::lead($agent));

		$view->page(array(
			"head" => $head->data(),
			"agent" => $agent,
			"editor" => self::field(),
			"rows" => $rows,
		));
	}

	/**
	 * field is the box that starts a thread, as components/editor.tpl reads
	 * it: a note is a document and takes every tool, and an `@` opens the list
	 * of users, because a note is written to someone.
	 */
	static function field() {
		return array(
			"name" => "body",
			"rows" => 3,
			"label" => "A new note",
			"placeholder" => "",
			"mentions" => true,
			"toolbar" => "full",
		);
	}

	/**
	 * lead is the sentence under the heading: what a thread is, and the agent
	 * that answers one where the instance has one.
	 *
	 * The sentence is HTML and built here, not in the template: a template
	 * writes the newline that sits before a tag, so a sentence cut across
	 * branches arrives cut. See docs/phpfront.md. The agent's name is a user
	 * name, so it is escaped where the prose around it is not.
	 */
	static function lead($agent) {
		if ($agent === "") {
			return self::LEAD;
		}

		return self::LEAD . sprintf(self::AGENT, htmlspecialchars($agent, ENT_QUOTES));
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_threads;

	$mod->serve();
}
