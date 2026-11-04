<?php

// @route GET /workflows/new

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workflow_new is the editor on a blank workflow: the same form the edit
// page draws, with nothing in it and Add workflow under it.
//
// The route is spelled before /workflows/{id} by the router's own ordering, a
// static segment winning over a parameter, so new is never read as an ID.
class mod_workflow_new implements Module {
	/** LEAD is the paragraph under the heading. */
	const LEAD = "Name the workflow, then add its steps in the order they run: commands, prompts for Claude and a summary. Each step starts when the one before it ends ok.";

	function name() {
		return "workflows";
	}

	function title() {
		return "Add workflow";
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
			"user",
			"key_info",
			"key_role",
			"key_user",
			"workspace",
		);
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("workflow/edit")) {
			$view->refused();

			return;
		}

		$chrome->heading("Add workflow", array(
			array("text" => "Workflows", "href" => "/workflows"),
			array("text" => "Add workflow", "href" => ""),
		));

		$store = new Task\Workflow($chrome->db);
		$flows = $store->visible($chrome);
		$space = new Task\StepSpace($chrome);
		$editor = new Task\Editor($chrome, $space, $flows);
		$head = new Task\Head("Add workflow", self::LEAD);

		$view->page(array(
			"head" => $head->data(),
			"editor" => $editor->data(null, "/api/workflows", "POST", "/workflows", "Add workflow"),
		), "workflow_new.tpl");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workflow_new;

	$mod->serve();
}
