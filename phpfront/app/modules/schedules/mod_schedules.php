<?php

// @route GET /schedules

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_schedules lists what runs on a cron expression, the jobs and the
// workflows, and takes a new one of a job; a workflow's comes from its edit
// page.
//
// The page is read only. Every form on it calls /api/schedules from the
// browser, so the markup carries the data-* attributes task-ui.js drives and
// nothing here writes.
class mod_schedules implements Module {
	/** LEAD is the paragraph under the heading. */
	const LEAD = "Jobs and workflows that run in the background on a cron expression. Each schedule runs as its user: whoever created it, or the user an admin picks under Run as. The scheduler skips a job if it's still running when its time comes, or if its user can't run it at that time. Schedule a workflow from its edit page.";

	/** NONE is what the block says with no schedules, and MORE what it adds where one can be added. */
	const NONE = "No schedules yet.";

	const MORE = " Pick a job and a cron expression above.";

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

		if (!$chrome->allowed("schedule/view")) {
			$view->refused();

			return;
		}

		$chrome->heading("Scheduled", array(array("text" => "Scheduled", "href" => "")));

		$lookup = new Task\Schedules\Lookup;
		$catalogue = $this->catalogue($chrome);
		$users = $chrome->runAsUsers();
		$any = count($catalogue["jobs"]) > 0 || count($catalogue["hosts"]) > 0;
		$rows = $this->rows($chrome, $lookup, $users, $any);
		$buttons = new Task\Buttons(true);
		$head = new Task\Head("Scheduled", self::LEAD);

		$view->page(array(
			"head" => $head->data(),
			"any" => $any,
			"folders" => $catalogue["folders"],
			"jobs" => $catalogue["jobs"],
			"hosts" => $catalogue["hosts"],
			"runAs" => array(
				"id" => "sch-user",
				"options" => array(
					"users" => $users,
					"selected" => $chrome->name(),
					"short" => false,
				),
			),
			"schedules" => $rows,
			"block" => array(
				"title" => "Schedules",
				"note" => "",
				"empty" => $this->said($rows, $any),
				"buttons" => $buttons->data(),
			),
		));
	}

	/** said is what the block says with no schedules, and "" with any. */
	function said($rows, $any) {
		if (count($rows) > 0) {
			return "";
		}

		return $any ? self::NONE . self::MORE : self::NONE;
	}

	/**
	 * catalogue is the folders and the jobs the Add schedule form offers: the
	 * jobs the reader may run, each with the folder it belongs to, and the
	 * folders that hold one, with the folder a query named selected.
	 *
	 * A job is offered where the reader may run it, which takes the allow list
	 * of the job and of its folder together, and a folder where it holds one
	 * such job. A schedule is written with the job's ID and the picker shows
	 * its name, which differ on a folder that follows more than one
	 * convention.
	 *
	 * The jobs that run a command on a host belong to no catalogue, so they
	 * come last and in a group of their own, and the root comes with them: a
	 * reader with the hosts permission is offered it even where it holds no
	 * other job.
	 */
	function catalogue($chrome) {
		$folders = array();
		$jobs = array();
		$pick = Task\Route::input("folder");
		$hosts = Task\HostJobs::of($chrome);

		foreach ($chrome->folders() as $folder) {
			$id = Task\Row::text($folder, "id");
			$allow = Task\Row::items($folder, "allow");
			$runnable = array();

			foreach (Task\Row::items($folder, "jobs") as $job) {
				if ($chrome->access->run($id, $allow, Task\Row::items($job, "allow"))) {
					$runnable[] = array(
						"id" => Task\Row::text($job, "id"),
						"name" => Task\Row::text($job, "name"),
						"folder" => $id,
					);
				}
			}

			if (count($runnable) === 0) {
				continue;
			}

			$folders[] = array("id" => $id, "selected" => $id === $pick);
			$jobs = array_merge($jobs, $runnable);
		}

		if (count($hosts) > 0) {
			$internal = Task\Route::INTERNAL;
			$folders[] = array("id" => $internal, "selected" => $internal === $pick);
		}

		return array(
			"folders" => $folders,
			"jobs" => $jobs,
			"hosts" => $hosts,
		);
	}

	/**
	 * rows is the schedules the reader may see, each as the table draws it: a
	 * workflow's where the reader sees the workflow, a job's where they see
	 * its folder.
	 *
	 * edits is whether the page drew the form that saves a schedule, which is
	 * where a row's Edit sends the reader; a reader with no job to schedule
	 * gets no form and no Edit.
	 */
	function rows($chrome, $lookup, $users, $edits = true) {
		$found = $lookup->workflows();
		$workflows = $this->visible($chrome, $found);
		$workspaces = $chrome->workspaces();
		$schedules = new Task\Schedules\Schedules;

		$out = array();

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

	/**
	 * visible is the workflows the reader may see: those every folder of whose
	 * steps they see, the steps of a workflow a step runs included.
	 */
	function visible($chrome, $workflows) {
		$out = array();

		foreach ($workflows as $id => $workflow) {
			$seen = true;

			foreach ($workflow["folders"] as $folder) {
				if (!Task\User::sees($chrome->user, $folder)) {
					$seen = false;
				}
			}

			if ($seen) {
				$out[$id] = $workflow;
			}
		}

		return $out;
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
		$own = array(
			"workflow" => $flow,
			"folder" => $schedule->getFolder(),
			"folderHref" => Task\Route::folderHref($schedule->getFolder(), $workspaces),
			"job" => $flow ? $what : $schedule->getJob(),
			"jobHref" => $this->jobHref($schedule, $workspaces),
			"args" => Task\Schedule::argLine($schedule->getJob(), $schedule->getArgs()),
			"menu" => $this->menu($schedule, $cells["on"], $what, $edits),
		);

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
	 * menu is a row's [...]: Edit, where the page drew the form that saves one,
	 * then Enable or Disable, by whether it runs, and Delete.
	 *
	 * It is the same [...] a job's page draws over its own schedules, so the
	 * entries are written once, where the table is read.
	 */
	function menu($schedule, $on, $what, $edits = true) {
		return Task\Schedule::menu($schedule, $on, $what, $edits);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_schedules;

	$mod->serve();
}
