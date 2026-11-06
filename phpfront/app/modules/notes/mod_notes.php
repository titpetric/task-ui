<?php

// @route GET /notes

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_notes is the notes board: the notes newest first, each with its
// replies, and the agent a note names to start work.
//
// The page reads and does not write. Both forms it draws hold
// data-note-form, which the script posts to /api/notes and then refreshes the
// list in place; a POST route here would reload the whole page instead, and
// the Go page has none either.
class mod_notes implements Module {
	/** LEAD is the sentence under the heading, as HTML: it names a tag. */
	const LEAD = 'Leave a note for someone with <code>@name</code>. If you can run jobs, the note is an instruction to the people it names, and naming several splits one request between them. It stays open until they mark it working, and then done or failed, here or over MCP.';

	/** AGENT is what the sentence says further where an agent answers notes. */
	const AGENT = ' Name <span class="mention">@%s</span> in a note to start the agent on the thread, and reply to follow up: a reply goes to it without naming it again. It works on the request and replies here when it finishes, or explains why it can\'t.';

	function name() {
		return "notes";
	}

	function title() {
		return "Notes";
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

		$chrome->heading("Notes", array(array("text" => "Notes", "href" => "")));

		$store = new Task\Note;
		$users = new Task\User;
		$roots = $store->threads(Task\Note::NOTES);
		$replies = $store->replies(Task\Note::ids($roots));
		$agents = $users->agents();
		$seen = new Task\Seen($chrome);
		$threads = array();

		foreach ($roots as $note) {
			$under = isset($replies[$note->getId()]) ? $replies[$note->getId()] : array();
			$answerer = Task\Note::answerer($note, $under, $agents);
			$threads[] = Task\Note::thread($note, $replies, $chrome, $answerer, $seen);
		}

		$agent = count($agents) > 0 ? $agents[0]->getName() : "";
		$head = new Task\Head("Notes", self::lead($agent));

		$view->page(array(
			"head" => $head->data(),
			"agent" => $agent,
			"editor" => self::field(),
			"threads" => $threads,
		));
	}

	/**
	 * field is the board's own field, as components/editor.tpl reads it: a
	 * note is a document and takes every tool, and an `@` opens the list of
	 * users, because a note is written to someone.
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
	 * lead is the sentence under the heading: how a note names someone, and
	 * the agent that answers one where the instance has one.
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
	$mod = new mod_notes;

	$mod->serve();
}
