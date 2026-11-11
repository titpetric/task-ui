<?php

namespace Task;

// PermissionPolicy reads and writes the permission_policy table: an effect per
// path, per subject. A subject is a role or a user, by name; the admin role
// holds no entries.
class PermissionPolicy {
	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
	}

	/** policy is a subject's entries. */
	function policy($kind, $subject) {
		$rows = $this->db->get_all("SELECT * FROM permission_policy WHERE subject_kind = ? AND subject = ?", $kind, $subject);
		$entries = array();

		if (is_array($rows)) {
			foreach ($rows as $row) {
				$entries[Row::text($row, "path")] = Row::text($row, "effect");
			}
		}

		return new Policy($entries);
	}

	/**
	 * put sets a subject's entry for a path, or removes it with an empty
	 * effect so the path inherits again.
	 */
	function put($kind, $subject, $path, $effect) {
		if ($kind !== Policy::ROLE && $kind !== Policy::USER) {
			throw new \Exception("no subject kind " . $kind);
		}

		if ($subject === "") {
			throw new \Exception("a policy needs a subject");
		}

		if ($kind === Policy::ROLE && $subject === User::ADMIN) {
			throw new \Exception("the admin role can do everything and has no policy to change");
		}

		if (!Permission::valid($path)) {
			throw new \Exception("no such permission: " . $path);
		}

		if ($effect !== "" && $effect !== Policy::ALLOW && $effect !== Policy::DENY) {
			throw new \Exception("no effect " . $effect . ": allow, deny or none");
		}

		$this->db->query("DELETE FROM permission_policy WHERE subject_kind = ? AND subject = ? AND path = ?", $kind, $subject, $path);

		if ($effect === "") {
			return true;
		}

		$row = new Model\PermissionPolicy;

		$row->setId(Id::make());
		$row->setSubjectKind($kind);
		$row->setSubject($subject);
		$row->setPath($path);
		$row->setEffect($effect);
		$row->setCreatedAt(Clock::stamp());

		$this->db->insert(Model\PermissionPolicy::TABLE, $row->values());

		return true;
	}

	/** forget removes every entry of a subject, as when its user goes. */
	function forget($kind, $subject) {
		$this->db->query("DELETE FROM permission_policy WHERE subject_kind = ? AND subject = ?", $kind, $subject);

		return true;
	}
}
