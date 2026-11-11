<?php

// @route GET /settings

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_settings is the index of the settings: the steps in the order an
// instance is set up in, each with how much it holds.
//
// The page reads and does not write. It is the one step that is not a step
// itself, so it draws no trail at the foot: a reader is at the start of the
// walk and every row is a way into it.
class mod_settings implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "Set up the instance in this order: each step builds on the ones before it. Every page ends with links to the step before and the step after.";

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
		return array("secret", "user", "workspace", "workflow", "workflow_step");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!Task\Sidebar::settings($chrome->access)) {
			$view->refused();

			return;
		}

		$chrome->heading("Settings", array(array("text" => "Settings", "href" => "")));

		$head = new Task\Head("Settings", self::LEAD);
		$steps = Task\Setup::steps($chrome, "");
		$counts = $this->counts($chrome);
		$rows = array();

		foreach ($steps as $step) {
			$name = $step["name"];
			$step["count"] = isset($counts[$name]) ? $counts[$name] : "";
			$rows[] = $step;
		}

		$view->page(array("head" => $head->data(), "steps" => $rows));
	}

	/**
	 * counts is how much each step holds, by the step's name: a step with
	 * nothing to count has no entry and its cell stays empty.
	 *
	 * A count is only read where the reader may see what it counts, so the
	 * question each step asks is the permission that opens it. The Instance,
	 * Data retention, Telemetry and Sidebar steps count nothing: what they
	 * hold is one page of settings and not a list.
	 */
	function counts($chrome) {
		$out = array();

		if ($chrome->allowed("key/view")) {
			$keystore = Task\Key::folder();
			$out[Task\Setup::KEYS] = Task\Setup::says(count($keystore["keys"]), "key", "keys");
		}

		if ($chrome->allowed("secret/view")) {
			$secrets = new Task\Secret;
			$out[Task\Setup::SECRETS] = Task\Setup::says($secrets->count(), "variable", "variables");
		}

		if ($chrome->allowed("user/view")) {
			$users = new Task\User;
			$out[Task\Setup::USERS] = Task\Setup::says($users->count(), "user", "users");
			$roles = count(Task\User::ROLES);
			$out[Task\Setup::PERMISSIONS] = Task\Setup::says($roles, "role", "roles");
		}

		if ($chrome->allowed("workflow/view")) {
			$workflows = new Task\Workflow;
			$seen = $workflows->visible($chrome);
			$out[Task\Setup::WORKFLOWS] = Task\Setup::says(count($seen), "workflow", "workflows");
		}

		$workspaces = new Task\Workspace;
		$folders = $workspaces->names();

		if ($chrome->allowed("workspace/view")) {
			$out[Task\Setup::WORKSPACES] = Task\Setup::says(count($folders), "workspace", "workspaces");
		}

		if ($chrome->allowed("project/view")) {
			$entries = Task\Host::entries($folders);
			$out[Task\Setup::PROJECTS] = Task\Setup::says(count($entries), "folder", "folders");
		}

		return $out;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_settings;

	$mod->serve();
}
