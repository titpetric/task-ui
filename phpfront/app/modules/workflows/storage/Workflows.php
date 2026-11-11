<?php

namespace Task\Workflows;

// Workflows reads the workflow table and its steps: every workflow by name,
// one by ID, the steps of one in order, and the workflows a step of one runs.
//
// The table is this module's. What a page asks of a workflow - who its runs
// run as, whether the reader may start it, which step stops them - is
// Task\Workflow, which asks here for the rows.
//
// Nothing here writes. Every form of the workflow pages calls /api/workflows
// from the browser, so Create, Update and Delete stay in
// storage/workflow_storage.go.
class Workflows {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** all is every workflow, by name. */
	function all() {
		return \Task\Model\Workflow::rows($this->db->get_all("SELECT * FROM workflow ORDER BY name COLLATE NOCASE, id"));
	}

	/** get is one workflow by ID, or null. */
	function get($id) {
		if ($id === "") {
			return null;
		}

		$row = $this->db->get("SELECT * FROM workflow WHERE id = ?", $id);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Workflow($row);
	}

	/** steps is a workflow's step rows, in order. */
	function steps($id) {
		if ($id === "") {
			return array();
		}

		return \Task\Model\WorkflowStep::rows($this->db->get_all("SELECT * FROM workflow_step WHERE workflow_id = ? ORDER BY position", $id));
	}

	/** refs is the workflows the steps of one run, in order. */
	function refs($id) {
		$rows = $this->db->get_all("SELECT workflow_ref FROM workflow_step WHERE workflow_id = ? AND workflow_ref != '' ORDER BY position", $id);
		$out = array();

		if (!is_array($rows)) {
			return $out;
		}

		foreach ($rows as $row) {
			$out[] = \Task\Row::text($row, "workflow_ref");
		}

		return $out;
	}

	/** allSteps is every workflow's steps, in position order, by workflow ID. */
	function allSteps() {
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
	 * named is every workflow by ID: its name, the user its runs run as, and
	 * the user row's name where the workflow names one.
	 *
	 * A workflow with no run as user, or one whose user is gone, runs as
	 * whoever starts it, which a schedule names itself.
	 */
	function named() {
		return $this->db->get_all("SELECT w.id AS id, w.name AS name, u.name AS run_as FROM workflow w LEFT JOIN user u ON u.id = w.run_as_user_id ORDER BY w.name, w.id");
	}
}
