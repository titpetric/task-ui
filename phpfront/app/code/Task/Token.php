<?php

namespace Task;

// Token reads the user_auth_token table: the tokens that sign a client in
// without a session, one row per user per name.
//
// Nothing here answers with the hash and nothing answers with the token. A
// token is shown once, by whoever made it, and a page that lists one says when
// it was made, when it runs out and when it was last used.
class Token {
	/** MCP names the token a user's settings enable for MCP, one per user. */
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
		$this->db = $db === null ? DB::reader() : $db;
	}

	/** get is one user's token by name, or null where they have none. */
	function get($user, $name) {
		$row = $this->db->get("SELECT * FROM user_auth_token WHERE user_id = ? AND name = ?", $user, $name);

		if (!is_array($row)) {
			return null;
		}

		return new Model\UserAuthToken($row);
	}

	/**
	 * expired reports whether an expiry has passed. A token from `task-ui
	 * token` has none and never expires, which reads as not expired.
	 */
	static function expired($expires) {
		$at = Clock::nanos($expires);

		return $at !== 0 && Clock::now()->unixnano() >= $at;
	}

	/**
	 * tone is the colour of the expiry badge: bad once it has passed, signal
	 * within a week of it, and the plain badge before that and for a token
	 * that never expires.
	 */
	static function tone($expires) {
		$at = Clock::nanos($expires);

		if ($at === 0) {
			return "";
		}

		$now = Clock::now()->unixnano();

		if ($now >= $at) {
			return "bad";
		}

		return $at - $now <= self::SOON ? "signal" : "";
	}
}
