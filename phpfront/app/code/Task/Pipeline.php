<?php

namespace Task;

// Pipeline is a workflow run as its pages draw it: the links of a chain as the
// nodes of the list, and the run itself as the thread of what it was asked and
// what came back.
//
// A node is shaped here and not in the template because it says which step it
// is by its number, and because a node that ran a workflow holds the nodes of
// that run under it: MiniTPL's {include} is a verbatim replace that cannot
// nest into itself, so the depth is unrolled into two components and the
// arithmetic is done once, here.
class Pipeline {
	/** WORKING is the reply of a run still going, which has nothing to say yet. */
	const WORKING = "Works through the steps; ";

	/** nodes are the links of a chain as the list draws them, in order. */
	static function nodes($links) {
		$out = array();

		foreach ($links as $i => $link) {
			$out[] = self::node($link, $i);
		}

		return $out;
	}

	/**
	 * node is one step of a pipeline: where its name links, how it ended, the
	 * screen its run ended on and the chain of the workflow run it made.
	 *
	 * small is the line under the name: the state for a step that ran, and
	 * which step it is for one that has not, a step that is off saying that
	 * runs skip it.
	 *
	 * expand is the button that opens the whole output of a run whose screen
	 * does not hold it, which a run still going never shows: its output may
	 * still scroll, so there is no count to offer.
	 *
	 * then says a step has one before it, which the chain of a table cell draws
	 * the arrow between its links with. The list of a pipeline ignores it.
	 */
	static function node($link, $i) {
		$state = $link["state"];
		$small = $state;

		if ($state === "") {
			$small = "step " . ($i + 1);

			if ($link["disabled"]) {
				$small .= ", off: runs skip it";
			}
		}

		return array(
			"pos" => $i,
			"then" => $i > 0,
			"state" => $state,
			"disabled" => $link["disabled"],
			"href" => $link["href"],
			"runId" => $link["run_id"],
			"workflow" => $link["workflow"],
			"summary" => $link["summary"],
			"check" => $link["check"],
			"prompt" => $link["prompt"],
			"text" => $link["text"],
			"small" => $small,
			"expand" => $link["run_id"] !== "" && !$link["complete"],
			"lines" => $link["lines"],
			"frame" => $link["frame"],
			"live" => $state === Run::RUNNING,
			"subId" => $link["workflow_run_id"],
			"workflowId" => $link["workflow_id"],
			"sub" => self::nodes($link["sub"]),
		);
	}

	/**
	 * thread is a workflow run as a thread: what it was given, or else the
	 * prompt of its first prompt step, under whoever started it, and the reply
	 * under Claude.
	 *
	 * A run of commands alone is answered by the workflow itself and not by an
	 * agent, so the reply is how the run ended and the author is the workflow's
	 * name.
	 *
	 * While the run goes the reply says it works, and task-ui.js puts the
	 * thread in again off the page once the workflow's topic says it ended,
	 * which data-follow and data-page name.
	 */
	static function thread($chrome, $run, $steps) {
		$id = $run->getId();
		$flow = $run->getWorkflowId();
		$state = $run->getState();
		$answer = $run->getAnswer();
		$message = $run->getMessage();
		$running = !Run::ended($state);
		$agent = self::agent($steps);
		$question = self::question($run, $steps);

		return array(
			"id" => $id,
			"workflowId" => $flow,
			"page" => "/workflows/" . $flow . "/runs/" . $id,
			"running" => $running,
			"noteState" => self::noteState($state),
			"name" => $run->getName(),
			"user" => $chrome->who()->of($run->getUserName()),
			"ago" => Clock::ago($run->getStartedAt()),
			"stamp" => Clock::read($run->getStartedAt()),
			"doneAgo" => Clock::ago($run->getFinishedAt()),
			"doneStamp" => Clock::read($run->getFinishedAt()),
			"state" => $state,
			"agent" => $agent === "" ? null : $chrome->who()->of($agent),
			"question" => $question === "" ? "" : $chrome->message($question),
			"answer" => $answer === "" ? "" : $chrome->message($answer),
			"waiting" => $agent === "" ? self::WORKING . "how the run ended shows here." : self::WORKING . "Claude replies here when the run ends.",
			"ended" => self::ended($run, $state, $message, $agent),
		);
	}

	/**
	 * noteState is the run's state as its thread shows it, a note's: working
	 * while it runs, done once it ended ok, else failed.
	 */
	static function noteState($state) {
		if (!Run::ended($state)) {
			return Note::WORKING;
		}

		return $state === Run::OK ? Note::DONE : Note::FAILED;
	}

	/**
	 * agent is the name Claude's reply to the run carries, "" for a run of
	 * commands alone, whose reply is how it ended.
	 */
	static function agent($steps) {
		foreach ($steps as $step) {
			if (Step::asks($step)) {
				return Note::CLAUDE;
			}
		}

		return "";
	}

	/**
	 * question is what the thread asks: the run's input, and for a run started
	 * without any the prompt of its first prompt step.
	 *
	 * A summary step is no question: what it asks is task-ui's own words about
	 * the steps before it, which no reader wrote.
	 */
	static function question($run, $steps) {
		$input = trim($run->getInput());

		if ($input !== "") {
			return $input;
		}

		foreach ($steps as $step) {
			if (!Step::asks($step) || Step::kind($step) === Step::SUMMARY) {
				continue;
			}

			$asked = trim($step["prompt"]);

			if ($asked !== "") {
				return $asked;
			}
		}

		return "";
	}

	/** ended is the reply of a run that ended with no answer: how it went. */
	static function ended($run, $state, $message, $agent) {
		$out = $agent === "" ? "The run" : "No answer: the run";
		$out .= " ended " . $state;

		if ($message !== "") {
			$out .= ", " . $message;
		}

		if ($run->getFinishedAt() !== "") {
			$out .= ", in " . Run::duration(Run::elapsed($run));
		}

		return $out . ".";
	}
}
