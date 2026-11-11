<?php

namespace Task;

// Project reads the project table: the checkouts the instance was told to
// keep, which is where a folder of the catalogue came from.
//
// The page that edits one is the Projects step of the settings. A project is
// a directory under a workspace's src, which the host walks; the row here is
// only what the instance keeps about it, and a directory somebody made by
// hand is a project with no row at all.
class Project {
	/** PAGE is the Projects step of the settings, which every project's page hangs under. */
	const PAGE = "/settings/projects";

	/**
	 * PULL is the job Pull now runs: the pull of the atkins skill git, as
	 * folder/skill.go spells a skill job's ID.
	 */
	const PULL = "@git:pull";

	/** store is the settings module, which owns the project table. */
	var $store;

	/** db is what the other modules' stores are asked on. */
	var $db;

	function __construct($db = null) {
		if ($db === null) {
			$db = DB::reader();
		}

		$this->db = $db;
		$this->store = new Settings\Projects($db);
	}

	/**
	 * says is what the page of a folder that is no folder reports, which is
	 * the sentence frontend/project.go writes into the shared 404.
	 */
	static function says($id) {
		return "No folder named " . $id . ".";
	}

	/**
	 * crumbs is the trail above a page of one project: the settings, the
	 * projects, the folder itself and the page being read.
	 */
	static function crumbs($id, $workspaces, $last) {
		$href = Route::folderHref($id, $workspaces);

		return array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Setup::PROJECTS, "href" => self::PAGE),
			array("text" => $id, "href" => $href),
			array("text" => $last, "href" => ""),
		);
	}

	/**
	 * hidden is the folder of every project set not to show in the sidebar.
	 *
	 * The folders inside one go with it, and its pages stay reachable: this is
	 * what the sidebar lists, not what the reader may see.
	 */
	function hidden() {
		return $this->store->hidden();
	}

	/**
	 * folders is the folder of every project row, which the workspaces count
	 * beside the directories the host walked.
	 *
	 * A row whose directory is not there yet is a project of its workspace all
	 * the same: the clone failed and the page lists it with git's output.
	 */
	function folders() {
		return $this->store->folders();
	}

	/**
	 * row is the project a folder names, null where no row does.
	 *
	 * A directory somebody made by hand has no row, and the page of it draws
	 * the settings of a project the instance keeps nothing about.
	 */
	function row($folder) {
		return $this->store->row($folder);
	}

	/**
	 * refs are the rows that name a folder or a folder inside it, each said as
	 * its kind, its folder and its name: what a move of the project leaves
	 * where it was.
	 *
	 * The select of the project's settings carries them, and the script warns
	 * with them before it moves anything: a schedule, a variable, an env file
	 * and a workflow's step each name a folder by its ID and none of them
	 * follows the checkout.
	 *
	 * Each module answers for the tables it owns, so this is four questions
	 * and not one statement over four tables. storage/folder_ref_storage.go.
	 */
	function refs($id) {
		$schedules = new Schedules\Schedules($this->db);
		$workflows = new Workflows\Workflows($this->db);
		$mine = new Settings\Refs($this->db);
		$found = array_merge($schedules->refs($id), $mine->of($id), $workflows->stepsIn($id));
		$out = array();

		foreach ($found as $row) {
			$kind = str_replace("_", " ", Row::text($row, "kind"));
			$out[] = $kind . " " . Row::text($row, "folder") . " · " . Row::text($row, "name");
		}

		sort($out);

		return $out;
	}

	/**
	 * moves is the workspace the project sits in, which the Workspace select
	 * of its settings stands on, and "" where the page draws no select.
	 *
	 * A named workspace itself does not move: it is not a project, and its own
	 * page is the one that edits it. Neither does a folder deeper inside a
	 * project: the directory is the project's and moves with it.
	 * frontend/project.go:moveChoices.
	 */
	static function moves($id, $named) {
		if (in_array($id, $named, true)) {
			return "";
		}

		$cut = strpos($id, "/");

		if ($cut === false) {
			return Route::INTERNAL;
		}

		$top = substr($id, 0, $cut);
		$rest = substr($id, $cut + 1);

		if (!in_array($top, $named, true) || strpos($rest, "/") !== false) {
			return "";
		}

		return $top;
	}

	/**
	 * groups is a block of the Projects page per workspace row, in the order
	 * the settings list them, each with the directories inside it.
	 *
	 * The default workspace's are the directories at the top of the root, the
	 * named workspaces' own left out: those are the Workspaces page's rows and
	 * not projects. A workspace with nothing in it keeps its block, which says
	 * so. frontend/view/project_group.go.
	 */
	static function groups($entries, $workspaces) {
		$out = array();
		$at = array();

		foreach ($workspaces as $folder) {
			if (isset($at[$folder])) {
				continue;
			}

			$at[$folder] = count($out);
			$out[] = array("workspace" => $folder, "entries" => array());
		}

		foreach ($entries as $entry) {
			$top = self::groupOf(Row::text($entry, "id"), $at);

			if (!isset($at[$top])) {
				continue;
			}

			$index = $at[$top];
			$out[$index]["entries"][] = $entry;
		}

		return $out;
	}

	/**
	 * groupOf is the workspace an entry is listed under: the one its ID starts
	 * with, the default workspace for a directory at the top of the root, and
	 * "" for a named workspace's own directory, which is no project.
	 */
	static function groupOf($id, $at) {
		$parts = explode("/", $id);

		if (count($parts) > 1) {
			return $parts[0];
		}

		if (isset($at[$parts[0]]) && $parts[0] !== Route::INTERNAL) {
			return "";
		}

		return Route::INTERNAL;
	}

	/** title is a block's heading: Projects for the root, the workspace's folder otherwise. */
	static function title($workspace) {
		if ($workspace === Route::INTERNAL) {
			return Setup::PROJECTS;
		}

		return $workspace;
	}

	/** note is what the heading says after the title: where the projects are. */
	static function note($workspace, $root) {
		if ($workspace === Route::INTERNAL) {
			return "in " . $root . "/src";
		}

		return "inside " . $workspace . "/src/";
	}

	/** empty is the line of a block with no project in it. */
	static function empty($workspace) {
		if ($workspace === Route::INTERNAL) {
			return "No projects yet: add one from a git repository, or put a folder with jobs under src.";
		}

		return "No projects in " . $workspace . " yet.";
	}

	/**
	 * block is the head of one group's block: its title, where its projects
	 * are, the line for none, and Add project for a named workspace.
	 *
	 * The root's block offers none: the page's own button adds there, and the
	 * form defaults to the first workspace the select lists.
	 */
	static function block($chrome, $group, $root) {
		$workspace = $group["workspace"];
		$buttons = new Buttons(true);
		$found = count($group["entries"]);
		$adds = $workspace !== Route::INTERNAL && $chrome->allowed("project/edit");

		if ($adds) {
			$href = Route::href(self::PAGE . "/new") . "?workspace=" . Route::queryEscape($workspace);

			$buttons->item("plus", "Add project", array("href" => $href));
		}

		return array(
			"title" => self::title($workspace),
			"note" => self::note($workspace, $root),
			"empty" => $found === 0 ? self::empty($workspace) : "",
			"buttons" => $buttons->data(),
		);
	}

	/**
	 * table is the projects table of the Projects page and of a workspace's,
	 * the same rows on both, as components/project-table.tpl reads it.
	 *
	 * sized says the page measured the directories: the Size column is drawn
	 * where a measure came back for them, and an instance with no meter draws
	 * no column at all rather than an empty one.
	 */
	static function table($chrome, $entries, $spaces, $next) {
		$rows = array();

		foreach ($entries as $entry) {
			$rows[] = self::entry($chrome, $entry, $spaces, $next);
		}

		return array("sized" => count($spaces) > 0, "rows" => $rows);
	}

	/**
	 * entry is one project of the table: the folder it links to, what its
	 * pipeline reads, the checkout it is, what it takes on disk and its
	 * [...].
	 */
	static function entry($chrome, $entry, $spaces, $next) {
		$id = Row::text($entry, "id");
		$dir = Row::text($entry, "dir");
		$space = isset($spaces[$dir]) ? $spaces[$dir] : array();
		$pull = self::pull($chrome, $id);
		$found = Row::flag($entry, "checkout");
		$when = Row::text($entry, "when");

		return array(
			"id" => $id,
			"href" => Route::href(Route::folderHref($id, $chrome->workspaces())),
			"bad" => Row::text($entry, "error") !== "",
			"sources" => implode(", ", Row::items($entry, "sources")),
			"jobs" => Row::number($entry, "jobs"),
			"checkout" => $found,
			"remote" => Row::text($entry, "remote"),
			"branch" => Row::text($entry, "branch"),
			"commit" => Row::text($entry, "commit"),
			"subject" => Row::text($entry, "subject"),
			"ago" => $when === "" ? "" : Clock::ago($when),
			"space" => Size::mini($space),
			"menu" => self::menu($chrome, $entry, $pull, $next),
		);
	}

	/**
	 * menu is a row's [...]: the folder, its settings, a pull of its checkout
	 * and the removal of it, each behind the reader's own permission.
	 *
	 * Remove checkout reloads the list it stands in rather than going
	 * anywhere, which is what a row's delete does: the row goes and the page
	 * stays.
	 */
	static function menu($chrome, $entry, $pull, $next) {
		$id = Row::text($entry, "id");
		$menu = new Menu("Actions on " . $id);
		$href = Route::folderHref($id, $chrome->workspaces());

		$menu->item("folder-open", "Open", array("href" => Route::href($href)));
		$menu->item("settings", "Edit project", array("href" => Route::href(self::PAGE . "/" . $id)));

		if ($pull !== null) {
			$menu->item("download", "Pull now", self::pulls($id, $pull));
		}

		if (Row::flag($entry, "checkout") && $chrome->allowed("project/delete")) {
			$menu->danger("trash", "Remove checkout", self::removes($entry, $next));
		}

		return $menu->data();
	}

	/** pulls is what Pull now carries: the run of the git skill's pull job. */
	static function pulls($id, $pull) {
		return array(
			"data-run" => "",
			"data-folder" => $id,
			"data-dir" => Row::text($pull, "dir"),
			"data-job" => Row::text($pull, "id"),
			"data-title" => $id . " / " . Row::text($pull, "name"),
		);
	}

	/**
	 * removes is what Remove checkout carries: the folder it deletes, what it
	 * asks first, and where it goes after.
	 *
	 * The delete is of the API at the root of the instance and carries no
	 * mount. next is reload in a list, which redraws the page the row was in.
	 */
	static function removes($entry, $next) {
		$id = Row::text($entry, "id");

		return array(
			"data-delete" => "/api/projects/" . $id,
			"data-confirm" => Place::removeConfirm($id, Row::text($entry, "remote")),
			"data-confirm-ok" => "Remove checkout",
			"data-next" => $next,
		);
	}

	/**
	 * pull is the pull job of a folder's git skill, null where the skill is
	 * not applied there, holds no such job, or the reader may not run it.
	 *
	 * The folder's own allow list comes back with the job, because the
	 * catalogue folder behind a listed directory is not the page's to hold.
	 */
	static function pull($chrome, $id) {
		$job = Host::job($id, "", self::PULL);

		if (!Row::flag($job, "found")) {
			return null;
		}

		$place = array("id" => $id, "allow" => Row::items($job, "folder_allow"));

		if (!Folders::allows($chrome->access, $place, $job)) {
			return null;
		}

		return $job;
	}
}
