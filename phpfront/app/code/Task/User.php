<?php

namespace Task;

// User reads and writes the user table, and answers what a row of it means:
// the role it acts with, the folders it sees, the initials the header draws.
//
// The row behaviour sits here rather than on Task\Model\User because the
// runtime merges no methods into a generated class: there is no trait and no
// inheritance, so one class owns the table and its rows.
class User {
	/** ADMIN, RUNNER and VIEWER are the roles. */
	const ADMIN = "admin";

	const RUNNER = "runner";

	const VIEWER = "viewer";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
	}

	/** get is one user by ID, or null. */
	function get($id) {
		$row = $this->db->get("SELECT * FROM user WHERE id = ?", $id);

		return is_array($row) ? new Model\User($row) : null;
	}

	/** byName is one user by name, or null. */
	function byName($name) {
		$row = $this->db->get("SELECT * FROM user WHERE name = ?", $name);

		return is_array($row) ? new Model\User($row) : null;
	}

	/** all is every user, by name. */
	function all() {
		return Model\User::rows($this->db->get_all("SELECT * FROM user ORDER BY name"));
	}

	/** count is how many users there are. */
	function count() {
		$row = $this->db->get("SELECT COUNT(*) AS found FROM user");

		return Row::number($row, "found");
	}

	/** admins is how many users carry the admin flag. */
	function admins() {
		$row = $this->db->get("SELECT COUNT(*) AS found FROM user WHERE admin = 1");

		return Row::number($row, "found");
	}

	/** agents is every user an agent may reply as. */
	function agents() {
		return Model\User::rows($this->db->get_all("SELECT * FROM user WHERE agent_reply = 1 AND mcp = 1 ORDER BY name"));
	}

	/** roles is the role each user acts with, by name. */
	function roles() {
		$out = array();

		foreach ($this->all() as $user) {
			$out[$user->getName()] = self::role($user);
		}

		return $out;
	}

	/** access is what a user may do, from the role's policy and the user's own. */
	function access($user) {
		if ($user === null || self::admin($user)) {
			return new Access($user);
		}

		$policies = new PermissionPolicy($this->db);

		return new Access($user, $policies->policy(Policy::ROLE, $user->getRole()), $policies->policy(Policy::USER, $user->getName()));
	}

	/** admin reports whether a row is an admin, by the flag or the role. */
	static function admin($user) {
		return $user !== null && ($user->getAdmin() || $user->getRole() === self::ADMIN);
	}

	/** role is the role a row acts with: admin for an admin, else its column. */
	static function role($user) {
		if ($user === null) {
			return "";
		}

		return self::admin($user) ? self::ADMIN : $user->getRole();
	}

	/** folders is the folders a row may see; an empty list means every one. */
	static function folders($user) {
		return $user === null ? array() : Text::items($user->getFolders());
	}

	/** sees reports whether a row may see a folder. */
	static function sees($user, $folder) {
		if ($user === null) {
			return false;
		}

		$list = self::folders($user);

		return count($list) === 0 || in_array($folder, $list, true);
	}

	/** listed reports whether both allow lists let a row run a job. */
	static function listed($user, $folderAllow = array(), $jobAllow = array()) {
		return $user !== null && self::onList($user, $folderAllow) && self::onList($user, $jobAllow);
	}

	/** initials are the one or two letters an avatar draws for a row. */
	static function initials($user) {
		return $user === null ? "?" : Who::initials($user->getName());
	}

	/** onList reports whether an empty list, or one naming the row or its role. */
	static function onList($user, $list) {
		if (!is_array($list) || count($list) === 0) {
			return true;
		}

		return in_array($user->getName(), $list, true) || in_array(self::role($user), $list, true);
	}
}
