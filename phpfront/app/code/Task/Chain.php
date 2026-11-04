<?php

namespace Task;

// Chain turns steps into the links a page draws them as: a folder named once
// for the steps that follow in it, each linked to its own page or to the run
// the step made of it.
//
// A link is an array and not a model because no table holds one: it is what a
// step looks like once the workspaces have placed it. blank() names every
// field, so a template never tests for a key.
class Chain {
	/** blank is a link with nothing set. */
	static function blank() {
		return array(
			"text" => "",
			"href" => "",
			"state" => "",
			"run_id" => "",
			"workflow" => false,
			"workflow_id" => "",
			"workflow_run_id" => "",
			"prompt" => "",
			"summary" => false,
			"check" => false,
			"disabled" => false,
			"sub" => array(),
			"frame" => "",
			"complete" => false,
			"lines" => 0,
		);
	}

	/**
	 * of is the steps as links, the workspaces the named ones, which place a
	 * job's folder.
	 *
	 * A step that made a run links to the run and not to the job: what a
	 * reader wants off a pipeline is the output.
	 */
	static function of($steps, $workspaces) {
		$links = array();
		$last = "";

		foreach ($steps as $step) {
			$kind = Step::kind($step);
			$link = self::blank();

			$link["state"] = $step["state"];
			$link["run_id"] = $step["run_id"];
			$link["disabled"] = $step["disabled"];

			if ($kind === Step::SUMMARY) {
				$link["summary"] = true;
				$link["text"] = "summary in " . Step::place($step);
				$link["href"] = Route::folderHref($step["folder"], $workspaces);
				$last = "";
			} elseif ($kind === Step::PROMPT) {
				$link["prompt"] = $step["prompt"];
				$link["text"] = "prompt in " . Step::place($step);
				$link["href"] = Route::folderHref($step["folder"], $workspaces);
				$last = "";
			} elseif ($kind === Step::CHECK) {
				$link["check"] = true;
				$link["text"] = "check for " . $step["markers"];
				$link["href"] = "/workflows";
				$last = "";
			} elseif ($kind === Step::WORKFLOW) {
				$link["workflow"] = true;
				$link["text"] = $step["workflow_name"] === "" ? "a workflow that is gone" : $step["workflow_name"];
				$link["href"] = "/workflows/" . $step["workflow_id"];
				$link["workflow_id"] = $step["workflow_id"];
				$link["workflow_run_id"] = $step["workflow_run_id"];
				$last = "";
			} else {
				$link["text"] = $step["folder"] === $last ? $step["job"] : $step["folder"] . " / " . $step["job"];
				$link["href"] = Route::folderHref($step["folder"], $workspaces) . "/" . $step["job"];
				$last = $step["folder"];
			}

			if ($step["run_id"] !== "") {
				$link["href"] = "/runs/" . $step["run_id"];
			}

			$links[] = $link;
		}

		return $links;
	}

	/**
	 * pipeline is the workflow's steps top to bottom, with the states of a
	 * latest run that has the same steps.
	 *
	 * A run's step keeps no directory of its prompt and the workflow's step
	 * has it, so the directory comes off the workflow either way.
	 */
	static function pipeline($steps, $last, $workspaces) {
		if (count($last) !== count($steps) || count($steps) === 0) {
			return self::of($steps, $workspaces);
		}

		foreach ($steps as $i => $step) {
			$ran = $last[$i];

			if ($step["folder"] !== $ran["folder"] || $step["job"] !== $ran["job"] || $step["workflow_id"] !== $ran["workflow_id"]) {
				return self::of($steps, $workspaces);
			}

			$last[$i]["dir"] = $step["dir"];
		}

		return self::of($last, $workspaces);
	}

	/**
	 * frames puts on each link the screen the step's run ended on, and under a
	 * step that ran a workflow the chain of that run, one level down.
	 *
	 * deeper is false on the links of that inner chain, which is what stops the
	 * walk: a workflow that runs a workflow draws two levels and no more,
	 * however deep the steps go.
	 *
	 * seen keeps a run the reader may not see out of that inner chain, as
	 * Step::seen does for the outer one.
	 */
	static function frames($links, $db, $names, $workspaces, $deeper, $seen) {
		$runs = new Run($db);
		$store = new WorkflowRun($db);

		foreach ($links as $i => $link) {
			if ($link["run_id"] !== "") {
				$shown = self::screen($runs, $link["run_id"]);
				$links[$i]["frame"] = $shown["svg"];
				$links[$i]["lines"] = $shown["lines"];
				$links[$i]["complete"] = $shown["complete"];
			}

			if ($deeper && $link["workflow_run_id"] !== "") {
				$ran = Step::ofRun($store->steps($link["workflow_run_id"]), $names);
				$steps = Step::seen($seen, $ran);
				$sub = self::of($steps, $workspaces);
				$links[$i]["sub"] = self::frames($sub, $db, $names, $workspaces, false, $seen);
			}
		}

		return $links;
	}

	/**
	 * screen is the last frame of a run, replayed off the recording the run
	 * wrote, with the lines its whole output has.
	 *
	 * The row is read here and not by the host because the recording's path and
	 * the columns it was written at are columns of it, and the binding answers
	 * no table.
	 */
	static function screen($runs, $id) {
		$run = $runs->get($id);

		if ($run === null) {
			return array(
				"svg" => "",
				"lines" => 0,
				"complete" => false,
			);
		}

		return Host::screen($id, $run->getCastPath(), (int)$run->getTermCols());
	}
}
