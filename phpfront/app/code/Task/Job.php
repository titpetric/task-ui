<?php

namespace Task;

// Job is the page of one job: what it is, what it runs, every run of it with
// the charts of when they ran, and the schedules that start it.
//
// It is a class here and not a module of its own because only one module can
// own /f/*, and the folders module owns it: the path alone does not say
// whether it names a folder, a directory or a job, so one route resolves all
// three and this draws the third answer.
//
// The place the page is about is the folder for a job of the folder and the
// directory inside it for a job of one, which is the directory the route
// resolved the job under. Every row of the panel is read off that one place
// and the one job, and the job itself is the row the route already read.
class Job {
	/** HISTORY is how many runs the table lists, newest first. */
	const HISTORY = 100;

	/** NOTE is the note beside the Schedules heading. */
	const NOTE = "cron expressions, in the server's time zone";

	/** NONE is what the Schedules block says with none, and MORE what it says to a reader who may add one. */
	const NONE = "Not scheduled.";

	const MORE = "Not scheduled. Add a cron expression below to run it on a schedule.";

	/** SUMMARIZE is the hover of Summarize, in a run's [...] and in the head of its page. */
	const SUMMARIZE = "Claude summarizes the run's output, in this window";

	var $chrome;
	var $view;
	var $workspaces;

	function __construct($chrome, $view) {
		$this->chrome = $chrome;
		$this->view = $view;
		$this->workspaces = $chrome->workspaces();
	}

	/**
	 * serve draws the page of the job name names, in the folder id and under
	 * the directory dir where it names one.
	 *
	 * job is the row the route resolved the path into, so the place is the one
	 * reading this page costs: a job of a directory inside a folder has that
	 * directory detected once and not twice.
	 *
	 * The table reads a hundred runs and the charts read every run of the
	 * window, so they are two queries over the one filter and not one: the
	 * charts count a month the table does not list.
	 */
	function serve($id, $dir, $name, $job) {
		if (!Row::flag($job, "found")) {
			$this->view->missing("No job named " . $name . " in " . $id . ".");

			return;
		}

		$place = Host::place($id, $dir);
		$jobID = Row::text($job, "id");
		$jobDir = Row::text($job, "dir");
		$jobName = Row::text($job, "name");
		$where = Route::place($id, $jobDir);

		$this->chrome->heading($where . " / " . $jobName, Route::crumbs($where, $jobName, $this->workspaces));
		$this->chrome->at($id, Row::text($job, "key"));
		$this->chrome->asks($id, $jobDir);

		$control = new Control($this->chrome);
		$can = Folders::allows($this->chrome->access, $place, $job);
		$live = $this->chrome->liveRuns();
		$key = Row::text($job, "key");
		$running = isset($live[$key]) ? $live[$key] : null;

		$filter = array(
			"folder" => $id,
			"dir" => $jobDir,
			"job" => $jobID,
		);
		$runs = new Runs\Runs($this->chrome->db);
		$seen = new Seen($this->chrome);
		$filter["limit"] = self::HISTORY;
		$found = $seen->of($runs->find($filter));
		$reports = $this->reports($found);
		$users = $jobDir === "" ? $this->chrome->runAsUsers() : array();

		$scope = array(
			"head" => $this->head($place, $job, $running, $can, $control),
			"running" => $this->alert($running, $can),
			"denied" => !$can,
			"summary" => Row::text($job, "summary"),
			"job" => $this->detail($place, $job),
			"count" => count($found),
			"activity" => $this->activity($runs, $seen, $filter),
			"runs" => $this->history($found, $reports, $job, $can, $control),
			"can" => $can,
			"schedules" => $jobDir === "" ? $this->schedules($id, $job, $can, $users) : null,
		);

		$this->view->page($scope, "job.tpl");
	}

	/**
	 * head is the heading and the actions on the right: the job's name, its
	 * kind and whether it takes keyboard input, then the Attach of a run in
	 * progress or the Run of a reader who may start one, with the Pin after
	 * either.
	 *
	 * The div on the right is drawn however empty it is, and the paragraph
	 * under the heading however little the job says of itself, because that is
	 * what the markup of this page says.
	 */
	function head($place, $job, $running, $can, $control) {
		$head = new Head(Row::text($job, "name"));
		$buttons = new Buttons;

		$head->said(Row::text($job, "description"));
		$head->leads();
		$head->always();
		$head->tag(Row::text($job, "kind"));

		if (Row::flag($job, "interactive")) {
			$head->tag("interactive");
		}

		if ($running !== null) {
			$id = $running->getId();

			$buttons->primary("", "Attach", array("href" => Route::href("/runs/" . $id), "data-attach" => $id));
			$buttons->whenDanger($can, "", "Kill", array("data-kill" => $id));
		} elseif ($can) {
			$head->run($control->runButtons($place, $job));
		}

		if ($this->chrome->user !== null) {
			$pins = $this->chrome->pins();

			Pin::button($buttons, $pins, Row::text($place, "id"), Row::text($job, "id"), Row::text($job, "dir"), Row::text($job, "name"));
		}

		$head->buttons($buttons->data());

		return $head->data();
	}

	/**
	 * alert is the line about the run in progress, null with none: who started
	 * it, how long ago, and how many windows are watching it.
	 *
	 * The windows are the runner's count and no column of the row, so the
	 * session is read for the one run the line is about.
	 */
	function alert($running, $can) {
		if ($running === null) {
			return null;
		}

		$session = Host::session($running->getId());

		return array(
			"who" => $running->getUserName(),
			"ago" => Clock::ago($running->getStartedAt()),
			"watching" => Row::number($session, "attached"),
			"kills" => $can,
		);
	}

	/**
	 * detail is the Job panel: where the job is, the file it was detected in,
	 * what it runs, what it runs first, who may run it and the hours it runs
	 * in.
	 *
	 * Every piece is named on its own and the template writes the markup,
	 * because the cells hold more than the kv component draws: an icon before
	 * a path, a command in code and a Show button.
	 */
	function detail($place, $job) {
		$id = Row::text($place, "id");
		$dir = Row::text($job, "dir");
		$kind = Row::text($job, "kind");
		$source = Row::text($job, "source");
		$href = Route::href(Route::folderHref($id, $this->workspaces));
		$shows = null;

		// The button that shows the source is drawn for a target of a pipeline
		// file of the place itself, so a script, a skill and a job of a
		// directory inside read no lines, which is the rule the binding counts
		// them under.
		if ($kind !== "script" && $kind !== "skill" && $dir === "") {
			$shows = Place::source($id, $source, Row::number($job, "source_lines"));
		}

		$includes = array();

		foreach (Row::items($place, "includes") as $read) {
			$file = Row::text($read, "name");
			$includes[] = array(
				"file" => $file,
				"shows" => Place::source($id, $file, Row::number($read, "lines")),
			);
		}

		$allow = array_merge(Row::items($place, "allow"), Row::items($job, "allow"));

		return array(
			"folder" => $id,
			"folderHref" => $href,
			"title" => Row::text($place, "title"),
			"dir" => $dir,
			"dirHref" => $href . "/" . Route::escapePath($dir),
			"source" => $source,
			"shows" => $shows,
			"includes" => $includes,
			"command" => implode(" ", Row::items($job, "command")),
			"depends" => $this->depends($id, $job),
			"steps" => Row::items($job, "steps"),
			"aliases" => Row::items($job, "aliases"),
			"who" => Place::who($allow),
			"hours" => Place::hours(Row::text($place, "hours")),
			"interactive" => Row::flag($job, "interactive"),
		);
	}

	/**
	 * depends is the targets the job runs first: each one's own page where the
	 * place holds a job by that name, and the text unchanged where it does
	 * not.
	 *
	 * A target names a job of the place and not of the catalogue, so the
	 * directory it is linked under is the one the binding resolved it in.
	 */
	function depends($id, $job) {
		$out = array();

		foreach (Row::items($job, "depends") as $dep) {
			$href = "";

			if (Row::flag($dep, "found")) {
				$place = Route::placeHref($id, Row::text($dep, "dir"), $this->workspaces);
				$href = Route::href($place) . "/" . Row::text($dep, "id");
			}

			$out[] = array(
				"text" => Row::text($dep, "text"),
				"name" => Row::text($dep, "name"),
				"href" => $href,
			);
		}

		return $out;
	}

	/**
	 * reports is the summary of each listed run, by run, which a row draws
	 * under itself and its [...] offers to ask for.
	 *
	 * A reader without prompt/view reads none of them: a summary is a thread
	 * of Claude's and not a column of the run.
	 */
	function reports($found) {
		if (!$this->chrome->allowed("prompt/view")) {
			return array();
		}

		$ids = array();

		foreach ($found as $run) {
			$ids[] = $run->getId();
		}

		$board = new Report($this->chrome->db);

		return $board->summaries($ids);
	}

	/**
	 * activity is the charts over the table: every run of the job in the
	 * window the charts count, which is more runs than the table lists.
	 *
	 * The duration line is drawn, because the runs are one job's and mix
	 * nothing. The clock is the server's and the charts count their days in
	 * its zone.
	 *
	 * The window is read as rows and stays rows, as in the runs module: a
	 * model per row for five numbers was the cost of the page.
	 */
	function activity($runs, $seen, $filter) {
		$filter["limit"] = Activity::LIMIT;

		$matched = $runs->charted($filter);
		$found = $seen->rows($matched);
		$days = Activity::rows($found);
		$now = Clock::now();
		$charts = new Activity($days, $now, Activity::DAYS, true);

		return $charts->data();
	}

	/** history is the listed runs, each as the table draws it, in the order the query read them. */
	function history($found, $reports, $job, $can, $control) {
		$out = array();

		foreach ($found as $run) {
			$out[] = $this->row($run, $reports, $job, $can, $control);
		}

		return $out;
	}

	/**
	 * row is one run of the table: when it started, who started it, how long
	 * it took, how it ended, and Claude's summary of it where one was asked
	 * for.
	 *
	 * The runner is ahead of the table, so the session is laid over the row
	 * before a cell is read, as every table of runs reads one.
	 */
	function row($run, $reports, $job, $can, $control) {
		$id = $run->getId();
		$session = Host::session($id);

		Run::overlay($run, $session);

		$ended = Run::ended($run->getState());
		$report = isset($reports[$id]) ? $reports[$id] : null;

		return array(
			"id" => $id,
			"href" => Route::href("/runs/" . $id),
			"ago" => Clock::ago($run->getStartedAt()),
			"stamp" => Clock::read($run->getStartedAt()),
			"who" => $this->chrome->who()->ofRun($run),
			"duration" => Run::duration(Run::elapsed($run)),
			"state" => Run::label($run),
			"attach" => $ended ? "View" : "Attach",
			"pending" => $report !== null && $report["pending"],
			"summary" => $this->summary($report),
			"menu" => $this->menu($run, $report, $job, $can, $control, $ended),
		);
	}

	/**
	 * summary is the row under a run that holds Claude's summary of it: what
	 * it says, who asked for it and its report page.
	 *
	 * Null for a run nobody asked about, and for one whose summary Claude has
	 * not answered yet: the row under the run says what the summary says, and
	 * a pending one says nothing. The run's own cell notes that it is coming.
	 */
	function summary($report) {
		if ($report === null || $report["body"] === "") {
			return null;
		}

		return array(
			"by" => $report["by"],
			"body" => Host::markdown($report["body"]),
			"href" => Route::href(Report::href($report)),
		);
	}

	/**
	 * menu is a run's [...]: its ID and when it started, its own page, a
	 * replay of the recording, Claude's summary of it where one may be asked
	 * for, and Kill while it runs.
	 */
	function menu($run, $report, $job, $can, $control, $ended) {
		$id = $run->getId();
		$menu = new Menu("Actions on the run of " . $run->getJob());

		$menu->detail("Run", $id, true);
		$menu->detail("Started", Clock::read($run->getStartedAt()));
		$menu->item("file", "Open the run's page", array("href" => Route::href("/runs/" . $id)));
		$menu->when($run->getCastPath() !== "" && $ended, "play", "Replay the recording", array(
			"data-replay" => $id,
			"data-replay-title" => Run::name($run),
			"data-replay-who" => $run->getUserName(),
		));
		// The summary is asked for over the window the page is read in, so the
		// question goes on the item and not to an endpoint of its own.
		$menu->when($this->summarizes($run, $job, $report, $control, $ended), "note", "Summarize", array(
			"data-summarize-run" => $id,
			"data-question" => Report::SUMMARY,
			"title" => self::SUMMARIZE,
		));
		$menu->whenDanger(!$ended && $can, "stop", "Kill", array("data-kill" => $id));

		return $menu->data();
	}

	/**
	 * summarizes reports whether a run's [...] offers Claude's summary of it:
	 * a run that ended and is not Claude's own, of a job that summarizes at
	 * all, that nobody has asked about yet.
	 */
	function summarizes($run, $job, $report, $control, $ended) {
		if (!$ended || Run::isClaude($run)) {
			return false;
		}

		if (!$control->summarizes($job)) {
			return false;
		}

		return $report === null;
	}

	/**
	 * schedules is the Schedules block: what runs the job on a cron
	 * expression, and the form that adds one.
	 *
	 * A schedule names a folder and a job and no directory, so a job that runs
	 * in a directory inside a folder has no block at all: one of it could not
	 * be written, and the caller leaves the block out rather than drawing an
	 * empty one.
	 *
	 * The schedules of the folder are read and the job's own kept, which is
	 * the one query the folder's page makes as well.
	 */
	function schedules($id, $job, $can, $users) {
		$store = new Schedules\Schedules($this->chrome->db);
		$jobID = Row::text($job, "id");
		$rows = array();

		foreach ($store->byFolder($id) as $schedule) {
			if ($schedule->getJob() === $jobID) {
				$rows[] = $schedule;
			}
		}

		$buttons = new Buttons(true);
		$said = $can ? self::MORE : self::NONE;

		return array(
			"block" => array(
				"title" => "Schedules",
				"note" => self::NOTE,
				"empty" => count($rows) === 0 ? $said : "",
				"buttons" => $buttons->data(),
			),
			"table" => Schedule::table($this->chrome, $rows, $users, $can, Row::text($job, "name")),
			"folder" => $id,
			"job" => $jobID,
			"runAs" => array(
				"id" => "sch-user",
				"options" => array(
					"users" => $users,
					"selected" => $this->chrome->name(),
					"short" => false,
				),
			),
		);
	}
}
