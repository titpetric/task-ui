<?php

namespace Task;

// Setup is the settings in the order an instance is set up in: the steps the
// Settings page lists, and the trail at the foot of each step's page that
// names the one being read.
//
// The order, the prose and the permission that opens each step are
// frontend/view/setup_step.go's. They are a constant table no binding
// carries, so they are written here; keep them in step with that file.
class Setup {
	/** The steps, by name, as a page names the one it is reading. */
	const INSTANCE = "Instance";

	const RETENTION = "Data retention";

	const TELEMETRY = "Telemetry";

	const KEYS = "Keys";

	const SECRETS = "Secrets";

	const INTEGRATIONS = "Integrations";

	const USERS = "Users";

	const PERMISSIONS = "Permissions";

	const WORKFLOWS = "Workflows";

	const WORKSPACES = "Workspaces";

	const PROJECTS = "Projects";

	const SIDEBAR = "Sidebar";

	/**
	 * order is every step in the order an admin sets an instance up in: the
	 * instance first and how long it keeps what it records, then the keys and
	 * the secrets they encrypt, the integrations, the users and what they may
	 * do, the workflows, and the workspaces and projects they run in.
	 *
	 * A step holds the page it opens, its sprite symbol, what the Settings
	 * page says it holds, and the permission that opens it. An empty rule is a
	 * step every reader of the settings opens.
	 */
	static function order() {
		return array(
			array(
				"name" => self::INSTANCE,
				"href" => "/settings/instance",
				"icon" => "settings",
				"desc" => "The server's settings, from the environment.",
				"rule" => "",
			),
			array(
				"name" => self::RETENTION,
				"href" => "/settings/retention",
				"icon" => "clock",
				"desc" => "How long runs, recordings, threads, artifact history and the read log stay.",
				"rule" => "",
			),
			array(
				"name" => self::TELEMETRY,
				"href" => "/settings/telemetry",
				"icon" => "pulse",
				"desc" => "What the instance records of its own requests, and who reads the traces.",
				"rule" => "",
			),
			array(
				"name" => self::KEYS,
				"href" => "/settings/keys",
				"icon" => "key",
				"desc" => "The encryption key that stores the secrets, and the SSH keys that connect runs and git to hosts.",
				"rule" => "key/view",
			),
			array(
				"name" => self::SECRETS,
				"href" => "/settings/secrets",
				"icon" => "secret",
				"desc" => "Run environment variables and secrets, for every workspace or one.",
				"rule" => "secret/view",
			),
			array(
				"name" => self::INTEGRATIONS,
				"href" => "/settings/integrations",
				"icon" => "plug",
				"desc" => "Claude Code, OpenRouter, Forgejo and Beszel, and their tokens.",
				"rule" => "integration/view",
			),
			array(
				"name" => self::USERS,
				"href" => "/settings/users",
				"icon" => "users",
				"desc" => "Who signs in, their role and folders.",
				"rule" => "user/view",
			),
			array(
				"name" => self::PERMISSIONS,
				"href" => "/settings/permissions",
				"icon" => "key",
				"desc" => "What each role may do.",
				"rule" => "user/view",
			),
			array(
				"name" => self::WORKFLOWS,
				"href" => "/workflows",
				"icon" => "flow",
				"desc" => "Chains of jobs, by hand, schedule or webhook.",
				"rule" => "workflow/view",
			),
			array(
				"name" => self::WORKSPACES,
				"href" => "/settings/workspaces",
				"icon" => "folder-open",
				"desc" => "Directories that hold projects and their tooling.",
				"rule" => "workspace/view",
			),
			array(
				"name" => self::PROJECTS,
				"href" => "/settings/projects",
				"icon" => "folder",
				"desc" => "Git checkouts under the root, pulled before each run.",
				"rule" => "project/view",
			),
			array(
				"name" => self::SIDEBAR,
				"href" => "/settings/sidebar",
				"icon" => "grip",
				"desc" => "The order of the sidebar, and what it leaves out.",
				"rule" => "",
			),
		);
	}

	/**
	 * opens reports whether a reader may open a step. The steps with no rule
	 * of their own open for whoever sees any section of the settings, which is
	 * what the sidebar's own Settings row asks.
	 */
	static function opens($chrome, $rule) {
		if ($rule === "") {
			return Sidebar::settings($chrome->access);
		}

		return $chrome->allowed($rule);
	}

	/**
	 * steps are the steps the reader may open, numbered from 1 in the order
	 * they are set up in, the step the page is reading marked current.
	 *
	 * The links carry the prefix, as every link to a page of this application
	 * does.
	 */
	static function steps($chrome, $step = "") {
		$out = array();

		foreach (self::order() as $entry) {
			if (!self::opens($chrome, $entry["rule"])) {
				continue;
			}

			$entry["number"] = count($out) + 1;
			$entry["current"] = $entry["name"] === $step;
			$entry["href"] = Route::href($entry["href"]);
			$out[] = $entry;
		}

		return $out;
	}

	/**
	 * trail is the foot of a step's page: the step before it and the one after
	 * among those the reader may open, and every step as a trail.
	 *
	 * A reader with one step alone walks nowhere and draws none, and so does a
	 * page whose own step they may not open.
	 */
	static function trail($steps) {
		$count = count($steps);

		if ($count < 2) {
			return array();
		}

		foreach ($steps as $at => $step) {
			if (!$step["current"]) {
				continue;
			}

			$previous = $at > 0 ? $steps[$at - 1] : array();
			$next = $at + 1 < $count ? $steps[$at + 1] : array();

			return array(
				"previous" => $previous,
				"next" => $next,
				"steps" => $steps,
			);
		}

		return array();
	}

	/** says how many of a thing a step holds: 1 key, 2 keys. */
	static function says($found, $one, $many) {
		if ($found === 1) {
			return $found . " " . $one;
		}

		return $found . " " . $many;
	}
}
