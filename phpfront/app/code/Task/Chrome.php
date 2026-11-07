<?php

namespace Task;

// Chrome is what every page shares: who is reading, what they may do, the
// sidebar, the crumbs and what the page reports.
//
// It is built once per request and handed to the theme as one scope, so a
// component reads its own slice of it and nothing else.
class Chrome {
	var $db;
	var $user = null;
	var $access = null;

	/** title is the page's own, before the instance's name. */
	var $title = "";

	/** crumbs are the trail above the content: text, and a href or "". */
	var $crumbs = array();

	/**
	 * error is what the page reports outside a form, for a page that answers
	 * rather than redirects: a run that is not there says so on the page it
	 * draws instead. The query's own is read where this is empty.
	 */
	var $error = "";

	/**
	 * notice is what the page says a change did, for a page that works one out
	 * rather than being redirected to with it: the editor opened on a
	 * workspace says what moving the steps there left behind. The query's own
	 * is read where this is empty.
	 */
	var $notice = "";

	var $users = null;
	var $people = null;

	/**
	 * prompt is where the header's prompt asks, which a page says with asks()
	 * or asksOfRun(); the root until one does, as a page that names no place
	 * asks there.
	 */
	var $prompt = null;

	/** catalogue, discovered and named are read once a request, by several readers each. */
	var $catalogue = null;
	var $discovered = null;
	var $named = null;

	/**
	 * folder and job are the place the page is about, which the sidebar marks
	 * as the row being read: a folder's ID and a job's key, both "" on a page
	 * that names no place.
	 */
	var $folder = "";
	var $job = "";

	/** tree is the folder tree, built once however often the theme reads it. */
	var $tree = null;

	/** built is the tree behind the sidebar, which the runs in progress hang on. */
	var $built = null;

	/** flows is the sidebar's workflows, read once however often the theme asks. */
	var $flows = null;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
		$this->users = new User($this->db);

		$name = Host::whoami();

		if ($name !== "") {
			$this->user = $this->users->byName($name);
		}

		$this->access = $this->users->access($this->user);
	}

	/** heading sets the page's title and its trail in one call. */
	function heading($title, $crumbs = array()) {
		$this->title = $title;
		$this->crumbs = $crumbs;

		return $this;
	}

	/** reports sets what the page says went wrong, which the alerts draw. */
	function reports($why) {
		$this->error = $why;

		return $this;
	}

	/** notes sets what the page says a change did, which the alerts draw. */
	function notes($what) {
		$this->notice = $what;

		return $this;
	}

	/** allowed reports whether the reader may do what a path names. */
	function allowed($path) {
		return $this->access->allowed($path);
	}

	/** name is the reader's name, "" when nobody is signed in. */
	function name() {
		return $this->user === null ? "" : $this->user->getName();
	}

	/** who draws any user: the avatar, the name and the menu beside them. */
	function who() {
		if ($this->people === null) {
			$roles = $this->users->roles();
			$this->people = new Who($this->access, $roles);
		}

		return $this->people;
	}

	/**
	 * message renders a note's text, linking the mentions the reader may open
	 * and drawing a ```form fence of it as the form it declares.
	 *
	 * The body comes out as a sequence - markdown, the form, markdown - so the
	 * renderer stays a markdown renderer and knows nothing of forms. The split
	 * is Task\Form::blocks(), the Go page's form.Blocks().
	 */
	function message($text) {
		$out = "";
		$at = 0;

		foreach (Form::blocks($text) as $block) {
			$out .= $this->markdown(substr($text, $at, $block["start"] - $at));
			$out .= Form::html($block, $this);
			$at = $block["end"];
		}

		return $out . $this->markdown(substr($text, $at));
	}

	/** markdown renders one stretch of a note's text, the mentions the reader may open linked. */
	function markdown($text) {
		return Host::message($text, Route::href("/users/"), $this->who()->linkable());
	}

	/**
	 * line renders a line of markdown as what goes inside a <small> or a
	 * cell: the host's renderer, with the paragraph it wraps a single block
	 * in taken off again.
	 *
	 * The tag has to come off because a <p> is a block and the places that
	 * draw one line of prose -- a workflow's description under its name --
	 * are phrasing content, which would draw the margin of a paragraph in
	 * the middle of a table cell.
	 *
	 * Text that is more than one block keeps its tags, because what it is
	 * then is a document and not a line, and the renderer's own markup is
	 * the only honest way to draw it.
	 */
	static function line($text) {
		if (!is_string($text) || trim($text) === "") {
			return "";
		}

		$html = trim(Host::markdown($text));

		if (substr($html, 0, 3) !== "<p>" || substr($html, -4) !== "</p>") {
			return $html;
		}

		$inner = substr($html, 3, strlen($html) - 7);

		return strpos($inner, "<p>") === false ? $inner : $html;
	}

	/**
	 * folders is the catalogue the reader sees: every folder the host
	 * discovered, less the ones the reader's row does not list.
	 *
	 * A folder nobody sees decides no link and no icon, so a page that names a
	 * place reads the catalogue through this and never through the binding.
	 */
	function folders() {
		if ($this->catalogue === null) {
			$this->catalogue = array();

			foreach ($this->all() as $folder) {
				if (User::sees($this->user, Row::text($folder, "id"))) {
					$this->catalogue[] = $folder;
				}
			}
		}

		return $this->catalogue;
	}

	/**
	 * all is the catalogue as the host discovered it, the folders the reader
	 * does not see included, read once a request.
	 *
	 * One page reads it: the page of a place, which resolves a path against it
	 * and then answers Not found for a folder the reader does not see. A
	 * folder they may not open is not a shorter one they may, so resolving
	 * through folders() would open the parent instead of refusing the child.
	 */
	function all() {
		if ($this->discovered === null) {
			$this->discovered = Host::folders();
		}

		return $this->discovered;
	}

	/**
	 * workspaces is the folder of every named workspace, in the sidebar's
	 * order, which is what turns a folder's ID into the page that holds it.
	 */
	function workspaces() {
		if ($this->named === null) {
			$workspaces = new Workspace($this->db);
			$this->named = $workspaces->folders();
		}

		return $this->named;
	}

	/**
	 * at is the place the page is about: the folder it names and the key of
	 * the job it names, which the sidebar draws as the row being read.
	 *
	 * The folder is the one the catalogue holds, not the one a row stores: a
	 * run of a directory inside a folder marks the folder's row.
	 */
	function at($folder, $job = "") {
		$this->folder = $folder;
		$this->job = $job;
		$this->tree = null;

		return $this;
	}

	/**
	 * folderTree is the sidebar's folder tree, built once a request: the
	 * theme reads the chrome's scope more than once a page, and the tree
	 * costs the host a directory, a git status and a commit per folder.
	 */
	function folderTree() {
		if ($this->tree === null) {
			$tree = new Folders($this->db, $this->access);
			$this->built = $tree;
			$this->tree = $tree->tree($this->folders(), $this->workspaces(), $this->folder, $this->job);
		}

		return $this->tree;
	}

	/**
	 * workflows is the list the sidebar draws under Workflows, built once a
	 * request: deciding whether the reader may start one asks the host for a
	 * job per step, and every page carries the list.
	 */
	function workflows() {
		if ($this->flows === null) {
			$store = new Workflow($this->db);
			$seen = $store->visible($this);
			$this->flows = Workflow::sidebar($seen, $this->workspaces(), "/" . Route::path());
		}

		return $this->flows;
	}

	/**
	 * live is the runs in progress the reader sees, by the key of their job,
	 * which the sidebar's dot pulses for and its rows draw a Stop for.
	 *
	 * It is the tree's own answer and not a second reading of the runner:
	 * the dot and the rows under it say the same thing on one page.
	 */
	function live() {
		$this->folderTree();

		return $this->built->running;
	}

	/**
	 * liveRuns is the same runs as their rows, by the key of their job, which
	 * the Jobs table reads for the title its button attaches to.
	 *
	 * The tree read the rows to filter them, so the table costs nothing for
	 * them: the ID alone does not say what a run is titled.
	 */
	function liveRuns() {
		$this->folderTree();

		return $this->built->rows;
	}

	/**
	 * pins is the reader's pins as the tree read them, which a Jobs row's
	 * [...] reads to know whether its Pin unpins.
	 */
	function pins() {
		$this->folderTree();

		return $this->built->held;
	}

	/**
	 * asks is where the header's prompt asks on this page: a folder and a
	 * directory inside it, which a folder's page, a directory's, a job's and a
	 * workspace's settings each name.
	 */
	function asks($folder, $dir = "") {
		$this->prompt = Prompt::at($folder, $dir, $this->folders(), $this->workspaces());

		return $this;
	}

	/**
	 * upload is the directory of the artifact tree the markdown fields of this
	 * page write into, which is the place the page is about.
	 *
	 * The body carries it and every field reads it off the body, so a field
	 * the script builds - the one an edit opens - writes where the field it
	 * replaced would have.
	 */
	function upload() {
		$prompt = $this->prompt === null ? Prompt::root() : $this->prompt;

		return Upload::ofPlace(Route::place($prompt["folder"], $prompt["dir"]));
	}

	/** asksOfRun is where it asks on a run's page, which is about the run. */
	function asksOfRun($run) {
		$this->prompt = Prompt::ofRun($run, $this->folders(), $this->workspaces());

		return $this;
	}

	/**
	 * roles is the role each user acts with, by name, which the body carries
	 * for the colour the script gives an avatar it draws.
	 *
	 * Nobody signed in reads none of them: who someone is goes no further than
	 * the page needs.
	 */
	function roles() {
		return $this->user === null ? array() : $this->who()->roles;
	}

	/**
	 * runAsUsers is who a run may be started as: every user with the instance
	 * SSH keys they bring, in the shape a select draws them.
	 *
	 * An admin alone picks; anyone else runs as themselves, and an empty list
	 * is a select the page leaves out. What else a pick needs is the asking
	 * page's: the header asks it of a reader who may prompt, the schedules
	 * page of any admin, and both draw the one list.
	 */
	function runAsUsers() {
		if (!User::admin($this->user)) {
			return array();
		}

		$users = $this->users->all();
		$store = new Key($this->db);
		$keys = $store->ofUsers($users);

		return User::options($users, $keys);
	}

	/**
	 * runAs is the select that picks who the header's prompt runs as: the
	 * users, the reader picked, in the short form the header draws.
	 *
	 * A reader who may not prompt at all picks nobody, not even themselves.
	 */
	function runAs() {
		if (!$this->allowed("prompt/edit")) {
			return array(
				"users" => array(),
				"selected" => "",
				"short" => true,
			);
		}

		return array(
			"users" => $this->runAsUsers(),
			"selected" => $this->name(),
			"short" => true,
		);
	}

	/**
	 * data is the scope the theme reads.
	 *
	 * roles and workspaces arrive as the JSON the body's two attributes carry
	 * and not as maps, because nothing in the theme reads them: they are there
	 * for the script, which parses the attributes.
	 */
	function data() {
		$path = "/" . Route::path();

		if ($this->user === null) {
			return $this->anonymous($path);
		}

		$sidebar = new Sidebar($this->db);

		return array(
			"base" => Host::base(),
			"version" => Host::version(),
			"assets" => Host::assets(),
			"assetVersion" => Host::assetVersion(),
			"title" => $this->title,
			"path" => $path,
			"crumbs" => $this->crumbs,
			"notice" => $this->notice !== "" ? $this->notice : Route::input("notice"),
			"error" => $this->error !== "" ? $this->error : Route::input("error"),
			"user" => $this->reader(),
			"can" => $this->access->effective(),
			"roles" => Text::object($this->roles()),
			"workspaces" => Text::object(Workspace::ranked($this->workspaces())),
			"loginRequired" => Host::login(),
			"settings" => Sidebar::settings($this->access),
			"sidebar" => $sidebar->items($this->access, $path),
			"folders" => $this->folderTree(),
			"workflows" => $this->workflows(),
			"prompt" => $this->prompt === null ? Prompt::root() : $this->prompt,
			"upload" => $this->upload(),
			"runAs" => $this->runAs(),
			"recents" => $this->recents(),
			"running" => count($this->live()) > 0,
		);
	}

	/**
	 * anonymous is the scope of a page nobody is signed in for: the head, the
	 * body's attributes, the content and the footer, and none of the sidebar,
	 * the folder tree, the workflows or the recents.
	 *
	 * The shell draws none of those for a null user, so reading them would cost
	 * a reader who cannot see them a directory, a git status and a commit per
	 * folder. The two JSON attributes are null and not {}, which is what Go
	 * writes for the maps its sign in page leaves unset.
	 */
	function anonymous($path) {
		return array(
			"base" => Host::base(),
			"version" => Host::version(),
			"assets" => Host::assets(),
			"assetVersion" => Host::assetVersion(),
			"title" => $this->title,
			"path" => $path,
			"crumbs" => array(),
			"notice" => "",
			"error" => $this->error,
			"user" => null,
			"can" => $this->access->effective(),
			"roles" => "null",
			"workspaces" => "null",
			"loginRequired" => Host::login(),
		);
	}

	/** recents is the right rail: the latest runs by anyone, newest first. */
	function recents() {
		$rail = new Recent($this->db);
		$people = $this->who();

		return $rail->data($this->access, $people->roles, $this->folders());
	}

	/** reader is the signed in user, as the header draws them. */
	function reader() {
		if ($this->user === null) {
			return null;
		}

		return array(
			"name" => $this->user->getName(),
			"role" => User::role($this->user),
			"initials" => User::initials($this->user),
		);
	}
}
