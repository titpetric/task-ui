<?php

namespace Task\Schedules;

// Lookup is what the schedules page reads of the one table it does not own:
// the workflows a schedule may start, with the user they run as and the
// folders their steps name.
//
// It belongs on Task\Workflow and moves there with the workflows module, whose
// page owns it. It sits here because this is the first page to read it, and
// every statement has to live in code/Task or in a module's storage.
//
// Two readers have already left: the chrome reads the named workspaces and the
// keys a user's runs get on every page, so they are Task\Workspace's and
// Task\Key's. That is where the pattern ends up.
class Lookup {
	/** STEPS are the kinds of a workflow step, as model/workflow.go names them. */
	const STEP_JOB = "job";

	const STEP_WORKFLOW = "workflow";

	const STEP_PROMPT = "prompt";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
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

		foreach ($this->db->get_all("SELECT w.id AS id, w.name AS name, u.name AS run_as FROM workflow w LEFT JOIN user u ON u.id = w.run_as_user_id ORDER BY w.name, w.id") as $row) {
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
		$out = array();

		foreach (\Task\Model\WorkflowStep::rows($this->db->get_all("SELECT * FROM workflow_step ORDER BY workflow_id, position")) as $step) {
			$id = $step->getWorkflowId();

			if (!isset($out[$id])) {
				$out[$id] = array();
			}

			$out[$id][] = $step;
		}

		return $out;
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
