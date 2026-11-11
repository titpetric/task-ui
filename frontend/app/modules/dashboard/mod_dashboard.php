<?php

// @route GET /

require "ext/autoload.php";

// mod_dashboard is the page the instance opens on: every terminal running
// right now, as a tile with the last screen of its output, and the menu of
// what the reader may add.
//
// The tiles follow the runs topic, so a run that starts adds one and one that
// ends marks it without a reload.
class mod_dashboard implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "Every running terminal, live, for everyone to see. Open one to attach, or pick a folder on the left to start a run.";

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
		return array();
	}

	function serve() {
		$chrome = new Task\Chrome;

		$chrome->heading("Dashboard");

		$head = new Task\Head("Dashboard", self::LEAD);

		$head->menu($this->add($chrome));

		$view = new Task\View($this, $chrome);
		$note = Task\Findings::note($chrome->folders(), $chrome->workspaces());

		$view->page(array(
			"head" => $head->data(),
			"note" => $note,
			"runs" => $this->tiles($chrome),
		));
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

		$menu->when($chrome->allowed("workspace/edit"), "folder-open", "Add workspace", array("href" => Task\Route::href("/settings/workspaces/new")));
		$menu->when($chrome->allowed("project/edit"), "folder", "Add project", array("href" => Task\Route::href("/settings/projects/new")));
		$menu->when($chrome->allowed("user/edit"), "users", "Add user", array("href" => Task\Route::href("/settings/users/new")));
		$menu->when($chrome->allowed("key/edit"), "key", "Generate key", array("href" => Task\Route::href("/settings/keys/generate")));
		$menu->when($chrome->allowed("key/edit"), "upload", "Upload key", array("href" => Task\Route::href("/settings/keys/upload")));
		$menu->when($chrome->allowed("secret/edit"), "secret", "Add variable", array("href" => Task\Route::href("/settings/secrets/new")));
		$menu->when($chrome->allowed("workflow/edit"), "flow", "Add workflow", array("href" => Task\Route::href("/workflows/new")));

		return $menu->data();
	}

	/**
	 * tiles is a tile per run in progress the reader sees, newest first.
	 *
	 * The runs are the ones the chrome already read for the sidebar's dot and
	 * its Stop buttons, filtered by the same rule the Go page filters by, so
	 * the list costs this page nothing: a session and a frame per tile is all
	 * it asks the host for, which is what the Go page asks too.
	 */
	function tiles($chrome) {
		$live = $chrome->liveRuns();
		$runs = array_values($live);
		$roles = $chrome->who()->roles;
		$out = array();

		usort($runs, function($one, $other) {
			return strcmp($other->getId(), $one->getId());
		});

		foreach ($runs as $run) {
			$out[] = $this->tile($run, $roles);
		}

		return $out;
	}

	/**
	 * tile is one run in progress: its state, what it is running, who started
	 * it and the last screen of its terminal.
	 *
	 * The runner is ahead of the row, so the session is laid over it before a
	 * field is read: the state the dot draws and the title beside the name are
	 * the session's.
	 */
	function tile($run, $roles) {
		$id = $run->getId();
		$session = Task\Host::session($id);
		$message = Task\Run::overlay($run, $session);
		$state = $run->getState();
		$name = $run->getUserName();
		$role = Task\Recent::role($roles, $name);
		$watching = Task\Row::number($session, "attached");

		return array(
			"id" => $id,
			"href" => Task\Route::href("/runs/" . $id),
			"state" => $state,
			"running" => $state === Task\Run::RUNNING,
			"name" => Task\Run::name($run),
			"title" => $run->getTitle(),
			"faint" => $this->faint($state, $message, $watching),
			"who" => Task\Who::avatar($name, $role),
			"frame" => Task\Host::frame($id),
		);
	}

	/**
	 * faint is the grey text on the right of a tile's head: why a queued run
	 * is waiting, else how many windows are watching it.
	 */
	function faint($state, $message, $watching) {
		if ($state === Task\Run::QUEUED && $message !== "") {
			return $message;
		}

		if ($watching > 0) {
			return $watching . " watching";
		}

		return "";
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_dashboard;

	$mod->serve();
}
