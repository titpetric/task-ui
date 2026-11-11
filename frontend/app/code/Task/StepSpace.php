<?php

namespace Task;

// StepSpace is the workspace a step of the workflow editor picks before a
// folder: a workflow belongs to no workspace, so every step names the
// workspace of the folder it runs in.
//
// It holds what the editor picks from, read once a page: the folders with a
// job the reader sees, the root's host jobs, and the named workspaces with
// their titles and their order. Off those it answers every question a step
// card asks about a place: which workspace a folder is in, what the folder is
// called inside it, and what the picker shows of either.
//
// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.
class StepSpace {
	/** TITLE is what the picker shows for the default workspace with no title of its own. */
	const TITLE = "Internal";

	/** ROOT is what a picker shows for the root of a workspace. */
	const ROOT = "the root";

	/** folders are the folders with a job the reader sees, which a step picks from. */
	var $folders = array();

	/** hostJobs are the root's jobs that run a command on a host. */
	var $hostJobs = array();

	/** titles is the title of each workspace, by its folder, the default one too. */
	var $titles = array();

	/** workspaces is the folder of each named workspace, in the sidebar's order. */
	var $workspaces = array();

	/** ranks is the place of each named workspace in that order. */
	var $ranks = array();

	/** none reports whether a step may name no workspace at all. */
	var $none = false;

	/** spaces is the workspace picker's list, built once. */
	var $spaces = null;

	function __construct($chrome) {
		$store = new Workspace($chrome->db);

		$this->workspaces = $store->folders();
		$this->ranks = Workspace::ranked($this->workspaces);
		$this->titles = $store->titles();
		$this->hostJobs = HostJobs::of($chrome);
		$this->none = count(User::folders($chrome->user)) === 0;

		foreach ($chrome->folders() as $folder) {
			if (count(Row::items($folder, "jobs")) > 0) {
				$this->folders[] = $folder;
			}
		}
	}

	/**
	 * spaces is the workspace picker's list: the default workspace, then by
	 * rank each named one that holds a folder a step may pick.
	 *
	 * A workspace with no folder of its own in the list is not offered: there
	 * would be nothing to pick after it.
	 */
	function spaces() {
		if ($this->spaces !== null) {
			return $this->spaces;
		}

		$seen = array();

		foreach ($this->folders as $folder) {
			$ws = $this->spaceOf(Row::text($folder, "id"));

			if ($ws !== Route::INTERNAL) {
				$seen[$ws] = true;
			}
		}

		$named = array_keys($seen);
		$ranks = $this->ranks;

		usort($named, function($one, $other) use ($ranks) {
			return Folders::byRank($one, $other, $ranks);
		});

		$out = array(
			array("id" => Route::INTERNAL, "title" => $this->titleOf(Route::INTERNAL, self::TITLE)),
		);

		foreach ($named as $ws) {
			$out[] = array("id" => $ws, "title" => $this->titleOf($ws, $ws));
		}

		$this->spaces = $out;

		return $out;
	}

	/** titleOf is a workspace's own title, or the fallback where its row has none. */
	function titleOf($ws, $fallback) {
		if (isset($this->titles[$ws]) && $this->titles[$ws] !== "") {
			return $this->titles[$ws];
		}

		return $fallback;
	}

	/** spaceOf is the workspace a folder is in, the default one for anything outside a named one. */
	function spaceOf($id) {
		$parts = explode("/", $id);

		if (in_array($parts[0], $this->workspaces, true)) {
			return $parts[0];
		}

		return Route::INTERNAL;
	}

	/**
	 * rel is a folder relative to its workspace, the key a step keeps when it
	 * moves between workspaces: the ID itself in the default workspace, the
	 * rest after the workspace's folder in a named one, and "." for the root
	 * of either.
	 */
	function rel($id) {
		$ws = $this->spaceOf($id);

		if ($ws === Route::INTERNAL) {
			if ($id === Route::INTERNAL || $id === "") {
				return ".";
			}

			return $id;
		}

		$rest = substr($id, strlen($ws));

		if (substr($rest, 0, 1) === "/") {
			$rest = substr($rest, 1);
		}

		return $rest === "" ? "." : $rest;
	}

	/** label is what the folder picker shows of a folder ID: rel, the root for ".". */
	function label($id) {
		$rel = $this->rel($id);

		return $rel === "." ? self::ROOT : $rel;
	}

	/** folderIn is the folder ID a relative folder has in a workspace. */
	static function folderIn($ws, $rel) {
		if ($rel === "." && $ws === Route::INTERNAL) {
			return Route::INTERNAL;
		}

		if ($rel === ".") {
			return $ws;
		}

		if ($ws === Route::INTERNAL) {
			return $rel;
		}

		return $ws . "/" . $rel;
	}

	/** folder is the step folder of an ID, null where the reader has none such with a job. */
	function folder($id) {
		foreach ($this->folders as $folder) {
			if (Row::text($folder, "id") === $id && count(Row::items($folder, "jobs")) > 0) {
				return $folder;
			}
		}

		return null;
	}

	/**
	 * hasJob reports whether a folder holds a job a command step picks; the
	 * root's are its host jobs.
	 */
	function hasJob($id, $job) {
		if ($id === Route::INTERNAL) {
			foreach ($this->hostJobs as $one) {
				if ($one["id"] === $job) {
					return true;
				}
			}

			return false;
		}

		$folder = $this->folder($id);

		if ($folder === null) {
			return false;
		}

		foreach (Row::items($folder, "jobs") as $one) {
			if (Row::text($one, "id") === $job) {
				return true;
			}
		}

		return false;
	}

	/** title is what the workspace picker shows for a workspace. */
	function title($ws) {
		foreach ($this->spaces() as $space) {
			if ($space["id"] === $ws) {
				return $space["title"];
			}
		}

		return $ws;
	}

	/**
	 * has reports whether a workspace is one the steps may move to, which is
	 * what the query field of the editor is checked against.
	 */
	function has($ws) {
		foreach ($this->spaces() as $space) {
			if ($space["id"] === $ws) {
				return true;
			}
		}

		return false;
	}

	/**
	 * move takes the steps to a workspace by their relative folders, and says
	 * under each what it did not find there: steps and missed.
	 *
	 * A workflow step and a step with no folder move as they are. Anything
	 * that asks Claude, and a script, keep their place inside the workspace and
	 * fall back to its root; a command loses its job where the folder there has
	 * none such, so nothing runs a job the reader did not pick.
	 */
	function move($steps, $ws) {
		$title = $this->title($ws);
		$moved = array();
		$missed = 0;

		foreach ($steps as $step) {
			$step["missing"] = "";
			$kind = Step::kind($step);

			if ($kind === Step::WORKFLOW || $step["folder"] === "") {
				$moved[] = $step;

				continue;
			}

			$rel = $this->rel($step["folder"]);
			$to = self::folderIn($ws, $rel);
			$name = $this->label($step["folder"]);

			if ($kind !== Step::JOB) {
				if ($rel !== "." && $this->folder($to) === null) {
					$step["missing"] = $name . " is not in " . $title . ".";
					$to = self::folderIn($ws, ".");
				}

				$step["folder"] = $to;
			} elseif ($to !== Route::INTERNAL && $this->folder($to) === null) {
				$step["missing"] = $name . " is not in " . $title . ".";
				$step["folder"] = "";
				$step["job"] = "";
				$step["args"] = "";
			} elseif (!$this->hasJob($to, $step["job"])) {
				$step["missing"] = $name . " has no job " . $step["job"] . " in " . $title . ".";
				$step["folder"] = $to;
				$step["job"] = "";
				$step["args"] = "";
			} else {
				$step["folder"] = $to;
			}

			if ($step["missing"] !== "") {
				$missed++;
			}

			$moved[] = $step;
		}

		return array("steps" => $moved, "missed" => $missed);
	}

	/**
	 * moveItems is the edited workflow's menu items that take its steps to
	 * each other workspace, between two separators; none where there is only
	 * one workspace to be in.
	 */
	function moveItems($menu, $id) {
		$spaces = $this->spaces();

		if ($id === "" || count($spaces) < 2) {
			return $menu;
		}

		$menu->sep();

		foreach ($spaces as $space) {
			$href = "/workflows/" . $id . "/edit?workspace=" . Route::queryEscape($space["id"]);

			$menu->item("folder", "Move to " . $space["title"], array("href" => Route::href($href)));
		}

		return $menu->sep();
	}
}
