<?php

// What a thread does, over the repository in storage/Refinement.php.
//
// No SQL is written here. Every statement the module runs is in the
// storage class, and this decides what to ask it for: how a thread
// opens, what a correction is counted against, and what a page load
// reconciles against the artifact tree.

/**
 * refinement_now is the stamp the tables keep.
 *
 * DateTime::now()->format() and not date(): date() in this runtime
 * answers with the layout it was handed, so every row written through it
 * carried the literal "2006-01-02 15:04:05" and nothing could be ordered
 * by when it happened.
 *
 * To the microsecond, because a thread's comments are ordered by this
 * and two of them land in one second. ulid() is not monotonic inside a
 * millisecond, so the identifier is no tiebreak: a second-resolution
 * stamp let three corrections render in any order.
 */
function refinement_now() {
	return DateTime::now()->format("2006-01-02 15:04:05.000000");
}

/** refinement_cut is one line as the thread list shows it, 120 at most. */
function refinement_cut($line) {
	$line = trim($line);

	if (strlen($line) > 120) {
		return substr($line, 0, 117) . "...";
	}

	return $line;
}

/** refinement_title is the first line of a comment, for the thread list. */
function refinement_title($body) {
	$lines = explode("\n", trim($body));
	$line = refinement_cut($lines[0]);

	return $line === "" ? "a thread" : $line;
}

/**
 * refinement_report_title is the first heading of the report, which is
 * what the thread is called once a run has written one.
 *
 * A thread opens named after the line the person typed, because that is
 * all there is at the time. The runs then go and find out what the thing
 * actually is, and README.md leads with it: the report's own heading is
 * a better name than the request that started it, and a correction that
 * changes the heading renames the thread with it.
 *
 * ATX and setext both, and the first one anywhere in the document, so a
 * report opening with a badge or a comment still answers.
 *
 * Empty when the report has no heading at all. That is not a title of
 * "", it is the absence of one: the caller keeps what the thread had.
 */
function refinement_report_title($markdown) {
	$lines = explode("\n", $markdown);
	$count = count($lines);

	for ($i = 0; $i < $count; $i++) {
		$line = trim($lines[$i]);

		if ($line === "") {
			continue;
		}

		// # The heading, with any hashes the writer balanced it with.
		if (substr($line, 0, 2) === "# ") {
			return refinement_cut(trim(substr($line, 2), " #"));
		}

		// The heading, with a line of = under it.
		if ($i + 1 < $count && preg_match('/^=+$/', trim($lines[$i + 1]))) {
			return refinement_cut($line);
		}
	}

	return "";
}

/**
 * refinement_start opens a thread from its first comment.
 *
 * The first comment is no correction: it has nothing to correct. Every
 * comment after it is one, which is what separates "design this" from
 * "use ULID as primary keys".
 */
function refinement_start($store, $user, $want) {
	$now = refinement_now();
	$thread = ulid();
	$comment = ulid();

	$store->createRefinementThread(array(
		"id" => $thread,
		"title" => refinement_title($want),
		"user" => $user,
		"report_markdown" => "",
		"created_at" => $now,
		"updated_at" => $now,
	));

	$store->createRefinementComment(array(
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
 * refinement_correct adds a comment that changes what the run before it
 * produced, and records the correction against that comment.
 */
function refinement_correct($store, $thread, $user, $body) {
	$now = refinement_now();
	$comment = ulid();
	$before = $store->getLatestCommentID($thread);

	$store->createRefinementComment(array(
		"id" => $comment,
		"thread_id" => $thread,
		"user" => $user,
		"body" => $body,
		"note_id" => "",
		"created_at" => $now,
	));

	$store->createRefinementCorrection(array(
		"id" => ulid(),
		"thread_id" => $thread,
		"comment_id" => $comment,
		"corrects_id" => $before,
		"created_at" => $now,
	));

	$store->updateRefinementThread(array("id" => $thread, "updated_at" => $now));

	return $comment;
}

/**
 * refinement_dispatch starts the Claude run that answers a comment, and
 * records the note it opened beside the comment.
 *
 * action/refinement.md says what to do with the comment; task-ui puts
 * that file, the comment and the run's own environment into the prompt.
 * Which module is asking is not an argument: the call reads it from the
 * context task-ui put on the runtime before the request.
 *
 * The thread is the third argument and it is the folder of the artifact
 * tree the run works in. A thread gets its own and nothing else writes
 * there, so a new thread opens on an empty directory rather than on
 * whatever every thread before it left.
 */
function refinement_dispatch($store, $thread, $comment, $body) {
	$note = Task::dispatch($body, "action/refinement.md", $thread);

	$store->updateRefinementComment(array("id" => $comment, "note_id" => $note));

	return $note;
}

/**
 * refinement_sync brings a thread's record in step with the artifact
 * tree, and answers the thread as it stands after.
 *
 * The tree is the truth: a run writes the files and rewrites README.md,
 * and neither goes through this module. The tables are the record the
 * thread keeps, which is what the page reads and what a later correction
 * is counted against, so a page load reconciles them.
 *
 * The files are not reconciled. They are a folder of the artifact tree
 * and the instance's own artifacts view browses one; this module links
 * that folder and keeps no second listing of it, which is a listing that
 * goes stale the moment a run renames something.
 */
function refinement_sync($store, $thread) {
	$folder = $thread["id"];
	$report = Task::artifact($folder . "/README.md");
	$changed = array();

	// A tree with no README has nothing to say, which is not the same as
	// a report that is empty: a run that has not written one yet must
	// not erase the one the run before it did.
	if ($report !== "" && $report !== $thread["report_markdown"]) {
		$changed["report_markdown"] = $report;
		$thread["report_markdown"] = $report;
	}

	// The report's heading names the thread, so any run that writes one
	// fills the title and any correction that rewrites it renames the
	// thread. A thread that was opened before this reconciles on its next
	// page load, which is the whole of the backfill.
	$title = refinement_report_title($thread["report_markdown"]);

	if ($title !== "" && $title !== $thread["title"]) {
		$changed["title"] = $title;
		$thread["title"] = $title;
	}

	if ($changed) {
		$changed["id"] = $thread["id"];
		$changed["updated_at"] = refinement_now();

		$store->updateRefinementThread($changed);
	}

	return $thread;
}
