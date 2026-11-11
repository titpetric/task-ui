<?php

namespace Task\Settings;

// Policies reads and writes the permission_policy table: an effect per path,
// per subject. A subject is a role or a user, by name; the admin role holds no
// entries.
//
// The table is this module's, which draws the permissions page. What an entry
// means to a reader is Task\Policy.
class Policies {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? \Task\DB::reader() : $db;
	}

	/** policy is a subject's entries. */
	function policy($kind, $subject) {
		$rows = $this->db->get_all("SELECT * FROM permission_policy WHERE subject_kind = ? AND subject = ?", $kind, $subject);
		$entries = array();

		if (is_array($rows)) {
			foreach ($rows as $row) {
				$entries[\Task\Row::text($row, "path")] = \Task\Row::text($row, "effect");
			}
		}

		return new \Task\Policy($entries);
	}

	/**
	 * put sets a subject's entry for a path, or removes it with an empty
	 * effect so the path inherits again.
	 */
	function put($kind, $subject, $path, $effect) {
		if ($kind !== \Task\Policy::ROLE && $kind !== \Task\Policy::USER) {
			throw new \Exception("no subject kind " . $kind);
		}

		if ($subject === "") {
			throw new \Exception("a policy needs a subject");
		}

		if ($kind === \Task\Policy::ROLE && $subject === \Task\User::ADMIN) {
			throw new \Exception("the admin role can do everything and has no policy to change");
		}

		if (!\Task\Permission::valid($path)) {
			throw new \Exception("no such permission: " . $path);
		}

		if ($effect !== "" && $effect !== \Task\Policy::ALLOW && $effect !== \Task\Policy::DENY) {
			throw new \Exception("no effect " . $effect . ": allow, deny or none");
		}

		$this->db->query("DELETE FROM permission_policy WHERE subject_kind = ? AND subject = ? AND path = ?", $kind, $subject, $path);

		if ($effect === "") {
			return true;
		}

		$row = new \Task\Model\PermissionPolicy;

		$row->setId(\Task\Id::make());
		$row->setSubjectKind($kind);
		$row->setSubject($subject);
		$row->setPath($path);
		$row->setEffect($effect);
		$row->setCreatedAt(\Task\Clock::stamp());

		$this->db->insert(\Task\Model\PermissionPolicy::TABLE, $row->values());

		return true;
	}

	/** forget removes every entry of a subject, as when its user goes. */
	function forget($kind, $subject) {
		$this->db->query("DELETE FROM permission_policy WHERE subject_kind = ? AND subject = ?", $kind, $subject);

		return true;
	}
}
