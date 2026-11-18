<?php

namespace Task\Orb;

// Meetings reads the orb's three tables: a room, who joined it, and the lines
// of its log.
//
// Nothing here writes. Joining, leaving, speaking and ending a room are
// /api/meetings, which writes with storage/meeting_storage.go.
class Meetings {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** all is every room, open ones above ended ones, newest first. */
	function all($limit = 40) {
		return \Task\Model\Meeting::rows($this->db->get_all("SELECT * FROM meeting ORDER BY (state = 'open') DESC, started_at DESC LIMIT ?", $limit));
	}

	/** get is one room by ID, or null. */
	function get($id) {
		$row = $this->db->get("SELECT * FROM meeting WHERE id = ?", $id);

		if (!is_array($row)) {
			return null;
		}

		return new \Task\Model\Meeting($row);
	}

	/** participants is everyone who joined a room, oldest first. */
	function participants($id) {
		return \Task\Model\MeetingParticipant::rows($this->db->get_all("SELECT * FROM meeting_participant WHERE meeting_id = ? ORDER BY joined_at", $id));
	}

	/** lines is a room's log, oldest first. */
	function lines($id) {
		return \Task\Model\MeetingLine::rows($this->db->get_all("SELECT * FROM meeting_line WHERE meeting_id = ? ORDER BY created_at, rowid", $id));
	}

	/** present is how many participants are in a room now. */
	function present($id) {
		$row = $this->db->get("SELECT COUNT(*) AS n FROM meeting_participant WHERE meeting_id = ? AND left_at IS NULL", $id);

		return is_array($row) ? (int)$row["n"] : 0;
	}

	/** counts is how many are present in each room, by meeting ID. */
	function counts() {
		$out = [];
		$rows = $this->db->get_all("SELECT meeting_id, COUNT(*) AS n FROM meeting_participant WHERE left_at IS NULL GROUP BY meeting_id");

		if (!is_array($rows)) {
			return $out;
		}

		foreach ($rows as $row) {
			$out[$row["meeting_id"]] = (int)$row["n"];
		}

		return $out;
	}
}
