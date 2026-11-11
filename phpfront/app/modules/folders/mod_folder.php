<?php

// @route GET /f/*

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_folder is everything under /f: a folder of the catalogue, a directory
// inside one with nothing of its own to run, and a job of either.
//
// The three are one route because the path alone does not say which: folders
// nest, so the longest folder that starts the path is the one meant, and what
// is left over is a directory, a job, a file that runs or nothing at all. The
// host resolves each of those and this decides between them in that order.
//
// A path starts with its workspace: a named one, or internal for the projects
// at the root. /f/internal is the root itself, which is the one page here that
// is not a folder of the catalogue at all: it is the tree the catalogue was
// discovered from, and its page lists the projects in it.
class mod_folder implements Module {
	/** ROOT is what the Files panel of the root's page is titled by. */
	const ROOT = "the root";

	/** HOME words the home of a workspace, which is the $HOME of every run in it. */
	const HOME = "/, the \$HOME of every run in the workspace";

	/** INTERNAL titles the root on an instance whose workspace table has no row for it. */
	const INTERNAL = "Internal";

	var $chrome;
	var $view;
	var $workspaces;

	function name() {
		return "folders";
	}

	function title() {
		return "Folders";
	}

	function icon() {
		return "folder";
	}

	function tables() {
		return array(
			"workspace",
			"schedule",
			"pin",
			"run",
			"run_summary",
			"note",
			"user",
			"key_info",
			"key_role",
			"key_user",
		);
	}

	/**
	 * serve resolves the path into one of the three pages, in the order
	 * frontend/folder.go resolves it.
	 *
	 * The path is read against the named workspaces first, because that is
	 * what says whether its first segment is a workspace or a project of the
	 * root; a path that names neither moves under the root, for good, and so
	 * does one that spells the name the root used to have.
	 */
	function serve() {
		$this->chrome = new Task\Chrome;
		$this->view = new Task\View($this, $this->chrome);
		$this->workspaces = $this->chrome->workspaces();

		$segments = Task\Route::segments();
		$path = implode("/", array_slice($segments, 1));
		$canonical = Task\Route::canonical($path);

		if ($canonical !== $path) {
			Task\Route::moved("/f/" . Task\Route::escapePath($canonical));

			return;
		}

		$cut = strpos($path, "/");
		$top = $cut === false ? $path : substr($path, 0, $cut);
		$rest = $cut === false ? "" : substr($path, $cut + 1);

		if ($top === Task\Route::INTERNAL && $rest === "") {
			$this->root();

			return;
		}

		if ($top === Task\Route::INTERNAL) {
			$inner = explode("/", $rest);

			if (in_array($inner[0], $this->workspaces, true) || Task\Route::reserved($inner[0])) {
				$this->view->missing("No folder named " . $path . ".");

				return;
			}

			$path = $rest;
		} elseif (!in_array($top, $this->workspaces, true)) {
			Task\Route::moved("/f/" . Task\Route::INTERNAL . "/" . Task\Route::escapePath($path));

			return;
		}

		$this->at($path);
	}

	/**
	 * at serves the path of a folder, its workspace's prefix taken off for one
	 * at the root.
	 *
	 * The path resolves against the catalogue as the host discovered it and
	 * not against the folders the reader sees: a folder they may not open is
	 * not the shorter one above it, so resolving through what they see would
	 * open the parent instead of refusing the child.
	 *
	 * A bare name no folder holds is one of the root's own jobs, which the
	 * root folder declares and no listed folder holds.
	 */
	function at($path) {
		$folders = $this->chrome->all();
		$folder = Task\Folders::resolve($folders, $path);
		$rest = "";

		if ($folder === null && strpos($path, "/") === false) {
			$root = Task\Host::job(Task\Route::INTERNAL, "", $path);

			if (Task\Row::flag($root, "found")) {
				$folder = Task\Folders::bare(Task\Route::INTERNAL);
				$rest = $path;
			}
		}

		if ($folder === null) {
			$this->view->missing("No folder named " . $path . ".");

			return;
		}

		$id = Task\Row::text($folder, "id");

		if (!Task\User::sees($this->chrome->user, $id)) {
			$this->view->missing("No folder named " . $path . ".");

			return;
		}

		if ($rest === "") {
			$rest = strlen($path) > strlen($id) ? substr($path, strlen($id) + 1) : "";
		}

		if ($rest === "") {
			$this->folder($id);

			return;
		}

		$bare = strpos($rest, "/") === false;
		$own = $bare ? Task\Host::job($id, "", $rest) : array();

		if ($bare && Task\Row::flag($own, "found")) {
			$this->job($id, "", $rest, $own);

			return;
		}

		if (Task\Row::flag(Task\Host::place($id, $rest), "found")) {
			$this->browse($id, $rest);

			return;
		}

		$file = "./" . preg_replace("#^\\./#", "", $rest);
		$held = Task\Host::job($id, "", $file);

		if (Task\Row::flag($held, "found")) {
			$this->job($id, "", $file, $held);

			return;
		}

		if (!$bare) {
			$cut = strrpos($rest, "/");
			$dir = substr($rest, 0, $cut);
			$name = substr($rest, $cut + 1);
			$inside = Task\Host::job($id, $dir, $name);

			if (Task\Row::flag($inside, "found")) {
				$this->job($id, $dir, $name, $inside);

				return;
			}
		}

		if ($bare) {
			$this->job($id, "", $rest, $own);

			return;
		}

		// The path the line names is the one that was asked for and not the one
		// that resolved: a folder of the root is read under the root's prefix
		// and that is what the reader typed.
		$this->view->missing("There is nothing at /" . Task\Route::path() . ".");
	}

	/**
	 * root is the page of the root itself: the projects in it the reader may
	 * see, with project/view what the tree holds, and the workspace's home and
	 * prompt where a row names the root at all.
	 *
	 * The root is not a folder of the catalogue, so almost nothing a folder's
	 * page draws is here: it has no checkout of its own, no pipeline file, no
	 * Jobs table and no Folder panel. What it does draw is the projects it
	 * holds, which is the one thing the catalogue cannot say, because a
	 * checkout with no pipeline file is in no catalogue.
	 */
	function root() {
		$id = Task\Route::INTERNAL;
		$store = new Task\Workspace($this->chrome->db);
		$row = $store->row($id);
		$named = $row !== null;
		$title = $named ? Task\Row::text($row, "title") : self::INTERNAL;
		$description = $named ? Task\Row::text($row, "description") : "";

		$this->chrome->heading($title, Task\Route::crumbs($id, "", $this->workspaces));
		$this->chrome->at($id);

		$folders = $this->chrome->folders();
		$subs = $this->projects($folders);
		$head = new Task\Head($title);

		$head->said($description);
		$head->icon("folder-open");
		$head->always();

		// A root no row names is a workspace nothing may be done to, so its
		// head holds neither the button nor the menu: there is no workspace to
		// add a project to, to edit, or to keep a key or a secret of.
		if ($named) {
			$place = Task\Place::root();
			$place->projects = count($subs);

			$head->buttons($place->buttons($this->chrome));
			$head->menu($place->menu($this->chrome, $title, array(), array(), true));
		}

		$tree = $this->chrome->allowed("project/view") ? Task\Host::rootTree($named, Task\Folders::ids($folders)) : array();
		$files = Task\Files::root($tree, $this->workspaces, $subs, $this->named($subs));
		$home = $named ? $this->home($id, $folders) : null;
		$prompt = null;

		if ($named) {
			$notes = new Task\Seen($this->chrome);
			$prompt = Task\Prompt::ofRoot($this->chrome, $notes);
		}

		$scope = array(
			"head" => $head->data(),
			"findings" => Task\Findings::alert($folders, $id, $this->workspaces),
			"git" => null,
			"error" => "",
			"control" => array(),
			"nothing" => false,
			"prompt" => $prompt,
			"files" => count($files["rows"]) === 0 ? null : array("note" => self::ROOT, "panel" => $files),
			"home" => $home,
			"detail" => null,
		);

		$this->view->page($scope, "folder.tpl");
	}

	/**
	 * projects is the folders of the root the page lists: the ones at the top
	 * that are no named workspace and no name task-ui keeps for itself.
	 *
	 * A folder deeper than the top is a folder of a project and not a project,
	 * and the page lists one row per project.
	 */
	function projects($folders) {
		$out = array();

		foreach ($folders as $folder) {
			$id = Task\Row::text($folder, "id");

			if (strpos($id, "/") !== false) {
				continue;
			}

			if (in_array($id, $this->workspaces, true) || Task\Folders::kept($id)) {
				continue;
			}

			$out[] = $folder;
		}

		return $out;
	}

	/**
	 * named is every ID a row of the root's Files panel may link by name: the
	 * projects the page lists and the named workspaces.
	 *
	 * A directory of the tree is linked by the ID its name would be discovered
	 * under, which is how a workspace's directory opens its page even though
	 * the workspace is no folder inside the root.
	 */
	function named($subs) {
		$out = array();

		foreach ($subs as $sub) {
			$out[Task\Row::text($sub, "id")] = true;
		}

		foreach ($this->workspaces as $workspace) {
			$out[$workspace] = true;
		}

		return $out;
	}

	/**
	 * home is the Files panel of a workspace's home, the $HOME of every run in
	 * it, null where the reader does not see it.
	 *
	 * Every path is written under ~/, because that is the name a run knows the
	 * home by. A directory inside it that is a folder of the catalogue is
	 * still a path under the home and not that folder's page: the panel lists
	 * what the home holds and not what the catalogue holds.
	 */
	function home($workspace, $folders) {
		$id = $workspace === Task\Route::INTERNAL ? Task\Folders::HOME : $workspace . "/" . Task\Folders::HOME;
		$found = null;

		foreach ($folders as $folder) {
			if (Task\Row::text($folder, "id") === $id) {
				$found = $folder;
			}
		}

		if ($found === null) {
			return null;
		}

		$entries = Task\Host::listing($id, "", array());
		$panel = Task\Files::home($entries, $id, $this->workspaces);

		return array(
			"folder" => $id,
			"note" => $id . self::HOME,
			"panel" => $panel,
		);
	}

	/**
	 * folder is the page of a folder of the catalogue: its jobs, the checkout
	 * it sits in, what it holds and what it says of itself.
	 *
	 * The place is read once, in the depth the page draws, and every row the
	 * page needs follows from that one answer: the sources it was detected in,
	 * the hours a run of it is allowed under and what atkins' linter said.
	 */
	function folder($id) {
		$place = Task\Host::place($id, "");
		$title = Task\Row::text($place, "title");

		$this->chrome->heading($title, Task\Route::crumbs($id, "", $this->workspaces));
		$this->chrome->at($id);
		$this->chrome->asks($id);

		$folders = $this->chrome->folders();
		$names = Task\Folders::ids($folders);
		$groups = $this->groups($id, "");
		$subs = $this->children($folders, $id);
		$status = Task\Host::status(Task\Row::text($place, "dir"));
		$checkout = Task\Host::checkout(Task\Row::text($place, "dir"));

		$where = Task\Place::of($id, $this->workspaces);
		$where->allow = Task\Row::items($place, "allow");
		$where->pull = Task\Place::pull($groups);
		$where->projects = count($subs);

		$pins = $this->chrome->user === null ? array() : $this->chrome->pins();
		$head = new Task\Head($title);

		$head->said(Task\Row::text($place, "description"));
		$head->icon("folder-open");
		$head->always();
		$head->buttons($where->buttons($this->chrome));
		$head->menu($where->menu($this->chrome, $title, $pins, $checkout, false));

		$control = new Task\Control($this->chrome);
		$rows = $control->of($groups);
		$tree = Task\Host::listing($id, "", $names);
		$up = $this->up($id);
		$files = Task\Files::panel($tree, $id, $this->workspaces, $up, $subs);
		$home = $where->workspace ? $this->home($id, $folders) : null;
		$notes = new Task\Seen($this->chrome);
		$open = $where->workspace;
		$errors = Task\Row::items($place, "errors");
		$error = count($errors) === 0 ? Task\Row::text($place, "error") : "";

		$scope = array(
			"head" => $head->data(),
			"findings" => Task\Findings::alert($folders, $id, $this->workspaces),
			"git" => Task\Place::git($status),
			"error" => $error,
			"control" => $rows,
			"nothing" => count($groups) === 0 && count($subs) === 0,
			"prompt" => Task\Prompt::inPlace($this->chrome, $id, "", $open, $notes),
			"files" => array("note" => $id . "/", "panel" => $files),
			"home" => $home,
			"detail" => $this->detail($id, $place, $checkout),
		);

		$this->view->page($scope, "folder.tpl");
	}

	/**
	 * browse is the page of a directory inside a folder that is no folder
	 * itself: the targets of its own that run there, and what it holds, to go
	 * on into.
	 *
	 * The jobs run under the rules of the folder above, whatever the directory
	 * was detected as, so the page says which of the three it is: a directory
	 * with a pipeline file of its own, one with none whose jobs are the
	 * folder's, and one with nothing to run at all.
	 *
	 * Only the first of the three is a place of its own, which is why both the
	 * lead and the alert about a source that would not read are read off the
	 * sources: a directory with none has no file to fail on, and a failure of
	 * the folder's own file is the folder's page to report.
	 */
	function browse($id, $rel) {
		$place = Task\Host::place($id, $rel);
		$name = Task\Folders::nameOf($rel);
		$where = $id . "/" . $rel;

		$this->chrome->heading($name, Task\Route::crumbs($where, "", $this->workspaces));
		$this->chrome->at($id);
		$this->chrome->asks($id, $rel);

		$folders = $this->chrome->folders();
		$groups = $this->groups($id, $rel);
		$sources = Task\Row::items($place, "sources");
		$status = Task\Host::status(Task\Row::text($place, "dir"));
		$head = new Task\Head($name, $this->leads($id, $rel, count($sources) > 0, count($groups) > 0));

		$head->icon("folder-open");
		$head->always();

		if ($this->chrome->user !== null) {
			$menu = new Task\Menu("Actions on " . $where);

			Task\Pin::item($menu, $this->chrome->pins(), $where, "", "", "folder");
			$head->menu($menu->data());
		}

		$control = new Task\Control($this->chrome);
		$rows = $control->of($groups);
		$tree = Task\Host::listing($id, $rel, Task\Folders::ids($folders));
		$files = Task\Files::panel($tree, $id, $this->workspaces, $this->above($id, $rel), array());
		$notes = new Task\Seen($this->chrome);

		$scope = array(
			"head" => $head->data(),
			"findings" => null,
			"git" => Task\Place::git($status),
			"error" => count($sources) === 0 ? "" : Task\Row::text($place, "error"),
			"control" => $rows,
			"nothing" => false,
			"prompt" => Task\Prompt::inPlace($this->chrome, $id, $rel, false, $notes),
			"files" => array("note" => $where . "/", "panel" => $files),
			"home" => null,
			"detail" => null,
		);

		$this->view->page($scope, "folder.tpl");
	}

	/**
	 * leads is the sentence under the heading of a directory's page, as HTML:
	 * what folder it is a directory of, and which of the three kinds of
	 * directory it is.
	 *
	 * The folder's ID is a link and the rest is text, and the ID is written by
	 * the reader of a pipeline file rather than typed here, so it is escaped.
	 */
	function leads($id, $rel, $own, $runs) {
		$href = Task\Route::href(Task\Route::folderHref($id, $this->workspaces));
		$name = htmlspecialchars($id, ENT_QUOTES);
		$lead = 'A directory of <a href="' . htmlspecialchars($href, ENT_QUOTES) . '">' . $name . "</a>.";

		if ($own) {
			return $lead . " Its jobs run in " . $name . "/" . htmlspecialchars($rel, ENT_QUOTES) . "/, under the rules of " . $name . ".";
		}

		if ($runs) {
			return $lead . " It has no pipeline file, Makefile, package scripts or scripts folder of its own; the jobs below run here, under the rules of " . $name . ".";
		}

		return $lead . " Nothing to run here: no pipeline file, Makefile, package scripts or scripts folder.";
	}

	/**
	 * groups is the Jobs table of a place, which costs the host one walk of
	 * the directory and a detection of each directory in it.
	 *
	 * The two permissions a terminal is gated on go into the call, and so do
	 * the folders the reader sees: a directory inside that is a folder nobody
	 * sees runs under the folder above it.
	 */
	function groups($id, $rel) {
		$seen = Task\Folders::ids($this->chrome->folders());
		$shell = $this->chrome->allowed("shell");
		$claude = $shell && Task\Host::claude();

		return Task\Host::groups($id, $rel, $seen, $shell, $claude);
	}

	/**
	 * children is the folders of the catalogue one step inside a folder, which
	 * the Files panel lists a row for where the listing holds none.
	 *
	 * A workspace's home is not one of them: it is listed as the home and not
	 * as a project of the workspace.
	 */
	function children($folders, $id) {
		$prefix = $id . "/";
		$workspace = in_array($id, $this->workspaces, true);
		$out = array();

		foreach ($folders as $folder) {
			$child = Task\Row::text($folder, "id");

			if (strpos($child, $prefix) !== 0) {
				continue;
			}

			$rest = substr($child, strlen($prefix));

			if (strpos($rest, "/") !== false || ($workspace && $rest === Task\Folders::HOME)) {
				continue;
			}

			$out[] = $folder;
		}

		return $out;
	}

	/** up is the page the .. row of a folder's Files panel opens: the folder above, or the root. */
	function up($id) {
		$cut = strrpos($id, "/");

		if ($cut === false) {
			return Task\Route::href("/f/" . Task\Route::INTERNAL);
		}

		return Task\Route::href(Task\Route::folderHref(substr($id, 0, $cut), $this->workspaces));
	}

	/** above is the page the .. row of a directory's Files panel opens: the directory over it, or the folder. */
	function above($id, $rel) {
		$href = Task\Route::href(Task\Route::folderHref($id, $this->workspaces));
		$cut = strrpos($rel, "/");

		if ($cut === false) {
			return $href;
		}

		return $href . "/" . Task\Route::escapePath(substr($rel, 0, $cut));
	}

	/**
	 * detail is the Folder panel: where the folder is, the pipeline files it
	 * was detected in, the checkout it was cloned from and what that takes,
	 * who may run its jobs, the hours they run in, the files its pipeline
	 * reads and how many schedules name it.
	 *
	 * Every piece is named on its own and the template writes the markup,
	 * because the cells hold more than the kv component draws: an icon before
	 * a path, a command in code, a Show button and a bar.
	 */
	function detail($id, $place, $checkout) {
		$schedules = new Task\Schedules\Schedules($this->chrome->db);
		$count = count($schedules->byFolder($id));
		$sources = array();

		foreach (Task\Row::items($place, "sources") as $source) {
			$kind = Task\Row::text($source, "kind");
			$file = Task\Row::text($source, "file");
			$command = Task\Row::items($source, "command");
			$shows = null;

			if ($kind !== "script") {
				$shows = Task\Place::source($id, $file, Task\Row::number($source, "lines"));
			}

			$sources[] = array(
				"kind" => $kind,
				"file" => $file,
				"command" => count($command) === 0 ? "" : implode(" ", $command),
				"shows" => $shows,
			);
		}

		$includes = array();

		foreach (Task\Row::items($place, "includes") as $include) {
			$file = Task\Row::text($include, "name");
			$includes[] = array(
				"file" => $file,
				"shows" => Task\Place::source($id, $file, Task\Row::number($include, "lines")),
			);
		}

		return array(
			"path" => $id . "/",
			"sources" => $sources,
			"checkout" => $this->cloned($checkout),
			"space" => Task\Size::space(Task\Row::items($checkout, "space")),
			"who" => Task\Place::who(Task\Row::items($place, "allow")),
			"hours" => Task\Place::hours(Task\Row::text($place, "hours")),
			"includes" => $includes,
			"count" => $count,
			"schedules" => $count === 1 ? "schedule" : "schedules",
			"page" => Task\Route::href(Task\Place::SCHEDULES) . "?folder=" . Task\Route::urlEscape($id),
		);
	}

	/**
	 * cloned is the Source row of the Folder panel: where the checkout came
	 * from, the commit it is at and when that was made. Null for a folder in
	 * no checkout.
	 */
	function cloned($checkout) {
		if (!Task\Row::flag($checkout, "found")) {
			return null;
		}

		$when = Task\Row::text($checkout, "when");

		return array(
			"remote" => Task\Row::text($checkout, "remote"),
			"branch" => Task\Row::text($checkout, "branch"),
			"commit" => Task\Row::text($checkout, "commit"),
			"subject" => Task\Row::text($checkout, "subject"),
			"ago" => $when === "" ? "" : Task\Clock::ago($when),
		);
	}

	/**
	 * job is the page of one job of a folder, or of a directory inside one.
	 *
	 * The page is a class of its own under code/Task, because only one module
	 * can own /f/* and this one owns it: a second mod_*.php here would declare
	 * a route chi already holds, and the autoloader does not reach a module's
	 * own file either way.
	 *
	 * job is the row the path resolved into and is handed down rather than read
	 * again: a job of a directory inside a folder had that directory detected
	 * to answer the path at all, and the page costs what the path cost.
	 */
	function job($id, $dir, $name, $job) {
		$page = new Task\Job($this->chrome, $this->view);

		$page->serve($id, $dir, $name, $job);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_folder;

	$mod->serve();
}
