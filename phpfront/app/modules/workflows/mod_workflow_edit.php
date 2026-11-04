<?php

// @route GET /workflows/{id}/edit

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workflow_edit is the form that changes a workflow, for its owner or an
// admin, and under it the schedules that start it with the form that adds one.
//
// ?workspace= opens the editor with the steps taken to that workspace. Nothing
// here writes, so the move is a draft the reader reads before Save keeps it,
// and the notice says how much of it did not land.
class mod_workflow_edit implements Module {
	/** LEAD is the paragraph under the heading. */
	const LEAD = "Change the name and the steps: fold a step open to edit it, drag it by its grip to move it. Earlier runs keep the steps they ran with.";

	/** MISSING is what the page says for an ID no workflow the reader sees holds. */
	const MISSING = "No workflow with that ID.";

	/** NOT_YOURS heads the refusal of a workflow that is someone else's. */
	const NOT_YOURS = "Not yours";

	/** DENIED is why the refusal happened, and ADMIN what the theme's own page adds to it. */
	const DENIED = "Its owner or an admin changes a workflow.";

	const ADMIN = " An admin grants permissions under Settings. Pick a folder on the left to carry on.";

	/** ANYONE and OWNER name who the schedules run as, in the Schedule note. */
	const ANYONE = "the user who adds the schedule";

	const OWNER = ", the user set on the workflow";

	/** SCHEDULE is the note after the Schedule heading, around who the runs are as. */
	const SCHEDULE = "Runs the workflow on a cron expression, as ";

	const TIMELY = ". Each run needs permission to run every step at that time.";

	/** NONE is what the Schedule block says with no schedule. */
	const NONE = "Not scheduled.";

	/** MOVED is what the page says once the steps moved to another workspace. */
	const MOVED = "The steps moved to another workspace; Save keeps it.";

	/** MISSED says the same where a step lost its place, around how many did. */
	const MISSED = "The steps moved to another workspace; ";

	const LACKS = " of them say under their card what it lacks. Save keeps it.";

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
			"schedule",
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

		if (!$flow["canChange"]) {
			$this->denied($chrome, $view, $flow);

			return;
		}

		$chrome->heading("Edit " . $flow["name"], array(
			array("text" => "Workflows", "href" => "/workflows"),
			array("text" => $flow["name"], "href" => "/workflows/" . $id),
			array("text" => "Edit", "href" => ""),
		));

		$space = new Task\StepSpace($chrome);
		$head = new Task\Head("Edit " . $flow["name"], self::LEAD);
		$menu = $this->menu($space, $flow);
		$flows = $this->steppable($store, $chrome, $id);
		$editor = new Task\Editor($chrome, $space, $flows);
		$clock = new Task\Schedule($chrome->db);
		$rows = $clock->byWorkflow($id);
		$buttons = new Task\Buttons(true);
		$can = $flow["canRun"] || Task\User::admin($chrome->user);

		$head->menu($menu);
		$flow = $this->moved($chrome, $space, $flow);

		$view->page(array(
			"head" => $head->data(),
			"editor" => $editor->data($flow, "/api/workflows/", "PUT", "/workflows/" . $id, "Save"),
			"id" => $id,
			"block" => array(
				"title" => "Schedule",
				"note" => self::SCHEDULE . $this->runsAs($flow) . self::TIMELY,
				"empty" => count($rows) === 0 ? self::NONE : "",
				"buttons" => $buttons->data(),
			),
			"schedules" => $this->rows($chrome, $rows, $flow, $can),
		), "workflow_edit.tpl");
	}

	/**
	 * denied answers for a workflow that is not the reader's to change: the
	 * theme's own refusal page, under the crumbs of the workflow itself.
	 *
	 * It is not View::refused(), which says a reader may not open the page at
	 * all: this reader reads the workflow and may not write it, so the trail
	 * still leads back to it and the heading names what was refused.
	 */
	function denied($chrome, $view, $flow) {
		$head = new Task\Head(self::NOT_YOURS, self::DENIED . self::ADMIN);

		Task\Route::refused();
		$chrome->heading($flow["name"], array(
			array("text" => "Workflows", "href" => "/workflows"),
			array("text" => $flow["name"], "href" => ""),
		));
		$chrome->reports(self::DENIED);
		$view->page(array("head" => $head->data()), "forbidden.tpl");
	}

	/**
	 * steppable are the workflows a step of this one may run: the ones the
	 * reader sees, less any that reaches this workflow through its own steps.
	 *
	 * A workflow reaches itself, so the one being edited is left out by the
	 * same test that leaves out a cycle through a third one.
	 */
	function steppable($store, $chrome, $id) {
		$out = array();

		foreach ($store->visible($chrome) as $flow) {
			if (!$store->reaches($flow["id"], $id)) {
				$out[] = $flow;
			}
		}

		return $out;
	}

	/**
	 * moved is the workflow with its steps taken to the workspace a query
	 * named, and the notice that says what the move left behind; the workflow
	 * as it is where no query names one the steps may move to.
	 */
	function moved($chrome, $space, $flow) {
		$ws = Task\Route::input("workspace");

		if ($ws === "" || !$space->has($ws)) {
			return $flow;
		}

		$move = $space->move($flow["steps"], $ws);
		$flow["steps"] = $move["steps"];

		$chrome->notes($move["missed"] === 0 ? self::MOVED : self::MISSED . $move["missed"] . self::LACKS);

		return $flow;
	}

	/**
	 * menu is the head's [...]: the workflow's own page, where to take its
	 * steps, and Delete, whose confirm says the runs stay.
	 */
	function menu($space, $flow) {
		$id = $flow["id"];
		$menu = new Task\Menu("Actions on " . $flow["name"]);

		$menu->item("flow", "Open", array("href" => Task\Route::href("/workflows/" . $id)));
		$space->moveItems($menu, $id);
		$menu->danger("trash", "Delete", array(
			"data-delete" => "/api/workflows/" . $id,
			"data-confirm" => "Delete the workflow " . $flow["name"] . "? Its runs stay in the history.",
			"data-confirm-ok" => "Delete",
			"data-next" => Task\Route::href("/workflows"),
		));

		return $menu->data();
	}

	/** runsAs names who the schedules run as, which the Schedule note says. */
	function runsAs($flow) {
		return $flow["runAs"] === "" ? self::ANYONE : $flow["runAs"] . self::OWNER;
	}

	/**
	 * rows are the workflow's schedules as the table draws them: the
	 * expression, when it runs next and last, who it runs as and whether it is
	 * on, with the menu that turns it off or removes it.
	 *
	 * A workflow with a user of its own runs its schedules as that user, so the
	 * column states it and offers no pick: changing the schedule would not
	 * change who the run is.
	 */
	function rows($chrome, $schedules, $flow, $can) {
		$fixed = $flow["runAs"];
		$users = $chrome->runAsUsers();
		$out = array();

		foreach ($schedules as $schedule) {
			$on = Task\Schedule::enabled($schedule);
			$name = $fixed === "" ? $schedule->getUserName() : $fixed;
			$last = $schedule->getLastRunAt();

			$out[] = array(
				"expression" => $schedule->getExpression(),
				"on" => $on,
				"next" => $on ? explode(" ", Task\Clock::read($schedule->getNextRunAt())) : array(),
				"last" => $last === "" ? "" : Task\Clock::ago($last),
				"user" => $chrome->who()->of($name),
				"pick" => $fixed === "" && count($users) > 0,
				"options" => array(
					"users" => $users,
					"selected" => $name,
					"short" => true,
				),
				"api" => "/api/schedules/" . $schedule->getId(),
				"triage" => Task\Schedule::triages($schedule),
				"can" => $can,
				"menu" => $this->menuOf($schedule, $on, $flow["name"]),
			);
		}

		return $out;
	}

	/**
	 * menuOf is a schedule's [...]: Edit, then Enable or Disable, by whether
	 * it runs, and Delete.
	 *
	 * Edit is always offered, because the form that saves a schedule is on
	 * this page whenever the table is: both are drawn for the reader who may
	 * change the workflow.
	 */
	function menuOf($schedule, $on, $what) {
		$id = $schedule->getId();
		$expression = $schedule->getExpression();
		$menu = new Task\Menu("Actions on the schedule " . $expression);

		$menu->item("edit", "Edit", array("data-sched-edit" => "/api/schedules/" . $id, "data-sched-expression" => $expression));
		$menu->item($on ? "stop" : "play", $on ? "Disable" : "Enable", array(
			"data-api" => "/api/schedules/" . $id,
			"data-method" => "PUT",
			"data-body" => $on ? '{"enabled":false}' : '{"enabled":true}',
			"data-next" => "reload",
			"data-notice" => $on ? "Schedule disabled" : "Schedule enabled",
		));
		$menu->danger("trash", "Delete", array(
			"data-delete" => "/api/schedules/" . $id,
			"data-confirm" => "Delete the schedule " . $expression . " for " . $what . "?",
			"data-confirm-ok" => "Delete",
			"data-next" => "reload",
			"data-notice" => "Schedule deleted",
		));

		return $menu->data();
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workflow_edit;

	$mod->serve();
}
