<?php

namespace Task\Refinement;

// Threads is the refinement module's own data model: the threads of the board,
// the comments of one, and the corrections that make a comment a rule.
//
// The tables are the module's and no other page reads them, so every statement
// the module runs lives here. A read answers rows; a write is a verb and the
// thing it writes, taking what the caller already composed, so the text of a
// title or of a flattened request is decided in code/ and this only stores it.
//
// There is no transaction. The instance is one sqlite connection and holding a
// transaction open holds it with them, so a write that spans two tables is two
// statements in the order a half-written thread survives: the row a reader
// needs first, last.
class Threads {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	// ---------------------------------------------------------------- reads

	/** all is every thread, newest first, each with how many comments it has. */
	function all() {
		return $this->db->get_all("SELECT t.*, (SELECT COUNT(*) FROM refinement_comment c WHERE c.thread_id = t.id) AS comments FROM refinement_thread t ORDER BY t.created_at DESC, t.id DESC");
	}

	/**
	 * find is one thread, or false where there is none: get() answers a
	 * missing row with false, the way PHP does, so a caller tests it with
	 * is_array().
	 */
	function find($id) {
		return $this->db->get("SELECT * FROM refinement_thread WHERE id = ?", $id);
	}

	/**
	 * comments is a thread's comments oldest first, each with whether it is a
	 * correction, which is what the page marks them with.
	 */
	function comments($thread) {
		return $this->db->get_all("SELECT c.*, (SELECT COUNT(*) FROM refinement_correction r WHERE r.comment_id = c.id) AS correction FROM refinement_comment c WHERE c.thread_id = ? ORDER BY c.created_at, c.id", $thread);
	}

	/**
	 * latest is the identifier of the comment a thread ends on, which a
	 * correction names, or "" for a thread with none.
	 */
	function latest($thread) {
		$row = $this->db->get("SELECT id FROM refinement_comment WHERE thread_id = ? ORDER BY created_at DESC, id DESC LIMIT 1", $thread);

		return is_array($row) ? $row["id"] : "";
	}

	/**
	 * design is the comment a thread opened on, which says what the artifact
	 * is, or false for a thread whose design was deleted.
	 *
	 * It is the oldest comment that corrects nothing and records no Apply all:
	 * a correction has a row in refinement_correction and an Apply all carries
	 * kind 'apply'.
	 */
	function design($thread) {
		return $this->db->get("SELECT c.* FROM refinement_comment c WHERE c.thread_id = ? AND c.kind <> 'apply' AND NOT EXISTS (SELECT 1 FROM refinement_correction r WHERE r.comment_id = c.id) ORDER BY c.created_at, c.id LIMIT 1", $thread);
	}

	/**
	 * rules is every correction of a thread, oldest first, folded ones
	 * included: a fold is what the page draws and deleting is what retires a
	 * rule, so what was asked for is still asked for.
	 */
	function rules($thread) {
		return $this->db->get_all("SELECT c.* FROM refinement_comment c WHERE c.thread_id = ? AND EXISTS (SELECT 1 FROM refinement_correction r WHERE r.comment_id = c.id) ORDER BY c.created_at, c.id", $thread);
	}

	// --------------------------------------------------------------- writes

	/**
	 * start opens a thread from its first comment, and answers both
	 * identifiers: the thread to redirect to and the comment to dispatch.
	 *
	 * The first comment is no correction: it has nothing to correct. Every
	 * comment after it is one, which is what separates "design this" from "use
	 * ULID as primary keys".
	 */
	function start($user, $want) {
		$now = \Task\Clock::stamp();
		$thread = \Task\Id::make();
		$comment = \Task\Id::make();

		$this->db->insert("refinement_thread", array(
			"id" => $thread,
			"title" => Thread::title($want),
			"user" => $user,
			"report_markdown" => "",
			"created_at" => $now,
			"updated_at" => $now,
		));

		$this->db->insert("refinement_comment", array(
			"id" => $comment,
			"thread_id" => $thread,
			"user" => $user,
			"body" => $want,
			"note_id" => "",
			"kind" => "",
			"collapsed_into" => "",
			"created_at" => $now,
		));

		return array("id" => $thread, "comment_id" => $comment);
	}

	/**
	 * correct adds a comment that changes what the run before it produced,
	 * records the correction against that comment, and answers the comment.
	 */
	function correct($thread, $user, $body) {
		$now = \Task\Clock::stamp();
		$comment = \Task\Id::make();
		$before = $this->latest($thread);

		$this->db->insert("refinement_comment", array(
			"id" => $comment,
			"thread_id" => $thread,
			"user" => $user,
			"body" => $body,
			"note_id" => "",
			"kind" => "",
			"collapsed_into" => "",
			"created_at" => $now,
		));

		$this->db->insert("refinement_correction", array(
			"id" => \Task\Id::make(),
			"thread_id" => $thread,
			"comment_id" => $comment,
			"corrects_id" => $before,
			"created_at" => $now,
		));

		$this->db->update("refinement_thread", array("id" => $thread, "updated_at" => $now), "id");

		return $comment;
	}

	/**
	 * apply records the flattened request the caller composed, folds every
	 * comment standing on its own into it, and answers the comment.
	 *
	 * The comment it writes carries kind 'apply' and no correction row, so it
	 * is neither the design nor a rule and the next Apply all composes from
	 * the same comments this one did. The rules themselves survive, because a
	 * fold is not a deletion.
	 */
	function apply($thread, $user, $body) {
		$now = \Task\Clock::stamp();
		$comment = \Task\Id::make();

		$this->db->insert("refinement_comment", array(
			"id" => $comment,
			"thread_id" => $thread,
			"user" => $user,
			"body" => $body,
			"note_id" => "",
			"kind" => "apply",
			"collapsed_into" => "",
			"created_at" => $now,
		));

		$this->db->query("UPDATE refinement_comment SET collapsed_into = ? WHERE thread_id = ? AND collapsed_into = '' AND id <> ?", $comment, $thread, $comment);
		$this->db->update("refinement_thread", array("id" => $thread, "updated_at" => $now), "id");

		return $comment;
	}

	/**
	 * carry copies the rules of one thread into another as rules of its own,
	 * each correcting the comment before it, which is what a fork is.
	 *
	 * The subject does not travel and the artifacts do not either: a fork
	 * starts on an empty folder with how the work is done already in it.
	 */
	function carry($from, $into, $after) {
		$before = $after;

		foreach ($this->rules($from) as $rule) {
			$now = \Task\Clock::stamp();
			$comment = \Task\Id::make();

			$this->db->insert("refinement_comment", array(
				"id" => $comment,
				"thread_id" => $into,
				"user" => $rule["user"],
				"body" => $rule["body"],
				"note_id" => "",
				"kind" => "",
				"collapsed_into" => "",
				"created_at" => $now,
			));

			$this->db->insert("refinement_correction", array(
				"id" => \Task\Id::make(),
				"thread_id" => $into,
				"comment_id" => $comment,
				"corrects_id" => $before,
				"created_at" => $now,
			));

			$before = $comment;
		}
	}

	/** note records the run a dispatch opened, beside the comment it answers. */
	function note($comment, $note) {
		return $this->db->update("refinement_comment", array("id" => $comment, "note_id" => $note), "id");
	}

	/**
	 * touch writes a thread back with what reconciling it against the artifact
	 * tree changed, and stamps it.
	 */
	function touch($thread, $changed) {
		$changed["id"] = $thread;
		$changed["updated_at"] = \Task\Clock::stamp();

		return $this->db->update("refinement_thread", $changed, "id");
	}

	/**
	 * remove retires one rule of one thread: the comment and the correction row
	 * that made it one.
	 *
	 * The thread is named as well as the comment, and both statements are
	 * scoped to it: the identifier arrives from the path of a form, and a
	 * comment of a thread the form was not drawn on is not that form's to
	 * delete.
	 *
	 * The artifacts it produced stay where they are. A rule is a thing the
	 * thread asks for, and dropping the ask is what this is; what the next run
	 * makes of the folder is the next run's business.
	 */
	function remove($thread, $comment) {
		$this->db->query("DELETE FROM refinement_correction WHERE thread_id = ? AND comment_id = ?", $thread, $comment);
		$this->db->query("DELETE FROM refinement_comment WHERE thread_id = ? AND id = ?", $thread, $comment);
	}

	/**
	 * drop removes a thread: its corrections, its comments and the thread row.
	 *
	 * Deleting the design is deleting the thread, because the rules are
	 * corrections to a design and without it they correct nothing.
	 *
	 * The folder in the artifact tree stays. This module reads that tree and
	 * does not write it, and what a run produced is the instance's to keep or
	 * to clear from the artifacts page.
	 */
	function drop($thread) {
		$this->db->query("DELETE FROM refinement_correction WHERE thread_id = ?", $thread);
		$this->db->query("DELETE FROM refinement_comment WHERE thread_id = ?", $thread);
		$this->db->query("DELETE FROM refinement_thread WHERE id = ?", $thread);
	}
}
