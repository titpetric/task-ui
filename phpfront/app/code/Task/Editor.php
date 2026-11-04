<?php

namespace Task;

// Editor is the workflow form, new and edit alike, as the step editor draws
// it: the name, the description, the sidebar toggle, who the runs run as and
// the workspace they run over, then the steps in order, each a card of its
// kind, and one blank card a kind for the script to clone.
//
// It is one class and not two pages because both pages draw the one form, and
// the blank cards are what task-ui.js adds a step from: a card the editor did
// not build is a step the script cannot read.
//
// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.
class Editor {
	/** BLANK is what the fields of a new workflow are named by, where there is no ID. */
	const BLANK = "new";

	/** NOBODY is what the Runs as field reads for a workflow that runs as whoever starts it. */
	const NOBODY = "Whoever starts it";

	/** FLOWS is the workspace of a step that runs another workflow, which belongs to none. */
	const FLOWS = "workflow:";

	var $chrome;

	/** space is what a step picks a place from, read once a page. */
	var $space;

	/** flows are the workflows a step may run, each an id and a name. */
	var $flows = array();

	/** users are who the runs may run as, empty for a reader who may not pick. */
	var $users = array();

	/** asks and shell are the kinds of step the reader may add. */
	var $asks = false;

	var $shell = false;

	function __construct($chrome, $space, $flows) {
		$this->chrome = $chrome;
		$this->space = $space;
		$this->flows = $flows;
		$this->users = $chrome->runAsUsers();
		$this->asks = $chrome->allowed("prompt/edit");
		$this->shell = $chrome->allowed("shell");
	}

	/**
	 * data is the scope components/stepeditor.tpl reads: the form's fields for
	 * the workflow, null for a new one, and where the script posts it.
	 *
	 * api takes the workflow's ID after it on edit, which is the one place the
	 * two pages differ in what they post to.
	 */
	function data($flow, $api, $method, $next, $label) {
		$runAs = $flow === null ? $this->chrome->name() : $flow["runAs"];
		$over = $flow === null ? "" : $flow["runsOver"];
		$steps = $flow === null ? array() : $flow["steps"];
		$cards = array();

		foreach ($steps as $step) {
			$cards[] = $this->card($step, false);
		}

		return array(
			"api" => $flow === null ? $api : $api . $flow["id"],
			"method" => $method,
			"next" => $next,
			"label" => $label,
			"id" => $flow === null ? self::BLANK : $flow["id"],
			"name" => $flow === null ? "" : $flow["name"],
			"description" => $flow === null ? "" : $flow["description"],
			"sidebar" => $flow === null || !$flow["sidebarHidden"],
			"users" => array(
				"users" => $this->users,
				"selected" => $runAs,
				"short" => false,
			),
			"runAs" => $runAs === "" ? self::NOBODY : $runAs,
			"runAsValue" => $runAs,
			"isNew" => $flow === null,
			"overNone" => $over === "",
			"overs" => $this->overs($over),
			"asks" => $this->asks,
			"shell" => $this->shell,
			"steps" => $cards,
			"emptyHidden" => $flow !== null && count($cards) > 0,
			"blankJob" => $this->blank(Step::JOB),
			"blankScript" => $this->blank(Step::SCRIPT),
			"blankPrompt" => $this->blank(Step::PROMPT),
			"blankSummary" => $this->blank(Step::SUMMARY),
		);
	}

	/** overs are the workspaces the steps may run over, the one picked marked. */
	function overs($over) {
		$out = array();

		foreach ($this->space->spaces() as $one) {
			$out[] = array(
				"id" => $one["id"],
				"title" => $one["title"],
				"selected" => $one["id"] === $over,
			);
		}

		return $out;
	}

	/** blank is the card a kind of step is added from, open and with nothing set. */
	function blank($kind) {
		$step = Step::added($kind);

		return $this->card($step, true);
	}

	/**
	 * card is one step of the form: its kind, the one line its head shows while
	 * folded, and the fields of its kind under it.
	 *
	 * Every field of every kind is answered, because the card is one template
	 * and the kind decides which half of it is drawn. A place a step names that
	 * the reader has no folder for is offered as an option of its own, so saving
	 * the form does not quietly move the step.
	 */
	function card($step, $open) {
		$kind = Step::kind($step);
		$folder = $step["folder"];
		$runs = $kind === Step::JOB || $kind === Step::WORKFLOW;
		$places = $this->places($folder);
		$found = $folder === Route::INTERNAL;

		foreach ($places as $place) {
			if ($place["selected"]) {
				$found = true;
			}
		}

		return array(
			"kind" => $kind,
			"disabled" => $step["disabled"],
			"open" => $open,
			"icon" => Step::icon($step),
			"kindLabel" => Step::kindLabel($step),
			"line" => Step::line($step),
			"menu" => $this->menu($step["disabled"]),
			"missing" => $step["missing"],
			"runs" => $runs,
			"script" => $kind === Step::SCRIPT,
			"asks" => $kind === Step::PROMPT,
			"sums" => $kind === Step::SUMMARY,
			"folderLabel" => $kind === Step::WORKFLOW ? "Workflow" : "Folder",
			"pickNone" => $folder === "" && $step["workflow_id"] === "",
			"places" => $places,
			"hosts" => count($this->space->hostJobs) > 0,
			"rootSelected" => $folder === Route::INTERNAL,
			"flows" => $this->flowOptions($step["workflow_id"]),
			"jobs" => $this->jobs($step),
			"hostJobs" => HostJobs::of($this->chrome, $step["job"]),
			"args" => Schedule::argLine($step["job"], $step["args"]),
			"command" => Step::command($step),
			"prompt" => $step["prompt"],
			"none" => !$runs && $kind !== Step::SCRIPT && $this->space->none,
			"noneSelected" => $folder === "",
			"extra" => $found ? null : $this->extra($folder),
			"dir" => $step["dir"],
			"question" => Step::SUMMARY_QUESTION,
			"instruction" => Step::SUMMARY_PROMPT,
			"space" => $this->spaceOf($kind, $folder),
		);
	}

	/**
	 * spaceOf is the Workspace field of a card: the workspace of the step's
	 * folder, the workflows for a step that runs one, and none at all for a
	 * prompt or a summary that asks at the workspaces root.
	 */
	function spaceOf($kind, $folder) {
		$runs = $kind === Step::JOB || $kind === Step::WORKFLOW;
		$sel = $this->space->spaceOf($folder);

		if ($kind === Step::WORKFLOW) {
			$sel = self::FLOWS;
		}

		if (!$runs && $kind !== Step::SCRIPT && $folder === "") {
			$sel = "";
		}

		$spaces = array();

		foreach ($this->space->spaces() as $one) {
			$spaces[] = array(
				"id" => $one["id"],
				"title" => $one["title"],
				"selected" => $one["id"] === $sel,
			);
		}

		return array(
			"none" => !$runs && $kind !== Step::SCRIPT && $this->space->none,
			"noneSelected" => $sel === "",
			"spaces" => $spaces,
			"flows" => $runs && count($this->flows) > 0,
			"flowsSelected" => $sel === self::FLOWS,
		);
	}

	/** places are the folders a step may run in, the one it names marked. */
	function places($selected) {
		$out = array();

		foreach ($this->space->folders as $folder) {
			$id = Row::text($folder, "id");
			$out[] = array(
				"id" => $id,
				"workspace" => $this->space->spaceOf($id),
				"rel" => $this->space->rel($id),
				"label" => $this->space->label($id),
				"selected" => $id === $selected,
			);
		}

		return $out;
	}

	/** extra is the option of a place the reader has no folder for, which a card keeps. */
	function extra($folder) {
		if ($folder === "") {
			return null;
		}

		return array(
			"id" => $folder,
			"workspace" => $this->space->spaceOf($folder),
			"rel" => $this->space->rel($folder),
			"label" => $this->space->label($folder),
		);
	}

	/** flowOptions are the workflows a step may run, the one it names marked. */
	function flowOptions($selected) {
		$out = array();

		foreach ($this->flows as $flow) {
			$out[] = array(
				"id" => $flow["id"],
				"name" => $flow["name"],
				"selected" => $flow["id"] === $selected,
			);
		}

		return $out;
	}

	/**
	 * jobs are every job of every folder a step may pick, each carrying the
	 * folder it belongs to: the script narrows the list to the folder picked,
	 * so the markup holds them all.
	 */
	function jobs($step) {
		$out = array();

		foreach ($this->space->folders as $folder) {
			$id = Row::text($folder, "id");

			foreach (Row::items($folder, "jobs") as $job) {
				$key = Row::text($job, "id");
				$out[] = array(
					"id" => $key,
					"folder" => $id,
					"name" => Row::text($job, "name"),
					"selected" => $id === $step["folder"] && $key === $step["job"],
				);
			}
		}

		return $out;
	}

	/**
	 * menu is the [...] menu of a step card: move it, turn it off or on,
	 * duplicate it and remove it, all of them the script's.
	 */
	function menu($disabled) {
		$menu = new Menu("Actions on the step");

		$menu->item("up", "Move up", array("data-step-move" => "up"));
		$menu->item("down", "Move down", array("data-step-move" => "down"));
		$menu->sep();
		$menu->item($disabled ? "play" : "stop", $disabled ? "Enable" : "Disable", array("data-step-off" => "1"));
		$menu->item("copy", "Duplicate", array("data-step-duplicate" => "1"));
		$menu->danger("trash", "Remove", array("data-step-remove" => "1"));

		return $menu->data();
	}
}
