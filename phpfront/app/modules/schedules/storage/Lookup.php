<?php

namespace Task\Schedules;

// Lookup is what the schedules page reads of the one table it does not own:
// the workflows a schedule may start, with the user they run as and the
// folders their steps name.
//
// The rows are the workflows module's and this asks that module for them;
// which workflow a schedule may start is the schedules page's question.
class Lookup {
	/** STEPS are the kinds of a workflow step, as model/workflow.go names them. */
	const STEP_JOB = "job";

	const STEP_WORKFLOW = "workflow";

	const STEP_PROMPT = "prompt";

	/** store is the workflows module, which owns the rows this reads. */
	var $store;

	function __construct($db = null) {
		$this->store = new \Task\Workflows\Workflows($db);
	}

	/**
	 * workflows is every workflow by ID: its name, the user its runs run as,
	 * and the folders its steps name.
	 *
	 * A workflow with no run as user, or one whose user is gone, runs as
	 * whoever starts it, which the schedule names itself.
	 */
	function workflows() {
		$steps = $this->steps();
		$out = array();

		foreach ($this->store->named() as $row) {
			$id = \Task\Row::text($row, "id");

			$out[$id] = array(
				"name" => \Task\Row::text($row, "name"),
				"runAs" => \Task\Row::text($row, "run_as"),
				"folders" => self::folders($steps, $id),
			);
		}

		return $out;
	}

	/** steps is every workflow's steps, in position order, by workflow ID. */
	function steps() {
		return $this->store->allSteps();
	}

	/**
	 * folders is the folders the steps of a workflow name, the steps of a
	 * workflow a step runs among them, each workflow walked once.
	 *
	 * The walk is a queue and not a recursion because the workflows already
	 * seen have to be shared across every branch, and a parameter written by a
	 * call is not visible to its caller. The order differs from the Go side's
	 * depth first walk; what reads it is a set.
	 */
	static function folders($steps, $id) {
		$out = array();
		$seen = array($id => true);
		$queue = array($id);

		while (count($queue) > 0) {
			$next = array_shift($queue);

			if (!isset($steps[$next])) {
				continue;
			}

			foreach ($steps[$next] as $step) {
				$ref = $step->getWorkflowRef();

				if (self::kind($step) !== self::STEP_WORKFLOW) {
					$out[] = $step->getFolder();

					continue;
				}

				if (!isset($seen[$ref])) {
					$seen[$ref] = true;
					$queue[] = $ref;
				}
			}
		}

		return $out;
	}

	/** kind is what a step does, from its fields for a row written without one. */
	static function kind($step) {
		if ($step->getKind() !== "") {
			return $step->getKind();
		}

		if ($step->getPrompt() !== "") {
			return self::STEP_PROMPT;
		}

		if ($step->getWorkflowRef() !== "") {
			return self::STEP_WORKFLOW;
		}

		return self::STEP_JOB;
	}
}
