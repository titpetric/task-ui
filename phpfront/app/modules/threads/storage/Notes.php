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

	/** CHAT is the channel a private chat thread carries. */
	const CHAT = "chat";

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

	/** timeline is every board thread, newest first; a chat is listed by chats(). */
	function timeline($limit = self::LIMIT) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND workspace_id = '' AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", self::NOTES, $limit));
	}

	/**
	 * feed is every board thread and prompt, newest first, narrowed by a
	 * place and the directories under it, a user, a state, an origin or words
	 * of the body. A chat is listed by chats().
	 */
	function feed($filter = [], $limit = self::LIMIT) {
		$where = "";
		$args = [self::NOTES];

		if (($filter["dir"] ?? "") !== "") {
			$where .= " AND (dir = ? OR dir LIKE ?)";
			$args[] = $filter["dir"];
			$args[] = $filter["dir"] . "/%";
		}

		foreach ([
			"user" => "user_name",
			"state" => "state",
			"origin" => "origin",
		] as $key => $column) {
			if (($filter[$key] ?? "") === "") {
				continue;
			}

			$where .= " AND " . $column . " = ?";
			$args[] = $filter[$key];
		}

		if (($filter["q"] ?? "") !== "") {
			$where .= " AND body LIKE ?";
			$args[] = "%" . $filter["q"] . "%";
		}

		$args[] = $limit;

		$sql = "SELECT * FROM note WHERE parent_id = '' AND channel = ? AND " . self::LIVE . $where . " ORDER BY rowid DESC LIMIT ?";

		return \Task\Model\Note::rows(\Task\DB::all($this->db, $sql, $args));
	}

	/** live is the board threads and prompts the agent is working on, newest first. */
	function live($limit = self::LIMIT) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND state IN ('open', 'working') AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", self::NOTES, $limit));
	}

	/** recent is the live threads and the ones that moved in the last hour, newest first. */
	function recent($limit = self::LIMIT) {
		return \Task\Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND (state IN ('open', 'working') OR updated_at >= datetime('now', '-1 hour')) AND " . self::LIVE . " ORDER BY rowid DESC LIMIT ?", self::NOTES, $limit));
	}

	/**
	 * chats is the chat threads a user is in, newest message first, each with
	 * its peer, its last message and how much of it the user has not read.
	 */
	function chats($userID, $limit = self::LIMIT) {
		return $this->db->get_all("SELECT p.note_id AS id, p.last_read_at AS last_read_at,
			m.id AS last_id, m.user_id AS last_user_id, m.user_name AS last_user_name,
			m.body AS last_body, m.spoken AS last_spoken, m.audio_path AS last_audio_path, m.created_at AS last_at,
			COALESCE((SELECT q.user_name FROM note_participant q
				WHERE q.note_id = p.note_id AND q.user_id <> p.user_id
				ORDER BY q.created_at, q.user_id LIMIT 1), '') AS peer,
			(SELECT COUNT(*) FROM note u
				WHERE (u.id = p.note_id OR u.parent_id = p.note_id) AND u.deleted_at IS NULL
				AND u.user_id <> p.user_id AND (p.last_read_at IS NULL OR u.created_at > p.last_read_at)) AS unread
			FROM note_participant p
			JOIN note n ON n.id = p.note_id AND n.channel = ? AND n.parent_id = '' AND n.deleted_at IS NULL
			JOIN note m ON m.rowid = (SELECT MAX(rowid) FROM note WHERE (id = p.note_id OR parent_id = p.note_id) AND deleted_at IS NULL)
			WHERE p.user_id = ?
			ORDER BY m.rowid DESC LIMIT ?", self::CHAT, $userID, $limit);
	}

	/** participant reports whether a user is in a chat thread. */
	function participant($id, $userID) {
		return is_array($this->db->get("SELECT note_id FROM note_participant WHERE note_id = ? AND user_id = ?", $id, $userID));
	}

	/** peer is the other user of a chat thread, or its first user for a reader who is not in it. */
	function peer($id, $userID) {
		foreach ($this->db->get_all("SELECT user_id, user_name FROM note_participant WHERE note_id = ? ORDER BY created_at, user_id", $id) as $row) {
			if ($row["user_id"] !== $userID) {
				return $row["user_name"];
			}
		}

		return "";
	}

	/** unread is how many messages in a chat thread a user has not read. */
	function unread($id, $userID) {
		$row = $this->db->get("SELECT COUNT(*) AS n FROM note n
			LEFT JOIN note_participant p ON p.note_id = ? AND p.user_id = ?
			WHERE (n.id = ? OR n.parent_id = ?) AND n.deleted_at IS NULL AND n.user_id <> ?
			AND (p.last_read_at IS NULL OR n.created_at > p.last_read_at)", $id, $userID, $id, $id, $userID);

		return is_array($row) ? (int)$row["n"] : 0;
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
			return [];
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

	/** events is what each run of a thread did, oldest first, by question and run. */
	function events($id) {
		return $this->db->get_all("SELECT * FROM note_event WHERE note_id = ? ORDER BY question_id, run_id, seq", $id);
	}

	/** lastEvent is a thread's newest event ID, the stream's replay cursor. */
	function lastEvent($id) {
		$row = $this->db->get("SELECT id FROM note_event WHERE note_id = ? ORDER BY rowid DESC LIMIT 1", $id);

		return is_array($row) && isset($row["id"]) ? (string)$row["id"] : "";
	}

	/** omits what each run of a thread did from every listing and every reader. */
	function forget($id) {
		$this->db->query("DELETE FROM note_event WHERE note_id = ?", $id);

		return true;
	}

	/** replies is the replies of many threads at once, oldest first, by thread. */
	function replies($ids) {
		$out = [];

		if (count($ids) === 0) {
			return $out;
		}

		$rows = \Task\DB::all($this->db, "SELECT * FROM note WHERE parent_id IN (" . \Task\Row::placeholders($ids) . ") AND " . self::LIVE . " ORDER BY rowid", $ids);

		foreach (\Task\Model\Note::rows($rows) as $note) {
			$parent = $note->getParentId();

			if (!isset($out[$parent])) {
				$out[$parent] = [];
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

	/** purge removes a note, every reply under it and their work, deleted or not. */
	function purge($id) {
		$this->forget($id);
		$this->db->query("DELETE FROM note WHERE id = ? OR parent_id = ?", $id, $id);

		return true;
	}
}
