<?php

// @route GET /schedules/edit/{id}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_schedule_edit is the form that changes a schedule: when it runs, who it
// runs as, its args and its triage, and whether it runs at all.
//
// What a schedule starts cannot change, because the API's update takes no
// folder and no job: the form names the target and offers no picker for it.
// The page reads and does not write; the form puts to /api/schedules/{id},
// which gates it again.
class mod_schedule_edit implements Module {
	/** TITLE is the heading, the last crumb and what the button says. */
	const TITLE = "Edit schedule";

	/** LEAD is the paragraph under the heading. */
	const LEAD = "When the schedule runs, who it runs as, and whether it runs. What it starts cannot change here: delete it and add another.";

	/** MISSING is what the page says for an ID no schedule the reader sees holds. */
	const MISSING = "No schedule with that ID. Someone may have deleted it.";

	/** NEXT is where the form goes once the schedule is saved. */
	const NEXT = "?notice=Schedule+saved";

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

		$store = new Task\Schedules\Schedules($chrome->db);
		$schedule = $store->get(Task\Route::param("id"));

		if ($schedule === null) {
			$view->missing(self::MISSING);

			return;
		}

		$users = $chrome->runAsUsers();
		$workspaces = $chrome->workspaces();

		if (Task\Schedule::isWorkflow($schedule)) {
			$target = $this->workflow($chrome, $view, $schedule);
		} else {
			$target = $this->job($chrome, $view, $schedule, $workspaces);
		}

		if ($target === null) {
			return;
		}

		$chrome->heading(self::TITLE, array(
			array("text" => "Scheduled", "href" => "/schedules"),
			array("text" => self::TITLE, "href" => ""),
		));

		$head = new Task\Head(self::TITLE, self::LEAD);

		$view->page(array(
			"head" => $head->data(),
			"edit" => true,
			"target" => $target["target"],
			"api" => "/api/schedules/" . rawurlencode($schedule->getId()),
			"method" => "PUT",
			"next" => Task\Route::href("/schedules") . self::NEXT,
			"back" => Task\Route::href("/schedules"),
			"label" => "Save schedule",
			"folder" => $schedule->getFolder(),
			"job" => $schedule->getJob(),
			"workflow" => $target["workflow"],
			"workflowName" => $target["name"],
			"workflowHref" => $target["href"],
			"targetHref" => $target["targetHref"],
			"expression" => $schedule->getExpression(),
			"args" => $target["args"],
			"triage" => $target["triage"],
			"enabled" => Task\Schedule::enabled($schedule),
			"pickUser" => $target["pickUser"],
			"runAs" => array(
				"id" => "sch-user",
				"options" => array(
					"users" => $users,
					"selected" => $target["runAs"],
					"short" => false,
				),
			),
			"folders" => array(),
			"jobs" => array(),
			"hosts" => array(),
			"workflows" => array(),
		), "edit.tpl");
	}

	/**
	 * workflow is the target of a schedule of a workflow: its name and page,
	 * the user its runs run as, and null once the page has answered for a
	 * workflow the reader may not see.
	 */
	function workflow($chrome, $view, $schedule) {
		$lookup = new Task\Schedules\Lookup;
		$flows = $lookup->workflows();
		$id = $schedule->getWorkflowId();

		if (!isset($flows[$id]) || !Task\Schedules\Catalogue::visible($chrome, array($id => $flows[$id]))) {
			$view->missing(self::MISSING);

			return null;
		}

		$name = $flows[$id]["name"] === "" ? $id : $flows[$id]["name"];
		$fixed = $flows[$id]["runAs"];
		$users = $chrome->runAsUsers();

		return array(
			"target" => "workflow",
			"workflow" => $id,
			"name" => $name,
			"href" => Task\Route::href("/workflows/" . Task\Route::pathEscape($id)),
			"targetHref" => "",
			"args" => "",
			"triage" => false,
			"pickUser" => $fixed === "" && count($users) > 0,
			"runAs" => $fixed === "" ? $schedule->getUserName() : $fixed,
		);
	}

	/**
	 * job is the target of a schedule of a job: its page, which a job of the
	 * instance has none of, and null once the page has answered for a folder
	 * the reader may not see.
	 */
	function job($chrome, $view, $schedule, $workspaces) {
		if (!Task\User::sees($chrome->user, $schedule->getFolder())) {
			$view->missing(self::MISSING);

			return null;
		}

		$href = "";

		if (!Task\Schedule::instance($schedule)) {
			$href = Task\Route::href(Task\Route::folderHref($schedule->getFolder(), $workspaces) . "/" . Task\Route::pathEscape($schedule->getJob()));
		}

		return array(
			"target" => "job",
			"workflow" => "",
			"name" => "",
			"href" => "",
			"targetHref" => $href,
			"args" => htmlspecialchars(Task\Schedule::argLine($schedule->getJob(), $schedule->getArgs()), ENT_QUOTES),
			"triage" => Task\Schedule::triages($schedule),
			"pickUser" => count($chrome->runAsUsers()) > 0,
			"runAs" => $schedule->getUserName(),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_schedule_edit;

	$mod->serve();
}
