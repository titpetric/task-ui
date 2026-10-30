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

/** refinement_title is the first line of a comment, for the thread list. */
function refinement_title($body) {
	$lines = explode("\n", trim($body));
	$line = trim($lines[0]);

	if (strlen($line) > 120) {
		return substr($line, 0, 117) . "...";
	}

	return $line === "" ? "a thread" : $line;
}

/**
 * refinement_kind says whether the page embeds an artifact or links it:
 * a diagram is read on the page, a SQL file is a link.
 */
function refinement_kind($path) {
	$dot = strrpos($path, ".");
	$ext = $dot === false ? "" : strtolower(substr($path, $dot));

	if (in_array($ext, array(".svg", ".png", ".jpg", ".jpeg", ".gif", ".webp"))) {
		return "image";
	}

	return "file";
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
 */
function refinement_dispatch($store, $comment, $body) {
	$note = Task::dispatch($body, "action/refinement.md");

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
 * The artifacts belong to the instance and not to the thread. A thread
 * lists what the tree holds now, which is what its runs left and what an
 * earlier thread left before them.
 */
function refinement_sync($store, $thread) {
	$report = Task::artifact("README.md");

	// A tree with no README has nothing to say, which is not the same as
	// a report that is empty: a run that has not written one yet must
	// not erase the one the run before it did.
	if ($report !== "" && $report !== $thread["report_markdown"]) {
		$store->updateRefinementThread(array(
			"id" => $thread["id"],
			"report_markdown" => $report,
			"updated_at" => refinement_now(),
		));
		$thread["report_markdown"] = $report;
	}

	$known = $store->getArtifactPaths($thread["id"]);
	$comment = $store->getLatestCommentID($thread["id"]);

	foreach (Task::artifacts() as $path) {
		if ($path === "README.md" || isset($known[$path])) {
			continue;
		}

		$store->createRefinementArtifact(array(
			"id" => ulid(),
			"thread_id" => $thread["id"],
			"comment_id" => $comment,
			"path" => $path,
			"kind" => refinement_kind($path),
			"created_at" => refinement_now(),
		));
	}

	return $thread;
}
