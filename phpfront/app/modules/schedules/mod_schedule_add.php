<?php

// @route GET /schedules/add

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_schedule_add is the form that adds a schedule: of a job the reader may
// run, or of a workflow whose every step they see.
//
// ?folder= and ?job= open it on a job's page, and ?workflow= on a workflow's:
// the pickers stand on what the query named. The page reads and does not
// write: the form posts to /api/schedules, which gates it again.
class mod_schedule_add implements Module {
	/** TITLE is the heading, the last crumb and what the button says. */
	const TITLE = "Add schedule";

	/** LEAD is the paragraph under the heading. */
	const LEAD = "A schedule for a job or a workflow, run on a cron expression. A job's runs as its user, whoever adds it or one an admin picks; a workflow's runs as the workflow's user, or the caller when it has none. The scheduler skips a job still running when its time comes, or a user who cannot run it then.";

	/** NEXT is where the form goes once the schedule is added. */
	const NEXT = "?notice=Schedule+added";

	function name() {
		return "schedules";
	}

	function title() {
		return "Scheduled";
	}

	function icon() {
		return "clock";
	}

	function tables() {
		return array(
			"schedule",
			"workflow",
			"workflow_step",
			"workspace",
			"user",
			"key_info",
			"key_role",
			"key_user",
		);
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("schedule/edit")) {
			$view->refused();

			return;
		}

		$folder = Task\Route::input("folder");
		$job = Task\Route::input("job");
		$workflow = Task\Route::input("workflow");
		$catalogue = Task\Schedules\Catalogue::of($chrome, $folder, $job);
		$workflows = Task\Schedules\Catalogue::workflows($chrome);
		$users = $chrome->runAsUsers();

		$chrome->heading(self::TITLE, array(
			array("text" => "Scheduled", "href" => "/schedules"),
			array("text" => self::TITLE, "href" => ""),
		));

		$head = new Task\Head(self::TITLE, self::LEAD);

		$view->page(array(
			"head" => $head->data(),
			"edit" => false,
			"target" => "",
			"api" => "/api/schedules",
			"method" => "POST",
			"next" => Task\Route::href("/schedules") . self::NEXT,
			"back" => Task\Route::href("/schedules"),
			"label" => self::TITLE,
			"folders" => $catalogue["folders"],
			"jobs" => $catalogue["jobs"],
			"hosts" => $catalogue["hosts"],
			"workflows" => $workflows,
			"workflow" => $workflow,
			"folder" => $folder,
			"job" => $job,
			"expression" => "",
			"args" => "",
			"triage" => false,
			"enabled" => true,
			"pickUser" => count($users) > 0,
			"runAs" => array(
				"id" => "sch-user",
				"options" => array(
					"users" => $users,
					"selected" => $chrome->name(),
					"short" => false,
				),
			),
			"workflowName" => "",
			"workflowHref" => "",
			"targetHref" => "",
		), "add.tpl");
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_schedule_add;

	$mod->serve();
}
