<?php

namespace Task\Threads;

// Notes reads and writes the note table: the notes that start a thread, newest
// first, and the replies under them.
//
// The table is this module's. A note of a channel is a board note; a note with
// a workspace is a prompt, which belongs to the page it was asked from and not
// to the board. What a page draws of one is Task\Note.
class Notes {
	/** NOTES is the channel the board's own notes carry. */
	const NOTES = "notes";

	/** LIVE is what a listing adds to its WHERE: a deleted thread is read by nothing. */
	const LIVE = "deleted_at IS NULL";

	/** LIMIT is how many threads the board lists. */
	const LIMIT = 100;

	/** PROMPTS is how many prompt threads a run page lists. */
	const PROMPTS = 20;

	/** PROFILE is how many notes and prompts of a user a profile lists. */
	const PROFILE = 10;

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** threads is the notes that start a thread of a channel, newest first. */
	function threads($channel, $limit = self::LIMIT) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND workspace_id = '' AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", $channel, $limit));
	}

	/** timeline is every thread of every channel, newest first. */
	function timeline($limit = self::LIMIT) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", $limit));
	}

	/** drop marks a thread and its replies deleted, which hides them from every listing. */
	function drop($id) {
		$now = \Task\Clock::stamp();

		$this->db->query("UPDATE note SET deleted_at = ?, updated_at = ? WHERE (id = ? OR parent_id = ?) AND " . self::LIVE, $now, $now, $id, $id);

		return true;
	}

	/**
	 * byUser is the board notes one user wrote, newest first, which a profile
	 * lists under Notes.
	 *
	 * The replies they wrote under someone else's note are not theirs to list:
	 * a profile lists threads, as the board does.
	 */
	function byUser($channel, $name, $limit = self::PROFILE) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND user_name = ? AND workspace_id = '' AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", $channel, $name, $limit));
	}

	/** promptsByUser is the prompts one user asked, newest first. */
	function promptsByUser($name, $limit = self::PROFILE) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND user_name = ? AND workspace_id != '' AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", $name, $limit));
	}

	/** prompts is the prompt threads asked about one run, newest first. */
	function prompts($runID, $limit = self::PROMPTS) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND about_run_id = ? AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", $runID, $limit));
	}

	/** inPlace is the prompts asked from one directory, about no run and no area. */
	function inPlace($dir, $limit = self::PROMPTS) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND dir = ? AND " . self::LIVE . " AND about_run_id = '' AND about_workflow_run_id = '' AND area = '' ORDER BY rowid DESC LIMIT ?", $dir, $limit));
	}

	/** ofArea is the prompts asked from one page of the instance. */
	function ofArea($area, $limit = self::PROMPTS) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND area = ? AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", $area, $limit));
	}

	/** ofWorkflowRun is the prompts asked about one workflow run. */
	function ofWorkflowRun($runID, $limit = 5) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND about_workflow_run_id = ? AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", $runID, $limit));
	}

	/** promptsAbout is the prompts asked about many runs at once, newest first. */
	function promptsAbout($runIDs) {
		if (count($runIDs) === 0) {
			return array();
		}

		$sql = "SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND " . self::LIVE . " AND about_run_id IN (" . \Task\Row::placeholders($runIDs) . ") ORDER BY rowid DESC";

		return \Task\Model\Note::rows(\Task\DB::all($this->db, $sql, $runIDs));
	}

	/** get is one note by ID, or null for one that is gone. */
	function get($id) {
		$row = $this->db->get("SELECT * FROM note WHERE id = ? AND " . self::LIVE, $id);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Note($row);
	}

	/** byRun is the thread a run belongs to, the parent for a reply's run. */
	function byRun($runID) {
		if ($runID === "") {
			return null;
		}

		$row = $this->db->get("SELECT * FROM note WHERE run_id = ? AND " . self::LIVE . " ORDER BY rowid LIMIT 1", $runID);

		if (!is_array($row)) {
			return null;
		}

		$note = new \Task\Model\Note($row);

		return $note->getParentId() === "" ? $note : $this->get($note->getParentId());
	}

	/** replies is the replies of many threads at once, oldest first, by thread. */
	function replies($ids) {
		$out = array();

		if (count($ids) === 0) {
			return $out;
		}

		$rows = \Task\DB::all($this->db, "SELECT * FROM note WHERE parent_id IN (" . \Task\Row::placeholders($ids) . ") AND " . self::LIVE . " ORDER BY rowid", $ids);

		foreach (\Task\Model\Note::rows($rows) as $note) {
			$parent = $note->getParentId();

			if (!isset($out[$parent])) {
				$out[$parent] = array();
			}

			$out[$parent][] = $note;
		}

		return $out;
	}

	/**
	 * post writes a note, or a reply when the parent is named.
	 *
	 * A reply holds the channel of the thread it sits under. The column takes
	 * the board's channel or the chat's and nothing else, and the thread list
	 * reads the parent to tell a thread from a reply.
	 */
	function post($user, $body, $parent = "") {
		$note = new \Task\Model\Note;

		$note->setId(\Task\Id::make());
		$note->setParentId($parent);
		$note->setChannel(self::NOTES);
		$note->setUserName($user);
		$note->setBody($body);
		$note->setMentions(\Task\Note::mentions($body));
		$note->setCreatedAt(\Task\Clock::stamp());
		$note->setUpdatedAt($note->getCreatedAt());

		$this->db->insert(\Task\Model\Note::TABLE, $note->values());

		return $note;
	}

	/** purge removes a note and every reply under it, deleted or not. */
	function purge($id) {
		$this->db->query("DELETE FROM note WHERE id = ? OR parent_id = ?", $id, $id);

		return true;
	}
}
