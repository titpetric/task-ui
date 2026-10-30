<?php

// The module's repository layer. Every SQL statement the module runs is
// in this file and nowhere else, the way a Go package keeps its queries
// in storage/ and nothing outside it issues one.
//
// The convention, which docs/modules.md states for every module:
//
//   - One class per module, named for it, holding a Database.
//   - A read is get<Thing>() and answers rows or a value. A read that
//     needs more than one statement composes them here rather than in
//     the page.
//   - A write is a CRUD verb and the table it writes:
//     createRefinementThread, updateRefinementThread. It takes one
//     array whose keys are the table's columns, so the call site reads
//     as the row it writes and a new column is a new key.
//   - Nothing here renders, routes or dispatches. The page decides what
//     to ask for; this answers.
class Refinement {
	var $db;

	function __construct($db) {
		$this->db = $db;
	}

	// ---------------------------------------------------------- reads

	/** getThreads is every thread, newest first, with its comment count. */
	function getThreads() {
		return $this->db->get_all("SELECT t.*, (SELECT COUNT(*) FROM refinement_comment c WHERE c.thread_id = t.id) AS comments FROM refinement_thread t ORDER BY t.created_at DESC, t.id DESC");
	}

	/**
	 * getThread is one thread, or false where there is none: get()
	 * answers a missing row with false, the way PHP does, and not null,
	 * so a caller tests it with is_array().
	 */
	function getThread($id) {
		return $this->db->get("SELECT * FROM refinement_thread WHERE id = ?", $id);
	}

	/**
	 * getComments is a thread's comments oldest first, each with whether
	 * it is a correction, which is what the page marks them with.
	 */
	function getComments($thread) {
		return $this->db->get_all("SELECT c.*, (SELECT COUNT(*) FROM refinement_correction r WHERE r.comment_id = c.id) AS correction FROM refinement_comment c WHERE c.thread_id = ? ORDER BY c.created_at, c.id", $thread);
	}

	/**
	 * getLatestCommentID is the identifier of the comment a thread ends
	 * on, which a correction names and a new artifact is counted
	 * against, or "" for a thread with none.
	 */
	function getLatestCommentID($thread) {
		$row = $this->db->get("SELECT id FROM refinement_comment WHERE thread_id = ? ORDER BY created_at DESC, id DESC LIMIT 1", $thread);

		return is_array($row) ? $row["id"] : "";
	}

	// --------------------------------------------------------- writes

	/** createRefinementThread writes a thread. */
	function createRefinementThread($data) {
		return $this->db->insert("refinement_thread", $data);
	}

	/** createRefinementComment writes a comment. */
	function createRefinementComment($data) {
		return $this->db->insert("refinement_comment", $data);
	}

	/** createRefinementCorrection writes a correction. */
	function createRefinementCorrection($data) {
		return $this->db->insert("refinement_correction", $data);
	}

	/** updateRefinementThread writes a thread back, keyed by its id. */
	function updateRefinementThread($data) {
		return $this->db->update("refinement_thread", $data, "id");
	}

	/** updateRefinementComment writes a comment back, keyed by its id. */
	function updateRefinementComment($data) {
		return $this->db->update("refinement_comment", $data, "id");
	}
}
