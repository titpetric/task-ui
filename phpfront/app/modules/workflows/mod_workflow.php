<?php

// @route GET /workflows/{id}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workflow is one workflow: the pipeline of its steps, live while a run is
// going, and under it the charts of its runs and the runs themselves, a page at
// a time.
//
// A reader who may not see every folder the steps run in is answered as if the
// workflow were not there: Task\Workflow::view decides it, and the page says
// the same thing for an ID nobody holds.
class mod_workflow implements Module {
	/** GONE is what the page says for an ID no workflow the reader sees holds. */
	const GONE = "No workflow with that ID.";

	/** COLUMNS is how wide the runs table is, which an answer row spans. */
	const COLUMNS = 4;

	/** SUMMARY is the hover on the menu entry that asks Claude for one. */
	const SUMMARY = "Claude's summary of every step's output";

	function name() {
		return "workflows";
	}

	function title() {
		return "Workflows";
	}

	function icon() {
		return "workflow";
	}

	function tables() {
		return array(
			"workflow",
			"workflow_step",
			"workflow_run",
			"workflow_run_step",
			"run",
			"user",
			"workspace",
		);
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("workflow/view")) {
			$view->refused();

			return;
		}

		$id = Task\Route::input("id");
		$store = new Task\Workflow($chrome->db);
		$row = $store->get($id);

		if ($row === null) {
			$view->missing(self::GONE);

			return;
		}

		$names = $store->names();
		$flow = $store->view($chrome, $row, $names);

		if ($flow === null) {
			$view->missing(self::GONE);

			return;
		}

		$chrome->heading($flow["name"], array(
			array("text" => "Workflows", "href" => "/workflows"),
			array("text" => $flow["name"], "href" => ""),
		));

		$workspaces = $chrome->workspaces();
		$runs = new Task\WorkflowRun($chrome->db);
		$size = Task\Pager::size();
		$before = Task\Route::input("before");
		$after = Task\Route::input("after");
		// One run more than the page holds: the row that says whether the
		// list goes on, which paging drops again.
		$filter = array(
			"limit" => $size + 1,
			"before" => $before,
			"after" => $after,
		);
		$found = $runs->listByWorkflow($id, $filter);
		$page = Task\Pager::paged($found, $size, $before, $after);
		$summarize = Task\Host::claude();
		// One reader of the runs for the whole page: the chain of a row, the
		// pipeline over them and the run going all name the same few runs.
		$seen = new Task\Seen($chrome);
		$rows = array();

		foreach ($page["rows"] as $run) {
			$rows[] = $this->row($chrome, $flow, $run, $runs, $names, $workspaces, $summarize, $seen);
		}

		$view->page(array(
			"head" => $this->head($flow),
			"id" => $flow["id"],
			"user" => $chrome->name(),
			"latest" => self::latest($flow["last"]),
			"pipeline" => $this->pipeline($chrome, $flow, $runs, $names, $workspaces, $seen),
			"activity" => $this->activity($id, $runs),
			"form" => array("n" => (string)$size),
			"sizes" => Task\Pager::SIZES,
			"summarize" => $summarize,
			"rows" => $rows,
			"pager" => Task\Pager::of("/workflows/" . $id, $page["prev"], $page["next"]),
		), "workflow.tpl");
	}

	/**
	 * head is the top of the page: what the workflow is, and the buttons that
	 * run it, stop the run going and open the editor.
	 *
	 * The line under the name is built here because it counts the steps and
	 * names the user the runs run as only where that is somebody else: a
	 * template that counted would hold the plural rule.
	 *
	 * The actions are live: a run that starts turns Run into Stop with no
	 * reload, which is the data-actions the script finds the div by. A reader
	 * who may not run it is told why where the button would be, so Edit stays
	 * to the right of the sentence.
	 */
	function head($flow) {
		$count = count($flow["steps"]);
		$steps = $count . ($count === 1 ? " step" : " steps") . ", by " . $flow["userName"];

		if ($flow["runAs"] !== "" && $flow["runAs"] !== $flow["userName"]) {
			$steps .= ". Runs as " . $flow["runAs"];
		}

		$head = new Task\Head($flow["name"]);
		$buttons = new Task\Buttons;

		$head->said($flow["description"]);
		$head->faint($steps);
		$head->live();
		$head->always();

		if ($flow["canRun"] && $flow["running"] !== "") {
			$buttons->item("stop", "Stop", array("data-workflow-stop" => $flow["running"]));
		} elseif ($flow["canRun"]) {
			$attrs = array("data-workflow-run" => $flow["id"]);

			if ($flow["prompted"]) {
				$attrs["data-workflow-prompt"] = $flow["name"];
			}

			$buttons->primary("play", "Run", $attrs);
		} elseif ($flow["runDenied"] !== "") {
			$buttons->faint($flow["runDenied"]);
		}

		if ($flow["canChange"]) {
			$edit = Task\Route::href("/workflows/" . $flow["id"] . "/edit");

			$buttons->item("edit", "Edit", array("href" => $edit));
		}

		$head->buttons($buttons->data());

		return $head->data();
	}

	/**
	 * latest is the line over the pipeline: how the newest run went. Null for a
	 * workflow that has not run, which the page says instead.
	 */
	static function latest($last) {
		if ($last === null) {
			return null;
		}

		return array(
			"state" => $last->getState(),
			"message" => $last->getMessage(),
			"ago" => Task\Clock::ago($last->getStartedAt()),
			"stamp" => Task\Clock::read($last->getStartedAt()),
			"userName" => $last->getUserName(),
		);
	}

	/**
	 * pipeline is the workflow's steps as the list draws them, with the states
	 * and the screens of the run going laid over them.
	 *
	 * Only a run that has not ended is laid over: a reader of this page wants
	 * the steps as they are now, and an ended run is opened by View under Runs,
	 * which draws the states it left.
	 */
	function pipeline($chrome, $flow, $runs, $names, $workspaces, $seen) {
		$last = $flow["last"];
		$latest = array();

		if ($last !== null && $last->getState() === Task\Run::RUNNING) {
			$steps = Task\Step::ofRun($runs->steps($last->getId()), $names);
			$latest = Task\Step::seen($seen, $steps);
		}

		$links = Task\Chain::pipeline($flow["steps"], $latest, $workspaces);
		$framed = Task\Chain::frames($links, $chrome->db, $names, $workspaces, true, $seen);

		return Task\Pipeline::nodes($framed);
	}

	/**
	 * activity is the charts over the table: every run of this workflow in the
	 * window and not the page's own, so paging through the list changes no
	 * chart.
	 *
	 * The duration line is drawn, which the runs list leaves out: the runs here
	 * are of one workflow, so their durations are of the same work.
	 */
	function activity($id, $runs) {
		$found = $runs->listByWorkflow($id, array("limit" => Task\Activity::LIMIT));
		$days = Task\Activity::of($found);
		$now = Task\Clock::now();
		$charts = new Task\Activity($days, $now, Task\Activity::DAYS);

		return $charts->data();
	}

	/**
	 * row is one run of the table: when it started and how it ended, who
	 * started it, the chain of its steps and the menu of what may be done to
	 * it, with Claude's answer in the row under it.
	 *
	 * The answer starts folded here and open on the Workflows page, which is
	 * what the chevron's open says: a reader of this page came for the runs and
	 * not for one answer.
	 */
	function row($chrome, $flow, $run, $runs, $names, $workspaces, $summarize, $seen) {
		$id = $run->getId();
		$state = $run->getState();
		$answer = $run->getAnswer();
		$href = "/workflows/" . $flow["id"] . "/runs/" . $id;
		$ran = Task\Step::ofRun($runs->steps($id), $names);
		$steps = Task\Step::seen($seen, $ran);
		$menu = new Task\Menu("Actions on the run of " . $flow["name"]);

		$menu->detail("Run", $id, true);
		$menu->detail("Started", Task\Clock::read($run->getStartedAt()));
		$menu->item("file", "Open the run's page", array("href" => Task\Route::href($href)));
		// The summary is read from the API at the root and not under the
		// theme's mount, as every data-* endpoint of a page is.
		$menu->when($summarize && $state !== Task\Run::RUNNING, "note", "Summary", array(
			"data-summary" => "/api/workflow-runs/" . $id . "/summary",
			"data-summary-page" => Task\Route::href($href),
			"data-summary-title" => $flow["name"],
			"title" => self::SUMMARY,
		));

		$chain = Task\Chain::of($steps, $workspaces);

		return array(
			"id" => $id,
			"href" => $href,
			"state" => $state,
			"message" => $run->getMessage(),
			"ago" => Task\Clock::ago($run->getStartedAt()),
			"stamp" => Task\Clock::read($run->getStartedAt()),
			"user" => $chrome->who()->of($run->getUserName()),
			"chain" => Task\Pipeline::nodes($chain),
			"fold" => array(
				"id" => $id,
				"answer" => $answer !== "",
				"open" => false,
			),
			"answer" => array(
				"id" => $id,
				"body" => $answer === "" ? "" : $chrome->message($answer),
				"cols" => self::COLUMNS,
				"open" => false,
			),
			"menu" => $menu->data(),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workflow;

	$mod->serve();
}
