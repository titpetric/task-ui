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

	/**
	 * of is one user as the user component reads them.
	 *
	 * The page of a user writes the name as it is and the list of their runs
	 * escapes it, which is what the `user-menu` define does: the one is a step
	 * of a path the router reads whole, the other a query field. A name that
	 * needs escaping in a path is a bug of that route and not of this link, so
	 * the two front ends spell it the same and it is fixed in one place.
	 */
	function of($name) {
		$role = isset($this->roles[$name]) ? $this->roles[$name] : "";
		$menu = new Menu("More on " . $name);

		$menu->when($this->profile($name), "users", "Profile", array("href" => Route::href("/users/" . $name)));
		$menu->item("list", "Runs by this user", array("href" => Route::href("/runs?user=" . Route::queryEscape($name))));
		$menu->when($this->edits($name), "edit", "Edit", array("href" => Route::href("/users/" . $name . "/edit")));

		$out = self::avatar($name, $role);
		$out["menu"] = $menu->data();

		return $out;
	}

	/**
	 * avatar is the scope the avatar component reads, for a place that draws a
	 * user by their avatar alone.
	 *
	 * of() answers a superset of it, so a page that draws the whole user hands
	 * the component that scope and does not build this one.
	 */
	static function avatar($name, $role) {
		return array(
			"name" => $name,
			"role" => $role,
			"initials" => self::initials($name),
		);
	}

	/**
	 * ofRun is who started a run and how, as the who component reads them: the
	 * user, and the badges that say the run was not started by hand.
	 *
	 * The table cell of a list and the detail row of a run's own page word a
	 * start the same way, so both read it from here.
	 */
	function ofRun($run) {
		return array(
			"user" => $this->of($run->getUserName()),
			"scheduled" => Run::isScheduled($run),
			"reason" => $run->getReason(),
			"forced" => Run::isForced($run),
		);
	}

	/** profile reports whether the reader may open a user's profile. */
	function profile($name) {
		if (!isset($this->roles[$name])) {
			return false;
		}

		return $this->mine($name) || $this->access->allowed("user/view");
	}

	/**
	 * edits reports whether the reader may open a user's edit page.
	 *
	 * The edit page is the user's permissions, which is not a reader's own to
	 * change: a name in this menu is a link for a reader holding user/edit and
	 * for nobody else, their own included. The profile is the other way about,
	 * because everyone may read theirs.
	 */
	function edits($name) {
		if (!isset($this->roles[$name])) {
			return false;
		}

		return $this->access !== null && $this->access->allowed("user/edit");
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
