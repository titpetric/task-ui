<?php

// @route GET /runs

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_runs lists every run the reader may see, newest first, a page at a
// time: the filters narrow it, and the table follows the runs topic so a run
// that starts goes on top and one that ends changes in place.
class mod_runs implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "Every run, by anyone, newest first. Attach to a running job from here; closing the window doesn't stop it.";

	function name() {
		return "runs";
	}

	function title() {
		return "Runs";
	}

	function icon() {
		return "list";
	}

	function tables() {
		return array("run", "user", "workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;

		$chrome->heading("Runs", array(array("text" => "Runs", "href" => "")));

		$runs = new Task\Run;
		$size = Task\Pager::size();
		$form = array(
			"folder" => Task\Route::input("folder"),
			"job" => Task\Route::input("job"),
			"state" => Task\Route::input("state"),
			"user" => Task\Route::input("user"),
			"n" => (string)$size,
			"batch" => Task\Route::input("batch"),
		);
		$before = Task\Route::input("before");
		$after = Task\Route::input("after");
		// One row more than the page holds: the row that tells whether the
		// list goes on, which paging drops again. The query narrows to the
		// folders the reader may see, so the page is not filled with rows
		// Seen is about to drop.
		//
		// The comments sit here and not among the keys: `phpscript fmt`
		// moves a comment out of an array literal and leaves it after the
		// statement. See docs/todo-correction.md.
		$filter = array(
			"folder" => $form["folder"],
			"folders" => Task\User::folders($chrome->user),
			"job" => $form["job"],
			"dir" => Task\Route::input("dir"),
			"state" => $form["state"],
			"user" => $form["user"],
			"batch" => $form["batch"],
			"before" => $before,
			"after" => $after,
			"limit" => $size + 1,
		);
		$matched = $runs->find($filter);
		$seen = new Task\Seen($chrome);
		$found = $seen->of($matched);
		$page = Task\Pager::paged($found, $size, $before, $after);
		$workspaces = $chrome->workspaces();
		$summarize = Task\Host::claude();
		$rows = array();

		foreach ($page["rows"] as $run) {
			$rows[] = $this->row($run, $chrome, $workspaces, $summarize);
		}

		$view = new Task\View($this, $chrome);
		$head = new Task\Head("Runs", self::LEAD);

		$view->page(array(
			"head" => $head->data(),
			"form" => $form,
			"folders" => $this->folders($chrome),
			"states" => Task\Run::states(),
			"users" => $this->users($chrome),
			"sizes" => Task\Pager::SIZES,
			"fold" => array("id" => "runs-activity", "label" => "Activity"),
			"activity" => $this->activity($seen, $runs, $filter),
			"toolFold" => array("id" => "runs-tooling", "label" => "Tooling"),
			"tooling" => $this->tooling(),
			"summarize" => $summarize,
			"rows" => $rows,
			"pager" => Task\Pager::of("/runs", $page["prev"], $page["next"]),
		));
	}

	/**
	 * activity is the charts above the table: the same filter over the whole
	 * window and not over the page, so a reader who pages through the list
	 * sees one chart of what the filter holds.
	 *
	 * The cursors come off and the limit becomes the charts' own, which is
	 * more runs than the days hold on a busy instance. The duration line is
	 * one job's alone and this page mixes jobs, so the charts draw without
	 * it.
	 *
	 * The clock is the server's, which the heading says, and the charts count
	 * their days in its zone.
	 *
	 * The window is read as rows and stays rows: five thousand of them, for
	 * five numbers each. Hydrating a model per row to read those cost 274ms of
	 * a 2s page and its stamps were parsed three times over, so charted() asks
	 * for the six columns the filter and the charts read and nothing else.
	 */
	function activity($seen, $runs, $filter) {
		$filter["before"] = "";
		$filter["after"] = "";
		$filter["limit"] = Task\Activity::LIMIT;

		$matched = $runs->charted($filter);
		$found = $seen->rows($matched);
		$days = Task\Activity::rows($found);
		$now = Task\Clock::now();
		$charts = new Task\Activity($days, $now, Task\Activity::DAYS, false);

		return $charts->data();
	}

	/**
	 * tooling is the panel under the charts: what the commands inside the
	 * runs spent their time on, which atkins recorded and not task-ui.
	 *
	 * It takes no filter. The rows are commands and the filters here are
	 * over runs, so a folder or a user narrows nothing atkins knows about,
	 * and the panel says the same thing on every page of the list.
	 */
	function tooling() {
		$report = Task\Host::tooling();
		$now = Task\Clock::now();
		$panel = new Task\Tooling($report, $now);

		return $panel->data();
	}

	/**
	 * row is one run of the table, its cells and the menu of what the reader
	 * may do to it.
	 *
	 * The runner is ahead of the table, so the session is laid over the row by
	 * Task\Run::overlay before a cell is read, and the windows watching the
	 * run are counted from the same session.
	 *
	 * summarize says the reader has a token for Claude, which the summary of
	 * an ended run needs.
	 */
	function row($run, $chrome, $workspaces, $summarize) {
		$id = $run->getId();
		$session = Task\Host::session($id);
		$live = Task\Row::flag($session, "live");
		$message = Task\Run::overlay($run, $session);
		$state = $run->getState();
		$ended = Task\Run::ended($state);
		$job = $run->getJob();
		$menu = new Task\Menu("Actions on the run of " . $job);

		$menu->detail("Run", $id, true);
		$menu->detail("Started", Task\Clock::read($run->getStartedAt()));
		$menu->item("file", "Open the run's page", array("href" => Task\Route::href("/runs/" . $id)));
		$menu->when($run->getCastPath() !== "" && $ended, "play", "Replay the recording", array(
			"data-replay" => $id,
			"data-replay-title" => Task\Run::name($run),
			"data-replay-who" => $run->getUserName(),
		));
		// The summary is read from the API at the root and not under the
		// theme's mount, as every data-* endpoint of a page is.
		$menu->when($summarize && $ended, "note", "Summary", array(
			"data-summary" => "/api/runs/" . $id . "/summary",
			"data-summary-page" => Task\Route::href("/runs/" . $id),
			"data-summary-title" => Task\Run::name($run),
			"title" => "Claude's summary of the output",
		));
		$menu->when($live && !$ended, "stop", "Stop", array("data-stop" => $id));
		$menu->whenDanger($live && !$ended, "trash", "Kill", array("data-kill" => $id));

		return array(
			"id" => $id,
			"ago" => Task\Clock::ago($run->getStartedAt()),
			"stamp" => Task\Clock::read($run->getStartedAt()),
			"place" => Task\Run::place($run),
			"folderHref" => Task\Run::runFolderHref($run, $workspaces),
			"job" => $run->getJob(),
			"jobHref" => Task\Run::runJobHref($run, $workspaces),
			"who" => $chrome->who()->ofRun($run),
			"started" => Task\Run::started($run->getStartedAt()),
			"duration" => Task\Run::duration(Task\Run::elapsed($run)),
			"state" => Task\Run::label($run),
			"message" => $state === Task\Run::QUEUED ? $message : "",
			"title" => $run->getTitle(),
			"watching" => Task\Row::number($session, "attached"),
			"attach" => $ended ? "View" : "Attach",
			"menu" => $menu->data(),
		);
	}

	/** folders is the folders of the catalogue the reader sees, for the filter. */
	function folders($chrome) {
		$out = array();

		foreach ($chrome->folders() as $folder) {
			$out[] = Task\Row::text($folder, "id");
		}

		return $out;
	}

	/** users is the names of the users, sorted, for the filter. */
	function users($chrome) {
		$out = array_keys($chrome->who()->roles);

		sort($out);

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_runs;

	$mod->serve();
}
