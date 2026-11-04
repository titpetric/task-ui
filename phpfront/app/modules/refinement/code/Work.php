<?php

namespace Task\Refinement;

// Work is what a thread does through the host: the run a request starts, and
// the record a page reconciles against the artifact tree.
//
// Neither is a statement and neither is text, so neither belongs in storage/ or
// in Thread. Both take the store and write through it, because a run that
// started is a note recorded beside its comment and a reconciled thread is a
// row.
//
// Static, and over the store it is handed: the five write endpoints are five
// files and a mod_ class is outside the classmap, so what they share lives on a
// class that is inside it.
class Work {
	/**
	 * ACTION is the markdown that says what to do with a request, as a path
	 * in the front end's own tree.
	 *
	 * The file ships beside the module it belongs to: a thread's correction and
	 * the instructions that apply it are one directory. The host resolves it
	 * inside the tree and refuses a path that leaves it, which is why this is a
	 * path and not a file the PHP read.
	 */
	const ACTION = "modules/refinement/action/refinement.md";

	/** REPORT is the file a run writes its report into, at the top of the thread's folder. */
	const REPORT = "README.md";

	/**
	 * start runs the request a comment holds, and records the note that holds
	 * the thread beside the comment.
	 *
	 * The folder is the thread, which is the folder of the artifact tree the
	 * run works in: a thread gets its own and nothing else writes there, so a
	 * new thread opens on an empty directory rather than on whatever every
	 * thread before it left.
	 *
	 * The run is the reader's. Starting one takes prompt/edit, which the board
	 * enforces and every endpoint here checks first, so a viewer is refused the
	 * page rather than the run.
	 */
	static function start($store, $thread, $comment, $body) {
		$note = \Task\Host::dispatch($body, self::ACTION, $thread);

		$store->note($comment, $note);

		return $note;
	}

	/**
	 * sync brings a thread's record in step with the artifact tree, and answers
	 * the thread as it stands after.
	 *
	 * The tree is the truth: a run writes the files and rewrites the report, and
	 * neither goes through this module. The tables are the record the thread
	 * keeps, which is what the page reads and what the board lists, so a page
	 * load reconciles them.
	 *
	 * The files are not reconciled. They are a folder of the artifact tree and
	 * the instance's own artifacts view browses one; this module links that
	 * folder and keeps no second listing of it, which is a listing that goes
	 * stale the moment a run renames something.
	 */
	static function sync($store, $thread) {
		$folder = $thread["id"];
		$report = \Task\Host::artifact($folder . "/" . self::REPORT);
		$changed = array();

		// A tree with no report has nothing to say, which is not the same as a
		// report that is empty: a run that has not written one yet must not
		// erase the one the run before it did.
		if ($report !== "" && $report !== $thread["report_markdown"]) {
			$changed["report_markdown"] = $report;
			$thread["report_markdown"] = $report;
		}

		// The report's heading names the thread, so any run that writes one
		// fills the title and any correction that rewrites it renames the
		// thread. A thread opened before this reconciles on its next page load,
		// which is the whole of the backfill.
		$title = Thread::reportTitle($thread["report_markdown"]);

		if ($title !== "" && $title !== $thread["title"]) {
			$changed["title"] = $title;
			$thread["title"] = $title;
		}

		if ($changed) {
			$store->touch($folder, $changed);
		}

		return $thread;
	}
}
