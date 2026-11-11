<?php

// @route GET /artifacts

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_artifacts is the artifact tree: the newest version of every file the
// runs left, a directory at a time, the space the files and the history
// behind them take, and the record of every version, newest first.
//
// The page reads and does not write. The buttons that remove something post to
// /api/artifacts, which the script reads and then reloads the page.
//
// The files themselves are not pages and are not here: a file of the tree, and
// the .tgz of a directory of it, are byte streams Go serves at the root of the
// instance, and they stay there. So a link to one is written without the mount
// and every link to a page of this application carries it, which is why a row
// arrives holding the whole link and the template adds no {base}.
class mod_artifacts implements Module {
	/** LEAD is the sentence under the heading, as HTML: it names the variable a run writes to. */
	const LEAD = 'Files that runs write to <code>$TASK_UI_ARTIFACTS</code>, collected into one tree. When a run writes the same file again, the new file replaces it, and the history keeps the earlier version.';

	/** LIMIT is how many versions the history lists. */
	const LIMIT = 200;

	/** FILES is where a file of the tree answers, at the root and outside the mount. */
	const FILES = "/artifacts/files/";

	/** API is where the page's data- endpoints answer, at the root as every page's do. */
	const API = "/api/artifacts";

	/** COPY is the hover and the label of the icon that copies the .tgz link. */
	const COPY = "Copy the link to the .tgz";

	function name() {
		return "artifacts";
	}

	function title() {
		return "Artifacts";
	}

	function icon() {
		return "box";
	}

	function tables() {
		return array("artifact", "user");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("artifact/view")) {
			$view->refused();

			return;
		}

		$dir = self::clean(Task\Route::input("path"));
		$form = array("folder" => Task\Route::input("folder"), "job" => Task\Route::input("job"));
		$browse = $dir === "" ? "" : $dir . "/";

		$chrome->heading("Artifacts", self::crumbs($dir));

		$artifacts = new Task\Artifact;
		$head = new Task\Head("Artifacts", self::LEAD);

		if ($chrome->allowed("artifact/delete")) {
			$head->menu(self::actions());
		}

		$view->page(array(
			"head" => $head->data(),
			"space" => self::space(),
			"block" => self::block($browse, self::archive($dir)),
			"files" => $this->files($chrome, $artifacts, $dir),
			"form" => $form,
			"folders" => $this->folders($chrome),
			"history" => $this->history($chrome, $artifacts, $form),
		));
	}

	/**
	 * space is the bar above the files: what the tree takes on disk, the files
	 * and the git history behind them apart.
	 *
	 * An instance with no meter answers nothing, and then the page says nothing
	 * about space rather than saying it takes none.
	 *
	 * A folder's page draws the same bar of its checkout, so the arithmetic is
	 * Task\Size's and not this page's.
	 */
	static function space() {
		return Task\Size::space(Task\Host::space());
	}

	/**
	 * block is the head of the Files table: which directory of the tree is
	 * being read, and the .tgz of it.
	 *
	 * The archive is a byte stream Go serves at the root, so its link carries
	 * no mount; the icon beside it copies that link rather than following it.
	 */
	static function block($browse, $archive) {
		$buttons = new Task\Buttons(true);

		$buttons->item("download", ($browse === "" ? "Everything" : "This folder") . " as .tgz", array("href" => $archive, "download" => ""));
		$buttons->icon("copy", array(
			"data-copy" => $archive,
			"title" => self::COPY,
			"aria-label" => self::COPY,
		));

		return array(
			"title" => "Files",
			"note" => "/" . $browse,
			"empty" => "",
			"buttons" => $buttons->data(),
		);
	}

	/**
	 * actions is the head's [...]: the one action on the whole tree, for a
	 * reader who may remove an artifact.
	 */
	static function actions() {
		$menu = new Task\Menu("Actions on every artifact");

		$menu->danger("trash", "Delete all", array(
			"data-delete" => self::API,
			"data-confirm" => "Delete all artifacts? This removes every file, its history and the record of every version.",
			"data-confirm-ok" => "Delete all",
			"data-next" => "reload",
		));

		return $menu->data();
	}

	/**
	 * files is the Files table: one directory of the tree, the entries the
	 * reader may see.
	 *
	 * An admin sees the directory as it is. Anyone else sees an entry the
	 * record covers: a file whose newest version they may see, and a directory
	 * holding one such file anywhere under it. The record is read once and not
	 * once per entry.
	 */
	function files($chrome, $artifacts, $dir) {
		$admin = Task\User::admin($chrome->user);
		$visible = $admin ? array() : self::visible($artifacts, $chrome);
		$edits = $chrome->allowed("artifact/edit");
		$out = array();

		foreach (Task\Host::tree($dir) as $item) {
			$name = Task\Row::text($item, "name");
			$folder = Task\Row::flag($item, "dir");
			$path = $dir === "" ? $name : $dir . "/" . $name;

			if (!$admin && !self::covers($visible, $path, $folder)) {
				continue;
			}

			$out[] = self::entry($item, $folder, $path, $edits);
		}

		return $out;
	}

	/**
	 * entry is one row of the Files table: the entry, and the menu of what the
	 * reader may do to it.
	 *
	 * A directory links to this page listing it, which carries the mount; a
	 * file links to the stream that serves its bytes, which does not. The
	 * delete reads the API at the root, as every data- endpoint of a page does,
	 * and it names a directory by the path as the tree holds it where it names
	 * a file by the link that reads it.
	 */
	static function entry($item, $dir, $path, $edits) {
		$name = Task\Row::text($item, "name");
		$stamp = Task\Row::text($item, "mod_time");
		$archive = self::archive($path);
		$href = $dir ? Task\Route::href("/artifacts?path=" . Task\Route::queryEscape($path)) : self::FILES . self::escape($path);
		$menu = new Task\Menu("Actions on " . $name);

		if ($dir) {
			$menu->item("download", "Download as .tgz", array("href" => $archive, "download" => ""));
			$menu->item("copy", self::COPY, array("data-copy" => $archive));
		} else {
			$menu->item("eye", "View and versions", array("href" => $href, "data-artifact" => $path));
			$menu->item("download", "Download", array("href" => $href . "?download"));
			$menu->item("copy", "Copy the link", array("data-copy" => $href));
		}

		$menu->whenDanger($edits, "trash", "Delete from the tree", array(
			"data-delete" => "/api" . ($dir ? self::FILES . $path : $href),
			"data-confirm" => "Delete " . $path . ($dir ? "/ and everything in it" : "") . " from the tree? The history keeps its versions.",
			"data-confirm-ok" => "Delete",
			"data-next" => "reload",
		));

		return array(
			"dir" => $dir,
			"name" => $name,
			"href" => $href,
			"path" => $path,
			"size" => Task\Size::bytes(Task\Row::number($item, "size")),
			"ago" => Task\Clock::ago($stamp),
			"stamp" => Task\Clock::read($stamp),
			"menu" => $menu->data(),
		);
	}

	/**
	 * visible is the files of the tree the reader may see, by path: the newest
	 * version of each one, which is the version the tree holds.
	 *
	 * A reader who may see no folder at all sees the files no folder holds,
	 * which is what was written over MCP or the API.
	 */
	static function visible($artifacts, $chrome) {
		$out = array();

		foreach ($artifacts->latest() as $artifact) {
			if (Task\User::seesArtifact($chrome->user, $artifact)) {
				$out[] = $artifact->getPath();
			}
		}

		return $out;
	}

	/**
	 * covers reports whether the files a reader may see hold an entry of the
	 * tree: the file itself, and for a directory any file under it.
	 *
	 * A directory is listed for what is under it, so one holding nothing the
	 * reader may see is not listed at all, and neither is a file of the
	 * directory that holds no version they may see.
	 */
	static function covers($visible, $path, $dir) {
		if (!$dir) {
			return in_array($path, $visible, true);
		}

		$under = $path . "/";

		foreach ($visible as $seen) {
			if (strpos($seen, $under) === 0) {
				return true;
			}
		}

		return false;
	}

	/**
	 * history is the History table: the record of every version, newest first,
	 * narrowed by the folder and the job the form names.
	 *
	 * A row of a run is its folder's, and the reader sees it where they see the
	 * folder; a row with neither a run nor a folder was written over MCP or the
	 * API and is nobody's folder, so anyone signed in sees it.
	 */
	function history($chrome, $artifacts, $form) {
		$filter = array(
			"folder" => $form["folder"],
			"job" => $form["job"],
			"limit" => self::LIMIT,
		);
		$found = $artifacts->find($filter);
		$workspaces = $chrome->workspaces();
		$deletes = $chrome->allowed("artifact/delete");
		$out = array();

		foreach ($found as $artifact) {
			if (!Task\User::seesArtifact($chrome->user, $artifact)) {
				continue;
			}

			$out[] = self::version($artifact, $chrome, $workspaces, $deletes);
		}

		return $out;
	}

	/**
	 * version is one row of the History table: when the file was written,
	 * where it came from, the version of it and the menu of the record.
	 *
	 * A run of Claude's own has no job page, which is the one place the folder
	 * links and the job beside it does not. Task\Run answers that of a row
	 * holding a folder and a job, which an artifact's row is, and the Go page
	 * reads it the same way.
	 *
	 * A row with no commit is kept apart from the tree and has no version to
	 * view, so its name links to the record itself.
	 */
	static function version($artifact, $chrome, $workspaces, $deletes) {
		$id = $artifact->getId();
		$run = $artifact->getRunId();
		$name = $artifact->getName();
		$path = $artifact->getPath();
		$commit = $artifact->getCommitId();
		$folder = $artifact->getFolder();
		$job = $artifact->getJob();
		$created = $artifact->getCreatedAt();
		$place = Task\Route::folderHref($folder, $workspaces);
		$claude = Task\Run::isPrompt($artifact) || Task\Run::isNote($artifact);
		$download = self::API . "/" . $id;
		// The path as the row holds it, which is how the Go page writes this
		// link; the tree's own link to the same file is escaped a step at a
		// time, because Go builds that one before the template sees it.
		$version = self::FILES . $path . "?version=" . $commit;
		$menu = new Task\Menu("Actions on " . $name);

		if ($commit !== "") {
			$menu->detail("Version", $commit, true);
		}

		if ($run !== "") {
			$menu->detail("Run", $run, true);
		}

		$menu->detail("Created", Task\Clock::read($created));
		$menu->when($commit !== "", "eye", "View this version", array(
			"href" => $version,
			"data-artifact" => $path,
			"data-version" => $commit,
		));
		$menu->item("download", "Download this version", array("href" => $download));
		$menu->whenDanger($deletes, "trash", "Delete the record", array(
			"data-delete" => $download,
			"data-confirm" => "Delete the record of " . $name . ($run === "" ? "" : " from " . $folder . " / " . $job) . "?",
			"data-confirm-ok" => "Delete",
			"data-next" => "reload",
		));

		return array(
			"ago" => Task\Clock::ago($created),
			"stamp" => Task\Clock::read($created),
			"run" => $run,
			"short" => substr($run, 0, 8),
			"runHref" => Task\Route::href("/runs/" . $run),
			"folder" => $folder,
			"folderHref" => $place === "" ? "" : Task\Route::href($place),
			"job" => $job,
			"jobHref" => $claude || $place === "" ? "" : Task\Route::href($place . "/" . $job),
			"user" => $run === "" ? $chrome->who()->of($artifact->getUserName()) : array(),
			"name" => $name,
			"nameHref" => $commit === "" ? $download : $version,
			"path" => $path,
			"commit" => $commit,
			"version" => substr($commit, 0, 7),
			"size" => Task\Size::bytes($artifact->getByteSize()),
			"menu" => $menu->data(),
		);
	}

	/** folders is the folders of the catalogue the reader sees, for the filter. */
	function folders($chrome) {
		$out = array();

		foreach ($chrome->folders() as $folder) {
			$out[] = Task\Row::text($folder, "id");
		}

		return $out;
	}

	/**
	 * crumbs is the trail of a directory of the tree: the top of the tree, then
	 * a step per directory down to the one being read, which is the last step
	 * and links nowhere.
	 */
	static function crumbs($dir) {
		if ($dir === "") {
			return array(array("text" => "Artifacts", "href" => ""));
		}

		$out = array(array("text" => "Artifacts", "href" => "/artifacts"));
		$parts = explode("/", $dir);
		$walked = "";

		foreach ($parts as $at => $part) {
			$walked = $walked === "" ? $part : $walked . "/" . $part;
			$last = $at === count($parts) - 1;
			$out[] = array(
				"text" => $part,
				"href" => $last ? "" : "/artifacts?path=" . Task\Route::queryEscape($walked),
			);
		}

		return $out;
	}

	/**
	 * clean is ?path= as a directory of the tree: the empty steps and the dots
	 * dropped, and a step of two dots taking off the step before it.
	 *
	 * This is what Go's path.Clean leaves of an absolute path, which is what
	 * makes it safe: a path that climbs past the top of the tree cleans to the
	 * top and not above it.
	 */
	static function clean($path) {
		$out = array();

		foreach (explode("/", $path) as $step) {
			if ($step === "" || $step === ".") {
				continue;
			}

			if ($step !== "..") {
				$out[] = $step;

				continue;
			}

			if (count($out) > 0) {
				array_pop($out);
			}
		}

		return implode("/", $out);
	}

	/** archive is the .tgz of a directory of the tree, the whole of it for the top. */
	static function archive($dir) {
		if ($dir === "") {
			return "/artifacts.tgz";
		}

		return "/artifacts/" . self::escape($dir) . ".tgz";
	}

	/** escape is a path as the steps of a URL path, which Task\Route spells. */
	static function escape($path) {
		return Task\Route::escapePath($path);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_artifacts;

	$mod->serve();
}
