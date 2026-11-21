<?php

// @route GET /schedules

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_schedules lists what runs on a cron expression, the jobs and the
// workflows; adding one and changing one are the add and edit pages.
//
// The page is read only. Every form on it calls /api/schedules from the
// browser, so the markup carries the data-* attributes task-ui.js drives and
// nothing here writes.
class mod_schedules implements Module {
	/** LEAD is the paragraph under the heading. */
	const LEAD = "Jobs and workflows that run in the background on a cron expression. Each schedule runs as its user: whoever created it, or the user an admin picks under Run as. The scheduler skips a job if it's still running when its time comes, or if its user can't run it at that time.";

	/** NONE is what the block says with no schedules, and MORE what it adds where one can be added. */
	const NONE = "No schedules yet.";

	const MORE = " Add one to run a job or a workflow on a schedule.";

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
		return [
			"schedule",
			"workflow",
			"workflow_step",
			"workspace",
			"user",
			"key_info",
			"key_role",
			"key_user",
		];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("schedule/view")) {
			$view->refused();

			return;
		}

		$chrome->heading("Scheduled", [["text" => "Scheduled", "href" => ""]]);

		$lookup = new Task\Schedules\Lookup;
		$catalogue = $this->catalogue($chrome);
		$workflows = $this->visible($chrome, $lookup->workflows());
		$users = $chrome->runAsUsers();
		$edits = $chrome->allowed("schedule/edit");
		$any = $edits && (count($catalogue["jobs"]) > 0 || count($catalogue["hosts"]) > 0 || count($workflows) > 0);
		$rows = $this->rows($chrome, $lookup, $users, $edits, $workflows);
		$buttons = new Task\Buttons(true);
		$head = new Task\Head("Scheduled", self::LEAD);
		$add = new Task\Buttons;

		if ($edits && $any) {
			$add->primary("plus", "Add schedule", ["href" => Task\Route::href("/schedules/add")]);
		}

		$head->buttons($add->data());

		$view->page([
			"head" => $head->data(),
			"schedules" => $rows,
			"can" => $edits,
			"block" => [
				"title" => "Schedules",
				"note" => "",
				"empty" => $this->said($rows, $any),
				"buttons" => $buttons->data(),
			],
		]);
	}

	/** said is what the block says with no schedules, and "" with any. */
	function said($rows, $any) {
		if (count($rows) > 0) {
			return "";
		}

		return $any ? self::NONE . self::MORE : self::NONE;
	}

	/** catalogue is what the add page offers, with the folder a query named selected. */
	function catalogue($chrome) {
		return Task\Schedules\Catalogue::of($chrome, Task\Route::input("folder"));
	}

	/**
	 * rows is the schedules the reader may see, each as the table draws it: a
	 * workflow's where the reader sees the workflow, a job's where they see
	 * its folder.
	 *
	 * edits is whether a row's Edit is offered, which is the permission to
	 * change a schedule. workflows is the visible workflows where the caller
	 * already read them.
	 */
	function rows($chrome, $lookup, $users, $edits = true, $workflows = null) {
		if ($workflows === null) {
			$workflows = $this->visible($chrome, $lookup->workflows());
		}

		$workspaces = $chrome->workspaces();
		$schedules = new Task\Schedules\Schedules;

		$out = [];

		foreach ($schedules->all() as $schedule) {
			if (Task\Schedule::isWorkflow($schedule)) {
				if (isset($workflows[$schedule->getWorkflowId()])) {
					$out[] = $this->row($chrome, $schedule, $workflows[$schedule->getWorkflowId()], $workspaces, $users, $edits);
				}

				continue;
			}

			if (Task\User::sees($chrome->user, $schedule->getFolder())) {
				$out[] = $this->row($chrome, $schedule, null, $workspaces, $users, $edits);
			}
		}

		return $out;
	}

	/** visible is the workflows the reader may see, out of every one the lookup read. */
	function visible($chrome, $workflows) {
		return Task\Schedules\Catalogue::visible($chrome, $workflows);
	}

	/**
	 * row is one schedule as this table draws it: the cells every table of a
	 * schedule holds, and the two columns this page heads of its own.
	 */
	function row($chrome, $schedule, $workflow, $workspaces, $users, $edits = true) {
		$flow = Task\Schedule::isWorkflow($schedule);
		$what = $this->what($schedule, $workflow);
		$fixed = $workflow === null ? "" : $workflow["runAs"];
		$cells = Task\Schedule::cells($chrome, $schedule, $fixed, $users);
		$own = [
			"workflow" => $flow,
			"folder" => $schedule->getFolder(),
			"folderHref" => Task\Route::folderHref($schedule->getFolder(), $workspaces),
			"job" => $flow ? $what : $schedule->getJob(),
			"jobHref" => $this->jobHref($schedule, $workspaces),
			"args" => Task\Schedule::argLine($schedule->getJob(), $schedule->getArgs()),
			"source" => $schedule->getSource(),
			"can" => $edits,
			"menu" => $this->menu($schedule, $what, $edits),
		];

		return array_merge($own, $cells);
	}

	/** what names what a schedule is for, which its Delete confirm asks about. */
	function what($schedule, $workflow) {
		if (!Task\Schedule::isWorkflow($schedule)) {
			return $schedule->getFolder() . " / " . $schedule->getJob();
		}

		if ($workflow !== null && $workflow["name"] !== "") {
			return $workflow["name"];
		}

		return $schedule->getWorkflowId();
	}

	/**
	 * jobHref is the page of what a schedule starts: the workflow, the job in
	 * its folder, or "" for a job of the instance itself, which has no page.
	 *
	 * The path is written without the prefix, as every path a template writes
	 * is: the template puts {base} in front of it.
	 */
	function jobHref($schedule, $workspaces) {
		if (Task\Schedule::isWorkflow($schedule)) {
			return "/workflows/" . $schedule->getWorkflowId();
		}

		if (Task\Schedule::instance($schedule)) {
			return "";
		}

		return Task\Route::folderHref($schedule->getFolder(), $workspaces) . "/" . $schedule->getJob();
	}

	/**
	 * menu is a row's [...]: Edit and Delete, where the reader may change a
	 * schedule; the flags it runs under are columns of the table.
	 *
	 * It is the same [...] a job's page draws over its own schedules, so the
	 * entries are written once, where the table is read.
	 */
	function menu($schedule, $what, $edits = true) {
		return Task\Schedule::menu($schedule, $what, $edits);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_schedules;

	$mod->serve();
}
