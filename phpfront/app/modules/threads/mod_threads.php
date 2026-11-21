<?php

// @route GET /threads

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_threads is the messenger: every thread the reader may see down the left,
// newest first, and the selected conversation on the right.
//
// A thread is a note of the board, a prompt asked from a page, a question put
// to the orb, or a chat with another user: one table, one split. The selection
// is the `id` parameter, and a reader with none gets the box that starts a
// thread. The right pane is the same template the thread page draws, and the
// script swaps it through /threads/{id}?pane=1 without leaving the page.
//
// The box is components/composer.tpl, which asks Claude or writes a note, and
// a chat's reply posts to /api/threads; the script refreshes the list and the
// pane in place.
class mod_threads implements Module {
	/** LEAD is the sentence under the heading, as HTML: it names a tag. */
	const LEAD = 'Everything said on this instance: a note to someone with <code>@name</code>, a prompt asked from a page, and what the orb heard. If you can run jobs, a note is an instruction to the people it names, and naming several splits one request between them.';

	/** AGENT is what the sentence says further where an agent answers notes. */
	const AGENT = ' Name <span class="mention">@%s</span> to start the agent on a thread, and reply to follow up: a reply goes to it without naming it again.';

	/** USAGE is the sentence of the usage table. */
	const USAGE = 'What every thread asked for: its replies, tool calls, the context it read and what it cost.';

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

		$chrome->heading("Threads", [["text" => "Threads", "href" => ""]]);

		$selectedID = trim(Task\Route::input("id"));

		if ($selectedID !== "") {
			Task\Route::go("/threads/" . Task\Route::pathEscape($selectedID));

			return;
		}

		$store = new Task\Threads\Notes;
		$users = new Task\User;
		$form = [
			"q" => trim(Task\Route::input("q")),
			"place" => trim(Task\Route::input("place")),
			"state" => Task\Route::input("state"),
			"user" => trim(Task\Route::input("user")),
		];
		$filter = [
			"q" => $form["q"],
			"dir" => $form["place"],
			"state" => $form["state"],
			"user" => $form["user"],
		];
		$roots = [];
		foreach ($store->feed($filter) as $note) {
			if (Task\Threads\Timeline::visible($note, $chrome)) {
				$roots[] = $note;
			}
		}

		$replies = $store->replies(Task\Note::ids($roots));
		$usage = $store->usage(Task\Note::ids($roots));
		$agents = $users->agents();
		$seen = new Task\Seen($chrome);
		$rows = [];
		$chats = [];
		$userID = $chrome->user === null ? "" : $chrome->user->getID();

		foreach ($roots as $note) {
			$under = isset($replies[$note->getId()]) ? $replies[$note->getId()] : [];
			$rows[] = Task\Threads\Timeline::row($note, $under, $chrome, $seen, "", $usage[$note->getId()] ?? null);
		}

		if (Task\Route::input("view") === "usage") {
			$replies = 0;
			$tools = 0;
			$tokens = 0;
			$cost = 0;

			foreach ($rows as $row) {
				$replies += $row["replies"];
				$tools += $row["tools"];
				$tokens += $row["tokens"];
				$cost += (float)str_replace("$", "", $row["cost"]);
			}

			$usageHead = new Task\Head("Prompt usage", self::USAGE);
			$back = new Task\Buttons;
			$back->item("note", "Threads", ["href" => Task\Route::href("/threads")]);
			$usageHead->buttons($back->data());
			$view->page([
				"head" => $usageHead->data(),
				"rows" => $rows,
				"count" => count($rows),
				"replies" => $replies,
				"tools" => $tools,
				"tokens" => Task\Threads\Work::size($tokens),
				"cost" => "$" . number_format($cost, 2),
			], "usage.tpl");

			return;
		}

		if ($chrome->allowed("note/view")) {
			foreach ($store->chats($userID) as $chat) {
				$chats[] = Task\Threads\Timeline::chat($chat, $chrome);
			}
		}

		$agent = count($agents) > 0 ? $agents[0]->getName() : "";
		$head = new Task\Head("Threads", self::lead($agent));
		$buttons = new Task\Buttons;
		$buttons->item("pulse", "Prompt usage", ["href" => Task\Route::href("/threads?view=usage")]);
		$head->buttons($buttons->data());
		$detailed = $form["place"] !== "" || $form["state"] !== "" || $form["user"] !== "";
		$filtered = $detailed || $form["q"] !== "";
		$empty = $filtered ? [
			"tag" => "li",
			"text" => "No thread matches.",
			"href" => Task\Route::href("/threads"),
			"label" => "Clear",
		] : ["tag" => "li", "text" => "No threads yet: ask Claude in the box beside this list."];
		$view->page([
			"head" => $head->data(),
			"agent" => $agent,
			"editor" => self::field(),
			"form" => $form,
			"states" => ["open", "working", "done", "failed"],
			"places" => Task\Composer::places($chrome),
			"rows" => $rows,
			"chats" => $chats,
			"composer" => Task\Composer::blank($chrome, $agent),
			"empty" => $empty,
			"filtered" => $filtered,
			"detailed" => $detailed,
		]);
	}

	/**
	 * field is the box that starts a thread, as components/editor.tpl reads
	 * it: a note is a document and takes every tool, and an `@` opens the list
	 * of users, because a note is written to someone.
	 */
	static function field() {
		return [
			"name" => "body",
			"rows" => 3,
			"label" => "A new note",
			"placeholder" => "",
			"mentions" => true,
			"toolbar" => "full",
		];
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
