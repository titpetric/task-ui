<?php

namespace Task;

// Report is the markdown a run stands for, without its terminal: the answer a
// run of Claude's wrote to a prompt, or the summary Claude wrote of another
// run.
//
// A report is an array, because the runtime merges no methods into a row and a
// report is not a row: it is one thread, one reply and the run either names.
// The fields are `run`, the run the report is of; `about`, the run Claude
// summarized or asked about; `thread`, the prompt thread that holds the reply;
// `question` and `by`, what was asked and who asked it; `body` and `at`,
// Claude's reply and when it was written; and `pending`, which says a question
// waits on that reply.
class Report {
	/** SUMMARY is the question the summary of one run asks. */
	const SUMMARY = "Summarize this run.";

	/** BATCH is the question the summary of a batch of host runs asks. */
	const BATCH = "Summarize the runs on these hosts.";

	/** PREVIEW is how much of a summary a Jobs row shows on a hover, in characters. */
	const PREVIEW = 240;

	var $db;

	var $notes;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
		$this->notes = new Note($this->db);
	}

	/**
	 * of is the report of a run, or null for a run that has none.
	 *
	 * A run of Claude's on a prompt reports the answer it wrote. Another run
	 * reports its summary, and a run of a batch reports the batch's, which the
	 * oldest run of the batch holds. A pending summary is no report.
	 */
	function of($run) {
		if ($run === null) {
			return null;
		}

		if (Run::isPrompt($run)) {
			return $this->answerOf($run);
		}

		$runs = new Runs\Runs($this->db);
		$about = $runs->root($run);
		$found = $this->summaries(array($about));

		if (!isset($found[$about]) || $found[$about]["pending"]) {
			return null;
		}

		$report = $found[$about];
		$report["run"] = $run->getId();

		return $report;
	}

	/**
	 * answerOf is the report of a run of Claude's on a prompt: the answer it
	 * wrote, null before the run ended.
	 */
	function answerOf($run) {
		if (!Run::ended($run->getState())) {
			return null;
		}

		$thread = $this->notes->byRun($run->getId());

		if ($thread === null) {
			return null;
		}

		$replies = $this->notes->replies(array($thread->getId()));
		$under = isset($replies[$thread->getId()]) ? $replies[$thread->getId()] : array();
		$answer = self::answer($under, $run->getId());

		if ($answer === null) {
			return null;
		}

		$asked = self::question($thread, $under, $answer);

		return array(
			"run" => $run->getId(),
			"about" => $thread->getAboutRunId(),
			"thread" => $thread->getId(),
			"question" => $asked->getBody(),
			"by" => $asked->getUserName(),
			"body" => $answer->getBody(),
			"at" => $answer->getCreatedAt(),
			"pending" => false,
		);
	}

	/**
	 * summaries is the reports of the runs that have a summary, or one under
	 * way, by run.
	 *
	 * A run's report is the oldest summary thread about it that did not fail,
	 * and where there is none, the summary the runs page kept.
	 */
	function summaries($runIDs) {
		$out = array();

		if (count($runIDs) === 0) {
			return $out;
		}

		$threads = $this->notes->promptsAbout($runIDs);
		$ids = array();

		foreach ($threads as $thread) {
			if (self::asks($thread)) {
				$ids[] = $thread->getId();
			}
		}

		$replies = $this->notes->replies($ids);

		// The threads come back newest first, so the oldest one is read last
		// and stays. A pending report gives way to one that holds an answer.
		foreach ($threads as $thread) {
			$id = $thread->getId();
			$report = self::summaryOf($thread, isset($replies[$id]) ? $replies[$id] : array());

			if ($report === null) {
				continue;
			}

			$kept = isset($out[$report["run"]]) ? $out[$report["run"]] : null;

			if ($kept === null || $kept["pending"] || $report["body"] !== "") {
				$out[$report["run"]] = $report;
			}
		}

		foreach ($runIDs as $id) {
			if (isset($out[$id])) {
				continue;
			}

			$kept = $this->kept($id);

			if ($kept !== null) {
				$out[$id] = $kept;
			}
		}

		return $out;
	}

	/**
	 * jobSummaries is where the summary of each of these runs is linked and
	 * what it says, by run, for the [...] of a Jobs row.
	 *
	 * It is not summaries(): a Jobs row links a thread that asked for a summary
	 * whether Claude has answered yet or not, a thread that failed counts, and
	 * no summary the runs page kept counts, because what the row links is the
	 * thread's report page.
	 *
	 * The threads come back newest first, so the oldest one is read last and
	 * stays, which is the thread summaries() keeps as well.
	 */
	function jobSummaries($runIDs) {
		$out = array();

		if (count($runIDs) === 0) {
			return $out;
		}

		$threads = array();
		$ids = array();

		foreach ($this->notes->promptsAbout($runIDs) as $thread) {
			if (!self::asks($thread)) {
				continue;
			}

			$threads[] = $thread;
			$ids[] = $thread->getId();
		}

		$replies = $this->notes->replies($ids);

		foreach ($threads as $thread) {
			$id = $thread->getId();
			$under = isset($replies[$id]) ? $replies[$id] : array();
			$body = count($under) === 0 ? "" : $under[0]->getBody();
			$out[$thread->getAboutRunId()] = self::link($thread->getAboutRunId(), $body);
		}

		return $out;
	}

	/**
	 * link is a summary as a row links it: the report page of the run, and the
	 * start of the text on one line, which is the hover.
	 */
	static function link($runID, $text) {
		$words = preg_split('/\s+/', $text, -1, PREG_SPLIT_NO_EMPTY);
		$preview = implode(" ", $words);

		if (mb_strlen($preview) > self::PREVIEW) {
			$preview = mb_substr($preview, 0, self::PREVIEW) . "…";
		}

		return array("href" => "/runs/" . $runID . "/report", "preview" => $preview);
	}

	/**
	 * kept is the summary the runs page wrote for a run, or null for a run
	 * with none. It is a report of its own: a summary with no thread.
	 */
	function kept($runID) {
		$summaries = new Runs\Summaries($this->db);
		$row = $summaries->of($runID);

		if ($row === null || Row::text($row, "summary") === "") {
			return null;
		}

		return array(
			"run" => $runID,
			"about" => $runID,
			"thread" => "",
			"question" => "",
			"by" => Row::text($row, "user_name"),
			"body" => Row::text($row, "summary"),
			"at" => Row::stamp($row, "created_at"),
			"pending" => false,
		);
	}

	/**
	 * summaryOf is the report of a summary thread with its replies: the first
	 * answer, or a pending report while the thread waits on one. A thread that
	 * is no summary, and one that failed without an answer, have none.
	 */
	static function summaryOf($thread, $replies) {
		if (!self::asks($thread)) {
			return null;
		}

		$report = array(
			"run" => $thread->getAboutRunId(),
			"about" => $thread->getAboutRunId(),
			"thread" => $thread->getId(),
			"question" => $thread->getBody(),
			"by" => $thread->getUserName(),
			"body" => "",
			"at" => "",
			"pending" => false,
		);
		$answer = self::answer($replies, self::summaryRun($thread, $replies));

		if ($answer !== null) {
			$report["body"] = $answer->getBody();
			$report["at"] = $answer->getCreatedAt();

			return $report;
		}

		if ($thread->getState() === Note::FAILED) {
			return null;
		}

		$report["pending"] = true;

		return $report;
	}

	/**
	 * answer is the reply of a thread that a run of Claude's wrote: the last
	 * reply naming the run that is not at work, null until the run answered.
	 */
	static function answer($replies, $runID) {
		if ($runID === "") {
			return null;
		}

		for ($i = count($replies) - 1; $i >= 0; $i--) {
			$reply = $replies[$i];

			if ($reply->getRunId() === $runID && $reply->getState() !== Note::WORKING && $reply->getBody() !== "") {
				return $reply;
			}
		}

		return null;
	}

	/**
	 * question is the note an answer replies to: the reply before it in the
	 * thread, or the thread itself.
	 */
	static function question($thread, $replies, $answer) {
		$count = count($replies);

		for ($i = 0; $i < $count; $i++) {
			if ($replies[$i]->getId() !== $answer->getId()) {
				continue;
			}

			return $i > 0 ? $replies[$i - 1] : $thread;
		}

		return $thread;
	}

	/** asks reports whether a thread is the summary a run page asks for. */
	static function asks($thread) {
		if ($thread === null || $thread->getParentId() !== "" || $thread->getAboutRunId() === "") {
			return false;
		}

		return $thread->getBody() === self::SUMMARY || $thread->getBody() === self::BATCH;
	}

	/**
	 * summaryRun is the run that answers a thread's question: the run its
	 * first reply names, and while there is none, the thread's own run. A
	 * reply asked later runs again; the first answer stays the summary.
	 */
	static function summaryRun($thread, $replies) {
		foreach ($replies as $reply) {
			if ($reply->getRunId() !== "") {
				return $reply->getRunId();
			}
		}

		return $thread === null ? "" : $thread->getRunId();
	}

	/** isSummary reports whether a report is a run's summary and not an answer. */
	static function isSummary($report) {
		if ($report["thread"] === "") {
			return true;
		}

		return $report["question"] === self::SUMMARY || $report["question"] === self::BATCH;
	}

	/** href is the report page of the run. */
	static function href($report) {
		return "/runs/" . $report["run"] . "/report";
	}
}
