<?php

namespace Task;

// Control is the Jobs table of a folder's page, and of a directory inside one:
// a row per row of Task\Host::groups(), the row's label linking the place its
// targets run in, then a button per job of it.
//
// A button is a Run, an eye on the run in progress, or a #! to the job's page
// for a reader who may not run it, and it takes the colour of the job's last
// run. What a row leaves out sits in its [...]: Run and summarize, the last
// run and its summary, History, Pin, Run force and Kill.
//
// A row reads nothing for itself. Every run, every summary and every pin the
// table draws is read here, once, however many jobs the page lists.
class Control {
	/** NEVER, OK and BAD are the colour a button takes from the last run. */
	const NEVER = "never";
	const OK = "ok";
	const BAD = "bad";

	var $db;
	var $chrome;
	var $access;
	var $workspaces;

	/** last is the run that ended last of a job, by the job's key. */
	var $last = array();

	/** live is the run in progress of a job as its row, by the job's key. */
	var $live = array();

	/** summaries is where the summary of a job's last run is linked, by key. */
	var $summaries = array();

	/** pins are the reader's, which decide whether a Pin item pins or unpins. */
	var $pins = array();

	/** pinning says whether a row's [...] carries a Pin at all. */
	var $pinning = false;

	/** summarizing says whether a Run offers Claude's summary beside it. */
	var $summarizing = false;

	function __construct($chrome) {
		$this->chrome = $chrome;
		$this->db = $chrome->db;
		$this->access = $chrome->access;
		$this->workspaces = $chrome->workspaces();
		$this->pinning = $chrome->user !== null;
		$this->summarizing = $chrome->allowed("prompt/edit") && Host::claude();
	}

	/**
	 * of is the rows of the table, one per group, in the groups' order.
	 *
	 * The runs, the summaries and the pins are read first and read once: a
	 * folder's page draws a couple of hundred buttons and asks the database for
	 * three statements, whatever they say.
	 */
	function of($groups) {
		$this->last = $this->latest($groups);
		$this->live = $this->chrome->liveRuns();
		$this->summaries = $this->summarized($groups);
		$this->pins = $this->pinning ? $this->chrome->pins() : array();

		$out = array();

		foreach ($groups as $group) {
			$out[] = $this->row($group);
		}

		return $out;
	}

	/**
	 * latest is the run that ended last of each job of the folders the rows run
	 * as, by the job's key.
	 *
	 * The colour of a button is not a permission: a run the reader may not open
	 * still says the job failed, which is what the Go page colours by.
	 */
	function latest($groups) {
		$ids = array();

		foreach ($groups as $group) {
			$folder = Row::items($group, "folder");
			$id = Row::text($folder, "id");

			if ($id !== "") {
				$ids[$id] = true;
			}
		}

		if (count($ids) === 0) {
			return array();
		}

		$runs = new Run($this->db);

		return $runs->latest(array_keys($ids));
	}

	/**
	 * summarized is where the summary of each job's last run is linked, by the
	 * job's key, for a reader who may read prompts.
	 */
	function summarized($groups) {
		if (!$this->access->allowed("prompt/view")) {
			return array();
		}

		$keys = array();

		foreach ($groups as $group) {
			foreach (Row::items($group, "jobs") as $job) {
				$key = Row::text($job, "key");

				if (!isset($this->last[$key])) {
					continue;
				}

				$keys[$this->last[$key]->getId()] = $key;
			}
		}

		if (count($keys) === 0) {
			return array();
		}

		$reports = new Report($this->db);
		$out = array();

		foreach ($reports->jobSummaries(array_keys($keys)) as $id => $link) {
			$out[$keys[$id]] = $link;
		}

		return $out;
	}

	/**
	 * row is one group as the table draws it: its label, what it says on a
	 * hover, the place a row of targets links to, its buttons and its [...].
	 *
	 * A row of targets is the one kind whose label is a link, and the one kind
	 * that says anything when it holds no job.
	 */
	function row($group) {
		$folder = Row::items($group, "folder");
		$jobs = Row::items($group, "jobs");
		$label = Row::text($group, "kind");
		$title = Row::text($group, "note");
		$href = "";
		$empty = "";

		if (Row::flag($group, "targets")) {
			$dir = Row::text($group, "dir");
			$source = Row::text($group, "source");
			$label = $dir === "" ? $source : $dir . "/" . $source;
			$title = Row::text($group, "kind") . " " . Row::text($group, "note");
			$href = Route::href(Route::placeHref(Row::text($folder, "id"), Row::text($folder, "rel"), $this->workspaces));
			$empty = count($jobs) > 0 ? "" : Row::text($group, "empty");
		}

		return array(
			"label" => $label,
			"title" => $title,
			"href" => $href,
			"empty" => $empty,
			"jobs" => $this->buttons($folder, $jobs),
			"menu" => $this->menu($label, $folder, $jobs),
		);
	}

	/** buttons is the row's jobs, each as the control-job component reads it. */
	function buttons($folder, $jobs) {
		$out = array();

		foreach ($jobs as $job) {
			$out[] = $this->button($folder, $job);
		}

		return $out;
	}

	/**
	 * button is one job of a row: what it says, what a click starts, and the
	 * one of the three shapes it takes.
	 *
	 * A job whose ID is not its name says its ID, which is what tells two
	 * sources that share a name apart, as the sidebar does.
	 */
	function button($folder, $job) {
		$id = Row::text($job, "id");
		$name = Row::text($job, "name");
		$dir = Row::text($job, "dir");
		$key = Row::text($job, "key");
		$args = Row::text($job, "args_command");
		$what = self::what($folder, $job);
		$desc = self::describes($job);
		$terminal = Row::flag($job, "interactive");
		$can = Folders::allows($this->access, $folder, $job);
		$run = isset($this->live[$key]) ? $this->live[$key] : null;

		return array(
			"key" => $key,
			"folder" => Row::text($folder, "id"),
			"job" => $id,
			"name" => $name,
			"label" => $id === $name ? $name : $id,
			"last" => $this->colour($key),
			"can" => $can,
			"dir" => $dir,
			"desc" => $desc,
			"args" => $args,
			"confirm" => Row::text($job, "confirm"),
			"path" => Row::text($job, "kind") === "file" ? Row::text($job, "source") : "",
			"terminal" => $terminal,
			"pinned" => Pin::of($this->pins, Row::text($folder, "id"), $id, $dir) !== null,
			"what" => $what,
			"run" => $run === null ? "" : $run->getId(),
			"href" => Route::href(Route::placeHref(Row::text($folder, "id"), $dir, $this->workspaces) . "/" . $id),
			"title" => self::says($run, $what, $desc, $args, $terminal, $can),
		);
	}

	/**
	 * colour is what a button's last run makes of it: never for a job that has
	 * not run, ok for one that ended well and bad for anything else.
	 */
	function colour($key) {
		if (!isset($this->last[$key])) {
			return self::NEVER;
		}

		return $this->last[$key]->getState() === Run::OK ? self::OK : self::BAD;
	}

	/**
	 * says is the hover of whichever shape the button takes: what it attaches
	 * to, what it runs, or why it will not run.
	 */
	static function says($run, $what, $desc, $args, $terminal, $can) {
		$detail = $desc === "" ? "" : ": " . $desc;

		if ($run !== null) {
			$title = $run->getTitle();

			return "Attach to " . $what . ", running" . ($title === "" ? "" : ": " . $title);
		}

		if (!$can) {
			return $what . $detail . ": no permission to run it";
		}

		$out = "Run " . $what . $detail;

		if ($args !== "") {
			$out .= ", with arguments after " . $args;
		}

		if ($terminal) {
			$out .= ", in a terminal";
		}

		return $out;
	}

	/** what is the title a run of a job is listed under: its place and its name. */
	static function what($folder, $job) {
		$place = Row::text($folder, "id");
		$dir = Row::text($job, "dir");

		if ($dir !== "") {
			$place .= "/" . $dir;
		}

		return $place . " / " . Row::text($job, "name");
	}

	/** describes is a job's description where it says more than its name, "" otherwise. */
	static function describes($job) {
		$desc = Row::text($job, "description");

		return $desc === Row::text($job, "name") ? "" : $desc;
	}

	/**
	 * menu is a row's [...]: per job, Run and summarize, its last run and the
	 * summary of it, History, Pin, Run force and Kill, each job's items a group
	 * of their own.
	 */
	function menu($label, $folder, $jobs) {
		$menu = new Menu("Actions on " . $label);
		$id = Row::text($folder, "id");
		$hours = Row::text($folder, "hours");

		foreach ($jobs as $job) {
			$this->items($menu, $id, $hours, $folder, $job);
			$menu->sep();
		}

		return $menu->data();
	}

	/** items is the entries of one job, which the row lists under its own. */
	function items($menu, $id, $hours, $folder, $job) {
		$job_id = Row::text($job, "id");
		$key = Row::text($job, "key");
		$dir = Row::text($job, "dir");
		$can = Folders::allows($this->access, $folder, $job);
		$idle = !isset($this->live[$key]);
		$start = $this->starts($id, $folder, $job);

		if ($idle && $can && $this->summarizes($job)) {
			$label = "Run and summarize " . $job_id;

			if (Row::flag($job, "takes_args")) {
				$label .= "...";
			}

			$menu->item("summary", $label, array_merge(array("data-summarize" => ""), $start));
		}

		if (isset($this->last[$key])) {
			$menu->item("eye", "View the last run of " . $job_id, array("data-attach" => $this->last[$key]->getId()));
		}

		if (isset($this->summaries[$key])) {
			$found = $this->summaries[$key];

			$menu->item("summary", "Summary of " . $job_id, array("href" => Route::href($found["href"]), "title" => $found["preview"]));
		}

		$menu->item("list", "History of " . $job_id, array("href" => Route::href(Route::placeHref($id, $dir, $this->workspaces) . "/" . $job_id)));

		if ($this->pinning) {
			Pin::item($menu, $this->pins, $id, $job_id, $dir, $job_id);
		}

		if ($idle && $can && $hours !== "") {
			$forced = array("data-force" => "", "title" => "Run out of hours, with a reason");

			$menu->danger("play", "Run force " . $job_id, array_merge($forced, $start));
		}

		if (!$idle && $can) {
			$menu->danger("stop", "Kill " . $job_id, array("data-kill" => $this->live[$key]->getId()));
		}

		return $menu;
	}

	/**
	 * starts is the attributes every entry that starts a run carries: what to
	 * run, under what title, what it asks first and the command its arguments
	 * follow.
	 */
	function starts($id, $folder, $job) {
		$out = array(
			"data-run" => "",
			"data-folder" => $id,
			"data-dir" => Row::text($job, "dir"),
			"data-job" => Row::text($job, "id"),
			"data-title" => self::what($folder, $job),
		);
		$confirm = Row::text($job, "confirm");

		if ($confirm !== "") {
			$out["data-confirm"] = $confirm;
			$out["data-confirm-ok"] = "Run " . Row::text($job, "name");
		}

		if (Row::flag($job, "takes_args")) {
			$out["data-args"] = Row::text($job, "args_command");
		}

		return $out;
	}

	/**
	 * summarizes reports whether a job's Run has Claude's summary beside it: a
	 * job that ends on its own, for a reader who may ask.
	 */
	function summarizes($job) {
		if (!$this->summarizing) {
			return false;
		}

		if (Row::flag($job, "shell") || Row::flag($job, "interactive")) {
			return false;
		}

		$id = Row::text($job, "id");
		$kind = Row::text($job, "kind");

		// A prompt's job and the Notes agent's are Claude's own and summarize
		// nothing: the two are the ID and the convention together, as the
		// folder package names them.
		if ($id === Run::PROMPT_JOB && $kind === "prompt") {
			return false;
		}

		return !($id === Run::NOTE_JOB && $kind === "note");
	}

	/**
	 * runButtons is the Run of a job's own page: the split button, with the
	 * other ways to run it under the caret.
	 *
	 * place is the title Run force gives the run, which leaves the job's name
	 * out, and what the Run itself gives it names the job.
	 */
	function runButtons($folder, $job) {
		$args = Row::text($job, "args_command");
		$takes = Row::flag($job, "takes_args");
		$summarize = $this->summarizes($job);
		$force = Row::text($folder, "hours") !== "";
		$dir = Row::text($job, "dir");

		return array(
			"folder" => Row::text($folder, "id"),
			"dir" => $dir,
			"job" => Row::text($job, "id"),
			"name" => Row::text($job, "name"),
			"what" => self::what($folder, $job),
			"place" => Route::place(Row::text($folder, "id"), $dir),
			"args" => $args,
			"takesArgs" => $takes,
			"summarize" => $summarize,
			"force" => $force,
			"menu" => $summarize || $force,
		);
	}
}
