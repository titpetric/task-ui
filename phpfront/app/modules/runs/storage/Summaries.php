<?php

namespace Task\Runs;

// Summaries reads the run_summary table: what Claude wrote about a run, kept
// beside the run it is about.
//
// The table is this module's. What a page makes of a summary is Task\Report,
// which asks here for the row.
class Summaries {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** of is the summary row of a run, or null for a run with none. */
	function of($runID) {
		$row = $this->db->get("SELECT * FROM run_summary WHERE run_id = ?", $runID);

		return is_array($row) ? $row : null;
	}
}
