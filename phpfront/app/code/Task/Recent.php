<?php

namespace Task;

// Recent is the right rail every page draws: the latest runs of jobs and of
// workflows by anyone, merged newest first, under a heading per the hour or
// the day they started in.
//
// The rail is chrome and not a page, so it is built from the same two tables
// the runs module reads and handed to the recents component as one scope.
class Recent {
	/** COUNT is how many entries the rail lists; the rail scrolls. */
	const COUNT = 50;

	var $db;
	var $runs = null;
	var $flows = null;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
		$this->runs = new Run($this->db);
		$this->flows = new WorkflowRun($this->db);
	}

	/**
	 * data is the scope the recents component reads: where the label goes,
	 * and the groups of entries.
	 */
	function data($access, $roles = array(), $folders = array()) {
		return array("href" => "/runs", "groups" => $this->groups($access, $roles, $folders));
	}

	/**
	 * groups is the entries under the heading of when they started, newest
	 * first.
	 *
	 * The headings read against the wall clock, as "ago" does: frontend.Page
	 * calls view.GroupRecents(time.Now(), ...) while the page around it carries
	 * the instance's Now, so a pinned clock does not move these headings.
	 */
	function groups($access, $roles = array(), $folders = array()) {
		return self::group($this->entries($access, $roles, $folders), Clock::real()->unixnano());
	}

	/**
	 * group puts entries, newest first, under the heading of the hour or the
	 * day they started in. The moment is nanoseconds.
	 *
	 * The moment is a parameter, as view.GroupRecents takes one, so the
	 * headings of a list are what a reading of the clock decides and not what
	 * the clock said while the list was built.
	 */
	static function group($entries, $now) {
		$out = array();

		foreach ($entries as $entry) {
			$label = self::hoursAgo($now - Clock::nanos($entry["at"]));
			$last = count($out) - 1;

			if ($last >= 0 && $out[$last]["label"] === $label) {
				$out[$last]["entries"][] = $entry;

				continue;
			}

			$out[] = array("label" => $label, "entries" => array($entry));
		}

		return $out;
	}

	/**
	 * entries is the runs and the workflow runs the reader sees, merged
	 * newest first, at most COUNT of them.
	 *
	 * Both lists arrive newest first, so the merge walks them rather than
	 * sorting: a run and a workflow run that started in the same instant keep
	 * the order the Go page's stable sort keeps, the run first. The instant is
	 * read to the nanosecond, because the Go page compares two times and not
	 * two seconds.
	 */
	function entries($access, $roles = array(), $folders = array()) {
		$user = $access === null ? null : $access->user;
		$prompts = $access !== null && $access->allowed("prompt/view");
		$notes = $access !== null && $access->allowed("note/view");
		$runs = array();
		$flows = array();

		foreach ($this->runs->find(array("limit" => 2 * self::COUNT)) as $run) {
			if (Run::seen($user, $run, $prompts, $notes)) {
				$runs[] = $this->ofRun($run, $access, $roles, $folders);
			}
		}

		foreach ($this->flows->visible($user, self::COUNT) as $flow) {
			$flows[] = $this->ofFlow($flow, $roles);
		}

		return self::merge($runs, $flows, self::COUNT);
	}

	/** merge walks two lists that are each newest first and keeps the first n. */
	static function merge($runs, $flows, $limit) {
		$out = array();
		$i = 0;
		$j = 0;

		while (count($out) < $limit && ($i < count($runs) || $j < count($flows))) {
			if ($j >= count($flows)) {
				$out[] = $runs[$i];
				$i++;

				continue;
			}

			if ($i >= count($runs) || Clock::nanos($flows[$j]["at"]) > Clock::nanos($runs[$i]["at"])) {
				$out[] = $flows[$j];
				$j++;

				continue;
			}

			$out[] = $runs[$i];
			$i++;
		}

		return $out;
	}

	/** ofRun is one run of a job as the component reads it. */
	function ofRun($run, $access, $roles, $folders) {
		$state = $run->getState();
		$name = Run::name($run);
		$user = $run->getUserName();
		$ended = Run::ended($state);
		$act = "eye";

		if (!$ended && self::runnable($run, $access, $folders)) {
			$act = "play";
		}

		return array(
			"at" => $run->getStartedAt(),
			"href" => "/runs/" . $run->getId(),
			"name" => $run->getJob(),
			"place" => Run::place($run),
			"state" => $state,
			"ended" => $ended,
			"act" => $act,
			"flow" => false,
			"title" => "Open " . $name . ", " . $user . ", " . $state,
			"avatar" => Who::avatar($user, self::role($roles, $user)),
		);
	}

	/** ofFlow is one run of a workflow as the component reads it. */
	function ofFlow($flow, $roles) {
		$state = $flow->getState();
		$user = $flow->getUserName();

		return array(
			"at" => $flow->getStartedAt(),
			"href" => "/workflows/" . $flow->getWorkflowId(),
			"name" => $flow->getName(),
			"place" => "workflow",
			"state" => $state,
			"ended" => Run::ended($state),
			"act" => "eye",
			"flow" => true,
			"title" => "Workflow " . $flow->getName() . ", " . $user . ", " . $state,
			"avatar" => Who::avatar($user, self::role($roles, $user)),
		);
	}

	/** role is the role a user acts with, "" for a name no row holds. */
	static function role($roles, $name) {
		return isset($roles[$name]) ? $roles[$name] : "";
	}

	/**
	 * runnable reports whether the reader may start a run's job again, which
	 * is the icon the entry of a run in progress carries. Only such an entry
	 * draws one, so this is asked of a run in progress alone.
	 *
	 * The run names its folder, the directory it ran in and its job, and the
	 * host resolves the job out of those three the way the run did: the jobs
	 * of a directory inside a folder are in no catalogue, and neither are the
	 * four a folder has without declaring them. The rule over the job is the
	 * one the sidebar's own rows are drawn by.
	 */
	static function runnable($run, $access, $folders) {
		if ($access === null) {
			return false;
		}

		$id = $run->getFolder();

		foreach ($folders as $folder) {
			if (Row::text($folder, "id") !== $id) {
				continue;
			}

			$job = Host::job($id, $run->getDir(), $run->getJob());

			return Row::flag($job, "found") && Folders::allows($access, $folder, $job);
		}

		return false;
	}

	/**
	 * hoursAgo is the heading of an entry that started a while ago: whole
	 * hours, then whole days. The since is nanoseconds, because
	 * view.HoursAgo takes a time.Duration and divides it whole.
	 */
	static function hoursAgo($since) {
		$hours = intdiv($since, 3600 * Clock::NANO);

		if ($hours < 1) {
			return "In the last hour";
		}

		if ($hours < 24) {
			return self::plural($hours, "hour") . " ago";
		}

		return self::plural(intdiv($hours, 24), "day") . " ago";
	}

	/** plural words a count: one hour, two hours. */
	static function plural($count, $word) {
		return $count === 1 ? "1 " . $word : $count . " " . $word . "s";
	}
}
