<?php

// @route GET /settings/permissions

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_permissions is what each role may do: a policy per role, the users who
// act with it and the instance SSH keys linked to it.
//
// The page reads and does not write. A matrix puts the cells that changed to
// /api/permissions/roles/<role>, which gates the change again.
class mod_permissions implements Module {
	/** LEAD is the sentence under the heading. */
	const LEAD = "What each role may do. Click a box to cycle it: inherit, ✓ allow, ✕ deny. A user's own allow or deny, on their profile, decides first.";

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
		return array("user", "permission_policy");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("user/view")) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::PERMISSIONS, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::PERMISSIONS, "href" => ""),
		));

		$head = new Task\Head(Task\Setup::PERMISSIONS, self::LEAD);
		$steps = Task\Setup::steps($chrome, Task\Setup::PERMISSIONS);
		$views = $chrome->allowed("key/view");
		$members = $this->members($chrome, $views);
		$matrices = array();

		foreach (Task\User::ROLES as $role) {
			$matrices[] = $this->matrix($chrome, $role, $members);
		}

		$view->page(array(
			"head" => $head->data(),
			"addUser" => $this->addUser($chrome),
			"viewsKeys" => $views,
			"generate" => $chrome->allowed("key/edit") ? Task\Route::href("/settings/keys/generate") : "",
			"matrices" => $matrices,
			"trail" => Task\Setup::trail($steps),
		), "permissions.tpl");
	}

	/**
	 * matrix is one role's block: its policy, and who holds the role.
	 *
	 * The admin role's is fixed and reads no policy row: an admin may do
	 * everything whatever a row says, so the table allows every cell and the
	 * heading says so instead of drawing a Save.
	 */
	function matrix($chrome, $role, $members) {
		$fixed = $role === Task\User::ADMIN;
		$policies = new Task\Settings\Policies($chrome->db);
		$policy = new Task\Policy;

		if (!$fixed) {
			$policy = $policies->policy(Task\Policy::ROLE, $role);
		}

		$acts = new Task\Model\User(array("role" => $role));
		$access = new Task\Access($acts, $policy);
		$effective = $access->effective();
		$matrix = new Task\Matrix($policy, $effective, $fixed);
		$held = $members[$role];

		$matrix->api = "/api/permissions/roles/" . $role;
		$matrix->editable = !$fixed && $chrome->allowed("user/edit") && ($role !== Task\User::role($chrome->user) || self::editsByRole($chrome));

		return array(
			"id" => "role-" . $role,
			"title" => $role,
			"initials" => Task\Who::initials($role),
			"fixed" => $fixed,
			"editable" => $matrix->editable,
			"users" => self::chips($held["users"], "/users/"),
			"keys" => self::chips($held["keys"], "/settings/keys/"),
			"matrix" => $matrix->data(),
		);
	}

	/** chips is a list of names as the links a role's line draws. */
	static function chips($names, $under) {
		$out = array();

		foreach ($names as $name) {
			$out[] = array("name" => $name, "href" => Task\Route::href($under . $name));
		}

		return $out;
	}

	/**
	 * members is what each role holds: the users who act with it and the
	 * instance SSH keys linked to it, each list by name.
	 *
	 * A user counts under the role they act with, so an admin flag puts them
	 * under admin. A workspace's key and the encryption key belong to no role,
	 * and the keys are left out of every list for a reader without key/view.
	 * frontend/view/role_members.go:21.
	 */
	function members($chrome, $views) {
		$users = new Task\User($chrome->db);
		$out = array();

		foreach (Task\User::ROLES as $role) {
			$out[$role] = array("users" => array(), "keys" => array());
		}

		foreach ($users->all() as $user) {
			$role = Task\User::role($user);

			if (isset($out[$role])) {
				$out[$role]["users"][] = $user->getName();
			}
		}

		if ($views) {
			$folder = Task\Key::folder();

			foreach ($folder["keys"] as $key) {
				if ($key["purpose"] !== Task\Key::SSH || $key["workspace"] !== "") {
					continue;
				}

				foreach ($key["roles"] as $role) {
					if (isset($out[$role])) {
						$out[$role]["keys"][] = $key["name"];
					}
				}
			}
		}

		foreach (Task\User::ROLES as $role) {
			sort($out[$role]["users"]);
			sort($out[$role]["keys"]);
		}

		return $out;
	}

	/**
	 * addUser is the Add user link each role's Users line ends with, "" for a
	 * reader who may not add one or an instance with no password file to add
	 * them to.
	 */
	function addUser($chrome) {
		if (!$chrome->allowed("user/edit")) {
			return "";
		}

		$instance = Task\Host::instance(false);

		if (Task\Row::text($instance, "htpasswd") === "") {
			return "";
		}

		return Task\Route::href("/settings/users/new");
	}

	/**
	 * editsByRole reports whether the reader holds user/edit through the role
	 * they act with, which a rule of their own alone does not.
	 */
	static function editsByRole($chrome) {
		if (Task\User::admin($chrome->user)) {
			return true;
		}

		return $chrome->access->role->effect("user/edit") === Task\Policy::ALLOW;
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_permissions;

	$mod->serve();
}
