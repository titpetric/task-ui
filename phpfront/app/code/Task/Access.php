<?php

namespace Task;

// Access is what one user may do: the user, the role's policy and the user's
// own. An admin may do everything and holds no policy.
class Access {
	var $user = null;
	var $role = null;
	var $own = null;

	function __construct($user = null, $role = null, $own = null) {
		$this->user = $user;
		$this->role = $role === null ? new Policy : $role;
		$this->own = $own === null ? new Policy : $own;
	}

	/** allowed reports whether the user may do what a path names. */
	function allowed($path) {
		if ($this->user === null) {
			return false;
		}

		if (User::admin($this->user)) {
			return true;
		}

		$own = $this->own->effect($path);

		if ($own !== "") {
			return $own === Policy::ALLOW;
		}

		return $this->role->effect($path) === Policy::ALLOW;
	}

	/** effective is every checkable path and whether the user may. */
	function effective() {
		$out = array();

		foreach (Permission::checks() as $path) {
			$out[$path] = $this->allowed($path);
		}

		return $out;
	}

	/** allows is the paths the user may, in file order. */
	function allows() {
		$out = array();

		foreach (Permission::checks() as $path) {
			if ($this->allowed($path)) {
				$out[] = $path;
			}
		}

		return $out;
	}

	/** summary is what the profile page says the user may do. */
	function summary() {
		if (User::admin($this->user)) {
			return "everything, as an admin";
		}

		$allows = $this->allows();

		if (count($allows) === 0) {
			return "nothing";
		}

		return implode(", ", $allows) . "; everything else is denied";
	}

	/** run reports whether the user may run a job of a folder. */
	function run($folder, $folderAllow = array(), $jobAllow = array()) {
		if (!User::sees($this->user, $folder)) {
			return false;
		}

		if (User::admin($this->user)) {
			return true;
		}

		return $this->allowed("run") && User::listed($this->user, $folderAllow, $jobAllow);
	}

	/** sees reports whether the user may see a folder. */
	function sees($folder) {
		return User::sees($this->user, $folder);
	}
}
