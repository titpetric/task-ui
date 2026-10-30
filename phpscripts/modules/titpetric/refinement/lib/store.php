<?php

// The module's own tables, and the run a comment starts.
//
// Every identifier is a ULID, which ulid() answers, and every table is
// task-ui's in task-ui's database: the module declared the SQL under
// schema/ and migrate.php applied it at startup.

/** refinement_now is the stamp the tables keep, as the database reads it. */
function refinement_now() {
	return date("2006-01-02 15:04:05");
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
 * refinement_start opens a thread from its first comment.
 *
 * The first comment is no correction: it has nothing to correct. Every
 * comment after it is one, which is what separates "design this" from
 * "use ULID as primary keys".
 */
function refinement_start($db, $user, $want) {
	$now = refinement_now();
	$thread = ulid();
	$comment = ulid();

	$db->insert("refinement_thread", array(
		"id" => $thread,
		"title" => refinement_title($want),
		"user" => $user,
		"report_markdown" => "",
		"created_at" => $now,
		"updated_at" => $now,
	));

	$db->insert("refinement_comment", array(
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
function refinement_correct($db, $thread, $user, $body) {
	$now = refinement_now();
	$comment = ulid();
	$before = $db->get("SELECT id FROM refinement_comment WHERE thread_id = ? ORDER BY created_at DESC, id DESC LIMIT 1", $thread);

	$db->insert("refinement_comment", array(
		"id" => $comment,
		"thread_id" => $thread,
		"user" => $user,
		"body" => $body,
		"note_id" => "",
		"created_at" => $now,
	));

	$db->insert("refinement_correction", array(
		"id" => ulid(),
		"thread_id" => $thread,
		"comment_id" => $comment,
		"corrects_id" => is_array($before) ? $before["id"] : "",
		"created_at" => $now,
	));

	$db->query("UPDATE refinement_thread SET updated_at = ? WHERE id = ?", $now, $thread);

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
function refinement_dispatch($db, $thread, $comment, $body) {
	$note = Task::dispatch($body, "action/refinement.md");

	$db->query("UPDATE refinement_comment SET note_id = ? WHERE id = ?", $note, $comment);

	return $note;
}

/** refinement_threads is every thread, newest first, with its comment count. */
function refinement_threads($db) {
	return $db->get_all("SELECT t.*, (SELECT COUNT(*) FROM refinement_comment c WHERE c.thread_id = t.id) AS comments" . " FROM refinement_thread t ORDER BY t.created_at DESC, t.id DESC");
}

/** refinement_thread is one thread, or false where there is none: get()
 * answers a missing row with false, the way PHP does, and not null. */
function refinement_thread($db, $id) {
	return $db->get("SELECT * FROM refinement_thread WHERE id = ?", $id);
}

/**
 * refinement_comments is a thread's comments oldest first, each with
 * whether it is a correction, which is what the page marks them with.
 */
function refinement_comments($db, $thread) {
	return $db->get_all("SELECT c.*, (SELECT COUNT(*) FROM refinement_correction r WHERE r.comment_id = c.id) AS correction" . " FROM refinement_comment c WHERE c.thread_id = ? ORDER BY c.created_at, c.id", $thread);
}

/** refinement_artifacts is what a thread's runs left, by path. */
function refinement_artifacts($db, $thread) {
	return $db->get_all("SELECT * FROM refinement_artifact WHERE thread_id = ? ORDER BY path", $thread);
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
function refinement_sync($db, $thread) {
	$report = Task::artifact("README.md");

	// A tree with no README has nothing to say, which is not the same as
	// a report that is empty: a run that has not written one yet must
	// not erase the one the run before it did.
	if ($report !== "" && $report !== $thread["report_markdown"]) {
		$db->query("UPDATE refinement_thread SET report_markdown = ?, updated_at = ? WHERE id = ?", $report, refinement_now(), $thread["id"]);
		$thread["report_markdown"] = $report;
	}

	$known = array();
	foreach (refinement_artifacts($db, $thread["id"]) as $row) {
		$known[$row["path"]] = true;
	}

	$last = $db->get("SELECT id FROM refinement_comment WHERE thread_id = ? ORDER BY created_at DESC, id DESC LIMIT 1", $thread["id"]);
	$comment = is_array($last) ? $last["id"] : "";

	foreach (Task::artifacts() as $path) {
		if ($path === "README.md" || isset($known[$path])) {
			continue;
		}

		$db->insert("refinement_artifact", array(
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
