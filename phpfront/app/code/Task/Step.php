<?php

namespace Task;

// Step is a step of a workflow, or of a run of one, as a page draws it: a job
// of a folder, another workflow, a prompt to Claude, or Claude's summary of
// the steps before it.
//
// A step is an array and not a model because one shape comes out of two
// tables: workflow_step is what the editor writes and workflow_run_step what
// a run recorded, and the one card and the one chain draw either. blank()
// names every field, so a reader of a step never tests for a key.
class Step {
	/** The kinds of a step, as model/workflow.go names them. */
	const JOB = "job";

	const WORKFLOW = "workflow";

	const PROMPT = "prompt";

	const SUMMARY = "summary";

	/**
	 * SCRIPT is the kind the editor shows a job step of SCRIPT_JOB as: a
	 * command the step names, run by bash in the folder's directory. The
	 * table keeps it as a job step and the editor keeps it apart, because
	 * what it asks for is a command and not a job.
	 */
	const SCRIPT = "script";

	const SCRIPT_JOB = "@script";

	/** SUMMARY_QUESTION is what a summary step asks Claude, as note/board_workflow.go writes it. */
	const SUMMARY_QUESTION = "Summarize this workflow run.";

	/**
	 * SUMMARY_PROMPT is the instruction beside it, claude.RunSummary. The
	 * summary card quotes both under Claude gets, so an editor shows what the
	 * step sends without a run of one.
	 */
	const SUMMARY_PROMPT = "Summarise the run in one or two short paragraphs: what it did and how it ended, and the cause if it failed. " . "Then, only if the output has values worth seeing at a glance (versions, images, counts, durations, failures per step, hosts), a markdown table of them. " . "No preamble, no headings, no advice unless it failed, don't explain the obvious. Change nothing.";

	/** blank is a step with nothing set, which every field of one. */
	static function blank() {
		return array(
			"kind" => "",
			"folder" => "",
			"job" => "",
			"workflow_id" => "",
			"workflow_name" => "",
			"disabled" => false,
			"args" => "",
			"prompt" => "",
			"dir" => "",
			"missing" => "",
			"state" => "",
			"run_id" => "",
			"workflow_run_id" => "",
		);
	}

	/**
	 * added is a blank step of a kind, as the editor's template for that kind
	 * holds it: a prompt or a summary asks in the root, and a script is a job
	 * step of SCRIPT_JOB.
	 */
	static function added($kind) {
		$step = self::blank();
		$step["kind"] = $kind;

		if ($kind === self::PROMPT || $kind === self::SUMMARY) {
			$step["folder"] = Route::INTERNAL;
		}

		if ($kind === self::SCRIPT) {
			$step["kind"] = self::JOB;
			$step["job"] = self::SCRIPT_JOB;
		}

		return $step;
	}

	/** of is a workflow's steps as a page draws them, names the name by workflow ID. */
	static function of($steps, $names) {
		$out = array();

		foreach ($steps as $row) {
			$ref = $row->getWorkflowRef();
			$step = self::blank();

			$step["kind"] = self::resolve($row->getKind(), $row->getPrompt(), $ref);
			$step["folder"] = $row->getFolder();
			$step["job"] = $row->getJob();
			$step["workflow_id"] = $ref;
			$step["workflow_name"] = isset($names[$ref]) ? $names[$ref] : "";
			$step["disabled"] = $row->getDisabled();
			$step["args"] = $row->getArgs();
			$step["prompt"] = $row->getPrompt();
			$step["dir"] = $row->getDir();
			$out[] = $step;
		}

		return $out;
	}

	/**
	 * ofRun is a run's steps as its page draws them: how each ended, the run
	 * it made of a job and the workflow run it made of a workflow.
	 *
	 * A run's step keeps no directory of its prompt, which is why the chain of
	 * a pipeline takes the workflow's directory over it.
	 */
	static function ofRun($rows, $names) {
		$out = array();

		foreach ($rows as $row) {
			$ref = $row->getWorkflowRef();
			$step = self::blank();

			$step["kind"] = self::resolve($row->getKind(), $row->getPrompt(), $ref);
			$step["folder"] = $row->getFolder();
			$step["job"] = $row->getJob();
			$step["workflow_id"] = $ref;
			$step["workflow_name"] = isset($names[$ref]) ? $names[$ref] : "";
			$step["state"] = $row->getState();
			$step["run_id"] = $row->getRunId();
			$step["workflow_run_id"] = $row->getWorkflowRunRef();
			$step["prompt"] = $row->getPrompt();
			$out[] = $step;
		}

		return $out;
	}

	/**
	 * seen blanks the run of a step whose run the reader may not see.
	 *
	 * A chain links a step's run as /runs/{id} and that page refuses what
	 * seeing the step's folder does not give: a prompt's run takes prompt/view
	 * and the Notes agent's takes note/view. A step with no run links its own
	 * page, and no frame of the run is drawn either, which a run the page
	 * refuses must not show.
	 *
	 * A run that is gone is blanked the same way, as a note's is: nothing is
	 * left to link to.
	 */
	static function seen($seen, $steps) {
		foreach ($steps as $i => $step) {
			if ($step["run_id"] !== "" && !$seen->run($step["run_id"])) {
				$steps[$i]["run_id"] = "";
			}
		}

		return $steps;
	}

	/** resolve is a row's kind, from its fields for a row written without one. */
	static function resolve($kind, $prompt, $ref) {
		if ($kind !== "") {
			return $kind;
		}

		if ($prompt !== "") {
			return self::PROMPT;
		}

		return $ref === "" ? self::JOB : self::WORKFLOW;
	}

	/** kind is what the step does, SCRIPT for a job step of SCRIPT_JOB. */
	static function kind($step) {
		$kind = self::resolve($step["kind"], $step["prompt"], $step["workflow_id"]);

		if ($kind === self::JOB && $step["job"] === self::SCRIPT_JOB) {
			return self::SCRIPT;
		}

		return $kind;
	}

	/** asks reports whether the step asks Claude: a prompt or a summary. */
	static function asks($step) {
		$kind = self::kind($step);

		return $kind === self::PROMPT || $kind === self::SUMMARY;
	}

	/** command is the command of a script step, its one argument. */
	static function command($step) {
		$list = Text::items($step["args"]);

		return implode(" ", $list);
	}

	/** kindLabel names the step's kind as the editor shows it. */
	static function kindLabel($step) {
		$kind = self::kind($step);

		if ($kind === self::WORKFLOW) {
			return "Workflow";
		}

		if ($kind === self::PROMPT) {
			return "Prompt";
		}

		if ($kind === self::SUMMARY) {
			return "Summary";
		}

		return $kind === self::SCRIPT ? "Script" : "Command";
	}

	/** icon is the symbol of the step's kind, #i-<icon> of the sprite. */
	static function icon($step) {
		$kind = self::kind($step);

		if ($kind === self::WORKFLOW) {
			return "flow";
		}

		if ($kind === self::PROMPT) {
			return "prompt";
		}

		if ($kind === self::SUMMARY) {
			return "summary";
		}

		return $kind === self::SCRIPT ? "script" : "terminal";
	}

	/**
	 * where is where a prompt or a summary step asks Claude: the folder and
	 * the directory inside it, "" for no workspace, which asks at the
	 * workspaces root.
	 */
	static function where($step) {
		if ($step["folder"] === "") {
			return "";
		}

		return Route::place($step["folder"], $step["dir"]);
	}

	/**
	 * label names a place for a reader: the workspace root for none, the root
	 * for the internal folder, else the place itself.
	 */
	static function label($place) {
		if ($place === "") {
			return "the workspace root";
		}

		return $place === Route::INTERNAL ? "the root" : $place;
	}

	/** place is where the step asks, as a reader reads it. */
	static function place($step) {
		return self::label(self::where($step));
	}

	/**
	 * line is the step in one line, its card's heading while folded: the job
	 * and its arguments, the workflow, the first line of the prompt, or where
	 * the summary asks.
	 */
	static function line($step) {
		$kind = self::kind($step);

		if ($kind === self::WORKFLOW) {
			return $step["workflow_name"] === "" ? "a workflow that is gone" : $step["workflow_name"];
		}

		if ($kind === self::PROMPT) {
			return self::first($step["prompt"]);
		}

		if ($kind === self::SUMMARY) {
			return "of the steps before it, in " . self::place($step);
		}

		if ($kind === self::SCRIPT) {
			$line = self::first(self::command($step));

			if ($step["folder"] === "" || $line === "") {
				return "";
			}

			return $step["folder"] . ": " . $line;
		}

		if ($step["job"] === "") {
			return "";
		}

		$line = $step["folder"] . " / " . $step["job"];
		$args = Schedule::argLine($step["job"], $step["args"]);

		return $args === "" ? $line : $line . " " . $args;
	}

	/** first is the first line of a text, trimmed, which a folded card shows. */
	static function first($text) {
		$parts = explode("\n", trim($text));

		return $parts[0];
	}
}
