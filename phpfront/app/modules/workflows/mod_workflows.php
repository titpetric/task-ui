<?php

// @route GET /workflows

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_workflows lists the workflows the reader may see, each with the
// schedules that start it and how its last run ended, and under them the
// newest runs of any of them.
//
// A workflow the reader may not see every folder of is not listed at all, so
// the page and the sidebar hold the same set: Task\Workflow::visible decides
// it once.
class mod_workflows implements Module {
	/** LEAD is the paragraph under the heading. */
	const LEAD = "Chains of steps: commands from any folders, prompts to Claude and summaries. Each step starts when the one before ends ok; a failure stops the rest, but for a summary. Running one needs permission to run every step.";

	/** RECENT is how many runs the Recent runs block lists. */
	const RECENT = 15;

	/** NONE is what each block says with nothing in it. */
	const NONE = "No workflows yet. Add one with a name and its steps, in order.";

	const NEVER = "No workflow has run yet.";

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
			"schedule",
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

		$chrome->heading("Workflows", array(array("text" => "Workflows", "href" => "")));

		$store = new Task\Workflow($chrome->db);
		$flows = $store->visible($chrome);
		$clock = new Task\Schedule($chrome->db);
		$schedules = $clock->all();
		$rows = array();

		foreach ($flows as $flow) {
			$rows[] = $this->row($chrome, $flow, $schedules);
		}

		$runs = $this->recent($chrome, $flows);
		$buttons = new Task\Buttons;
		$head = new Task\Head("Workflows", self::LEAD);

		$buttons->primary("plus", "Add workflow", array("href" => Task\Route::href("/workflows/new")));
		$head->buttons($buttons->data());

		$view->page(array(
			"head" => $head->data(),
			"prompt" => $this->prompt($chrome),
			"block" => $this->block("Workflows", count($rows) === 0 ? self::NONE : ""),
			"rows" => $rows,
			"runsBlock" => $this->block("Recent runs", count($runs) === 0 ? self::NEVER : ""),
			"runs" => $runs,
		));
	}

	/** block is the head of one of the page's two blocks, with what it says when empty. */
	function block($title, $empty) {
		$buttons = new Task\Buttons(true);

		return array(
			"title" => $title,
			"note" => "",
			"empty" => $empty,
			"buttons" => $buttons->data(),
		);
	}

	/**
	 * row is one workflow of the table: its name, the schedules that start it,
	 * how it last ran, who it runs as and the menu of what may be done to it.
	 *
	 * The Run button carries the workflow for task-ui.js to post, and the name
	 * as well where a step prompts Claude, which is what the script asks the
	 * input for before it starts the run.
	 */
	function row($chrome, $flow, $schedules) {
		$id = $flow["id"];
		$last = $flow["last"];
		$menu = new Task\Menu("Actions on " . $flow["name"]);

		$menu->when($flow["canChange"], "edit", "Edit", array("href" => Task\Route::href("/workflows/" . $id . "/edit")));
		$menu->when($flow["running"] !== "", "eye", "The run going", array("href" => Task\Route::href("/workflows/" . $id . "/runs/" . $flow["running"])));

		return array(
			"id" => $id,
			"name" => $flow["name"],
			"description" => $flow["description"],
			"schedules" => self::hints($flow, $schedules),
			"running" => $flow["running"],
			"last" => $last !== null,
			"state" => $last === null ? "" : $last->getState(),
			"ago" => $last === null ? "" : Task\Clock::ago($last->getStartedAt()),
			"stamp" => $last === null ? "" : Task\Clock::read($last->getStartedAt()),
			"user" => $chrome->who()->of($flow["userName"]),
			"runAs" => $flow["runAs"] !== "" && $flow["runAs"] !== $flow["userName"] ? $flow["runAs"] : "",
			"canRun" => $flow["canRun"],
			"prompted" => $flow["prompted"],
			"runDenied" => $flow["runDenied"],
			"menu" => $menu->data(),
		);
	}

	/**
	 * hints are the schedules that start the workflow, as the lines under its
	 * name: the expression, whether it is on, and who its runs run as.
	 *
	 * A workflow that names the user it runs as names it for its schedules
	 * too, however they were made, so the hover says who the run would be.
	 */
	static function hints($flow, $schedules) {
		$out = array();

		foreach ($schedules as $schedule) {
			if ($schedule->getWorkflowId() !== $flow["id"]) {
				continue;
			}

			$on = Task\Schedule::enabled($schedule);
			$who = $flow["runAs"] === "" ? $schedule->getUserName() : $flow["runAs"];

			$out[] = array(
				"expression" => $schedule->getExpression(),
				"enabled" => $on,
				"title" => ($on ? "Scheduled" : "A schedule, disabled") . ", as " . $who,
			);
		}

		return $out;
	}

	/**
	 * recent is the newest runs of the workflows the reader sees, at most
	 * RECENT of them, each as the row the table draws with its answer open
	 * under it.
	 *
	 * Four times the page's own is read because a run of a workflow the reader
	 * may not see is dropped after the query and not by it: the table is filled
	 * from what is left rather than left short.
	 */
	function recent($chrome, $flows) {
		if (count($flows) === 0) {
			return array();
		}

		$names = array();

		foreach ($flows as $flow) {
			$names[$flow["id"]] = $flow["name"];
		}

		$store = new Task\WorkflowRun($chrome->db);
		$out = array();

		foreach ($store->latest(4 * self::RECENT) as $run) {
			if (count($out) === self::RECENT) {
				break;
			}

			if (isset($names[$run->getWorkflowId()]) && $store->sees($chrome->user, $run->getId())) {
				$out[] = self::thread($chrome, $run);
			}
		}

		return $out;
	}

	/**
	 * thread is one run of the Recent runs table: its state on the dot beside
	 * the workflow's name, and Claude's answer in the row under it.
	 *
	 * The answer is open from the start here and folded on the workflow's own
	 * page, which is what the chevron's `open` says.
	 */
	static function thread($chrome, $run) {
		$id = $run->getId();
		$answer = $run->getAnswer();
		$message = $run->getMessage();

		return array(
			"id" => $id,
			"workflowId" => $run->getWorkflowId(),
			"name" => $run->getName(),
			"state" => $run->getState(),
			"title" => $message === "" ? $run->getState() : $run->getState() . ", " . $message,
			"user" => $chrome->who()->of($run->getUserName()),
			"ago" => Task\Clock::ago($run->getStartedAt()),
			"stamp" => Task\Clock::read($run->getStartedAt()),
			"fold" => array(
				"id" => $id,
				"answer" => $answer !== "",
				"open" => true,
			),
			"answer" => array(
				"id" => $id,
				"body" => $answer === "" ? "" : $chrome->message($answer),
				"cols" => 3,
				"open" => true,
			),
		);
	}

	/**
	 * prompt is the box that has Claude build or change a workflow, for a
	 * reader who may prompt on an instance with a token for it, and the threads
	 * already asked from this page under it. Null where the page draws none.
	 *
	 * The box asks in the root unless its select picks a folder, so the pick is
	 * the catalogue the reader sees and the root comes first, named as the root
	 * because a workflow belongs to no folder.
	 */
	function prompt($chrome) {
		if (!$chrome->allowed("prompt/edit") || !Task\Host::claude()) {
			return null;
		}

		$notes = new Task\Note($chrome->db);
		$users = new Task\User($chrome->db);
		$roots = $notes->ofArea(Task\Note::AREA_WORKFLOWS);
		$replies = $notes->replies(Task\Note::ids($roots));
		$agents = $users->agents();
		$threads = array();

		foreach ($roots as $note) {
			$under = isset($replies[$note->getId()]) ? $replies[$note->getId()] : array();
			$answerer = Task\Note::answerer($note, $under, $agents);

			if (self::seesPlace($chrome, $note->getDir())) {
				$threads[] = Task\Note::thread($note, $replies, $chrome, $answerer, false, null);
			}
		}

		return array(
			"title" => "Prompt",
			"note" => "Claude Code builds or changes a workflow with the task tools, runs it and answers with its link",
			"open" => true,
			"hidden" => false,
			"watch" => false,
			"api" => Task\Prompt::API,
			"fields" => array(array("name" => "area", "value" => Task\Note::AREA_WORKFLOWS)),
			"placeholder" => "Describe a workflow to build or change",
			"lead" => "Runs @claude-prompt as ",
			"tail" => " in ",
			"picks" => $this->picks($chrome),
			"after" => " with the task tools: it saves the workflow and runs it up to 3 times.",
			"options" => $chrome->runAs(),
			"place" => Task\Route::INTERNAL,
			"area" => Task\Note::AREA_WORKFLOWS,
			"threads" => $threads,
		);
	}

	/**
	 * seesPlace reports whether the reader may read a thread asked in a place:
	 * the root only for a reader with every folder, and anywhere else the
	 * folder of theirs that holds it.
	 *
	 * A thread of this page's own asks in the root unless its select picked a
	 * folder, so the root is the common case and it is the one a reader with a
	 * folder list does not read.
	 */
	static function seesPlace($chrome, $place) {
		if ($place === Task\Route::INTERNAL) {
			return count(Task\User::folders($chrome->user)) === 0;
		}

		return Task\Route::heldHref($place, $chrome->folders(), $chrome->workspaces()) !== "";
	}

	/** picks are the folders the prompt box offers to run in, the root first. */
	function picks($chrome) {
		$out = array(
			array("value" => Task\Route::INTERNAL, "text" => Task\Route::INTERNAL . " (the root)"),
		);

		foreach ($chrome->folders() as $folder) {
			$id = Task\Row::text($folder, "id");
			$out[] = array("value" => $id, "text" => $id);
		}

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_workflows;

	$mod->serve();
}
