<?php

namespace Task;

// Policy is a subject's entries: an effect per path, as the
// permission_policy rows hold them. No entry inherits.
class Policy {
	/** ALLOW and DENY are the two effects an entry may carry. */
	const ALLOW = "allow";

	const DENY = "deny";

	/** ROLE and USER are the two kinds of subject. */
	const ROLE = "role";

	const USER = "user";

	/** ALL is the rule that stands for every rule of a section. */
	const ALL = "all";

	var $entries = array();

	function __construct($entries = array()) {
		$this->entries = is_array($entries) ? $entries : array();
	}

	/**
	 * effect is what this policy says about a path, "" where it says
	 * nothing. A section's all stands for every rule of the section, and
	 * where all and the rule disagree the deny decides.
	 */
	function effect($path) {
		$effect = isset($this->entries[$path]) ? $this->entries[$path] : "";

		if (Permission::extra($path)) {
			return $effect;
		}

		$parts = explode("/", $path, 2);

		if (count($parts) !== 2) {
			return $effect;
		}

		$all = isset($this->entries[$parts[0] . "/" . self::ALL]) ? $this->entries[$parts[0] . "/" . self::ALL] : "";

		if ($all === self::DENY || $effect === "") {
			$effect = $all;
		}

		return $effect;
	}

	/** set writes an entry, or removes it when the effect is "". */
	function set($path, $effect) {
		if ($effect === "") {
			unset($this->entries[$path]);

			return $this;
		}

		$this->entries[$path] = $effect;

		return $this;
	}
}
