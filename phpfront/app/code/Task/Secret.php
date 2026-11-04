<?php

namespace Task;

// Secret reads the secret table: the environment variables and secrets a
// workspace's runs are given.
//
// No value is here. A secret's value is encrypted with the instance's key and
// the key does not enter PHP, so nothing on this class answers more than what
// a page counts or lists.
class Secret {
	/**
	 * OWNED starts the scope of a variable an integration owns, as
	 * integration/store.go spells it.
	 *
	 * Those rows are a service's own settings and not a user's variables, so
	 * the settings leave them out wherever they count them.
	 */
	const OWNED = "integration:%";

	/** PAGE is the Secrets step of the settings, which a variable's pages hang under. */
	const PAGE = "/settings/secrets";

	/** ALL is what the scope of every workspace is called in the form's picker. */
	const ALL = "All workspaces";

	/** READS is how many reads of one variable a page of its log shows. */
	const READS = 50;

	/**
	 * MISSING is what the page of a variable that is not there reports, which
	 * is the sentence frontend/secret.go writes into the shared 404.
	 *
	 * The ID is not repeated in it. A reader who followed a stale link learns
	 * no more than that the link is stale.
	 */
	const MISSING = "No secret with that ID.";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
	}

	/** count is how many variables the settings list, the integrations' apart. */
	function count() {
		$row = $this->db->get("SELECT COUNT(*) AS found FROM secret WHERE scope NOT LIKE ?", self::OWNED);

		return Row::number($row, "found");
	}

	/**
	 * scopes is the scope of every variable the settings list, one entry per
	 * variable, which the keys page counts by workspace.
	 */
	function scopes() {
		$out = array();

		foreach ($this->db->get_all("SELECT scope FROM secret WHERE scope NOT LIKE ?", self::OWNED) as $row) {
			$out[] = Row::text($row, "scope");
		}

		return $out;
	}

	/**
	 * all is every variable the settings list, in the order the page draws
	 * them: those of every workspace first, then each workspace's, then a
	 * folder's, and by name inside each. storage/secret_storage.go:81.
	 *
	 * No content is read here. The value of a plain env entry is the vault's,
	 * which Task::variables() answers by row ID, and a secret has none to draw.
	 */
	function all() {
		$sql = "SELECT * FROM secret WHERE scope NOT LIKE ? ORDER BY CASE WHEN scope = ? THEN 0 WHEN scope LIKE ? THEN 1 ELSE 2 END, scope, name";

		return Model\Secret::rows($this->db->get_all($sql, self::OWNED, Workspace::EVERY, Workspace::SCOPE . "%"));
	}

	/** counts is how many runs got each variable, by its ID. */
	function counts() {
		$out = array();

		foreach ($this->db->get_all("SELECT secret_id, COUNT(*) AS found FROM secret_read GROUP BY secret_id") as $row) {
			$out[Row::text($row, "secret_id")] = Row::number($row, "found");
		}

		return $out;
	}

	/**
	 * last is the newest read of each variable, by its ID.
	 *
	 * The newest is the greatest rowid and not the greatest timestamp, for the
	 * reason reads() gives: the reads of one run share a millisecond.
	 */
	function last() {
		$sql = "SELECT * FROM secret_read WHERE rowid IN (SELECT MAX(rowid) FROM secret_read GROUP BY secret_id)";
		$out = array();

		foreach (Model\SecretRead::rows($this->db->get_all($sql)) as $read) {
			$out[$read->getSecretId()] = $read;
		}

		return $out;
	}

	/**
	 * titles names each scope the list prints: every workspace, then one entry
	 * per workspace row. frontend/secret.go:180.
	 *
	 * A scope the map does not name is printed as it is written, which is a
	 * folder's own scope and a workspace that is gone alike.
	 */
	static function titles($rows) {
		$out = array(Workspace::EVERY => self::ALL);

		foreach ($rows as $row) {
			$out[Workspace::SCOPE . $row->getFolder()] = $row->getTitle();
		}

		return $out;
	}

	/**
	 * reaching keeps the variables a workspace's runs are given: those of every
	 * workspace, the workspace's own, and its folders'.
	 * frontend/secret.go:190.
	 *
	 * A scope is a list, so one term reaching the workspace keeps the row.
	 */
	static function reaching($rows, $workspace, $folders) {
		$want = Route::canonical($workspace);
		$out = array();

		foreach ($rows as $row) {
			foreach (explode(",", $row->getScope()) as $term) {
				$term = trim($term);

				if ($term !== Workspace::EVERY && Route::canonical(Workspace::of($term, $folders)) !== $want) {
					continue;
				}

				$out[] = $row;

				break;
			}
		}

		return $out;
	}

	/**
	 * byId is the variable an ID names, null where none does.
	 *
	 * The row carries its encrypted content, which no reader here opens: the
	 * pages that read a variable by ID read its name and its scope. An
	 * integration's own setting is no variable of the settings, so an ID that
	 * names one answers as an ID that names nothing, which is what the scan in
	 * frontend/secret.go does with the list it filtered.
	 */
	function byId($id) {
		$row = $this->db->get("SELECT * FROM secret WHERE id = ? AND scope NOT LIKE ?", $id, self::OWNED);

		if (!is_array($row)) {
			return null;
		}

		return new Model\Secret($row);
	}

	/**
	 * edited is the variable a form was opened on: the row an ID names, or the
	 * row a scope and a name name where no ID was given, null where neither
	 * names one. frontend/secret.go:127.
	 *
	 * The integration scope is not left out here, which byId() leaves out: the
	 * form opens whatever row it was pointed at, and a service's own setting
	 * has a form like any other variable.
	 */
	function edited($id, $scope, $name) {
		if ($id !== "") {
			$row = $this->db->get("SELECT * FROM secret WHERE id = ?", $id);
		} else {
			$row = $this->db->get("SELECT * FROM secret WHERE scope = ? AND name = ?", $scope, $name);
		}

		if (!is_array($row)) {
			return null;
		}

		return new Model\Secret($row);
	}

	/**
	 * says is what the form of a variable that is not there reports, which is
	 * the sentence frontend/secret.go:134 writes into the shared 404.
	 *
	 * The name and the scope are the query's and not a row's, so a form opened
	 * by ID alone names neither: a reader who followed a stale link learns the
	 * link is stale and nothing else.
	 */
	static function says($name, $scope) {
		return "No variable named " . $name . " for " . $scope . ".";
	}

	/**
	 * reads is a page of the runs that got a variable, newest first, or oldest
	 * first for a page read after a cursor, which is reversed here.
	 *
	 * The order is the row's and not the timestamp's: the reads of one run are
	 * written in one transaction and share a millisecond, so a timestamp
	 * orders none of them and rowid orders all of them. The cursor is a read's
	 * ID, and a cursor whose row is gone keeps nothing.
	 */
	function reads($id, $before, $after, $limit) {
		$sql = "SELECT * FROM secret_read WHERE secret_id = ?";
		$args = array($id);
		$order = " ORDER BY rowid DESC";

		if ($before !== "") {
			$sql .= " AND rowid < (SELECT rowid FROM secret_read WHERE id = ?)";
			$args[] = $before;
		} elseif ($after !== "") {
			$sql .= " AND rowid > (SELECT rowid FROM secret_read WHERE id = ?)";
			$args[] = $after;
			$order = " ORDER BY rowid ASC";
		}

		$sql .= $order;

		if ($limit > 0) {
			$sql .= " LIMIT ?";
			$args[] = $limit;
		}

		$out = Model\SecretRead::rows(DB::all($this->db, $sql, $args));

		if ($before === "" && $after !== "") {
			$out = array_reverse($out);
		}

		return $out;
	}

	/**
	 * options is the Applies-to picker of a variable's form: every workspace
	 * first, then one entry per workspace row, the scope the variable is in
	 * marked.
	 *
	 * The title is the row's own and a row that carries none draws an empty
	 * option, which is what frontend/secret.go hands the template. The rows
	 * are Task\Workspace::all(), the default workspace among them: a variable
	 * of the root's runs is scoped to it like any other.
	 */
	static function options($scope, $rows) {
		$out = array();
		$out[] = array(
			"value" => Workspace::EVERY,
			"title" => self::ALL,
			"selected" => $scope === Workspace::EVERY,
		);

		foreach ($rows as $row) {
			$value = Workspace::SCOPE . $row->getFolder();
			$out[] = array(
				"value" => $value,
				"title" => $row->getTitle(),
				"selected" => $scope === $value,
			);
		}

		return $out;
	}
}
