<?php

namespace Task;

// Host declares the host calls for the linter and for the test fixtures.
// Nothing requires this at runtime: the real calls are Go functions whose
// registered names hold `::`, which is what a static call resolves first.
//
// Each answer is a static a fixture writes, so a fixture drives the host it
// runs without one.
//
// Keep it in step with phpfront/bind.go.
class Host {
	public static $base = "";
	public static $now = "";
	public static $version = "";
	public static $login = false;
	public static $assetVersion = "";
	public static $whoami = "";
	public static $folders = array();
	public static $jobs = array();
	public static $projects = array();
	public static $folderDirs = array();

	/**
	 * folderDirCalls counts the crossings a tree spent asking for them, which
	 * is what the batch answer is for: the number and not the answer is the
	 * thing a fixture holds.
	 */
	public static $folderDirCalls = 0;
	public static $isDir = array();
	public static $commits = array();
	public static $running = array();
	public static $flows = array();
	public static $dirty = array();
	public static $rules = array();
	public static $published = array();
	public static $dispatched = array();
	public static $runs = array();
	public static $artifacts = array();
	public static $sessions = array();
	public static $frames = array();
	public static $screens = array();
	public static $expired = array();
	public static $claude = false;
	public static $tree = array();
	public static $space = array();
	public static $spaces = array();
	public static $inventory = array();
	public static $hosts = array();
	public static $reachable = array();
	public static $tooling = array();
	public static $places = array();
	public static $groups = array();
	public static $listings = array();
	public static $rootTrees = array();
	public static $modTimes = array();
	public static $statuses = array();
	public static $checkouts = array();
	public static $hours = array();
	public static $keystore = array();
	public static $keyFolders = array();
	public static $instance = array();
	public static $usage = array();
	public static $telemetry = array();
	public static $integrations = array();
	public static $modules = array();
	public static $entries = array();
	public static $variables = array();
	public static $variable = array();
	public static $integrationEnv = array();

	/** The path the router is mounted at, "" once the prefix is gone. */
	static function base() {
		return self::$base;
	}

	/**
	 * This moment on the instance's clock, in the shape Task\Clock calls
	 * CLOCK: RFC 3339, with the zone the instance reads in. Empty is a clock a
	 * fixture has not pinned, which Task\Clock reads as the runtime's own.
	 */
	static function now() {
		return self::$now;
	}

	/** The instance's version. */
	static function version() {
		return self::$version;
	}

	/** Whether the instance requires a sign in, so there is a session to end. */
	static function login() {
		return self::$login;
	}

	/** Where the static files answer, outside the base. */
	static function assets() {
		return "/assets";
	}

	/** The content hash of the static files, which their URLs carry as ?v=. */
	static function assetVersion() {
		return self::$assetVersion;
	}

	/** The name the Guard resolved, "" when nobody is signed in. */
	static function whoami() {
		return self::$whoami;
	}

	/**
	 * The folder catalogue: each folder as `id`, `title`, `dir`, its `allow`
	 * list, the `sources` its conventions were detected in, the `error` the
	 * sources failed to parse with and the `errors` atkins' linter found in
	 * its pipelines, each as `path`, `job`, `issue` and `detail`, and the
	 * `jobs` it declares, each job as `id`, `name`, `dir`, `shell`, `host`
	 * and its own `allow` list.
	 */
	static function folders() {
		return self::$folders;
	}

	/**
	 * The job a run names, as the catalogue resolves it: `found` says whether
	 * it is there at all, and both allow lists come with it, because the
	 * folder of a run is not the page's to hold.
	 *
	 * A fixture writes the answers by folder, directory and job, the way the
	 * host answers them, so no fixture resolves a catalogue of its own.
	 */
	static function job($folder, $dir, $id) {
		$key = $folder . "|" . $dir . "|" . $id;

		if (!isset(self::$jobs[$key])) {
			return array(
				"found" => false,
				"id" => "",
				"name" => "",
				"dir" => "",
				"shell" => false,
				"host" => false,
				"allow" => array(),
				"folder_allow" => array(),
				"kind" => "",
				"source" => "",
				"source_lines" => 0,
				"description" => "",
				"summary" => "",
				"command" => array(),
				"steps" => array(),
				"aliases" => array(),
				"depends" => array(),
				"interactive" => false,
				"takes_args" => false,
				"args_command" => "",
				"key" => "",
			);
		}

		return self::$jobs[$key];
	}

	/**
	 * The folder IDs of the directories right inside a named workspace's
	 * src: a checkout with no pipeline file is in no catalogue and the
	 * sidebar lists it anyway.
	 */
	static function projects($workspace) {
		return isset(self::$projects[$workspace]) ? self::$projects[$workspace] : array();
	}

	/**
	 * The directory a folder ID names: the catalogue's for a folder it holds,
	 * else where the layout puts the project among the named workspaces.
	 */
	static function folderDir($id, $workspaces) {
		self::$folderDirCalls++;

		return isset(self::$folderDirs[$id]) ? self::$folderDirs[$id] : "";
	}

	/**
	 * The same for many IDs at once, keyed by the ID, which is what a tree
	 * asks: an ID the layout places nowhere is absent from the answer.
	 */
	static function folderDirs($ids, $workspaces) {
		self::$folderDirCalls++;
		$out = array();

		foreach ($ids as $id) {
			if (isset(self::$folderDirs[$id]) && self::$folderDirs[$id] !== "") {
				$out[$id] = self::$folderDirs[$id];
			}
		}

		return $out;
	}

	/** Whether a directory inside a folder of the catalogue is there. */
	static function isDir($folder, $rel) {
		$key = $folder . "|" . $rel;

		return isset(self::$isDir[$key]) ? self::$isDir[$key] : false;
	}

	/** The last commit in a directory, in the shape Task\Clock calls STORED. */
	static function commit($dir) {
		return isset(self::$commits[$dir]) ? self::$commits[$dir] : "";
	}

	/** The runs in progress, by the key their job is listed under. */
	static function running() {
		return self::$running;
	}

	/**
	 * The workflow runs in progress, by the workflow each belongs to: the
	 * workflow runner's memory, which is what running means for a workflow.
	 * A workflow_run row left behind by a process that died is not one.
	 */
	static function flows() {
		return self::$flows;
	}

	/** The changed and untracked files under a checkout directory. */
	static function dirty($dir) {
		return isset(self::$dirty[$dir]) ? self::$dirty[$dir] : 0;
	}

	/**
	 * The permission titles and path lists of config/permissions.yml: rules,
	 * sections, permissions and extras are each a titled key, paths and
	 * checks a list of paths. An extra's key is the path a check names,
	 * <section>/<rule>, because a section's all does not stand for it.
	 */
	static function rules() {
		return self::$rules;
	}

	/**
	 * A run as the runner holds it, which is ahead of the row: a page reads
	 * the row and lays this over the fields a session decides. live is false
	 * where the runner holds no session, and then the row is the whole
	 * answer.
	 */
	static function session($id) {
		if (isset(self::$sessions[$id])) {
			return self::$sessions[$id];
		}

		return array(
			"live" => false,
			"attached" => 0,
			"message" => "",
			"state" => "",
			"exit_code" => 0,
			"title" => "",
			"finished_at" => "",
		);
	}

	/** The last screen of a live run, as the SVG a page embeds. */
	static function frame($id) {
		return isset(self::$frames[$id]) ? self::$frames[$id] : "";
	}

	/**
	 * The last screen of a run that has ended, replayed off its recording:
	 * the SVG a pipeline draws under a step, the lines the whole output has
	 * and whether the screen already shows them all.
	 */
	static function screen($id, $castPath, $cols) {
		if (isset(self::$screens[$id])) {
			return self::$screens[$id];
		}

		return array(
			"svg" => "",
			"lines" => 0,
			"complete" => false,
		);
	}

	/** Whether a run names a recording the retention sweep has removed. */
	static function expired($castPath, $state) {
		return isset(self::$expired[$castPath]) ? self::$expired[$castPath] : false;
	}

	/** Whether the reader has a token for Claude, which a summary needs. */
	static function claude() {
		return self::$claude;
	}

	/** Sends an event on the stream every live label reads. */
	static function publish($topic, $payload) {
		self::$published[] = array("topic" => $topic, "payload" => $payload);

		return true;
	}

	/**
	 * Starts a run over a comment and the markdown saying what to do with
	 * it, and answers the note that holds the thread. The action is a path
	 * in this tree and the folder is where in the artifact tree the run
	 * writes, "" for the top of it.
	 */
	static function dispatch($comment, $action, $folder = "") {
		self::$dispatched[] = array(
			"comment" => $comment,
			"action" => $action,
			"folder" => $folder,
		);

		return "01NOTE";
	}

	/**
	 * The run a dispatch turned into, "" where there is none yet: a
	 * dispatch opens the note first, so the run arrives after it answered.
	 */
	static function run($note) {
		return isset(self::$runs[$note]) ? self::$runs[$note] : "";
	}

	/**
	 * Markdown as HTML, through the renderer that owns the content policy,
	 * with what a relative link in it resolves against: the prefix the files
	 * sit under and the directory the text was written in. Both empty leaves
	 * a relative link as text.
	 *
	 * A blank line starts a paragraph here as it does there, because that is
	 * the difference a reader of the answer decides on: Task\Chrome::line
	 * takes the tag off one paragraph and leaves two alone.
	 */
	static function markdown($src, $prefix = "", $dir = "") {
		$out = array();

		foreach (preg_split('/\n\s*\n/', $src) as $block) {
			$out[] = "<p>" . htmlspecialchars($block, ENT_QUOTES) . "</p>";
		}

		return implode("\n\n", $out);
	}

	/**
	 * A note's text: the same renderer with a line break kept and @name
	 * marked. A name in link becomes a link under prefix.
	 */
	static function message($src, $prefix = "", $link = array()) {
		return "<p>" . htmlspecialchars($src, ENT_QUOTES) . "</p>";
	}

	/** The files under one folder of the artifact tree, "" for the whole tree. */
	static function artifacts($folder = "") {
		return isset(self::$artifacts[$folder]) ? self::$artifacts[$folder] : array();
	}

	/** The text of one artifact, from the top of the tree, or "". */
	static function artifact($path) {
		return "";
	}

	/**
	 * One directory of the artifact tree, which a page lists a level at a
	 * time: each entry as `name`, `dir`, `size` and `mod_time`. A directory's
	 * size is the space the files under it take.
	 */
	static function tree($dir) {
		return isset(self::$tree[$dir]) ? self::$tree[$dir] : array();
	}

	/**
	 * The space the artifact tree takes: `files`, the `history` behind them
	 * and what git `ignored`, which the total leaves out. `measured` is false
	 * on an instance with no meter, where a page says nothing about space.
	 */
	static function space() {
		if (self::$space) {
			return self::$space;
		}

		return array(
			"measured" => false,
			"files" => 0,
			"history" => 0,
			"ignored" => 0,
		);
	}

	/**
	 * The space each of the directories takes, by the directory, in the shape
	 * space() answers: the Projects settings measure every project they list.
	 *
	 * A directory nothing measured is not in the answer, so an instance with
	 * no meter answers none of them and a page says nothing about space.
	 */
	static function spaces($dirs) {
		$out = array();

		foreach ($dirs as $dir) {
			if (isset(self::$spaces[$dir])) {
				$out[$dir] = self::$spaces[$dir];
			}
		}

		return $out;
	}

	/**
	 * The hosts page's facts: `off` for an instance with no Beszel, the
	 * `error` the hub answered, the `hosts` largest first, the `logins` by
	 * host address, the `cluster` of pools and the `history` by host ID.
	 *
	 * $hosts is the reader's hosts permission, which is what makes task-ui
	 * read a host rather than answer out of what it kept.
	 */
	static function inventory($hosts) {
		if (self::$inventory) {
			return self::$inventory;
		}

		return array(
			"off" => true,
			"error" => "",
			"hosts" => array(),
			"logins" => array(),
			"cluster" => array(
				"up" => 0,
				"down" => 0,
				"left" => 0,
				"pools" => array(),
				"hosts" => array(),
			),
			"history" => array(),
		);
	}

	/**
	 * What atkins has recorded for the tree: a scope per database it records
	 * to, the project's first, each with its `commands`, the `usage` and
	 * `command_usage` series by day and the `samples` one by one.
	 *
	 * Empty on a tree nobody has run a job in through atkins, where a page
	 * draws no panel at all.
	 */
	static function tooling() {
		return self::$tooling;
	}

	/**
	 * The hosts a command can be run on, largest capacity first: the ones
	 * Beszel monitors that task-ui has an ssh login to, by the address it
	 * signs in to. A job picker offers one job per host and one over all of
	 * them, and reads nothing else.
	 */
	static function reachable() {
		return self::$reachable;
	}

	/**
	 * One host of the inventory in full: its `containers` as Beszel and
	 * docker ps report them together, its durable `filesystems`, and the
	 * `samples` its charts draw. `found` is false where Beszel monitors no
	 * such host.
	 */
	static function host($name, $hosts) {
		if (isset(self::$hosts[$name])) {
			return self::$hosts[$name];
		}

		return array(
			"found" => false,
			"error" => "",
			"system" => array(),
			"login" => "",
			"containers" => array(),
			"measured" => false,
			"sized" => false,
			"unsized" => "",
			"filesystems" => array(),
			"from_df" => false,
			"read" => "",
			"samples" => array(),
			"container_samples" => array(),
		);
	}

	/**
	 * A folder of the catalogue, or a directory inside one, in the depth the
	 * page about it draws: what it says of itself as `id`, `title`,
	 * `description`, `dir` and `rel`, the `allow` list and `hours` that gate
	 * a run of it, the `sources` it was detected in each as `kind`, `file`,
	 * `command` and `lines`, the `includes` its pipeline reads each as `name`
	 * and `lines`, and the `error` and `errors` its pipelines failed with.
	 *
	 * `found` is false where the catalogue holds no such folder, or the
	 * directory inside it is not there to detect.
	 */
	static function place($folder, $rel) {
		$key = $folder . "|" . $rel;

		if (isset(self::$places[$key])) {
			return self::$places[$key];
		}

		return array(
			"found" => false,
			"id" => "",
			"title" => "",
			"description" => "",
			"dir" => "",
			"rel" => "",
			"allow" => array(),
			"hours" => "",
			"error" => "",
			"errors" => array(),
			"sources" => array(),
			"includes" => array(),
		);
	}

	/**
	 * The Jobs table of a place: a row per pipeline file of it, the files at
	 * its top that run, the skills atkins applies there, the terminals, and a
	 * row per directory inside it that runs jobs of its own. Each row is
	 * `kind`, `note`, `source`, `dir`, `empty`, `targets`, the `folder` its
	 * jobs run as and the `jobs` themselves.
	 *
	 * $seen is the folders the reader sees, $shell and $claude the two
	 * permissions a terminal is gated on.
	 */
	static function groups($folder, $rel, $seen, $shell, $claude) {
		$key = $folder . "|" . $rel;

		return isset(self::$groups[$key]) ? self::$groups[$key] : array();
	}

	/**
	 * A directory as the Files panel draws it: a folder's own, or a directory
	 * inside it, each entry `name`, `dir`, `size`, `mod_time`, `path`,
	 * `folder`, `sealed` and `exec`.
	 *
	 * `path` is empty for an entry the folder refuses to open, and `folder`
	 * the ID of the catalogue folder a directory entry is in its own right,
	 * of the ones $seen names.
	 */
	static function listing($folder, $rel, $seen) {
		$key = $folder . "|" . $rel;

		return isset(self::$listings[$key]) ? self::$listings[$key] : array();
	}

	/**
	 * The root's own directories as the Files panel of the root's page draws
	 * them, in the shape listing() answers, which no folder of the catalogue
	 * answers: the root is what the catalogue was read from.
	 *
	 * $named says a workspace row names the root, which moves the listing
	 * into the src of the tree. `path` is always empty.
	 */
	static function rootTree($named, $seen) {
		$key = $named ? "1" : "0";

		return isset(self::$rootTrees[$key]) ? self::$rootTrees[$key] : array();
	}

	/**
	 * When a directory last changed, in the shape Task\Clock calls STORED:
	 * for a folder under a folder that the listing of the one above does not
	 * hold.
	 */
	static function modTime($dir) {
		return isset(self::$modTimes[$dir]) ? self::$modTimes[$dir] : "";
	}

	/**
	 * What git reports of the checkout a place is in: the `checkout` under
	 * the root, its `branch` or the commit it is `detached` at, the
	 * `upstream` it tracks, how far `ahead` and `behind` it is, how many
	 * files are `dirty` and the `changes` themselves, each as `code`,
	 * `state`, `path` and `from`.
	 *
	 * `found` is false where the place is in no checkout, or the instance has
	 * no root to name one against.
	 */
	static function status($dir) {
		if (isset(self::$statuses[$dir])) {
			return self::$statuses[$dir];
		}

		return array(
			"found" => false,
			"checkout" => "",
			"branch" => "",
			"detached" => "",
			"upstream" => "",
			"ahead" => 0,
			"behind" => 0,
			"dirty" => 0,
			"changes" => array(),
		);
	}

	/**
	 * The commit a folder's checkout is at and the `space` that checkout
	 * takes: the `remote`, the `branch`, the `commit` and its `subject`, and
	 * `when` it was made in the shape Task\Clock calls STORED.
	 *
	 * `found` is false for a folder in no checkout, and `space` is unmeasured
	 * where nothing measured it.
	 */
	static function checkout($dir) {
		if (isset(self::$checkouts[$dir])) {
			return self::$checkouts[$dir];
		}

		return array(
			"found" => false,
			"remote" => "",
			"branch" => "",
			"commit" => "",
			"subject" => "",
			"when" => "",
			"space" => array(
				"measured" => false,
				"files" => 0,
				"history" => 0,
				"ignored" => 0,
			),
		);
	}

	/**
	 * An hours rule read: the one `line` a page prints it as, and whether it
	 * is `open` at the moment of the reading. A rule the parser refuses
	 * prints as itself and is always open, as a run of it is never refused.
	 */
	static function hours($text) {
		if (isset(self::$hours[$text])) {
			return self::$hours[$text];
		}

		return array("line" => "", "open" => true);
	}

	/**
	 * The key folder: the `dir` it is, the name of the `instance` encryption
	 * key and every key in it as `keys`.
	 *
	 * A key row holds `name`, `purpose`, `scope`, `algo`, `bits`, `size`,
	 * `fingerprint`, `comment`, `public_key`, `created_at`, the `roles` and
	 * `users` who get it, `git` for the one git uses, `encrypts` for one that
	 * stores secrets, the `workspace` it belongs to, and `backup`,
	 * `backup_fingerprint` and `rotated_at` of a rotation that kept the
	 * previous material. No private material is here.
	 */
	static function keys() {
		if (self::$keystore) {
			return self::$keystore;
		}

		return array(
			"dir" => "",
			"instance" => "",
			"keys" => array(),
		);
	}

	/**
	 * The `.ssh` folder of every project that holds a private key of its own:
	 * the `project` it belongs to, the `dir` it is and the names of its `keys`.
	 */
	static function keyFolders() {
		return self::$keyFolders;
	}

	/**
	 * What the server runs with: the `url`, `listen`, `root`, `data_dir`,
	 * `ssh_dir`, `htpasswd` and `host_data` paths, `discover`, `auto_login`,
	 * `time_zone`, the `session_ttl`, `session_max`, `keep` and `grace` durations,
	 * `artifact_history`, `compose_project`, `beszel`, `running` and the
	 * `backlog` of slots, the `parent_env` mode with the `env` names,
	 * `env_base` and `env_sensitive` lists, and `measured` with the `storage`
	 * table.
	 *
	 * $admin is the reader's flag: `env` and `storage` are an admin's, and an
	 * instance read by anyone else holds neither.
	 */
	static function instance($admin) {
		if (self::$instance) {
			return self::$instance;
		}

		return array(
			"url" => "",
			"listen" => "",
			"root" => "",
			"data_dir" => "",
			"ssh_dir" => "",
			"htpasswd" => "",
			"host_data" => "",
			"discover" => false,
			"auto_login" => "",
			"time_zone" => "",
			"session_ttl" => "",
			"session_max" => "",
			"keep" => "",
			"grace" => "",
			"artifact_history" => 0,
			"compose_project" => "",
			"beszel" => false,
			"running" => false,
			"backlog" => array(
				"parallel" => 0,
				"going" => 0,
				"waiting" => 0,
				"auto" => 0,
				"configured" => 0,
				"length" => 0,
			),
			"parent_env" => "",
			"env" => array(),
			"env_base" => array(),
			"env_sensitive" => array(),
			"measured" => false,
			"storage" => array(
				"rows" => array(),
				"total" => array(
					"name" => "",
					"where" => "",
					"size" => 0,
					"files" => 0,
					"pending" => false,
				),
				"pending" => false,
			),
		);
	}

	/**
	 * What the retention rules apply to: the `areas` keyed by kind, each with
	 * its `kind`, the `count` of items, their `bytes` where they have a size
	 * and the `oldest` of them.
	 *
	 * `off` is an instance with no retention sweep, whose page draws no table
	 * at all, and `error` what the walk answered when it failed, which the
	 * page says instead of a table of nothing.
	 */
	static function usage() {
		if (self::$usage) {
			return self::$usage;
		}

		return array(
			"off" => true,
			"error" => "",
			"areas" => array(),
		);
	}

	/**
	 * The options the tracer runs on, the stored ones over the defaults:
	 * `enabled`, `service_name`, `sample_rate`, `ring_buffer_size`,
	 * `top_requests`, `max_spans_per_trace`, `debug_span_min_duration`,
	 * `ignore_paths`, `live_stream`, `refresh_interval`, `allowed_networks`,
	 * `option_path`, `capture_logs`, `track_memory_use` and
	 * `trust_request_id`.
	 *
	 * `debug_span_min_duration` is spelled as a duration, "1ms" or "500us",
	 * which is what the form posts back: the shortest span marked as detail
	 * that a trace keeps.
	 *
	 * `wired` says the instance has telemetry at all, which is what makes the
	 * page editable; `path` is the mount the router read at startup, which is
	 * what the link opens. `error` is what reading the stored options
	 * answered when it failed.
	 */
	static function telemetry() {
		if (self::$telemetry) {
			return self::$telemetry;
		}

		return array(
			"error" => "",
			"wired" => false,
			"path" => "",
			"enabled" => false,
			"service_name" => "",
			"sample_rate" => 0,
			"ring_buffer_size" => 0,
			"top_requests" => 0,
			"max_spans_per_trace" => 0,
			"debug_span_min_duration" => "",
			"ignore_paths" => array(),
			"live_stream" => false,
			"refresh_interval" => 0,
			"allowed_networks" => array(),
			"option_path" => "",
			"capture_logs" => false,
			"track_memory_use" => false,
			"trust_request_id" => false,
		);
	}

	/**
	 * The integrations as the Integrations settings draw them, in the order
	 * the page lists them: the `id`, `title` and `description` of a service,
	 * the `credential` its `steps` lead to, its `fields`, and what its
	 * settings hold now — `configured`, the plain `values`, `set` for a
	 * secret that has a value, the roles and users `checked` by a Who
	 * setting, `enabled` for the switches, and the `error` of a read that
	 * failed.
	 *
	 * A field holds `name`, `label`, `help`, `placeholder`, `secret`,
	 * `optional`, `who`, `options`, `generated`, `env`, `default`, `toggle`,
	 * `default_on`, `empty` and `source`.
	 *
	 * No secret value is here: a form shows whether one is set, and an empty
	 * field on save keeps it.
	 */
	static function integrations() {
		return self::$integrations;
	}

	/**
	 * The plugins the instance serves under /m, which the Sidebar settings
	 * list beside the pages: the `name` in the path, the row's `title`, the
	 * `description` under its heading and its `icon` symbol.
	 */
	static function modules() {
		return self::$modules;
	}

	/**
	 * The project directories under the src of each workspace: the `id` a
	 * page of one answers under, the `dir` itself, the `sources` its
	 * conventions were read from, how many `jobs` it holds, the `error` of a
	 * source that does not parse, and `checkout` with the `remote`, `branch`,
	 * `commit`, `subject` and `when` of a directory that is one.
	 *
	 * $workspaces is the folder of each workspace row, in the order the
	 * settings list them.
	 */
	static function entries($workspaces) {
		return self::$entries;
	}

	/**
	 * What the vault holds of the plain env variables: `values`, the value of
	 * each by row ID, and the `error` the walk of them stopped at.
	 *
	 * A secret is not in it, since no page prints one, and a row the walk did
	 * not reach past is absent rather than empty: the page prints what the row
	 * itself carries for one, which is its armour.
	 */
	static function variables() {
		return self::$variables;
	}

	/**
	 * One variable as the form that edits it opens the row: `content`, and the
	 * `error` the vault gave.
	 *
	 * A secret answers no content, because the form never draws one, and an
	 * error is the shared 404 rather than a value the page does not have.
	 */
	static function variable($id) {
		return self::$variable;
	}

	/**
	 * What the integrations add to a run of the reader, in name order: the
	 * `name` of each, its `value`, and `redacted` for a name that holds a
	 * credential, whose value is not here.
	 *
	 * It is the reader's own and not the instance's: a Who setting names the
	 * readers a service reaches, so two readers see two environments.
	 */
	static function integrationEnv() {
		return self::$integrationEnv;
	}
}
