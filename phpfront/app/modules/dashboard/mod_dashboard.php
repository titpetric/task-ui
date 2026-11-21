<?php

// @route GET /

require "ext/autoload.php";

// mod_dashboard is the page the instance opens on: every run in progress as
// one row of a process table, and the menu of what the reader may add.
//
// The rows follow the runs topic, so a run that starts adds one and one that
// ends leaves it without a reload.
class mod_dashboard implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "Every run in progress, one row each: the command or the prompt it runs, who started it, and where its work is read.";

	/** ADD is the label of the head's menu. */
	const ADD = "Add a workspace, project, user, key, secret or workflow";

	function name() {
		return "dashboard";
	}

	function title() {
		return "Dashboard";
	}

	function icon() {
		return "grid";
	}

	function tables() {
		return [];
	}

	function serve() {
		$chrome = new Task\Chrome;

		$chrome->heading("Dashboard");

		$head = new Task\Head("Dashboard", self::LEAD);

		$head->menu($this->add($chrome));

		$view = new Task\View($this, $chrome);
		$note = Task\Findings::note($chrome->folders(), $chrome->workspaces());

		$view->page([
			"head" => $head->data(),
			"note" => $note,
			"runs" => $this->processes($chrome),
		]);
	}

	/**
	 * add is the head's menu: what the reader may create anywhere, in the
	 * order of the setup, each behind the rule of the page that adds it.
	 *
	 * Adding a user also needs the instance to have a password file, which the
	 * Go page gates on as well. It has one: the configuration fills the path
	 * in where it is unset, so the flag is set on every instance that serves a
	 * page and no binding carries it here.
	 */
	function add($chrome) {
		$menu = new Task\Menu(self::ADD);

		$menu->when($chrome->allowed("workspace/edit"), "folder-open", "Add workspace", ["href" => Task\Route::href("/settings/workspaces/new")]);
		$menu->when($chrome->allowed("project/edit"), "folder", "Add project", ["href" => Task\Route::href("/settings/projects/new")]);
		$menu->when($chrome->allowed("user/edit"), "users", "Add user", ["href" => Task\Route::href("/settings/users/new")]);
		$menu->when($chrome->allowed("key/edit"), "key", "Generate key", ["href" => Task\Route::href("/settings/keys/generate")]);
		$menu->when($chrome->allowed("key/edit"), "upload", "Upload key", ["href" => Task\Route::href("/settings/keys/upload")]);
		$menu->when($chrome->allowed("secret/edit"), "secret", "Add secret", ["href" => Task\Route::href("/settings/secrets/new")]);
		$menu->when($chrome->allowed("workflow/edit"), "flow", "Add workflow", ["href" => Task\Route::href("/workflows/new")]);

		return $menu->data();
	}

	/**
	 * processes is a row per run in progress the reader sees, newest first, as
	 * ps lists one line per process.
	 *
	 * The runner's memory holds every run in progress, whoever started it and
	 * whatever its job, which is why the list is read here and not off the
	 * tree: the tree keys a run by its job and keeps one of each.
	 */
	function processes($chrome) {
		$runs = new Task\Runs\Runs($chrome->db);
		$threads = new Task\Threads\Notes($chrome->db);
		$prompts = $chrome->allowed("prompt/view");
		$notes = $chrome->allowed("note/view");
		$out = [];
		$seen = [];

		foreach (Task\Host::running() as $each) {
			$id = Task\Row::text($each, "id");

			if ($id === "" || isset($seen[$id])) {
				continue;
			}

			$seen[$id] = true;
			$run = $runs->get($id);

			if (!Task\Run::seen($chrome->user, $run, $prompts, $notes)) {
				continue;
			}

			Task\Run::overlay($run, Task\Host::session($id));
			$out[] = $this->process($run, $threads, $prompts || $notes, $chrome->workspaces());
		}

		usort($out, function($one, $other) {
			return strcmp($other["id"], $one["id"]);
		});

		return $out;
	}

	/** process is one row: the run as ps lists it and where its work is read. */
	function process($run, $threads, $threadsOK, $workspaces) {
		$id = $run->getId();
		$thread = null;

		if (Task\Run::isPrompt($run) || Task\Run::isNote($run)) {
			$thread = $threads->byRun($id);
		}

		return [
			"id" => $id,
			"pid" => substr($id, 0, 8),
			"run_href" => Task\Route::href("/runs/" . rawurlencode($id)),
			"user" => $run->getUserName(),
			"user_href" => Task\Route::href("/users/" . rawurlencode($run->getUserName())),
			"state" => $run->getState(),
			"start" => Task\Clock::format($run->getStartedAt(), "15:04"),
			"started" => Task\Run::started($run->getStartedAt()),
			"time" => Task\Run::duration(Task\Run::elapsed($run)),
			"task" => Task\Run::name($run),
			"href" => $this->href($run, $thread, $threadsOK, $workspaces),
			"command" => $this->command($run, $thread),
		];
	}

	/** href is where a run's work is read: its thread, its host, or the page of its job. */
	function href($run, $thread, $threadsOK, $workspaces) {
		if ($thread !== null && $threadsOK) {
			return Task\Route::href("/threads/" . rawurlencode($thread->getId()));
		}

		if (Task\Run::isPrompt($run) || Task\Run::isNote($run)) {
			return $threadsOK ? Task\Route::href("/threads") : "";
		}

		if (Task\Run::isHost($run)) {
			$host = substr($run->getJob(), strlen(Task\Run::HOST_PREFIX));

			return Task\Route::href($host === Task\Run::HOST_ALL ? "/inventory" : "/inventory/hosts/" . rawurlencode($host));
		}

		$href = Task\Run::runJobHref($run, $workspaces);

		return $href === "" ? "" : Task\Route::href($href);
	}

	/** command is what a run runs: the thread's question, else the title a job gave itself. */
	function command($run, $thread) {
		if ($thread !== null) {
			return Task\Note::excerpt($thread->getBody());
		}

		$title = $run->getTitle();

		if ($title !== "") {
			return $title;
		}

		return implode(" ", Task\Run::argList($run));
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_dashboard;

	$mod->serve();
}
