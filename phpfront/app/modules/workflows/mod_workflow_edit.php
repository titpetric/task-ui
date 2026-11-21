<?php

// @route GET /workflows/{id}/edit

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workflow_edit is the form that changes a workflow, for its owner or an
// admin, and under it the schedules that start it.
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

	const ADMIN = " An admin grants permissions under Settings.";

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
		return [
			"workflow",
			"workflow_step",
			"workflow_run",
			"schedule",
			"user",
			"key_info",
			"key_role",
			"key_user",
			"workspace",
		];
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

		$chrome->heading("Edit " . $flow["name"], [
			["text" => "Workflows", "href" => "/workflows"],
			["text" => $flow["name"], "href" => "/workflows/" . $id],
			["text" => "Edit", "href" => ""],
		]);

		$space = new Task\StepSpace($chrome);
		$head = new Task\Head("Edit " . $flow["name"], self::LEAD);
		$menu = $this->menu($space, $flow);
		$flows = $this->steppable($store, $chrome, $id);
		$editor = new Task\Editor($chrome, $space, $flows);
		$clock = new Task\Schedules\Schedules($chrome->db);
		$rows = $clock->byWorkflow($id);
		$buttons = new Task\Buttons(true);
		$can = $flow["canRun"] || Task\User::admin($chrome->user);
		$canSchedule = $can && $chrome->allowed("schedule/edit");

		if ($canSchedule) {
			$buttons->primary("plus", "Add schedule", [
				"href" => Task\Route::href("/schedules/add") . "?workflow=" . Task\Route::queryEscape($id),
			]);
		}

		$head->menu($menu);
		$flow = $this->moved($chrome, $space, $flow);
		$view->page([
			"head" => $head->data(),
			"editor" => $editor->data($flow, "/api/workflows/", "PUT", "/workflows/" . $id, "Save"),
			"id" => $id,
			"block" => [
				"title" => "Schedule",
				"note" => self::SCHEDULE . $this->runsAs($flow) . self::TIMELY,
				"empty" => count($rows) === 0 ? self::NONE : "",
				"buttons" => $buttons->data(),
			],
			"schedules" => $this->rows($chrome, $rows, $flow, $canSchedule),
			"can" => $canSchedule,
		], "workflow_edit.tpl");
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
		$chrome->heading($flow["name"], [
			["text" => "Workflows", "href" => "/workflows"],
			["text" => $flow["name"], "href" => ""],
		]);
		$chrome->reports(self::DENIED);
		$view->page(["head" => $head->data()], "forbidden.tpl");
	}

	/**
	 * steppable are the workflows a step of this one may run: the ones the
	 * reader sees, less any that reaches this workflow through its own steps.
	 *
	 * A workflow reaches itself, so the one being edited is left out by the
	 * same test that leaves out a cycle through a third one.
	 */
	function steppable($store, $chrome, $id) {
		$out = [];

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
		$menu->item("flow", "Open", ["href" => Task\Route::href("/workflows/" . $id)]);
		$space->moveItems($menu, $id);
		$menu->danger("trash", "Delete", [
			"data-delete" => "/api/workflows/" . $id,
			"data-confirm" => "Delete the workflow " . $flow["name"] . "? Its runs stay in the history.",
			"data-confirm-ok" => "Delete",
			"data-next" => Task\Route::href("/workflows"),
		]);

		return $menu->data();
	}

	/** runsAs names who the schedules run as, which the Schedule note says. */
	function runsAs($flow) {
		return $flow["runAs"] === "" ? self::ANYONE : $flow["runAs"] . self::OWNER;
	}

	/**
	 * rows are the workflow's schedules as the table draws them: the
	 * expression, when it runs next and last, who it runs as and whether it is
	 * on, a toggle the table saves, with the menu that removes it.
	 *
	 * A workflow with a user of its own runs its schedules as that user, so the
	 * column states it and offers no pick: changing the schedule would not
	 * change who the run is.
	 */
	function rows($chrome, $schedules, $flow, $can) {
		$fixed = $flow["runAs"];
		$users = $chrome->runAsUsers();
		$out = [];

		foreach ($schedules as $schedule) {
			$on = Task\Schedule::enabled($schedule);
			$name = $fixed === "" ? $schedule->getUserName() : $fixed;
			$last = $schedule->getLastRunAt();

			$out[] = [
				"id" => $schedule->getId(),
				"expression" => $schedule->getExpression(),
				"on" => $on,
				"next" => $on ? explode(" ", Task\Clock::read($schedule->getNextRunAt())) : [],
				"last" => $last === "" ? "" : Task\Clock::ago($last),
				"user" => $chrome->who()->of($name),
				"pick" => $fixed === "" && count($users) > 0,
				"options" => [
					"users" => $users,
					"selected" => $name,
					"short" => true,
				],
				"api" => "/api/schedules/" . $schedule->getId(),
				"triage" => Task\Schedule::triages($schedule),
				"can" => $can,
				"menu" => $this->menuOf($schedule, $flow["name"], $can),
			];
		}

		return $out;
	}

	/** menuOf is a schedule's [...]: Edit and Delete; its flags are columns of the table. */
	function menuOf($schedule, $what, $edits = true) {
		return Task\Schedule::menu($schedule, $what, $edits);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workflow_edit;
	$mod->serve();
}
