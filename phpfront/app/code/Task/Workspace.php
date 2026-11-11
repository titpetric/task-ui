<?php

namespace Task;

// Workspace reads the workspace table: the named workspaces whose folders the
// sidebar lists, in the order it lists them.
//
// The default workspace is not one of them. Its projects are the root's, so a
// folder of it has no workspace of its own in a link, and every reader here
// leaves it out.
class Workspace {
	/** SCOPE starts the scope of a variable or a key of one workspace. */
	const SCOPE = "workspace:";

	/** EVERY is the scope of a variable every workspace's runs are given. */
	const EVERY = "all";

	/** PAGE is the Workspaces step of the settings, which every workspace's page hangs under. */
	const PAGE = "/settings/workspaces";

	/** store is the settings module, which owns the table. */
	var $store;

	function __construct($db = null) {
		$this->store = new Settings\Workspaces($db);
	}

	/**
	 * says is what the page of a workspace that is not there reports, which is
	 * the sentence frontend/workspace.go writes into the shared 404.
	 */
	static function says($folder) {
		return "No workspace named " . $folder . ".";
	}

	/**
	 * crumbs is the trail above a page of one workspace: the settings, the
	 * workspaces, the workspace itself and the page being read.
	 *
	 * The default workspace is Internal in a crumb and internal everywhere
	 * else, which Route::crumbName says; $last has no link, as the last crumb
	 * of any trail has none.
	 */
	static function crumbs($folder, $last) {
		return array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Setup::WORKSPACES, "href" => self::PAGE),
			array("text" => Route::crumbName($folder), "href" => self::PAGE . "/" . $folder),
			array("text" => $last, "href" => ""),
		);
	}

	/** folders is the folder of every named workspace, in the sidebar's order. */
	function folders() {
		return $this->store->folders();
	}

	/**
	 * names is the folder of every workspace row, the default one first and
	 * the named ones in the sidebar's order after it.
	 *
	 * It is folders() with the default workspace, which the project walk reads:
	 * the projects under the root belong to a workspace row as much as a named
	 * workspace's do, and the settings count the rows themselves.
	 */
	function names() {
		return $this->store->names();
	}

	/**
	 * all is every workspace row, the default one first and the named ones in
	 * the sidebar's order after it, which is storage/workspace_storage.go:69.
	 */
	function all() {
		return $this->store->all();
	}

	/**
	 * of is the folder of the workspace a scope belongs to, "" for the scope
	 * all, which is a variable of every workspace. secret/workspaces.go:26.
	 *
	 * folders is the folder of every workspace row, which the caller read.
	 */
	static function of($scope, $folders) {
		if (strpos($scope, self::SCOPE) === 0) {
			return Route::canonical(substr($scope, strlen(self::SCOPE)));
		}

		if ($scope === self::EVERY) {
			return "";
		}

		$parts = explode("/", Route::canonical($scope));
		$top = $parts[0];

		if ($top === Route::INTERNAL || in_array($top, $folders, true)) {
			return $top;
		}

		return Route::INTERNAL;
	}

	/**
	 * row is the workspace a folder names, null where none does.
	 *
	 * The root reads its own, which no other reader here does: whether a row
	 * names the root decides what its page is titled, where its directories
	 * are listed from, and whether it draws a home and a prompt at all. An
	 * instance that never had one is a real instance, and its root page is the
	 * shorter one.
	 */
	function row($folder) {
		return $this->store->row($folder);
	}

	/**
	 * homes is the repository of each workspace's home, by its folder, from the
	 * rows the page already read.
	 *
	 * A workspace with no home repository has an empty URL and its page says
	 * the home is an empty directory; the Workspaces table says empty.
	 */
	static function homes($rows) {
		$out = array();

		foreach ($rows as $row) {
			$out[$row->getFolder()] = $row->getHomeUrl();
		}

		return $out;
	}

	/**
	 * split is the workspace a folder ID belongs to and the ID's path under it,
	 * "" for the workspace itself, which is layout.Split.
	 *
	 * An ID whose first segment names no workspace belongs to the default one,
	 * where the whole ID is the path: the projects under the root are the
	 * default workspace's.
	 */
	static function split($named, $id) {
		$id = trim($id, "/");

		if ($id === "" || $id === Route::INTERNAL) {
			return array(Route::INTERNAL, "");
		}

		$at = strpos($id, "/");

		if ($at === false) {
			return in_array($id, $named, true) ? array($id, "") : array(Route::INTERNAL, $id);
		}

		$top = substr($id, 0, $at);

		if (in_array($top, $named, true)) {
			return array($top, substr($id, $at + 1));
		}

		return array(Route::INTERNAL, $id);
	}

	/**
	 * dir is the directory of a workspace under the root, which its settings
	 * print: the root itself for the default one, which is no directory of its
	 * own. model.Workspace.Dir.
	 */
	static function dir($root, $folder) {
		if ($folder === Route::INTERNAL) {
			return $root;
		}

		return $root . "/" . $folder;
	}

	/**
	 * homeId is the folder ID of a workspace's home, the HOME of its runs:
	 * home at the root and <folder>/home inside a named workspace. layout.HomeID.
	 */
	static function homeId($folder) {
		if ($folder === "" || $folder === Route::INTERNAL) {
			return Folders::HOME;
		}

		return $folder . "/" . Folders::HOME;
	}

	/**
	 * inside is the projects of one workspace out of the whole walk, which its
	 * page lists as the Projects page lists them.
	 *
	 * A named workspace's are the directories one step under its src, which is
	 * what an ID of two segments starting with its folder is. The default
	 * workspace's are the directories at the top of the root, less the named
	 * workspaces themselves: those are rows of the Workspaces page and no
	 * project of anything. frontend/workspace.go:workspaceProjects.
	 */
	static function inside($entries, $folder, $named) {
		$out = array();

		foreach ($entries as $entry) {
			$parts = explode("/", Row::text($entry, "id"));
			$top = $parts[0];
			$deeper = count($parts) > 1;

			if ($folder === Route::INTERNAL && !$deeper && !in_array($top, $named, true)) {
				$out[] = $entry;

				continue;
			}

			if ($deeper && $top === $folder) {
				$out[] = $entry;
			}
		}

		return $out;
	}

	/**
	 * nested is the projects inside each workspace, by folder: the directories
	 * the walk found under its src and the project rows there, the direct ones,
	 * each once, in name order.
	 *
	 * They are what the Projects column counts and what a delete of the
	 * workspace takes with it. The default workspace holds the root's. A row
	 * whose clone failed has no directory and is counted all the same, which is
	 * why the rows are read beside the walk.
	 *
	 * entries is the walk as Task\Host::entries() answers it, projects the
	 * folder of every project row, and named the folders of the named
	 * workspaces.
	 */
	static function nested($entries, $projects, $named) {
		$out = array();

		foreach ($entries as $entry) {
			self::hold($out, $named, Row::text($entry, "id"));
		}

		foreach ($projects as $folder) {
			$parts = explode("/", $folder);

			if (count($parts) === 1 || in_array($parts[0], $named, true)) {
				self::hold($out, $named, $folder);
			}
		}

		foreach ($out as $folder => $ids) {
			usort($ids, "strcmp");
			$out[$folder] = $ids;
		}

		return $out;
	}

	/**
	 * hold adds a folder ID to the projects of the workspace that holds it,
	 * where the ID names a project of one at all.
	 *
	 * A workspace itself is no project of itself, nor is a directory deeper
	 * than a project, nor one under a name task-ui keeps for the root: those
	 * are the three the ID is refused for, and an ID already counted is not
	 * counted twice.
	 */
	static function hold(&$out, $named, $id) {
		$found = self::split($named, $id);
		$folder = $found[0];
		$rel = $found[1];

		if ($rel === "" || strpos($rel, "/") !== false || Route::reserved($rel)) {
			return;
		}

		if (isset($out[$folder]) && in_array($id, $out[$folder], true)) {
			return;
		}

		$out[$folder][] = $id;
	}

	/**
	 * service reports whether a scope is a service's own settings rather than
	 * somebody's variables, which is secret.Service: a scope with a colon in it
	 * that does not name a workspace.
	 */
	static function service($scope) {
		if (strpos($scope, ":") === false) {
			return false;
		}

		return strpos($scope, self::SCOPE) !== 0;
	}

	/**
	 * counts is how many variables reach each workspace, by its folder: its
	 * shared ones and its folders', the integrations' settings apart.
	 *
	 * It is secret.Store.Encrypted counted per workspace, which the Workspaces
	 * table draws and a delete of a workspace is refused for: the variables go
	 * first, by hand, because nothing here removes somebody's credential as a
	 * side effect of removing a directory.
	 *
	 * The variables of every workspace belong to none of them and are counted
	 * under no folder, which is what the empty name means.
	 */
	static function counts($scopes, $folders) {
		$out = array();

		foreach ($scopes as $scope) {
			if (self::service($scope)) {
				continue;
			}

			$name = self::of($scope, $folders);

			if ($name === "") {
				continue;
			}

			$found = isset($out[$name]) ? $out[$name] : 0;
			$out[$name] = $found + 1;
		}

		return $out;
	}

	/** blocked is the menu line of a workspace whose variables stop its delete. */
	static function blocked($variables) {
		if ($variables === 1) {
			return "Delete its 1 variable first";
		}

		return "Delete its " . $variables . " variables first";
	}

	/**
	 * confirm is what Delete of a workspace asks before it removes it, and it
	 * names the projects that go with it.
	 *
	 * The default workspace is the root: its home and the projects at the root
	 * go, and the instance keeps running without a workspace at all, which Add
	 * project then needs one of again.
	 */
	static function confirm($folder, $nested) {
		$named = implode(", ", $nested);

		if ($folder === Route::INTERNAL) {
			$at = count($nested) > 0 ? " and the projects at the root with their checkouts: " . $named : "";

			return "Delete the default workspace " . $folder . " with its keys, its home" . $at . "? The root's built-in jobs keep running; Add project then needs a workspace.";
		}

		$projects = count($nested) > 0 ? ", and the projects inside it with their checkouts: " . $named : "";

		return "Delete the workspace " . $folder . " with its keys and secrets" . $projects . "? Its directory goes once nothing else is in it.";
	}

	/**
	 * titles is what each workspace is called, by its folder.
	 *
	 * The default workspace is, unlike folders(): the workflow editor's
	 * workspace picker shows every step's workspace, the root among them, and
	 * it is the row's title that names it where the row carries one.
	 */
	function titles() {
		return $this->store->titles();
	}

	/**
	 * limits is how many projects the sidebar lists for a workspace that
	 * limits them, by folder; a workspace that lists every one is not here.
	 *
	 * The default workspace is, unlike everywhere else in this class: its
	 * projects are the root's and its row carries their limit.
	 */
	function limits() {
		return $this->store->limits();
	}

	/**
	 * ranked is the rank of each named workspace by its folder, which the body
	 * carries for the script that orders what it draws.
	 *
	 * A rank is a workspace's place in the order folders() answers, so the
	 * ranking is handed the list the caller already read rather than reading
	 * the table a second time.
	 */
	static function ranked($folders) {
		$out = array();

		foreach ($folders as $rank => $folder) {
			$out[$folder] = $rank;
		}

		return $out;
	}
}
