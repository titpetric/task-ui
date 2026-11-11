<?php

namespace Task;

// Place is what a page says about the folder or the directory it is about,
// beyond the jobs and the files: what kind of place it is, the actions on it,
// the checkout it sits in and the rule a run of it is allowed under.
//
// A folder of the catalogue is one of three things, and the three draw
// different actions: a named workspace, a project inside one or at the root,
// or a folder deeper than that, which pins and nothing else. The depth of the
// ID decides, against the named workspaces.
class Place {
	/** WORKSPACES, PROJECTS, SECRETS and SCHEDULES are the pages the actions open. */
	const WORKSPACES = "/settings/workspaces/";
	const PROJECTS = "/settings/projects";
	const SECRETS = "/settings/secrets";
	const SCHEDULES = "/schedules";

	/** GIT is the kind of the Jobs table group of the git skill, which pulls. */
	const GIT = "@git";

	/** id is the folder's, internal for the root. */
	var $id = "";

	/** workspace marks a named workspace's page, project a project's. */
	var $workspace = false;
	var $project = false;

	/**
	 * projects is how many projects a workspace's page lists: none puts Add
	 * project in the head as the page's own button.
	 */
	var $projects = 0;

	/** pull is the pull job of the git skill of the place, null without one. */
	var $pull = null;

	/** allow is who the place lets run its jobs, beyond admin and runner. */
	var $allow = array();

	/**
	 * of is the place a folder's ID names, read against the named workspaces:
	 * a workspace at the top, a project one step under a named one or at the
	 * root, and neither below that.
	 */
	static function of($id, $workspaces) {
		$place = new Place;
		$place->id = $id;
		$parts = explode("/", $id);
		$named = in_array($parts[0], $workspaces, true);
		$depth = count($parts) - 1;

		if ($named && $depth === 0) {
			$place->workspace = true;
		} elseif (($named && $depth === 1) || (!$named && $depth === 0)) {
			$place->project = true;
		}

		return $place;
	}

	/**
	 * root is the place the root's own page is about, which of() cannot read
	 * off an ID.
	 *
	 * internal names no row of the workspace table, so the depth of it reads as
	 * a project at the root; the root is the default workspace itself, and its
	 * page draws a workspace's actions.
	 */
	static function root() {
		$place = new Place;
		$place->id = Route::INTERNAL;
		$place->workspace = true;

		return $place;
	}

	/** kind names what the place is a folder of, which is what a pin records. */
	function kind() {
		if ($this->workspace) {
			return "workspace";
		}

		if ($this->project) {
			return "project";
		}

		return "folder";
	}

	/** settings is the page that edits the workspace or the project, "" for neither. */
	function settings() {
		if ($this->workspace) {
			return Route::href(self::WORKSPACES . $this->id);
		}

		if ($this->project) {
			return Route::href(self::PROJECTS . "/" . $this->id);
		}

		return "";
	}

	/** keysPage is the page of a workspace's SSH keys. */
	function keysPage() {
		return Route::href(self::WORKSPACES . $this->id . "/keys");
	}

	/** secretsPage is the secrets page, narrowed to the workspace. */
	function secretsPage() {
		return Route::href(self::SECRETS) . "?workspace=" . Route::queryEscape($this->id);
	}

	/** addProjectPage is the form that clones a project into the workspace. */
	function addProjectPage() {
		return Route::href(self::PROJECTS . "/new") . "?workspace=" . Route::queryEscape($this->id);
	}

	/** schedulesPage is the schedules page, narrowed to the place. */
	function schedulesPage() {
		return Route::href(self::SCHEDULES) . "?folder=" . Route::queryEscape($this->id);
	}

	/**
	 * menu is the [...] of the head of a place's page: the settings of a
	 * workspace or a project, a workspace's keys and secrets, the pin, a
	 * schedule, a pull of the checkout and the removal of it, each behind the
	 * reader's own permission.
	 *
	 * title is what the page calls the place, checkout what git answered of
	 * the directory, and root says the page is the root's own: the root pins
	 * nothing and schedules nothing, because its jobs are the instance's and
	 * not a project's.
	 */
	function menu($chrome, $title, $pins, $checkout, $root) {
		$menu = new Menu("Actions on " . $title);
		$edits = $chrome->allowed("project/edit");
		$reads = $chrome->allowed("workspace/view");

		$menu->when($this->workspace && $edits, "plus", "Add project", array("href" => $this->addProjectPage()));
		$menu->when($this->workspace && $reads, "settings", "Edit workspace", array("href" => $this->settings()));
		$menu->when($this->workspace && $reads, "key", "SSH keys", array("href" => $this->keysPage()));
		$menu->when($this->workspace && $chrome->allowed("secret/view"), "secret", "Secrets", array("href" => $this->secretsPage()));
		$menu->when($this->project && $chrome->allowed("project/view"), "settings", "Edit project", array("href" => $this->settings()));

		if (!$root && $chrome->user !== null) {
			Pin::item($menu, $pins, $this->id, "", "", $this->kind());
		}

		$placed = $this->workspace || $this->project;

		$menu->when(!$root && $placed && $chrome->allowed("schedule/edit"), "clock", "Add schedule", array("href" => $this->schedulesPage()));

		if ($this->pull !== null && Folders::allows($chrome->access, $this->place(), $this->pull)) {
			$menu->item("download", "Pull now", $this->pulls());
		}

		if ($this->project && Row::flag($checkout, "found") && $chrome->allowed("project/delete")) {
			$menu->danger("trash", "Remove checkout", $this->removes($checkout));
		}

		return $menu->data();
	}

	/**
	 * place is the place as a run of it is gated: the ID and the allow list
	 * that decides who may run its jobs.
	 */
	function place() {
		return array("id" => $this->id, "allow" => $this->allow);
	}

	/** pulls is what Pull now carries: the run of the git skill's pull job. */
	function pulls() {
		return array(
			"data-run" => "",
			"data-folder" => $this->id,
			"data-dir" => Row::text($this->pull, "dir"),
			"data-job" => Row::text($this->pull, "id"),
			"data-title" => $this->id . " / " . Row::text($this->pull, "name"),
		);
	}

	/**
	 * removes is what Remove checkout carries: the folder it deletes, what it
	 * asks first, and the page it goes to after.
	 *
	 * The delete is of the API at the root of the instance and carries no
	 * mount; the page it lands on is of this application and carries one.
	 */
	function removes($checkout) {
		$said = "Removed " . $this->id;

		return array(
			"data-delete" => "/api/projects/" . $this->id,
			"data-confirm" => self::removeConfirm($this->id, Row::text($checkout, "remote")),
			"data-confirm-ok" => "Remove checkout",
			"data-next" => Route::href(self::PROJECTS) . "?notice=" . Route::queryEscape($said),
		);
	}

	/**
	 * buttons is the row on the right of the head: Add project, which a
	 * workspace with no project in it yet offers as the page's own action.
	 */
	function buttons($chrome) {
		$buttons = new Buttons;

		if ($this->workspace && $this->projects === 0 && $chrome->allowed("project/edit")) {
			$buttons->primary("plus", "Add project", array("href" => $this->addProjectPage()));
		}

		return $buttons->data();
	}

	/** removeConfirm is what Remove checkout asks before it removes the folder. */
	static function removeConfirm($id, $remote) {
		if ($remote === "") {
			return "Remove the folder " . $id . " and its checkout? The checkout has no remote to clone it again from.";
		}

		return "Remove the folder " . $id . " and its checkout? You can clone it again from " . $remote . ".";
	}

	/**
	 * pull is the pull job of the git skill of a place, from the Jobs table
	 * groups the page asked for anyway, null where the skill is not applied
	 * there or holds no such job.
	 */
	static function pull($groups) {
		foreach ($groups as $group) {
			if (Row::text($group, "kind") !== self::GIT) {
				continue;
			}

			foreach (Row::items($group, "jobs") as $job) {
				if (Row::text($job, "name") === "pull") {
					return $job;
				}
			}
		}

		return null;
	}

	/**
	 * git is the line about the checkout a place sits in, as
	 * components/git-line.tpl reads it, null for a place in none.
	 */
	static function git($status) {
		if (!Row::flag($status, "found")) {
			return null;
		}

		$checkout = Row::text($status, "checkout");
		$upstream = Row::text($status, "upstream");
		$ahead = Row::number($status, "ahead");
		$behind = Row::number($status, "behind");
		$dirty = Row::number($status, "dirty");
		$title = "The checkout " . ($checkout === "" ? "at the root" : $checkout);

		if ($upstream !== "") {
			$title .= ", tracking " . $upstream;
		}

		return array(
			"title" => $title,
			"branch" => Row::text($status, "branch"),
			"detached" => Row::text($status, "detached"),
			"upstream" => $upstream,
			"ahead" => $ahead,
			"behind" => $behind,
			"dirty" => $dirty,
			"files" => self::files($dirty),
			"clean" => $upstream !== "" && $ahead === 0 && $behind === 0 && $dirty === 0,
			"where" => $checkout === "" ? "the root" : $checkout,
			"changes" => Row::items($status, "changes"),
		);
	}

	/** files words how many files of a checkout git reports as changed. */
	static function files($dirty) {
		if ($dirty === 1) {
			return "1 file changed or untracked";
		}

		return $dirty . " files changed or untracked";
	}

	/**
	 * source is the button that opens a file a place's pipeline reads, as
	 * components/show-source.tpl reads it.
	 */
	static function source($folder, $file, $lines) {
		$label = $lines === 1 ? "Show 1 line" : "Show " . $lines . " lines";

		return array(
			"folder" => $folder,
			"file" => $file,
			"label" => $label,
		);
	}

	/**
	 * hours is the rule a run of a place is allowed under, as
	 * components/hours.tpl reads it: the line it prints as, and whether it is
	 * open at the moment of the reading.
	 *
	 * A place with no rule is open always and says so, which is what the empty
	 * text means; so does one whose rule prints as nothing at all.
	 */
	static function hours($text) {
		$read = Host::hours($text);

		return array(
			"text" => $text,
			"line" => Row::text($read, "line"),
			"open" => Row::flag($read, "open"),
		);
	}

	/** who is who may run the jobs of a place, which a runner always may. */
	static function who($allow) {
		if (count($allow) === 0) {
			return "admin, runner";
		}

		return "admin, " . implode(", ", $allow);
	}
}
