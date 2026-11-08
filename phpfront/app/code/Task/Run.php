<?php

namespace Task;

// Run reads the run table and answers what a row of it means: the name a
// page calls it by, the place it ran, how long it took and the label that
// follows it over the event stream.
//
// The row behaviour sits here rather than on Task\Model\Run because the
// runtime merges no methods into a generated class: one class owns the table
// and its rows.
class Run {
	/** The states a run passes through and ends in, queued and running first. */
	const QUEUED = "queued";

	const RUNNING = "running";

	const OK = "ok";

	const FAILED = "failed";

	const STOPPED = "stopped";

	const KILLED = "killed";

	/** The job IDs of Claude's own runs, which no prompt asks about. */
	const PROMPT_JOB = "@claude-prompt";

	const NOTE_JOB = "@claude-note";

	const CLAUDE_JOB = "@claude";

	/** BATCH_LIMIT is how many runs of a batch a page looks over. */
	const BATCH_LIMIT = 200;

	/** HOST_PREFIX starts the ID of a command on a host; HOST_ALL is every host. */
	const HOST_PREFIX = "@host:";

	const HOST_ALL = "@host:*";

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
		$this->db = $db === null ? DB::reader() : $db;
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
		$before = self::pick($filter, "before");
		$after = self::pick($filter, "after");
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

		$out = self::rows(DB::all($this->db, $sql, $args));

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

		return self::rows(DB::all($this->db, $sql, array($id)));
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

		$sql = "SELECT " . self::CHILD_COLUMNS . " FROM run WHERE parent_id IN (" . Row::placeholders($ids) . ")";
		$found = DB::all($this->db, $sql, $ids);

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
		$args = array(self::RUNNING);

		if (count($folders) > 0) {
			$where .= " AND folder IN (" . Row::placeholders($folders) . ")";

			foreach ($folders as $folder) {
				$args[] = $folder;
			}
		}

		$ranked = "SELECT id, ROW_NUMBER() OVER (PARTITION BY folder, dir, job ORDER BY started_at DESC, id DESC) AS latest FROM run WHERE " . $where;
		$sql = "SELECT " . self::COLUMNS . " FROM run WHERE id IN (SELECT id FROM (" . $ranked . ") WHERE latest = 1) ORDER BY folder, dir, job";
		$out = array();

		foreach (self::rows(DB::all($this->db, $sql, $args)) as $run) {
			$out[Folders::key($run->getFolder(), $run->getDir(), $run->getJob())] = $run;
		}

		return $out;
	}

	/**
	 * narrowed is what a filter says about the table, as conditions and the
	 * arguments under them, with the cursors left out: find reads a page of
	 * the rows and charted reads the window, and the rows either one answers
	 * are the same rows.
	 */
	static function narrowed($filter, &$where, &$args) {
		$job = self::pick($filter, "job");

		self::narrow($where, $args, "folder", self::pick($filter, "folder"));

		// @host:* is every host job and not a job by that name.
		if ($job === self::HOST_ALL) {
			$where[] = "job LIKE ?";
			$args[] = self::HOST_PREFIX . "%";
		} else {
			self::narrow($where, $args, "job", $job);
		}

		if ($job !== "") {
			$where[] = "dir = ?";
			$args[] = self::pick($filter, "dir");
		}

		self::narrow($where, $args, "state", self::pick($filter, "state"));
		self::narrow($where, $args, "user_id", self::pick($filter, "user_id"));
		self::narrow($where, $args, "user_name", self::pick($filter, "user"));
		self::narrow($where, $args, "batch_id", self::pick($filter, "batch"));
		self::narrow($where, $args, "parent_id", self::pick($filter, "parent"));

		// `roots` keeps the runs nothing dispatched, which is the rail.
		if (isset($filter["roots"]) && $filter["roots"]) {
			$where[] = "parent_id = ''";
		}

		$folders = isset($filter["folders"]) && is_array($filter["folders"]) ? $filter["folders"] : array();

		if (count($folders) > 0) {
			$where[] = "folder IN (" . Row::placeholders($folders) . ")";

			foreach ($folders as $folder) {
				$args[] = $folder;
			}
		}
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

		$limit = isset($filter["limit"]) ? (int)$filter["limit"] : Activity::LIMIT;
		$sql = "SELECT " . self::CHART_COLUMNS . " FROM run";

		if (count($where) > 0) {
			$sql .= " WHERE " . implode(" AND ", $where);
		}

		$sql .= " ORDER BY started_at DESC, id DESC";

		if ($limit > 0) {
			$sql .= " LIMIT ?";
			$args[] = $limit;
		}

		$found = DB::all($this->db, $sql, $args);

		return is_array($found) ? $found : array();
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
		$run = new Model\Run($row);

		$run->setStartedAt(Row::stamp($row, "started_text"));
		$run->setFinishedAt(Row::stamp($row, "finished_text"));

		return $run;
	}

	/**
	 * overlay lays the runner's session over the row and answers what the run
	 * reports.
	 *
	 * While the runner holds the run it is ahead of the table: the session
	 * decides the state, the exit code, the title and when the run finished,
	 * and says why it waits. Where it holds none, the row is the whole answer,
	 * which is most of them.
	 *
	 * The run is written through, so a page lays the session over a row once
	 * and reads the run after it.
	 */
	static function overlay($run, $session) {
		$message = $run->getMessage();

		if (!Row::flag($session, "live")) {
			return $message;
		}

		$said = Row::text($session, "message");

		$run->setState(Row::text($session, "state"));
		$run->setExitCode(Row::number($session, "exit_code"));
		$run->setTitle(Row::text($session, "title"));
		$run->setFinishedAt(Row::text($session, "finished_at"));

		return $said === "" ? $message : $said;
	}

	/** states is every state, queued and running first. */
	static function states() {
		return array(self::QUEUED, self::RUNNING, self::OK, self::FAILED, self::STOPPED, self::KILLED);
	}

	/** ended reports whether a state is one a run ended in. */
	static function ended($state) {
		return $state !== self::QUEUED && $state !== self::RUNNING && $state !== "";
	}

	/**
	 * label is the state a page draws beside a run, as the state component
	 * reads it: the exit code follows a run that ended as anything but ok.
	 */
	static function label($run) {
		if ($run === null) {
			return array(
				"id" => "",
				"state" => "",
				"ended" => false,
				"code" => "",
			);
		}

		$state = $run->getState();
		$ended = self::ended($state);

		return array(
			"id" => $run->getId(),
			"state" => $state,
			"ended" => $ended,
			"code" => $ended && $state !== self::OK ? (string)$run->getExitCode() : "",
		);
	}

	/**
	 * seen reports whether a user may see a run: one of a folder they see, a
	 * prompt's with prompt/view, the Notes agent's with note/view, and such a
	 * run they started themselves.
	 */
	static function seen($user, $run, $prompts, $notes) {
		if ($user === null || $run === null) {
			return false;
		}

		return self::allowed($user, $run->getFolder(), $run->getJob(), $run->getUserId(), $prompts, $notes);
	}

	/**
	 * seenRow is seen of a row of Run::charted, for the five thousand rows of
	 * a chart's window, which are read as rows and never hydrated.
	 */
	static function seenRow($user, $row, $prompts, $notes) {
		if ($user === null || !is_array($row)) {
			return false;
		}

		return self::allowed($user, Row::text($row, "folder"), Row::text($row, "job"), Row::text($row, "user_id"), $prompts, $notes);
	}

	/**
	 * allowed is the rule itself, over the three columns it reads, so a run
	 * and a row of one are answered by the same sentence.
	 */
	static function allowed($user, $folder, $job, $userID, $prompts, $notes) {
		$prompt = self::isPromptJob($folder, $job);
		$note = self::isNoteJob($folder, $job);

		if (($prompt || $note) && $userID === $user->getId()) {
			return true;
		}

		return User::sees($user, $folder) && ($prompts || !$prompt) && ($notes || !$note);
	}

	/** tone is the badge class of a state: ok, signal while running, bad for a run that broke. */
	static function tone($state) {
		if ($state === self::OK) {
			return "ok";
		}

		if ($state === self::RUNNING) {
			return "signal";
		}

		if ($state === self::FAILED || $state === self::KILLED) {
			return "bad";
		}

		return "";
	}

	/** where is the folder a title names a run by: folder, or folder/dir. */
	static function where($run) {
		$dir = $run->getDir();

		return $dir === "" ? $run->getFolder() : $run->getFolder() . "/" . $dir;
	}

	/** place is where, with the default workspace dropped from a directory of it. */
	static function place($run) {
		if ($run->getFolder() === Route::INTERNAL && $run->getDir() !== "") {
			return $run->getDir();
		}

		return self::where($run);
	}

	/** name is the run's title on a page: place / job, or the job alone for the root. */
	static function name($run) {
		if ($run->getFolder() === Route::INTERNAL && $run->getDir() === "") {
			return $run->getJob();
		}

		return self::place($run) . " / " . $run->getJob();
	}

	/**
	 * lead is what a list of runs calls one: the user a prompt or a note runs
	 * as, @miguel, which is the thread it answers; else the job.
	 */
	static function lead($run) {
		if ((self::isPrompt($run) || self::isNote($run)) && $run->getUserName() !== "") {
			return "@" . $run->getUserName();
		}

		return $run->getJob();
	}

	/** trail is what a list says under lead: the job of a thread's run without the @, else the place. */
	static function trail($run) {
		if (self::isPrompt($run) || self::isNote($run)) {
			return ltrim($run->getJob(), "@");
		}

		return self::place($run);
	}

	/** icon names the sprite symbol a list draws before trail, "" for a run that is not a thread's. */
	static function icon($run) {
		if (self::isPrompt($run)) {
			return "prompt";
		}

		if (self::isNote($run)) {
			return "note";
		}

		return "";
	}

	/** isPrompt reports whether the run is a prompt's. */
	static function isPrompt($run) {
		return self::isPromptJob($run->getFolder(), $run->getJob());
	}

	/** isNote reports whether the run is the Notes agent's. */
	static function isNote($run) {
		return self::isNoteJob($run->getFolder(), $run->getJob());
	}

	/** isPromptJob is isPrompt of the two columns it reads, which a row holds too. */
	static function isPromptJob($folder, $job) {
		return $job === self::PROMPT_JOB || ($folder === Route::INTERNAL && $job === "claude-prompt");
	}

	/** isNoteJob is isNote of the two columns it reads. */
	static function isNoteJob($folder, $job) {
		return $job === self::NOTE_JOB || ($folder === Route::INTERNAL && $job === "claude-note");
	}

	/** isClaude reports whether the run is Claude's own, which no prompt asks about. */
	static function isClaude($run) {
		return self::isPrompt($run) || self::isNote($run) || $run->getJob() === self::CLAUDE_JOB || ($run->getFolder() === Route::INTERNAL && $run->getJob() === "claude");
	}

	/** isHost reports whether the run is a command on a host of the inventory. */
	static function isHost($run) {
		return strpos($run->getJob(), self::HOST_PREFIX) === 0;
	}

	/** isScheduled reports whether the scheduler started the run. */
	static function isScheduled($run) {
		return $run->getScheduled() !== 0;
	}

	/** isForced reports whether the run killed the one before it. */
	static function isForced($run) {
		return $run->getForced() !== 0;
	}

	/** argList is the arguments the run passed the job, after --. */
	static function argList($run) {
		return Text::items($run->getArgs());
	}

	/**
	 * elapsed is how long a run took, in milliseconds, or has taken so far.
	 *
	 * A run still going is measured against the process clock, because
	 * model.Run.Elapsed is time.Since and takes no clock. See Clock::real.
	 */
	static function elapsed($run) {
		return self::took(Clock::parse($run->getStartedAt()), Clock::parse($run->getFinishedAt()));
	}

	/**
	 * took is elapsed of two stamps already read, in milliseconds: start is
	 * null for a run that never started and end for one still going.
	 *
	 * Task\Activity::rows reads it, because a chart of the window has both
	 * stamps in hand and parsing either a second time is a crossing into the
	 * host for a number it holds.
	 */
	static function took($start, $end) {
		if ($start === null) {
			return 0;
		}

		$until = $end === null ? Clock::real()->unixMilli() : $end->unixMilli();
		$took = $until - $start->unixMilli();

		return $took < 0 ? 0 : $took;
	}

	/** duration words a length in milliseconds the way a stopwatch shows it. */
	static function duration($ms) {
		if ($ms < 0) {
			$ms = 0;
		}

		$ms = (int)(round($ms / 100) * 100);

		if ($ms < 60000) {
			return sprintf("%.1fs", $ms / 1000);
		}

		if ($ms < 3600000) {
			return sprintf("%dm%02ds", (int)($ms / 60000), (int)($ms / 1000) % 60);
		}

		return sprintf("%dh%02dm", (int)($ms / 3600000), (int)($ms / 60000) % 60);
	}

	/** started is a timestamp as the script counting a running row reads it: milliseconds. */
	static function started($value) {
		$at = Clock::parse($value);

		return $at === null ? "" : (string)$at->unixMilli();
	}

	/** runFolderHref is the page of the place a run ran in. */
	static function runFolderHref($run, $workspaces) {
		return Route::placeHref($run->getFolder(), $run->getDir(), $workspaces);
	}

	/** runJobHref is the page of a run's job, or "" for a run of Claude's that has none. */
	static function runJobHref($run, $workspaces) {
		if (self::isPrompt($run) || self::isNote($run)) {
			return "";
		}

		return self::runFolderHref($run, $workspaces) . "/" . $run->getJob();
	}

	/**
	 * pick is a filter field as a string, "" for one the caller left out.
	 * It does not trim: storage.RunFilter narrows on the query value as it
	 * arrived, so a padded filter matches nothing on both sides.
	 */
	static function pick($filter, $name) {
		return isset($filter[$name]) ? (string)$filter[$name] : "";
	}
}
