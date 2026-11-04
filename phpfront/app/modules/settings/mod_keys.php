<?php

// @route GET /settings/keys

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_keys is the key folder: the encryption key that stores the secrets, the
// SSH keys that connect runs and git to hosts, who gets each of them, and the
// private keys a checkout keeps that no run is given.
//
// The page reads and does not write. Every change is a menu item that puts to
// the API, which gates it again; no private material is here, because the
// keystore holds it and it does not enter PHP.
class mod_keys implements Module {
	/** LEAD is the sentence under the heading, as HTML: it emphasises the two kinds. */
	const LEAD = "There are two kinds of key, and a key can't be both. The <b>encryption key</b> protects secrets: the instance has one, and it encrypts every secret, the workspaces' and the integrations'. <b>SSH keys</b> let runs and git connect to hosts and remotes; git runs over ssh. Instance SSH keys go to the runs of the roles and users they link; a workspace's SSH keys go to every run of that workspace, with nothing to link.";

	/** PAGE is this page, which a change made in a key's menu returns to. */
	const PAGE = "/settings/keys";

	/** EMPTY_LINE is what the Every key head says where the folder holds none. */
	const EMPTY_LINE = "No keys added.";

	function name() {
		return "settings";
	}

	function title() {
		return "Settings";
	}

	function icon() {
		return "settings";
	}

	function tables() {
		return array("user", "workspace", "secret");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("key/view")) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::KEYS, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::KEYS, "href" => ""),
		));

		$folder = Task\Key::folder();
		$keys = $folder["keys"];
		$steps = Task\Setup::steps($chrome, Task\Setup::KEYS);
		$users = new Task\User($chrome->db);
		$everyone = $users->all();

		$view->page(array(
			"head" => $this->head($chrome),
			"scopes" => $this->scopes($chrome, $folder),
			"block" => $this->block($folder, $keys),
			"keys" => Task\Key::rows($chrome, $keys, self::PAGE),
			"holders" => $this->holders($chrome, $keys, $everyone),
			"folders" => $this->folders(),
			"trail" => Task\Setup::trail($steps),
		), "keys.tpl");
	}

	/** head is the page head: the heading, the lead and the two add buttons. */
	function head($chrome) {
		$buttons = new Task\Buttons;
		$head = new Task\Head(Task\Setup::KEYS, self::LEAD);

		if ($chrome->allowed("key/edit")) {
			$buttons->item("upload", "Upload key", array("href" => Task\Route::href("/settings/keys/upload")));
			$buttons->primary("key", "Generate key", array("href" => Task\Route::href("/settings/keys/generate")));
		}

		$head->buttons($buttons->data());

		return $head->data();
	}

	/**
	 * block is the head of the Every key section: what it is called, the key
	 * folder it reads, and the line a folder with no key in it draws instead of
	 * a table.
	 */
	function block($folder, $keys) {
		return array(
			"title" => "Every key",
			"note" => Task\Row::text($folder, "dir"),
			"empty" => count($keys) === 0 ? self::EMPTY_LINE : "",
			"buttons" => array(),
		);
	}

	/**
	 * scopes is the Scopes table: the instance's row and a row per workspace,
	 * each with its variables, its encryption key and its SSH keys.
	 *
	 * Every workspace's secrets are stored on the instance's encryption key, so
	 * the column says so and names no key of the workspace's own.
	 */
	function scopes($chrome, $folder) {
		$keys = $folder["keys"];
		$workspaces = new Task\Workspace($chrome->db);
		$rows = $workspaces->all();
		$folders = array();

		foreach ($rows as $row) {
			$folders[] = $row->getFolder();
		}

		$out = array();
		$index = array();

		foreach ($rows as $at => $row) {
			$index[$row->getFolder()] = $at;
			$out[] = $this->workspace($row, $keys);
		}

		$out = $this->projects($chrome, $out, $index, $folders);
		$out = $this->variables($chrome, $out, $index, $folders);

		return array("instance" => $this->instance($folder), "workspaces" => $out);
	}

	/**
	 * instance is the Instance row: its encryption key, which stores every
	 * secret of every workspace, and the instance SSH keys a policy hands to
	 * runs.
	 *
	 * The folder names its encryption key, so the row is the key of that name
	 * and the page reads no scope of its own to find it.
	 */
	function instance($folder) {
		$named = Task\Row::text($folder, "instance");
		$chips = array();
		$found = array();

		foreach ($folder["keys"] as $key) {
			if ($key["name"] === $named) {
				$found = $key;
			}

			if ($key["encrypts"] || $key["workspace"] !== "") {
				continue;
			}

			$name = $key["name"];
			$chips[] = array(
				"name" => $name,
				"href" => Task\Route::href("/settings/keys/" . $name),
				"title" => "Links of " . $name,
			);
		}

		$set = count($found) > 0;

		return array(
			"set" => $set,
			"name" => $set ? $found["name"] : "",
			"fingerprint" => $set ? $found["fingerprint"] : "",
			"backup" => $set ? $found["backup"] : "",
			"anchor" => "#key-" . $named,
			"chips" => $chips,
			"count" => count($chips),
		);
	}

	/** workspace is one workspace's row, before its projects and variables. */
	function workspace($row, $keys) {
		$folder = $row->getFolder();
		$title = $row->getTitle();
		$base = "/settings/workspaces/" . $folder;
		$scope = Task\Workspace::SCOPE . $row->getId();
		$ssh = array();

		foreach ($keys as $key) {
			if ($key["purpose"] !== Task\Key::SSH || $key["scope"] !== $scope) {
				continue;
			}

			$name = $key["name"];
			$ssh[] = array(
				"name" => $name,
				"href" => Task\Route::href($base . "/keys#key-" . $name),
				"title" => $name . ", to every run of " . $folder,
			);
		}

		return array(
			"folder" => $folder,
			"title" => $title === "" ? $folder : $title,
			"badge" => $title !== "" && $title !== $folder ? $folder : "",
			"href" => Task\Route::href($base),
			"keysHref" => Task\Route::href($base . "#keys"),
			"secretsHref" => Task\Route::href("/settings/secrets?workspace=" . $folder),
			"projects" => "",
			"secrets" => 0,
			"ssh" => $ssh,
			"count" => count($ssh),
		);
	}

	/**
	 * projects names the checkouts of each workspace: the folders one level
	 * inside a named workspace, and the ones at the top of the root that are no
	 * workspace of their own for the default workspace. frontend/key.go:96.
	 */
	function projects($chrome, $rows, $index, $folders) {
		$held = array();

		foreach ($chrome->folders() as $folder) {
			$id = Task\Row::text($folder, "id");
			$name = Task\Workspace::of($id, $folders);
			$depth = substr_count($id, "/");

			if (!isset($index[$name]) || !self::project($id, $depth, $name, $index)) {
				continue;
			}

			$at = $index[$name];

			if (!isset($held[$at])) {
				$held[$at] = array();
			}

			$held[$at][] = $id;
		}

		foreach ($rows as $at => $row) {
			$rows[$at]["projects"] = isset($held[$at]) ? implode(", ", $held[$at]) : "";
		}

		return $rows;
	}

	/**
	 * project reports whether a folder is a project of the workspace it belongs
	 * to: one level inside a named workspace, or at the top of the root and no
	 * named workspace itself for the default one.
	 */
	static function project($id, $depth, $workspace, $index) {
		if ($workspace !== Task\Route::INTERNAL) {
			return $depth === 1;
		}

		return $depth === 0 && !isset($index[$id]);
	}

	/**
	 * variables counts the run variables of each workspace, one per variable.
	 *
	 * An integration's own settings are not a workspace's variables and are
	 * left out, and the ones every workspace's runs are given belong to no row.
	 * frontend/key.go:82.
	 */
	function variables($chrome, $rows, $index, $folders) {
		$secrets = new Task\Secret($chrome->db);

		foreach ($secrets->scopes() as $scope) {
			$name = Task\Workspace::of($scope, $folders);

			if (!isset($index[$name])) {
				continue;
			}

			$at = $index[$name];
			$rows[$at]["secrets"] = $rows[$at]["secrets"] + 1;
		}

		return $rows;
	}

	/**
	 * holders is the Who gets which key table: a row per role and then one per
	 * user, each with the instance SSH keys their runs get.
	 *
	 * A user's own row says both: the keys linked to them by name, and the ones
	 * their role brings, which a change of role takes away again.
	 * frontend/view/key_holder.go:21.
	 */
	function holders($chrome, $keys, $users) {
		$who = $chrome->who();
		$out = array();

		foreach (Task\User::ROLES as $role) {
			$brought = self::ofRole($keys, $role);

			$out[] = array(
				"user" => array(),
				"role" => $role,
				"direct" => self::chips($brought, $role, false),
				"viaRole" => array(),
			);
		}

		foreach ($users as $user) {
			$role = Task\User::role($user);
			$name = $user->getName();
			$linked = self::ofUser($keys, $name);
			$brought = self::ofRole($keys, $role);

			$out[] = array(
				"user" => $who->of($name),
				"role" => $role,
				"direct" => self::chips($linked, $role, false),
				"viaRole" => self::chips($brought, $role, true),
			);
		}

		return $out;
	}

	/** ofRole is the instance SSH keys a role brings, in the folder's order. */
	static function ofRole($keys, $role) {
		$out = array();

		foreach ($keys as $key) {
			if ($key["purpose"] === Task\Key::SSH && $key["workspace"] === "" && in_array($role, $key["roles"], true)) {
				$out[] = $key;
			}
		}

		return $out;
	}

	/** ofUser is the instance SSH keys linked to one user by name. */
	static function ofUser($keys, $name) {
		$out = array();

		foreach ($keys as $key) {
			if ($key["purpose"] === Task\Key::SSH && $key["workspace"] === "" && in_array($name, $key["users"], true)) {
				$out[] = $key;
			}
		}

		return $out;
	}

	/**
	 * chips is a list of keys as the links a holder's row draws: a key of their
	 * own says its fingerprint, and one their role brings says the role.
	 */
	static function chips($keys, $role, $via) {
		$out = array();

		foreach ($keys as $key) {
			$name = $key["name"];
			$out[] = array(
				"name" => $name,
				"href" => Task\Route::href("/settings/keys/" . $name),
				"title" => $via ? $name . ", from the " . $role . " role" : $name . ", " . $key["fingerprint"],
			);
		}

		return $out;
	}

	/**
	 * folders are the .ssh folders of the checkouts that hold a private key of
	 * their own, which the alerts at the foot warn about: a run is given none of
	 * them. frontend/key.go:143.
	 */
	function folders() {
		$out = array();

		foreach (Task\Host::keyFolders() as $found) {
			$out[] = array(
				"project" => Task\Row::text($found, "project"),
				"dir" => Task\Row::text($found, "dir"),
				"keys" => implode(", ", Task\Row::items($found, "keys")),
			);
		}

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_keys;

	$mod->serve();
}
