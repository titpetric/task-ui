<?php

namespace Task;

// WhoField is the one "who gets it" element of a form, which
// components/who-field.tpl draws: a box per role with how many users act with
// it, a box per user with the role a run checks for them, and the sentence for
// nothing checked. frontend/view/who.go.
//
// The same element picks whose runs get an SSH key, who may run a project or a
// job, and whose runs get the Claude token; what differs is the boxes' name,
// their data-type and the sentence for nothing checked.
class WhoField {
	/** name is the roles' boxes' name, and the users' until named() says otherwise. */
	var $name = "";

	/** userName is the users' boxes' name: a key's users beside its policies. */
	var $userName = "";

	/** type is the boxes' data-type: set for a list, joined for one setting. */
	var $type = "";

	/** roles and users are the names the setting holds, checked. */
	var $roles = array();

	var $users = array();

	/** all are the users to list, as rows of the user table. */
	var $all = array();

	/** empty is what nothing checked means; "" means no run gets it then. */
	var $empty = "";

	/** compact draws the roles and the users on one line, with no summary. */
	var $compact = false;

	function __construct($name, $type, $roles, $users, $all, $empty = "") {
		$this->name = $name;
		$this->userName = $name;
		$this->type = $type;
		$this->roles = self::names($roles);
		$this->users = self::names($users);
		$this->all = $all;
		$this->empty = $empty;
	}

	/** named puts the users' boxes under another name. */
	function named($name) {
		$this->userName = $name;

		return $this;
	}

	/** tight is the compact form, for a table cell. */
	function tight() {
		$this->compact = true;

		return $this;
	}

	/** data is the scope the component reads. */
	function data() {
		$roles = array();
		$users = array();

		foreach (User::ROLES as $role) {
			$roles[] = array(
				"name" => $role,
				"checked" => in_array($role, $this->roles, true),
				"count" => $this->count($role),
			);
		}

		foreach ($this->all as $user) {
			$name = $user->getName();

			$users[] = array(
				"name" => $name,
				"role" => User::role($user),
				"checked" => in_array($name, $this->users, true),
			);
		}

		return array(
			"compact" => $this->compact,
			"empty" => $this->empty,
			"name" => $this->name,
			"userName" => $this->userName,
			"type" => $this->type,
			"roles" => $roles,
			"users" => $users,
			"extra" => $this->extra(),
		);
	}

	/** count is how many of the listed users act with a role. */
	function count($role) {
		$found = 0;

		foreach ($this->all as $user) {
			if (User::role($user) === $role) {
				$found++;
			}
		}

		return $found;
	}

	/**
	 * extra is the checked names that are neither a role nor a listed user, so
	 * a save keeps them.
	 */
	function extra() {
		$out = array();

		foreach ($this->users as $name) {
			if (in_array($name, User::ROLES, true) || in_array($name, $out, true)) {
				continue;
			}

			if ($this->listed($name)) {
				continue;
			}

			$out[] = $name;
		}

		return $out;
	}

	/** listed reports whether a name is one of the users the field draws. */
	function listed($name) {
		foreach ($this->all as $user) {
			if ($user->getName() === $name) {
				return true;
			}
		}

		return false;
	}

	/** names is a list of names from what a caller holds them in, the empty out. */
	static function names($value) {
		if (!is_array($value)) {
			return array();
		}

		$out = array();

		foreach ($value as $name) {
			if ($name !== "" && $name !== null) {
				$out[] = $name;
			}
		}

		return $out;
	}
}
