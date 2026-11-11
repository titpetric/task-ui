<?php

namespace Task\Schedules;

// Schedules reads the schedule table: every schedule, the ones of a folder and
// the ones of a workflow.
//
// The table is this module's. What a row of it means to a page - whether it is
// on, what it starts, the cells a table draws - is Task\Schedule.
//
// Nothing here writes. The schedules page is read only and its forms call
// /api/schedules from the browser, so Create, Update, Delete and Due stay in
// storage/schedule_storage.go.
class Schedules {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** all is every schedule, by folder and job. */
	function all() {
		return \Task\Model\Schedule::rows($this->db->get_all("SELECT * FROM schedule ORDER BY folder, job, id"));
	}

	/**
	 * byFolder is the schedules of one folder, by job: the page of a folder
	 * counts them and the page of a job lists its own.
	 *
	 * A schedule of a workflow names no folder, so this is the schedules of
	 * jobs and of hosts and nothing else.
	 */
	function byFolder($folder) {
		return \Task\Model\Schedule::rows($this->db->get_all("SELECT * FROM schedule WHERE folder = ? ORDER BY job, id", $folder));
	}

	/** byWorkflow is the schedules of one workflow, by ID. */
	function byWorkflow($id) {
		return \Task\Model\Schedule::rows($this->db->get_all("SELECT * FROM schedule WHERE workflow_id = ? ORDER BY id", $id));
	}

	/** get is one schedule by ID, or null. */
	function get($id) {
		$row = $this->db->get("SELECT * FROM schedule WHERE id = ?", $id);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Schedule($row);
	}
}
