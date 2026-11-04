<?php

namespace Task;

// Permission reads config/permissions.yml, which arrives over Host::rules(),
// and decides what a user may do: the user's own entry first, the role's
// where the user has none, and a refusal where neither sets one.
class Permission {
	/** rules is the file, read once per request. */
	public static $file = null;

	/** file is config/permissions.yml. */
	static function file() {
		if (self::$file === null) {
			self::$file = Host::rules();
		}

		return self::$file;
	}

	/** titles is one of the file's titled lists: rules, sections or permissions. */
	static function titles($which) {
		$file = self::file();

		return isset($file[$which]) ? $file[$which] : array();
	}

	/** paths is every path a policy may set, in file order, the all rules included. */
	static function paths() {
		$file = self::file();

		return isset($file["paths"]) ? $file["paths"] : array();
	}

	/** checks is the paths a check may name: paths without the all rules. */
	static function checks() {
		$file = self::file();

		return isset($file["checks"]) ? $file["checks"] : array();
	}

	/**
	 * extras is the rules the sections added of their own, which no column of
	 * every section holds: a key of <section>/<rule>, and the rule's own
	 * title, which the matrix composes with the section's.
	 */
	static function extras() {
		$file = self::file();

		return isset($file["extras"]) ? $file["extras"] : array();
	}

	/** extra reports whether a path is a rule a section added of its own. */
	static function extra($path) {
		foreach (self::extras() as $each) {
			if (Row::text($each, "key") === $path) {
				return true;
			}
		}

		return false;
	}

	/** valid reports whether a policy may set a path. */
	static function valid($path) {
		return in_array($path, self::paths(), true);
	}

	/** checkable reports whether a check may name a path. */
	static function checkable($path) {
		return in_array($path, self::checks(), true);
	}
}
