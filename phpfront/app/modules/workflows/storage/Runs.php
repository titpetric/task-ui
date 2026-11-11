<?php

namespace Task\Workflows;

// Runs reads the workflow_run table and its steps: the runs of every workflow,
// newest first, one workflow's own a page at a time, and which folders a run
// touched.
//
// The table is this module's, as are the workflow and its steps, which
// Task\Workflows\Workflows reads.
class Runs {
	/**
	 * COLUMNS is every column, with the timestamps as the text the column
	 * holds. Task\Runs\Runs::COLUMNS says why.
	 */
	const COLUMNS = "*, started_at || '' AS started_text, finished_at || '' AS finished_text";

	/** DEPTH is how far a workflow run that ran another is followed. */
	const DEPTH = 8;

	/**
	 * SCAN is how many times the window a profile reads is the list it draws:
	 * the runs of one user are found among the latest of everyone's.
	 */
	const SCAN = 50;

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** get is one workflow run by ID, or null. */
	function get($id) {
		if ($id === "") {
			return null;
		}

		$row = $this->db->get("SELECT " . self::COLUMNS . " FROM workflow_run WHERE id = ?", $id);

		return is_array($row) ? self::hydrate($row) : null;
	}

	/** latest is the runs of every workflow, newest first. */
	function latest($limit = 20) {
		$limit = (int)$limit;

		if ($limit <= 0) {
			$limit = 20;
		}

		return self::rows($this->db->get_all("SELECT " . self::COLUMNS . " FROM workflow_run ORDER BY started_at DESC, id DESC LIMIT ?", $limit));
	}

	/**
	 * listByWorkflow is one workflow's runs, newest first, a page at a time:
	 * limit of them, older than the run `before` names or newer than the one
	 * `after` names.
	 *
	 * A page read after a cursor reads up from it, oldest first, and is turned
	 * back over here: the rows either side of a cursor are the newer ones, and
	 * LIMIT takes from the top. A cursor on a row that is gone keeps none.
	 */
	function listByWorkflow($id, $filter = array()) {
		if ($id === "") {
			return array();
		}

		$limit = isset($filter["limit"]) ? (int)$filter["limit"] : 20;

		if ($limit <= 0) {
			$limit = 20;
		}

		$before = \Task\Run::pick($filter, "before");
		$after = \Task\Run::pick($filter, "after");
		$sql = "SELECT " . self::COLUMNS . " FROM workflow_run WHERE workflow_id = ?";
		$args = array($id);
		$order = "started_at DESC, id DESC";

		if ($before !== "") {
			$sql .= " AND (started_at, id) < (SELECT started_at, id FROM workflow_run WHERE id = ?)";
			$args[] = $before;
		} elseif ($after !== "") {
			$sql .= " AND (started_at, id) > (SELECT started_at, id FROM workflow_run WHERE id = ?)";
			$args[] = $after;
			$order = "started_at ASC, id ASC";
		}

		$args[] = $limit;
		$out = self::rows(\Task\DB::all($this->db, $sql . " ORDER BY " . $order . " LIMIT ?", $args));

		return $before === "" && $after !== "" ? array_reverse($out) : $out;
	}

	/** steps is a workflow run's steps, in order. */
	function steps($id) {
		if ($id === "") {
			return array();
		}

		return \Task\Model\WorkflowRunStep::rows($this->db->get_all("SELECT * FROM workflow_run_step WHERE workflow_run_id = ? ORDER BY position", $id));
	}

	/**
	 * sees reports whether a user sees every folder a workflow run ran in,
	 * through the steps of the workflow runs it started.
	 *
	 * A user who sees every folder is answered without a query: the rail
	 * reads a few hundred runs and the recursion is one statement per run.
	 */
	function sees($user, $id, $depth = 0) {
		if (count(\Task\User::folders($user)) === 0) {
			return \Task\User::sees($user, "");
		}

		if ($depth > self::DEPTH) {
			return false;
		}

		foreach ($this->steps($id) as $step) {
			if ($step->getFolder() !== "" && !\Task\User::sees($user, $step->getFolder())) {
				return false;
			}

			if ($step->getWorkflowRunRef() !== "" && !$this->sees($user, $step->getWorkflowRunRef(), $depth + 1)) {
				return false;
			}
		}

		return true;
	}

	/** visible is the latest runs a user sees, newest first, at most limit of them. */
	function visible($user, $limit) {
		$out = array();

		foreach ($this->latest(5 * $limit) as $run) {
			if (count($out) < $limit && $this->sees($user, $run->getId())) {
				$out[] = $run;
			}
		}

		return $out;
	}

	/**
	 * byUser is the latest workflow runs one user started that another sees,
	 * newest first, at most limit of them, which a profile lists.
	 *
	 * The user is not a column of the query but a test over the window, as the
	 * Go page reads it: the window is the latest SCAN runs by anyone, and a
	 * user who started none of them has an empty list rather than an older run
	 * of theirs.
	 */
	function byUser($viewer, $name, $limit) {
		$out = array();

		foreach ($this->latest(self::SCAN * $limit) as $run) {
			if (count($out) === $limit) {
				return $out;
			}

			if ($run->getUserName() === $name && $this->sees($viewer, $run->getId())) {
				$out[] = $run;
			}
		}

		return $out;
	}

	/** rows turns what the database answered into workflow runs. */
	static function rows($found) {
		$out = array();

		if (!is_array($found)) {
			return $out;
		}

		foreach ($found as $row) {
			$out[] = self::hydrate($row);
		}

		return $out;
	}

	/** hydrate is one row as a workflow run, its timestamps the text the column holds. */
	static function hydrate($row) {
		$run = new \Task\Model\WorkflowRun($row);

		$run->setStartedAt(\Task\Row::stamp($row, "started_text"));
		$run->setFinishedAt(\Task\Row::stamp($row, "finished_text"));

		return $run;
	}
}
