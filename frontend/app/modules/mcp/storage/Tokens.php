<?php

namespace Task\Mcp;

// Tokens reads the user_auth_token table: the tokens that sign a client in
// without a session.
//
// A row's kind says what issued it, which is this module's column: mcp for the
// ones a user's MCP access issues, which expire and which their page lists and
// deletes, and empty for the ones `task-ui token` writes, which never expire
// and which no page shows. A name is unique per user and kind, so a user holds
// as many MCP tokens as they care to name.
//
// Nothing here answers with the hash and nothing answers with the token. A
// token is shown once, by whoever made it, and a page that lists one says when
// it was made, when it runs out and when it was last used.
class Tokens {
	/** MCP is the kind of a token a user's MCP access issued. */
	const MCP = "mcp";

	/**
	 * SOON is how close to its expiry a token shows in amber, in nanoseconds:
	 * a week, as view.expiresSoon is. Nanoseconds because the two moments are
	 * compared the way view.MCPAccess compares two times, which the second one
	 * either side of the boundary reads the other way.
	 */
	const SOON = 604800000000000;

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/**
	 * ofKind is a user's tokens of one kind, newest first: the order their page
	 * lists them in, because the one they just issued is the one they came for.
	 */
	function ofKind($user, $kind) {
		$rows = $this->db->get_all("SELECT * FROM user_auth_token WHERE user_id = ? AND kind = ? ORDER BY created_at DESC, name", $user, $kind);
		$out = array();

		if (!is_array($rows)) {
			return $out;
		}

		foreach ($rows as $row) {
			$out[] = new \Task\Model\UserAuthToken($row);
		}

		return $out;
	}

	/**
	 * expired reports whether an expiry has passed. A token from `task-ui
	 * token` has none and never expires, which reads as not expired.
	 */
	static function expired($expires) {
		$at = \Task\Clock::nanos($expires);

		return $at !== 0 && \Task\Clock::now()->unixnano() >= $at;
	}

	/**
	 * tone is the colour of the expiry badge: bad once it has passed, signal
	 * within a week of it, and the plain badge before that and for a token
	 * that never expires.
	 */
	static function tone($expires) {
		$at = \Task\Clock::nanos($expires);

		if ($at === 0) {
			return "";
		}

		$now = \Task\Clock::now()->unixnano();

		if ($now >= $at) {
			return "bad";
		}

		return $at - $now <= self::SOON ? "signal" : "";
	}
}
