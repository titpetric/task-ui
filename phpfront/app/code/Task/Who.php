<?php

namespace Task;

// Who draws a user: the avatar, the name and the menu beside them.
//
// Every page that names a user draws the same thing, so the roles are read
// once per request and each name is built from that.
class Who {
	var $roles = array();
	var $access = null;

	function __construct($access, $roles) {
		$this->access = $access;
		$this->roles = is_array($roles) ? $roles : array();
	}

	/** of is one user as the user component reads them. */
	function of($name) {
		$role = isset($this->roles[$name]) ? $this->roles[$name] : "";
		$menu = new Menu("More on " . $name);

		$menu->when($this->profile($name), "users", "Profile", array("href" => Route::href("/users/" . rawurlencode($name))));
		$menu->item("list", "Runs by this user", array("href" => Route::href("/runs?user=" . rawurlencode($name))));
		$menu->when($this->edits($name), "edit", "Edit", array("href" => Route::href("/users/" . rawurlencode($name) . "/edit")));

		return array(
			"name" => $name,
			"role" => $role,
			"initials" => self::initials($name),
			"menu" => $menu->data(),
		);
	}

	/** profile reports whether the reader may open a user's profile. */
	function profile($name) {
		if (!isset($this->roles[$name])) {
			return false;
		}

		return $this->mine($name) || $this->access->allowed("user/view");
	}

	/** edits reports whether the reader may open a user's edit page. */
	function edits($name) {
		if (!isset($this->roles[$name])) {
			return false;
		}

		return $this->mine($name) || $this->access->allowed("user/edit");
	}

	/** mine reports whether a name is the reader's own. */
	function mine($name) {
		return $this->access !== null && $this->access->user !== null && $this->access->user->getName() === $name;
	}

	/** linkable is every name whose profile the reader may open, for a mention. */
	function linkable() {
		$out = array();

		foreach (array_keys($this->roles) as $name) {
			if ($this->profile($name)) {
				$out[] = $name;
			}
		}

		return $out;
	}

	/**
	 * initials are the one or two letters an avatar draws: the first letters
	 * of the first two words of a name, or the first two of a single word.
	 */
	static function initials($name) {
		if ($name === "") {
			return "?";
		}

		$words = preg_split('/[ .\-_@]+/', $name, -1, PREG_SPLIT_NO_EMPTY);

		if (!is_array($words) || count($words) === 0) {
			return "?";
		}

		$initials = substr($words[0], 0, 1);

		if (count($words) >= 2) {
			$initials .= substr($words[1], 0, 1);
		} elseif (strlen($words[0]) > 1) {
			$initials = substr($words[0], 0, 2);
		}

		return strtoupper($initials);
	}
}
