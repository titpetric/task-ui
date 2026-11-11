<?php

// @route GET /users/{name}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_profile is one user's page: who they are, and the latest runs, workflow
// runs, notes and prompts of theirs, as far as the reader sees them.
//
// Your own page is yours to read. Anybody else's takes user/view, and a name
// no row holds answers Not found rather than saying the name is free.
//
// Nothing here is the whole of what a user did: each block is the latest few
// of a kind the reader may see, and the Runs block links the list that holds
// the rest.
class mod_profile implements Module {
	/** COUNT is how many of each kind of activity the page lists. */
	const COUNT = 10;

	/**
	 * WINDOW is how many times the list the prompts are read in: a prompt
	 * asked in a place the reader does not see is dropped after the query, so
	 * the query reads more than the block draws and the block fills from what
	 * is left.
	 */
	const WINDOW = 5;

	/** MINE is the lead of your own page. */
	const MINE = "Your latest runs, notes and prompts.";

	/** THEIRS is the lead of anybody else's, which says the list is narrowed. */
	const THEIRS = "The latest runs, notes and prompts by %s that you can see.";

	/** NOBODY is what a name no row holds answers. */
	const NOBODY = "No user by that name.";

	/** DROPS is what deleting a user costs, which the confirmation says. */
	const DROPS = "? This removes them from the password file and ends their sessions and tokens right away. Their runs, notes and history keep the name.";

	function name() {
		return "profile";
	}

	function title() {
		return "Profile";
	}

	function icon() {
		return "users";
	}

	function tables() {
		return array("user", "run", "workflow_run", "workflow_run_step", "note", "workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);
		$name = Task\Route::param("name");
		$own = $name !== "" && $name === $chrome->name();

		if (!$own && !$chrome->allowed("user/view")) {
			$view->refused();

			return;
		}

		$users = new Task\User($chrome->db);
		$user = $users->byName($name);

		if ($user === null) {
			$view->missing(self::NOBODY);

			return;
		}

		$chrome->heading($name, $this->crumbs($chrome, $name));

		$seen = new Task\Seen($chrome);
		$workspaces = $chrome->workspaces();
		$edits = $own || $chrome->allowed("user/edit");

		$view->page(array(
			"who" => $this->who($chrome, $user, $own),
			"buttons" => $this->buttons($chrome, $user, $own, $edits),
			"menu" => $this->menu($chrome, $user),
			"runs" => $this->runs($chrome, $seen, $name, $workspaces),
			"flows" => $this->flows($chrome, $name),
			"notes" => $this->notes($chrome, $name),
			"prompts" => $this->prompts($chrome, $name, $workspaces),
			"keys" => $this->keys($chrome, $name, $own),
		), "profile.tpl");
	}

	/**
	 * crumbs is the trail above the page: the name alone for a reader who
	 * came to one user, and the path through the settings for one who may
	 * read the list of them.
	 */
	function crumbs($chrome, $name) {
		if (!$chrome->allowed("user/view")) {
			return array(array("text" => $name, "href" => ""));
		}

		return array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => "Users", "href" => "/settings/users"),
			array("text" => $name, "href" => ""),
		);
	}

	/** who is the head of the page: the avatar, the name and the lead under it. */
	function who($chrome, $user, $own) {
		$name = $user->getName();
		$created = $user->getCreatedAt();
		$lead = $own ? self::MINE : sprintf(self::THEIRS, $name);

		if ($created !== "") {
			$lead .= " Joined " . Task\Clock::ago($created) . ".";
		}

		return array(
			"name" => $name,
			"role" => Task\User::role($user),
			"initials" => Task\User::initials($user),
			"lead" => $lead,
		);
	}

	/**
	 * buttons is the head's actions: New MCP token where the reader may make
	 * one, and the Edit your own page always carries.
	 *
	 * New MCP token is one click: the script posts it, takes the token out of
	 * the answer and puts the `claude mcp add` line that carries it on the
	 * clipboard. Nothing is drawn, because nothing keeps the token.
	 */
	function buttons($chrome, $user, $own, $edits) {
		$buttons = new Task\Buttons;
		$name = $user->getName();

		if ($this->manages($chrome, $user, $own)) {
			$buttons->item("copy", "New MCP token", array(
				"data-api" => "/api/users/" . $user->getId() . "/mcp",
				"data-method" => "POST",
				"data-copy-fill" => Task\Mcp\Connect::add(Task\Route::site()),
				"data-notice" => "MCP token copied as the claude mcp add command",
				"title" => "Make an MCP token for " . $name . " and copy the claude mcp add command",
			));
		}

		if ($edits) {
			$buttons->primary("edit", "Edit", array("href" => Task\Route::href("/users/" . $name . "/edit")));
		}

		return $buttons->data();
	}

	/**
	 * manages reports whether the reader may make the user an MCP token: with
	 * user/edit and the MCP agent user flag, or on their own page once they
	 * hold one, which is the same answer the edit page's block draws from.
	 */
	function manages($chrome, $user, $own) {
		if (!$user->getMcp()) {
			return false;
		}

		if ($chrome->allowed("user/edit")) {
			return true;
		}

		if (!$own) {
			return false;
		}

		$store = new Task\Mcp\Tokens($chrome->db);

		return count($store->ofKind($user->getId(), Task\Mcp\Tokens::MCP)) > 0;
	}

	/**
	 * menu is the [...] of the head: the list of their runs, and removing the
	 * user where there is a password file to remove them from.
	 *
	 * An admin is not deletable here and neither is the reader themselves: the
	 * one would leave an instance nobody administers and the other would end
	 * the session reading the page.
	 *
	 * Whether a password file is configured is the last thing asked, because
	 * asking costs the host an instance read and the cheap conditions settle
	 * most pages before it.
	 */
	function menu($chrome, $user) {
		$name = $user->getName();
		$id = $user->getId();
		$menu = new Task\Menu("Actions on " . $name);

		$menu->item("list", "Runs by this user", array("href" => Task\Route::href("/runs?user=" . Task\Route::queryEscape($name))));
		$menu->whenDanger($this->deletes($chrome, $user), "trash", "Delete", array(
			"data-delete" => "/api/users/" . $id,
			"data-confirm" => "Delete the user " . $name . self::DROPS,
			"data-confirm-ok" => "Delete",
			"data-next" => "/settings/users?notice=User+deleted",
		));

		return $menu->data();
	}

	/** deletes reports whether the reader may remove the user the page shows. */
	function deletes($chrome, $user) {
		if (!$chrome->allowed("user/delete") || Task\User::admin($user)) {
			return false;
		}

		if ($chrome->user === null || $chrome->user->getId() === $user->getId()) {
			return false;
		}

		$instance = Task\Host::instance(false);

		return Task\Row::text($instance, "htpasswd") !== "";
	}

	/**
	 * runs is the Runs block: the latest runs the user started that the reader
	 * sees, and the link to every one of them.
	 *
	 * The query narrows to the folders the reader may see, so the list is not
	 * filled with rows Seen is about to drop.
	 */
	function runs($chrome, $seen, $name, $workspaces) {
		$store = new Task\Runs\Runs($chrome->db);
		$filter = array(
			"user" => $name,
			"folders" => Task\User::folders($chrome->user),
			"limit" => self::COUNT,
		);
		$found = $seen->of($store->find($filter));
		$rows = array();

		foreach ($found as $run) {
			$rows[] = $this->run($run, $workspaces);
		}

		return array(
			"block" => self::block("Runs", "the latest, in folders you can see", count($rows) === 0 ? "No runs yet." : ""),
			"rows" => $rows,
			"more" => array(
				"href" => Task\Route::href("/runs?user=" . Task\Route::queryEscape($name)),
				"label" => "Every run by " . $name,
			),
		);
	}

	/**
	 * run is one row of the Runs block.
	 *
	 * The runner is ahead of the table, so the session is laid over the row by
	 * Task\Run::overlay before a cell is read: a run that just ended says so
	 * here and its button says View rather than Attach.
	 *
	 * A run of Claude's has no job page, and then the name links the run.
	 */
	function run($run, $workspaces) {
		$id = $run->getId();
		$session = Task\Host::session($id);

		// The overlay is read for what it writes into the run and not for
		// what it answers: the row says nothing about a queued run here, so
		// the message it returns is nobody's.
		Task\Run::overlay($run, $session);

		$ended = Task\Run::ended($run->getState());
		$job = $run->getJob();
		$label = Task\Run::name($run);
		$href = Task\Run::runJobHref($run, $workspaces);
		$menu = new Task\Menu("Actions on the run of " . $job);

		$menu->detail("Run", $id, true);
		$menu->detail("Started", Task\Clock::read($run->getStartedAt()));
		$menu->item("file", "Open the run's page", array("href" => Task\Route::href("/runs/" . $id)));
		$menu->when($run->getCastPath() !== "" && $ended, "play", "Replay the recording", array(
			"data-replay" => $id,
			"data-replay-title" => $label,
			"data-replay-who" => $run->getUserName(),
		));

		return array(
			"id" => $id,
			"ago" => Task\Clock::ago($run->getStartedAt()),
			"stamp" => Task\Clock::read($run->getStartedAt()),
			"name" => $label,
			"href" => $href === "" ? "/runs/" . $id : $href,
			"duration" => Task\Run::duration(Task\Run::elapsed($run)),
			"state" => Task\Run::label($run),
			"attach" => $ended ? "View" : "Attach",
			"menu" => $menu->data(),
		);
	}

	/**
	 * flows is the Workflow runs block: the latest workflow runs the user
	 * started that the reader sees every folder of.
	 *
	 * The row draws its own state and not the state component, as the Go page
	 * does: a workflow run is not on the runs topic, so no script follows it
	 * and the span carries no run to follow.
	 */
	function flows($chrome, $name) {
		$store = new Task\WorkflowRun($chrome->db);
		$rows = array();

		foreach ($store->byUser($chrome->user, $name, self::COUNT) as $run) {
			$flow = $run->getWorkflowId();

			$rows[] = array(
				"ago" => Task\Clock::ago($run->getStartedAt()),
				"stamp" => Task\Clock::read($run->getStartedAt()),
				"name" => $run->getName(),
				"href" => "/workflows/" . $flow,
				"state" => $run->getState(),
				"view" => "/workflows/" . $flow . "/runs/" . $run->getId(),
			);
		}

		return array(
			"block" => self::block("Workflow runs", "the latest you can see", count($rows) === 0 ? "No workflow runs yet." : ""),
			"rows" => $rows,
		);
	}

	/**
	 * notes is the Notes block: the board notes the user started, which takes
	 * note/view.
	 *
	 * A reader without it gets the block with nothing in it and not the block
	 * left out, because the empty line of a block says what is there and not
	 * what the reader may read.
	 */
	function notes($chrome, $name) {
		$rows = array();

		if ($chrome->allowed("note/view")) {
			$store = new Task\Note($chrome->db);

			foreach ($store->byUser(Task\Note::NOTES, $name, self::COUNT) as $note) {
				$rows[] = $this->note($note, Task\Route::href("/notes"));
			}
		}

		return array(
			"block" => self::block("Notes", "the latest " . $name . " wrote", count($rows) === 0 ? "No notes yet." : ""),
			"rows" => $rows,
		);
	}

	/**
	 * prompts is the Prompts block: the prompt threads the user asked in a
	 * place the reader sees, which takes prompt/view.
	 *
	 * A prompt about a run opens that run and any other opens the place it was
	 * asked in; a place the reader's folders do not hold has no page and the
	 * link is the note's anchor alone.
	 */
	function prompts($chrome, $name, $workspaces) {
		$rows = array();

		if ($chrome->allowed("prompt/view")) {
			$rows = $this->asked($chrome, $name, $workspaces);
		}

		return array(
			"block" => self::block("Prompts", "the latest " . $name . " asked, in places you can see", count($rows) === 0 ? "No prompts yet." : ""),
			"rows" => $rows,
		);
	}

	/**
	 * asked is the rows of the Prompts block: the threads of the window that
	 * were asked in a place the reader sees, at most COUNT of them.
	 */
	function asked($chrome, $name, $workspaces) {
		$store = new Task\Note($chrome->db);
		$folders = $chrome->folders();
		$rows = array();

		foreach ($store->promptsByUser($name, self::WINDOW * self::COUNT) as $note) {
			if (count($rows) === self::COUNT) {
				return $rows;
			}

			$dir = $note->getDir();

			if (!self::seesPlace($chrome, $dir)) {
				continue;
			}

			$about = $note->getAboutRunId();
			$place = $about === "" ? Task\Route::heldHref($dir, $folders, $workspaces) : "/runs/" . $about;
			$page = $place === "" ? "" : Task\Route::href($place);

			$rows[] = $this->note($note, $page, $dir);
		}

		return $rows;
	}

	/**
	 * note is one item of the Notes or Prompts list: the one line the body
	 * reads as, the state it is in and when it was written.
	 *
	 * page is where the item links, the note's own anchor added here: the
	 * board for a note, and the run or the place for a prompt. dir is the
	 * place a prompt was asked in, which the item says in small type, and ""
	 * for a note of the board.
	 */
	function note($note, $page, $dir = "") {
		$id = $note->getId();

		return array(
			"id" => $id,
			"href" => $page . "#note-" . $id,
			"excerpt" => Task\Note::excerpt($note->getBody()),
			"state" => $note->getState(),
			"ago" => Task\Clock::ago($note->getCreatedAt()),
			"stamp" => Task\Clock::read($note->getCreatedAt()),
			"dir" => $dir,
		);
	}

	/**
	 * keys is the Keys block, for your own page and for a reader with
	 * key/view, and null for anybody else: which keys a user's runs get is
	 * what a key grants, and that is not everyone's to read.
	 */
	function keys($chrome, $name, $own) {
		if (!$own && !$chrome->allowed("key/view")) {
			return null;
		}

		$folder = Task\Key::folder();
		$linked = Task\Key::linked($folder["keys"], $name);

		return Task\Key::ofUser($chrome, $name, $linked, "/users/" . $name);
	}

	/**
	 * seesPlace reports whether the reader sees the place a prompt was asked
	 * in, as a note's dir holds it: the root, or a directory of the deepest
	 * folder of the catalogue that holds it.
	 *
	 * The catalogue is the whole of it and not the reader's own, because the
	 * deepest folder decides the answer: a folder they may not see is not a
	 * shorter one they may, and reading through theirs would let a prompt of a
	 * checkout inside it pass as the parent's.
	 */
	static function seesPlace($chrome, $place) {
		if ($place === Task\Route::INTERNAL) {
			return true;
		}

		$best = "";

		foreach ($chrome->all() as $folder) {
			$id = Task\Row::text($folder, "id");

			if ($place !== $id && strpos($place, $id . "/") !== 0) {
				continue;
			}

			if (strlen($id) > strlen($best)) {
				$best = $id;
			}
		}

		return $best !== "" && Task\User::sees($chrome->user, $best);
	}

	/** block is the head of one block of the page, which draws no buttons. */
	static function block($title, $note, $empty) {
		$buttons = new Task\Buttons(true);

		return array(
			"title" => $title,
			"note" => $note,
			"empty" => $empty,
			"buttons" => $buttons->data(),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_profile;

	$mod->serve();
}
