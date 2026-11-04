<?php

// @route GET /workflows/{id}/runs/{run}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workflow_run is one run of a workflow: what it was asked over what came
// back, then the steps as the run left them, each with the screen its own run
// ended on.
//
// The workflow's page opens this in a window, which data-pipeline-view on the
// block is the contract for, so the block is the whole page and the chrome
// around it is dropped by the script.
class mod_workflow_run implements Module {
	/** GONE is what the page says for an ID this workflow holds no run of. */
	const GONE = "No run of this workflow with that ID.";

	/** MISSING is what the page says for an ID no workflow the reader sees holds. */
	const MISSING = "No workflow with that ID.";

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
			"note",
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
			$view->missing(self::MISSING);

			return;
		}

		$names = $store->names();
		$flow = $store->view($chrome, $row, $names);

		if ($flow === null) {
			$view->missing(self::MISSING);

			return;
		}

		$runs = new Task\WorkflowRun($chrome->db);
		$run = $runs->get(Task\Route::input("run"));

		if ($run === null || $run->getWorkflowId() !== $id) {
			$view->missing(self::GONE);

			return;
		}

		$chrome->heading($flow["name"], array(
			array("text" => "Workflows", "href" => "/workflows"),
			array("text" => $flow["name"], "href" => "/workflows/" . $id),
			array("text" => "Run", "href" => ""),
		));

		$workspaces = $chrome->workspaces();
		$seen = new Task\Seen($chrome);
		$ran = Task\Step::ofRun($runs->steps($run->getId()), $names);
		$steps = Task\Step::seen($seen, $ran);
		$links = Task\Chain::of($steps, $workspaces);
		$framed = Task\Chain::frames($links, $chrome->db, $names, $workspaces, true, $seen);

		$view->page(array(
			"flow" => array("name" => $flow["name"]),
			"run" => self::run($run),
			"thread" => Task\Pipeline::thread($chrome, $run, $steps),
			"pipeline" => Task\Pipeline::nodes($framed),
			"prompt" => Task\Prompt::ofWorkflowRun($chrome, $flow, $run),
		), "workflow_run.tpl");
	}

	/**
	 * run is the line under the name: how the run ended, when it started and
	 * who started it, and how long it took once it ended.
	 *
	 * A run still going has no time to show, which finished says: elapsed
	 * would count against now and the line reads as if it were over.
	 */
	static function run($run) {
		$done = $run->getFinishedAt() !== "";

		return array(
			"state" => $run->getState(),
			"message" => $run->getMessage(),
			"ago" => Task\Clock::ago($run->getStartedAt()),
			"stamp" => Task\Clock::read($run->getStartedAt()),
			"userName" => $run->getUserName(),
			"finished" => $done,
			"took" => $done ? Task\Run::duration(Task\Run::elapsed($run)) : "",
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workflow_run;

	$mod->serve();
}
