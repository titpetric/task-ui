<?php

namespace Task\Runs;

// Runs reads the run table: every run of a job, the runs one run dispatched,
// and the window an activity chart counts.
//
// The table is this module's. What a row of it means to a page - its name, the
// place it ran, the label that follows it over the event stream - is
// Task\Run, which asks here for the rows.
class Runs {
	/** BATCH_LIMIT is how many runs of a batch a page looks over. */
	const BATCH_LIMIT = 200;

	/**
	 * COLUMNS is every column, with the timestamps as the text the column
	 * holds.
	 *
	 * A timestamp column is declared DATETIME, so the driver parses it and the
	 * bridge renders the time it got: ISO seconds, with the milliseconds the
	 * Go side wrote dropped. An expression carries no declared type, so the
	 * text arrives whole and a duration reads to a tenth of a second the way
	 * the Go page reads it.
	 */
	const COLUMNS = "*, started_at || '' AS started_text, finished_at || '' AS finished_text";

	/**
	 * CHART_COLUMNS is what a chart of a window reads and nothing else: when a
	 * run started and ended, how it ended, and the three columns the
	 * permission filter asks about.
	 *
	 * A chart counts the window and not the page, which is five thousand rows
	 * of it: hydrating each one into a model for five numbers was 274ms of a
	 * 2s /runs, and the columns it did not read crossed into the runtime
	 * anyway. See charted.
	 */
	const CHART_COLUMNS = "id, folder, job, user_id, state, started_at || '' AS started_text, finished_at || '' AS finished_text";

	/** CHILD_COLUMNS is what a count of a run's children reads: the parent, and the columns the rule that hides a run asks about. */
	const CHILD_COLUMNS = "id, parent_id, folder, job, user_id";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** get is one run by ID, or null. */
	function get($id) {
		if ($id === "") {
			return null;
		}

		$row = $this->db->get("SELECT " . self::COLUMNS . " FROM run WHERE id = ?", $id);

		return is_array($row) ? self::hydrate($row) : null;
	}

	/**
	 * root is the run a batch is named by: the oldest of it, and the run
	 * itself where it is in no batch.
	 *
	 * A batch is one job run over several places at once, and the thing a
	 * reader asks about or a summary is written about is the batch and not the
	 * page they happen to be on, so every page of a batch answers the same
	 * run.
	 */
	function root($run) {
		$batch = $run->getBatchId();

		if ($batch === "") {
			return $run->getId();
		}

		$rows = $this->find(array("batch" => $batch, "limit" => self::BATCH_LIMIT));

		return count($rows) === 0 ? $run->getId() : $rows[count($rows) - 1]->getId();
	}

	/**
	 * find is the runs a filter lists, newest first, or oldest first for a
	 * page read after a cursor.
	 *
	 * The filter is an array and a field it leaves out narrows nothing. Two
	 * fields are not that: `dir` is read only alongside `job`, where it tells
	 * the folder's own job from the job of a directory inside it with the same
	 * name, and `folders` is the list a reader may see, which keeps the page
	 * from being filled with rows the reader is about to be denied.
	 */
	function find($filter = array()) {
		$before = \Task\Run::pick($filter, "before");
		$after = \Task\Run::pick($filter, "after");
		$limit = isset($filter["limit"]) ? (int)$filter["limit"] : 15;
		$where = array();
		$args = array();

		self::narrowed($filter, $where, $args);

		// A page read after a cursor reads up from it, oldest first, for the
		// caller to reverse: the rows either side of a cursor are the newer
		// ones, and LIMIT takes from the top. A cursor either side of a row
		// that is gone keeps none.
		$order = "started_at DESC, id DESC";

		if ($before !== "") {
			$where[] = "(started_at, id) < (SELECT started_at, id FROM run WHERE id = ?)";
			$args[] = $before;
		} elseif ($after !== "") {
			$where[] = "(started_at, id) > (SELECT started_at, id FROM run WHERE id = ?)";
			$args[] = $after;
			$order = "started_at ASC, id ASC";
		}

		$sql = "SELECT " . self::COLUMNS . " FROM run";

		if (count($where) > 0) {
			$sql .= " WHERE " . implode(" AND ", $where);
		}

		$sql .= " ORDER BY " . $order;

		if ($limit > 0) {
			$sql .= " LIMIT ?";
			$args[] = $limit;
		}

		$out = self::rows(\Task\DB::all($this->db, $sql, $args));

		if ($before === "" && $after !== "") {
			$out = array_reverse($out);
		}

		return $out;
	}

	/** children is the runs one run dispatched, oldest first. */
	function children($id) {
		if ($id === "") {
			return array();
		}

		$sql = "SELECT " . self::COLUMNS . " FROM run WHERE parent_id = ? ORDER BY started_at ASC, id ASC";

		return self::rows(\Task\DB::all($this->db, $sql, array($id)));
	}

	/**
	 * childrenOf is the children of many runs at once, as rows and not as
	 * runs: the rail counts the children of every row it lists.
	 *
	 * The columns are the ones Task\Seen::rows filters on, so a count costs no
	 * hydration: the rail is chrome and is drawn on every page.
	 */
	function childrenOf($ids) {
		if (count($ids) === 0) {
			return array();
		}

		$sql = "SELECT " . self::CHILD_COLUMNS . " FROM run WHERE parent_id IN (" . \Task\Row::placeholders($ids) . ")";
		$found = \Task\DB::all($this->db, $sql, $ids);

		return is_array($found) ? $found : array();
	}

	/**
	 * latest is the last run of each job of the folders, by the key the job is
	 * listed under; every folder's for an empty list.
	 *
	 * A run still queued or going is not the last one: the Jobs table draws the
	 * colour of what a job did last beside the run it is doing now. Which runs
	 * the reader may see is not asked here, as the Go page does not ask it
	 * either: the colour of a dot is not a permission.
	 */
	function latest($folders = array()) {
		$where = "state != ?";
		$args = array(\Task\Run::RUNNING);

		if (count($folders) > 0) {
			$where .= " AND folder IN (" . \Task\Row::placeholders($folders) . ")";

			foreach ($folders as $folder) {
				$args[] = $folder;
			}
		}

		$ranked = "SELECT id, ROW_NUMBER() OVER (PARTITION BY folder, dir, job ORDER BY started_at DESC, id DESC) AS latest FROM run WHERE " . $where;
		$sql = "SELECT " . self::COLUMNS . " FROM run WHERE id IN (SELECT id FROM (" . $ranked . ") WHERE latest = 1) ORDER BY folder, dir, job";
		$out = array();

		foreach (self::rows(\Task\DB::all($this->db, $sql, $args)) as $run) {
			$out[\Task\Folders::key($run->getFolder(), $run->getDir(), $run->getJob())] = $run;
		}

		return $out;
	}

	/**
	 * charted is the rows the activity charts count, newest first, as rows and
	 * not as runs: the same filter find reads, with the cursors off, over the
	 * columns of CHART_COLUMNS.
	 *
	 * A chart is the window and not the page, so the list is Activity::LIMIT
	 * long where the filter names no limit of its own. Task\Seen::rows filters
	 * it and Task\Activity::rows reads it; nothing else may, because the rows
	 * carry seven columns and a page reads a run.
	 */
	function charted($filter) {
		$where = array();
		$args = array();

		self::narrowed($filter, $where, $args);

		$limit = isset($filter["limit"]) ? (int)$filter["limit"] : \Task\Activity::LIMIT;
		$sql = "SELECT " . self::CHART_COLUMNS . " FROM run";

		if (count($where) > 0) {
			$sql .= " WHERE " . implode(" AND ", $where);
		}

		$sql .= " ORDER BY started_at DESC, id DESC";

		if ($limit > 0) {
			$sql .= " LIMIT ?";
			$args[] = $limit;
		}

		$found = \Task\DB::all($this->db, $sql, $args);

		return is_array($found) ? $found : array();
	}

	/**
	 * narrowed is what a filter says about the table, as conditions and the
	 * arguments under them, with the cursors left out: find reads a page of
	 * the rows and charted reads the window, and the rows either one answers
	 * are the same rows.
	 */
	static function narrowed($filter, &$where, &$args) {
		$job = \Task\Run::pick($filter, "job");

		self::narrow($where, $args, "folder", \Task\Run::pick($filter, "folder"));

		// @host:* is every host job and not a job by that name.
		if ($job === \Task\Run::HOST_ALL) {
			$where[] = "job LIKE ?";
			$args[] = \Task\Run::HOST_PREFIX . "%";
		} else {
			self::narrow($where, $args, "job", $job);
		}

		if ($job !== "") {
			$where[] = "dir = ?";
			$args[] = \Task\Run::pick($filter, "dir");
		}

		self::narrow($where, $args, "state", \Task\Run::pick($filter, "state"));
		self::narrow($where, $args, "user_id", \Task\Run::pick($filter, "user_id"));
		self::narrow($where, $args, "user_name", \Task\Run::pick($filter, "user"));
		self::narrow($where, $args, "batch_id", \Task\Run::pick($filter, "batch"));
		self::narrow($where, $args, "parent_id", \Task\Run::pick($filter, "parent"));

		// `roots` keeps the runs nothing dispatched, which is the rail.
		if (isset($filter["roots"]) && $filter["roots"]) {
			$where[] = "parent_id = ''";
		}

		$folders = isset($filter["folders"]) && is_array($filter["folders"]) ? $filter["folders"] : array();

		if (count($folders) > 0) {
			$where[] = "folder IN (" . \Task\Row::placeholders($folders) . ")";

			foreach ($folders as $folder) {
				$args[] = $folder;
			}
		}
	}

	/** narrow adds a condition on a column for a value the filter named. */
	static function narrow(&$where, &$args, $column, $value) {
		if ($value === "") {
			return;
		}

		$where[] = $column . " = ?";
		$args[] = $value;
	}

	/** rows turns what the database answered into runs. */
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

	/** hydrate is one row as a run, its timestamps the text the column holds. */
	static function hydrate($row) {
		$run = new \Task\Model\Run($row);

		$run->setStartedAt(\Task\Row::stamp($row, "started_text"));
		$run->setFinishedAt(\Task\Row::stamp($row, "finished_text"));

		return $run;
	}
}
