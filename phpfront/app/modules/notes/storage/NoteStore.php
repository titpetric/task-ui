<?php

namespace Task\Notes;

// NoteStore reads and writes the note table: the notes that start a thread,
// newest first, and the replies under them.
//
// A note of a channel is a board note. A note with a workspace is a prompt,
// which belongs to the page the prompt was asked from and not to the board.
class NoteStore {
	/** NOTES is the channel the board's own notes carry. */
	const NOTES = "notes";

	/** LIMIT is how many threads the board lists. */
	const LIMIT = 100;

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** threads is the notes that start a thread of a channel, newest first. */
	function threads($channel, $limit = self::LIMIT) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND workspace_id = ''" . " ORDER BY rowid DESC LIMIT ?", $channel, $limit));
	}

	/** get is one note by ID, or null. */
	function get($id) {
		$row = $this->db->get("SELECT * FROM note WHERE id = ?", $id);

		return is_array($row) ? new \Task\Model\Note($row) : null;
	}

	/**
	 * replies is the replies to the threads of a channel, oldest first, by
	 * note.
	 *
	 * The threads are named by the same subquery threads() runs rather than by
	 * the ids it answered: argument unpacking is unavailable, so an IN list of
	 * a length only the caller knows cannot be passed. See
	 * docs/todo-correction.md.
	 */
	function replies($channel, $limit = self::LIMIT) {
		$out = array();
		$rows = $this->db->get_all("SELECT * FROM note WHERE parent_id IN (" . "SELECT id FROM note WHERE parent_id = '' AND channel = ? AND workspace_id = ''" . " ORDER BY rowid DESC LIMIT ?" . ") ORDER BY rowid", $channel, $limit);

		foreach (\Task\Model\Note::rows($rows) as $note) {
			$parent = $note->getParentId();

			if (!isset($out[$parent])) {
				$out[$parent] = array();
			}

			$out[$parent][] = $note;
		}

		return $out;
	}

	/** post writes a note, or a reply when the parent is named. */
	function post($user, $body, $mentions, $parent = "") {
		$note = new \Task\Model\Note;

		$note->setId(\Task\Id::make());
		$note->setParentId($parent);
		$note->setChannel($parent === "" ? self::NOTES : "");
		$note->setUserName($user);
		$note->setBody($body);
		$note->setMentions($mentions);
		$note->setCreatedAt(\Task\Clock::stamp());
		$note->setUpdatedAt($note->getCreatedAt());

		$this->db->insert(\Task\Model\Note::TABLE, $note->values());

		return $note;
	}

	/** edit rewrites a note's body and the names it mentions. */
	function edit($id, $body, $mentions) {
		$now = \Task\Clock::stamp();

		$this->db->query("UPDATE note SET body = ?, mentions = ?, edited_at = ?, updated_at = ? WHERE id = ?", $body, $mentions, $now, $now, $id);

		return true;
	}

	/** drop removes a note and every reply under it. */
	function drop($id) {
		$this->db->query("DELETE FROM note WHERE id = ? OR parent_id = ?", $id, $id);

		return true;
	}
}
