<?php

// @route GET /runs/{id}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_run is the page of one run: the terminal while the run is live, its
// output once it ended, the prompt threads asked about it, its details and the
// files it left.
//
// The page reads the row and lays the runner's session over it, as the table
// of runs does, so a run that ended a moment ago reads the same on both.
class mod_run implements Module {
	/** EXPIRED is what the Output block says of a recording the sweep removed. */
	const EXPIRED = "The recording of this run expired under the Data retention settings: its output is gone. The run's details stay.";

	/** ARTIFACT_API is where a file of a run is read, at the root. */
	const ARTIFACT_API = "/api/artifacts/";

	/** SUMMARIZING stands where Summarize stood, while Claude writes the summary. */
	const SUMMARIZING = "summarizing";

	function name() {
		return "runs";
	}

	function title() {
		return "Runs";
	}

	function icon() {
		return "list";
	}

	function tables() {
		return ["run", "artifact", "note", "run_summary", "user", "workspace"];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);
		$runs = new Task\Runs\Runs;
		$seen = new Task\Seen($chrome);
		$run = $runs->get(Task\Route::param("id"));

		if ($run === null || !$seen->run($run->getId())) {
			$view->missing("No run with that ID.");

			return;
		}

		$name = Task\Run::name($run);

		$chrome->heading($name, [["text" => "Runs", "href" => "/runs"], ["text" => $name, "href" => ""]]);
		$this->at($chrome, $run);
		$chrome->asksOfRun($run);

		$scope = $this->scope($chrome, $run, $seen);

		$view->page($scope, "run.tpl");
	}

	/**
	 * at marks in the sidebar where the run ran: the folder of the run, and
	 * the job itself where the directory it ran in still holds one.
	 *
	 * Both are read off the catalogue and not off the row, as the Go page
	 * reads them: a run of a folder that is gone marks nothing, and a run of a
	 * job that is gone marks the folder alone. The folder is read of the whole
	 * catalogue rather than the sidebar's, so a run the reader may open in a
	 * folder they may not marks a row the sidebar does not draw.
	 */
	function at($chrome, $run) {
		$id = $run->getFolder();
		$found = false;

		foreach ($chrome->all() as $folder) {
			if (Task\Row::text($folder, "id") === $id) {
				$found = true;
			}
		}

		if (!$found) {
			return;
		}

		$job = Task\Host::job($id, $run->getDir(), $run->getJob());
		$key = Task\Row::flag($job, "found") ? Task\Row::text($job, "key") : "";

		$chrome->at($id, $key);
	}

	/**
	 * scope is what the template reads: the head, the live terminal, the
	 * output, the prompt box, the details and the files.
	 *
	 * The report and the expired recording are read of the row and not of the
	 * session, as the Go page reads them: the session goes over the row after
	 * both.
	 */
	function scope($chrome, $run, $seen) {
		$report = null;

		if ($chrome->allowed("prompt/view")) {
			$reports = new Task\Report;
			$report = $reports->of($run);
		}

		$expired = Task\Host::expired($run->getCastPath(), $run->getState());
		$session = Task\Host::session($run->getId());
		$message = Task\Run::overlay($run, $session);
		$name = Task\Run::name($run);
		$state = $run->getState();
		$ended = Task\Run::ended($state);
		$can = $this->can($chrome, $run);
		$summarizes = $this->summarizes($chrome, $run, $report, $ended);
		$pending = $this->pending($chrome, $run);
		$asks = $chrome->allowed("prompt/edit");
		$thread = $this->thread($run);
		$buttons = self::buttons($run, $report, $expired, $can, $session, $summarizes, $pending, $asks, $thread);
		$code = $ended && $state !== Task\Run::OK ? (string)$run->getExitCode() : "";
		$head = new Task\Head($name);

		$head->said(self::lead($run, $expired, $can));
		$head->badge($code === "" ? $state : $state . " " . $code, Task\Run::tone($state));
		$head->note($run->getTitle());
		$head->buttons($buttons->data());

		return [
			"head" => $head->data(),
			"terminal" => self::terminal($run, $session, $name),
			"output" => [
				"expired" => $expired,
				"recording" => !$expired && $run->getCastPath() !== "" && $ended,
				"text" => self::EXPIRED,
				"id" => $run->getId(),
				"cols" => $run->getTermCols(),
				"rows" => $run->getTermRows(),
			],
			"prompt" => Task\Prompt::panel($chrome, $run, $seen),
			"children" => $this->children($run, $seen),
			"kv" => ["rows" => $this->details($run, $chrome, $message)],
			"artifacts" => $this->artifacts($chrome, $run),
		];
	}

	/**
	 * buttons are the actions of the page head: attach to the run, run its job
	 * again once it ended, stop and kill it while the runner holds it, open its
	 * report or ask Claude for one, and replay the recording of one that ended.
	 *
	 * A summary Claude is still writing says so where Summarize stood, as a
	 * faint item of the row: the button would post the question a second time.
	 *
	 * Run again posts the run's own ID: the endpoint reads the inputs off the
	 * row, and asks the thread's question again for a prompt's run, which no
	 * job can repeat. A prompt's run takes prompt/edit and not the job's gate.
	 */
	static function buttons($run, $report, $expired, $can, $session, $summarizes, $pending, $asks, $thread = null) {
		$id = $run->getId();
		$ended = Task\Run::ended($run->getState());
		$stops = Task\Row::flag($session, "live") && !$ended && $can;
		$claude = Task\Run::isPrompt($run) || Task\Run::isNote($run);
		$again = $ended && ($claude ? $asks : $can);
		$href = $report === null ? "" : Task\Route::href(Task\Report::href($report));
		$buttons = new Task\Buttons;

		$buttons->primary("", $ended ? "View" : "Attach", ["data-attach" => $id]);
		$buttons->when($thread !== null, "note", "Open thread", ["href" => $thread["href"], "title" => $thread["title"]]);
		$buttons->when($again, "", "Run again", ["data-rerun" => $id, "data-title" => Task\Run::name($run)]);
		$buttons->when($stops, "", "Stop", ["data-stop" => $id]);
		$buttons->whenDanger($stops, "", "Kill", ["data-kill" => $id]);
		$buttons->when($href !== "", "", "Report", ["href" => $href]);
		$buttons->when($summarizes, "", "Summarize", [
			"data-summarize-run" => $id,
			"data-question" => Task\Report::SUMMARY,
			"title" => Task\Job::SUMMARIZE,
		]);

		if ($pending) {
			$buttons->faint(self::SUMMARIZING);
		}

		$buttons->when($run->getCastPath() !== "" && $ended && !$expired, "", "|>", [
			"data-replay" => $id,
			"data-replay-title" => Task\Run::name($run),
			"data-replay-who" => $run->getUserName(),
			"title" => "Replay the recording",
			"aria-label" => "Replay the recording",
		]);

		return $buttons;
	}

	/**
	 * summarizes reports whether the head holds Summarize: a run that ended, of
	 * a job that summarizes at all, with no summary read yet nor one under way.
	 *
	 * The job comes off the catalogue, as the Go page's p.Job does: a run of
	 * Claude's own and a run of a job that is gone each summarize nothing.
	 */
	function summarizes($chrome, $run, $report, $ended) {
		if (!$ended || $report !== null || Task\Run::isClaude($run)) {
			return false;
		}

		if ($this->pending($chrome, $run)) {
			return false;
		}

		$job = Task\Host::job($run->getFolder(), $run->getDir(), $run->getJob());

		if (!$job["found"]) {
			return false;
		}

		$control = new Task\Control($chrome);

		return $control->summarizes($job);
	}

	/**
	 * pending reports whether Claude is still writing the run's summary.
	 *
	 * A pending summary is no report: Task\Report::of answers null for it. The
	 * question goes to the summaries themselves, read under prompt/view alone
	 * as the Go page reads them.
	 */
	function pending($chrome, $run) {
		if (!$chrome->allowed("prompt/view")) {
			return false;
		}

		$id = $run->getId();
		$reports = new Task\Report;
		$found = $reports->summaries([$id]);

		return isset($found[$id]) && $found[$id]["pending"];
	}

	/** thread is the thread a run belongs to, or null for a run nobody asked about. */
	function thread($run) {
		$notes = new Task\Threads\Notes;
		$note = $notes->byRun($run->getId());

		if ($note === null) {
			return null;
		}

		return [
			"id" => $note->getId(),
			"href" => Task\Route::href("/threads/" . Task\Route::pathEscape($note->getId())),
			"title" => Task\Threads\Conversation::title($note, []),
		];
	}

	/** parent is the cell that names what started a run, linked where the page is. */
	function parent($run, $id) {
		$kind = $run->getParentKind();

		if ($kind === "workflow_run") {
			$flows = new Task\Workflows\Runs;
			$flow = $flows->get($id);
			$href = $flow === null ? "" : "/workflows/" . $flow->getWorkflowId() . "/runs/" . $id;

			return $href === "" ? ["text" => $id, "after" => "a workflow run"] : [
				"link" => self::link(Task\Route::href($href), "flow", $flow->getName()),
				"after" => "workflow run " . substr($id, 0, 8),
			];
		}

		$runs = new Task\Runs\Runs;
		$found = $runs->get($id);

		return $found === null ? ["text" => $id, "after" => "a run"] : [
			"link" => self::link(Task\Route::href("/runs/" . $id), "play", Task\Run::name($found)),
			"after" => "run " . substr($id, 0, 8),
		];
	}

	/**
	 * children is the runs this run dispatched that the reader sees, oldest
	 * first: each a link, its job and how it ended.
	 *
	 * The rows are read as they stand and no session goes over them, as the Go
	 * page reads them: a run of the list is somewhere else on the page of its
	 * own.
	 */
	function children($run, $seen) {
		$runs = new Task\Runs\Runs;
		$found = $runs->children($run->getId());
		$rows = [];

		foreach ($seen->of($found) as $child) {
			$rows[] = [
				"href" => Task\Route::href("/runs/" . $child->getId()),
				"name" => Task\Run::name($child),
				"state" => Task\Run::label($child),
			];
		}

		return $rows;
	}

	/** terminal is the live screen of a run the runner still holds, or null. */
	static function terminal($run, $session, $name) {
		if (!Task\Row::flag($session, "live") || Task\Run::ended($run->getState())) {
			return null;
		}

		return [
			"id" => $run->getId(),
			"name" => $name,
			"watching" => Task\Row::number($session, "attached"),
			"frame" => Task\Host::frame($run->getId()),
		];
	}

	/**
	 * details are the rows of the Run table: where the run ran, what it was
	 * given, who started it, how long it took and how it ended.
	 */
	function details($run, $chrome, $message) {
		$workspaces = $chrome->workspaces();
		$folder = $run->getFolder();
		$dir = $run->getDir();
		$job = $run->getJob();
		$jobHref = Task\Run::runJobHref($run, $workspaces);
		$args = Task\Run::argList($run);
		$rows = [];

		$rows[] = self::detail("Folder", [
			"link" => self::link(Task\Route::href(Task\Route::folderHref($folder, $workspaces)), "folder", $folder),
		]);

		if ($dir !== "") {
			$rows[] = self::detail("Directory", [
				"link" => self::link(Task\Route::href(Task\Route::placeHref($folder, $dir, $workspaces)), "folder-open", $dir . "/"),
				"after" => "the job ran here",
			]);
		}

		$rows[] = self::detail("Job", $jobHref === "" ? ["text" => $job] : ["link" => self::link(Task\Route::href($jobHref), "", $job)]);

		$parentID = $run->getParentId();

		if ($parentID !== "") {
			$rows[] = self::detail("Started by", $this->parent($run, $parentID));
		}

		if (count($args) > 0) {
			$rows[] = self::detail("Arguments", ["code" => $args, "after" => "after --, one argument each"]);
		}

		$rows[] = self::detail("By", ["who" => $chrome->who()->ofRun($run)]);
		$rows[] = self::detail("Started", ["text" => Task\Clock::read($run->getStartedAt())]);
		$rows[] = self::detail("Finished", $run->getFinishedAt() === "" ? ["faint" => "in progress"] : ["text" => Task\Clock::read($run->getFinishedAt())]);
		$rows[] = self::detail("Duration", ["text" => Task\Run::duration(Task\Run::elapsed($run))]);
		$rows[] = self::detail("Result", ["state" => Task\Run::label($run), "after" => $message]);

		if ($run->getReason() !== "") {
			$rows[] = self::detail("Reason", [
				"text" => $run->getReason(),
				"badge" => ["tone" => "bad", "text" => "out of hours"],
			]);
		}

		$rows[] = self::detail("Terminal", ["text" => $run->getTermCols() . "×" . $run->getTermRows()]);
		$rows[] = self::detail("ID", ["class" => "faint", "text" => $run->getId()]);

		return $rows;
	}

	/**
	 * artifacts is the Artifacts block: the files the run left under the
	 * directory its job was given, newest first.
	 *
	 * The table is the module's own markup, as the table of schedules is. The
	 * theme's kv component is the detail table beside it, and no other page
	 * lists a run's files.
	 *
	 * The links read the API at the root and not under the theme's mount, as
	 * every data- endpoint of a page does.
	 */
	function artifacts($chrome, $run) {
		$artifacts = new Task\Artifacts\Artifacts;
		$found = $artifacts->find(["run" => $run->getId()]);
		$deletes = $chrome->allowed("artifact/delete");
		$rows = [];

		foreach ($found as $file) {
			$name = $file->getName();
			$href = self::ARTIFACT_API . $file->getId();
			$menu = new Task\Menu("Actions on " . $name);

			$menu->whenDanger($deletes, "trash", "Delete", [
				"data-delete" => $href,
				"data-confirm" => "Delete " . $name . "?",
				"data-confirm-ok" => "Delete",
				"data-next" => "reload",
			]);

			$rows[] = [
				"name" => $name,
				"href" => $href,
				"size" => Task\Size::bytes($file->getByteSize()),
				"type" => $file->getContentType(),
				"menu" => $menu->data(),
			];
		}

		$buttons = new Task\Buttons(true);

		return [
			"block" => [
				"title" => "Artifacts",
				"note" => 'files left under $TASK_UI_ARTIFACTS',
				"empty" => count($rows) === 0 ? 'None: a job keeps files by writing them to the directory in $TASK_UI_ARTIFACTS.' : "",
				"buttons" => $buttons->data(),
			],
			"rows" => $rows,
		];
	}

	/** can reports whether the reader may stop or kill the run. */
	function can($chrome, $run) {
		return Task\Run::canEnd($chrome, $run);
	}

	/**
	 * lead is the sentence under the heading: who started the run and how, and
	 * then what the reader may do about it where it stands.
	 *
	 * The whole sentence is built here and not in the template. A template
	 * writes the newline that sits before a tag, so a sentence cut across
	 * branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($run, $expired, $can) {
		$said = "Started by " . $run->getUserName() . " " . Task\Clock::ago($run->getStartedAt());

		if (Task\Run::isScheduled($run)) {
			$said .= " by the scheduler";
		}

		if (Task\Run::isForced($run)) {
			$said .= ", with Run force";
		}

		$said .= ".";
		$state = $run->getState();

		if ($state === Task\Run::QUEUED) {
			$said .= " It waits, queued, for the run of its job before it to end or for a free slot: attach to follow it once it starts";

			return $said . ($can ? ", or stop it to take it out of the queue." : ".");
		}

		if (!Task\Run::ended($state)) {
			$said .= " It's still running: attach to follow it";

			return $said . ($can ? ", or stop it." : ".");
		}

		return $said . ($expired ? " The run has finished; its recording expired." : " The run has finished; its full output is below.");
	}

	/** detail is one row of the Run table: its label, and the pieces of its cell. */
	static function detail($label, $cell) {
		$row = [
			"label" => $label,
			"class" => "",
			"who" => [],
			"state" => [],
			"link" => self::link("", "", ""),
			"code" => [],
			"text" => "",
			"faint" => "",
			"after" => "",
			"badge" => ["tone" => "", "text" => ""],
		];

		foreach ($cell as $key => $value) {
			$row[$key] = $value;
		}

		return $row;
	}

	/** link is the link of a cell: where it goes, its icon and what it says. */
	static function link($href, $icon, $text) {
		return [
			"href" => $href,
			"icon" => $icon,
			"text" => $text,
		];
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_run;

	$mod->serve();
}
