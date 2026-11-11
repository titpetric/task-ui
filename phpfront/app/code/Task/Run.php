<?php

namespace Task;

// Run answers what a row of the run table means: the name a page calls it by,
// the place it ran, how long it took and the label that follows it over the
// event stream.
//
// The row behaviour sits here rather than on Task\Model\Run because the
// runtime merges no methods into a generated class: one class owns the table
// and its rows. The rows are the runs module's, in Task\Runs\Runs.
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

	/** HOST_PREFIX starts the ID of a command on a host; HOST_ALL is every host. */
	const HOST_PREFIX = "@host:";

	const HOST_ALL = "@host:*";

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
	 * seenRow is seen of a row of Task\Runs\Runs::charted, for the five
	 * thousand rows of a chart's window, which are read as rows and never
	 * hydrated.
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
