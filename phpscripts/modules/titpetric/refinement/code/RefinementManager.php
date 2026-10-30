<?php

// What a thread does, over the store in RefinementStore.
//
// No SQL is written here. Every statement the module runs is in the
// store, and this decides what to ask it for: how a thread opens, what a
// correction is counted against, what an Apply all folds, and what a
// page load reconciles against the artifact tree.
//
// The shape is phpscript's own: a Manager over a store, made once per
// request and asked for things. The module's pages hold no logic of
// their own and nothing here is a free function.
class RefinementManager {
	var $store;

	function __construct($store) {
		$this->store = $store;
	}

	/**
	 * now is the stamp the tables keep.
	 *
	 * DateTime::now()->format() and not date(): date() in this runtime
	 * answers with the layout it was handed, so every row written
	 * through it carried the literal "2006-01-02 15:04:05" and nothing
	 * could be ordered by when it happened.
	 *
	 * To the microsecond, because a thread's comments are ordered by
	 * this and two of them land in one second. ulid() is not monotonic
	 * inside a millisecond, so the identifier is no tiebreak: a
	 * second-resolution stamp let three corrections render in any order.
	 */
	static function now() {
		return DateTime::now()->format("2006-01-02 15:04:05.000000");
	}

	/** cut is one line as the thread list shows it, 120 at most. */
	static function cut($line) {
		$line = trim($line);

		if (strlen($line) > 120) {
			return substr($line, 0, 117) . "...";
		}

		return $line;
	}

	/** title is the first line of a comment, for the thread list. */
	static function title($body) {
		$lines = explode("\n", trim($body));
		$line = RefinementManager::cut($lines[0]);

		return $line === "" ? "a thread" : $line;
	}

	/**
	 * reportTitle is the first heading of the report, which is what the
	 * thread is called once a run has written one.
	 *
	 * A thread opens named after the line the person typed, because that
	 * is all there is at the time. The runs then go and find out what the
	 * thing actually is, and README.md leads with it: the report's own
	 * heading is a better name than the request that started it, and a
	 * correction that changes the heading renames the thread with it.
	 *
	 * ATX and setext both, and the first one anywhere in the document, so
	 * a report opening with a badge or a comment still answers.
	 *
	 * Empty when the report has no heading at all. That is not a title of
	 * "", it is the absence of one: the caller keeps what the thread had.
	 */
	static function reportTitle($markdown) {
		$lines = explode("\n", $markdown);
		$count = count($lines);

		for ($i = 0; $i < $count; $i++) {
			$line = trim($lines[$i]);

			if ($line === "") {
				continue;
			}

			// # The heading, with any hashes the writer balanced it with.
			if (substr($line, 0, 2) === "# ") {
				return RefinementManager::cut(trim(substr($line, 2), " #"));
			}

			// The heading, with a line of = under it.
			if ($i + 1 < $count && preg_match('/^=+$/', trim($lines[$i + 1]))) {
				return RefinementManager::cut($line);
			}
		}

		return "";
	}

	/**
	 * designID is the comment a thread opened on, out of the rows the
	 * page already has, or "" for a thread with none left.
	 *
	 * The same rule the store uses: the oldest comment that corrects
	 * nothing and records no Apply all. The page needs it because the
	 * delete button on the design has to say that it takes the thread.
	 */
	static function designID($comments) {
		foreach ($comments as $comment) {
			if ($comment["kind"] !== "apply" && !$comment["correction"]) {
				return $comment["id"];
			}
		}

		return "";
	}

	/**
	 * entries groups a thread's comments the way the page draws them: a
	 * comment that stands on its own, or an Apply all with the ones it
	 * folded under it.
	 *
	 * Folded comments are not dropped from the page. They are the record
	 * of what was asked and of the run that answered each one, and the
	 * page keeps them behind a disclosure rather than behind nothing.
	 */
	static function entries($comments) {
		$folded = array();

		foreach ($comments as $comment) {
			$into = $comment["collapsed_into"];

			if ($into !== "") {
				if (!isset($folded[$into])) {
					$folded[$into] = array();
				}

				$folded[$into][] = $comment;
			}
		}

		$entries = array();

		foreach ($comments as $comment) {
			if ($comment["collapsed_into"] !== "") {
				continue;
			}

			$under = isset($folded[$comment["id"]]) ? $folded[$comment["id"]] : array();

			$entries[] = array(
				"comment" => $comment,
				"folded" => $under,
				"folded_count" => count($under),
				"applied" => $comment["kind"] === "apply",
			);
		}

		return $entries;
	}

	/**
	 * start opens a thread from its first comment.
	 *
	 * The first comment is no correction: it has nothing to correct.
	 * Every comment after it is one, which is what separates "design
	 * this" from "use ULID as primary keys".
	 */
	function start($user, $want) {
		$now = RefinementManager::now();
		$thread = ulid();
		$comment = ulid();

		$this->store->createRefinementThread(array(
			"id" => $thread,
			"title" => RefinementManager::title($want),
			"user" => $user,
			"report_markdown" => "",
			"created_at" => $now,
			"updated_at" => $now,
		));

		$this->store->createRefinementComment(array(
			"id" => $comment,
			"thread_id" => $thread,
			"user" => $user,
			"body" => $want,
			"note_id" => "",
			"created_at" => $now,
		));

		return array("id" => $thread, "comment_id" => $comment);
	}

	/**
	 * correct adds a comment that changes what the run before it
	 * produced, and records the correction against that comment.
	 */
	function correct($thread, $user, $body) {
		$now = RefinementManager::now();
		$comment = ulid();
		$before = $this->store->getLatestCommentID($thread);

		$this->store->createRefinementComment(array(
			"id" => $comment,
			"thread_id" => $thread,
			"user" => $user,
			"body" => $body,
			"note_id" => "",
			"created_at" => $now,
		));

		$this->store->createRefinementCorrection(array(
			"id" => ulid(),
			"thread_id" => $thread,
			"comment_id" => $comment,
			"corrects_id" => $before,
			"created_at" => $now,
		));

		$this->store->updateRefinementThread(array("id" => $thread, "updated_at" => $now));

		return $comment;
	}

	/**
	 * apply folds a thread's rules into one request and answers with the
	 * comment that holds it, for the caller to dispatch.
	 *
	 * The comment it writes carries kind 'apply' and no correction row,
	 * so it is neither the design nor a rule and the next Apply all
	 * composes from the same comments this one did. Everything standing
	 * before it is folded into it, which is what the page draws; the
	 * rules themselves survive, because a fold is not a deletion.
	 *
	 * A thread whose design was deleted has nothing to apply rules to,
	 * and that is the caller's to refuse: this answers an empty array
	 * and writes nothing.
	 */
	function apply($thread, $user) {
		$design = $this->store->getDesignComment($thread);

		if (!is_array($design)) {
			return array();
		}

		$rules = RefinementPrompt::rules($this->store->getRuleComments($thread));
		$body = RefinementPrompt::flatten($design["body"], $rules);
		$now = RefinementManager::now();
		$comment = ulid();

		$this->store->createRefinementComment(array(
			"id" => $comment,
			"thread_id" => $thread,
			"user" => $user,
			"body" => $body,
			"note_id" => "",
			"kind" => "apply",
			"collapsed_into" => "",
			"created_at" => $now,
		));

		$this->store->collapseComments($thread, $comment);
		$this->store->updateRefinementThread(array("id" => $thread, "updated_at" => $now));

		return array("id" => $comment, "body" => $body);
	}

	/**
	 * dispatch starts the Claude run that answers a comment, and records
	 * the note it opened beside the comment.
	 *
	 * action/refinement.md says what to do with the request; task-ui puts
	 * that file, the request and the run's own environment into the
	 * prompt. Which module is asking is not an argument: the call reads
	 * it from the context task-ui put on the runtime before the request.
	 *
	 * The thread is the third argument and it is the folder of the
	 * artifact tree the run works in. A thread gets its own and nothing
	 * else writes there, so a new thread opens on an empty directory
	 * rather than on whatever every thread before it left.
	 */
	function dispatch($thread, $comment, $body) {
		$note = Task::dispatch($body, "action/refinement.md", $thread);

		$this->store->updateRefinementComment(array("id" => $comment, "note_id" => $note));

		return $note;
	}

	/**
	 * remove retires one rule: the comment and the correction row that
	 * made it one.
	 *
	 * The artifacts it produced stay where they are. A rule is a thing
	 * the thread asks for, and dropping the ask is what this is; what the
	 * next run makes of the folder is the next run's business.
	 */
	function remove($comment) {
		$this->store->deleteRefinementCorrection($comment);
		$this->store->deleteRefinementComment($comment);
	}

	/**
	 * drop removes a thread: its corrections, its comments and the thread
	 * row.
	 *
	 * Deleting the design is deleting the thread, because the rules are
	 * corrections to a design and without it they correct nothing.
	 *
	 * The folder in the artifact tree stays. This module reads that tree
	 * and does not write it, and what a run produced is the instance's to
	 * keep or to clear from the artifacts page.
	 */
	function drop($thread) {
		$this->store->deleteThreadCorrections($thread);
		$this->store->deleteThreadComments($thread);
		$this->store->deleteRefinementThread($thread);
	}

	/**
	 * sync brings a thread's record in step with the artifact tree, and
	 * answers the thread as it stands after.
	 *
	 * The tree is the truth: a run writes the files and rewrites
	 * README.md, and neither goes through this module. The tables are the
	 * record the thread keeps, which is what the page reads and what a
	 * later correction is counted against, so a page load reconciles
	 * them.
	 *
	 * The files are not reconciled. They are a folder of the artifact
	 * tree and the instance's own artifacts view browses one; this module
	 * links that folder and keeps no second listing of it, which is a
	 * listing that goes stale the moment a run renames something.
	 */
	function sync($thread) {
		$folder = $thread["id"];
		$report = Task::artifact($folder . "/README.md");
		$changed = array();

		// A tree with no README has nothing to say, which is not the same
		// as a report that is empty: a run that has not written one yet
		// must not erase the one the run before it did.
		if ($report !== "" && $report !== $thread["report_markdown"]) {
			$changed["report_markdown"] = $report;
			$thread["report_markdown"] = $report;
		}

		// The report's heading names the thread, so any run that writes
		// one fills the title and any correction that rewrites it renames
		// the thread. A thread opened before this reconciles on its next
		// page load, which is the whole of the backfill.
		$title = RefinementManager::reportTitle($thread["report_markdown"]);

		if ($title !== "" && $title !== $thread["title"]) {
			$changed["title"] = $title;
			$thread["title"] = $title;
		}

		if ($changed) {
			$changed["id"] = $thread["id"];
			$changed["updated_at"] = RefinementManager::now();

			$this->store->updateRefinementThread($changed);
		}

		return $thread;
	}
}
