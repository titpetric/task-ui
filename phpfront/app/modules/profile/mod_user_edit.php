<?php

// @route GET /users/{name}/edit

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_user_edit holds the forms of one user: the roles they act with, the
// folders they see, their own permissions, the keys their runs get and their
// MCP access.
//
// It opens as the profile does, for your own page or with user/view, and then
// each block decides for itself what the reader may change: the roles are an
// admin's to set, the folders take user/edit, the permissions take user/edit
// through a role, and the MCP token is the user's own once it is on.
//
// Nothing here writes. Every form puts to the API, which gates the change
// again, so a block drawn to a reader who may not change it is a block the API
// would refuse.
class mod_user_edit implements Module {
	/** LEAD is the sentence under the heading, which names the blocks. */
	const LEAD = "The roles, folders, permissions, keys and MCP access of %s.";

	/** NOBODY is what a name no row holds answers. */
	const NOBODY = "No user by that name.";

	/**
	 * FILL and UNTIL are the placeholders task-ui.js fills from the answer to
	 * Enable or Rotate: the token itself and the day it runs out.
	 *
	 * They are constants of the module and not markup of the template, because
	 * MiniTPL compiles any {word} into an echo: a template cannot write the
	 * literal, so it arrives as a field of the scope.
	 */
	const FILL = "{token}";

	const UNTIL = "{expires_at}";

	/** ADD is the line that adds this instance to Claude Code, with FILL in it. */
	const ADD = 'claude mcp add --transport http --scope user task %s/api/mcp --header "Authorization: Bearer {token}"';

	function name() {
		return "profile";
	}

	function title() {
		return "Profile";
	}

	function icon() {
		return "users";
	}

	function tables() {
		return array("user", "permission_policy", "user_auth_token");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);
		$name = Task\Route::param("name");
		$own = $name !== "" && $name === $chrome->name();

		if (!$own && !$chrome->allowed("user/view")) {
			$view->refused();

			return;
		}

		$users = new Task\User($chrome->db);
		$user = $users->byName($name);

		if ($user === null) {
			$view->missing(self::NOBODY);

			return;
		}

		$chrome->heading("Edit " . $name, $this->crumbs($chrome, $name));

		$view->page(array(
			"head" => $this->head($name),
			"name" => $name,
			"api" => "/api/users/" . $user->getId(),
			"view" => Task\Route::href("/users/" . $name),
			"roles" => $this->roles($chrome, $user),
			"folders" => $this->folders($chrome, $user),
			"permissions" => $this->permissions($chrome, $users, $user, $own),
			"keys" => $this->keys($chrome, $name, $own),
			"mcp" => $this->mcp($chrome, $user, $own),
		), "user_edit.tpl");
	}

	/**
	 * crumbs is the trail above the page: the user and then Edit for a reader
	 * who came to one user, and the path through the settings for one who may
	 * read the list of them.
	 */
	function crumbs($chrome, $name) {
		$profile = "/users/" . Task\Route::pathEscape($name);

		if (!$chrome->allowed("user/view")) {
			return array(array("text" => $name, "href" => $profile), array("text" => "Edit", "href" => ""));
		}

		return array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => "Users", "href" => "/settings/users"),
			array("text" => $name, "href" => $profile),
			array("text" => "Edit", "href" => ""),
		);
	}

	/** head is the page head: the heading, the lead and the way back to the profile. */
	function head($name) {
		$buttons = new Task\Buttons;
		$head = new Task\Head("Edit " . $name);

		$buttons->item("users", "View profile", array("href" => Task\Route::href("/users/" . $name)));
		$head->said(sprintf(self::LEAD, $name));
		$head->buttons($buttons->data());

		return $head->data();
	}

	/**
	 * roles is the Roles block: the admin flag, the role without it, and the
	 * two switches that let an agent act as the user.
	 *
	 * The form is an admin's alone, because what it sets is who may administer
	 * the instance; everybody else reads the same facts as a sentence.
	 */
	function roles($chrome, $user) {
		$name = $user->getName();
		$edits = Task\User::admin($chrome->user);
		$next = "/users/" . $name . "/edit?notice=Roles+saved#roles";

		return array(
			"edits" => $edits,
			"next" => Task\Route::href($next),
			"options" => Task\User::ASSIGNABLE,
			"role" => $user->getRole(),
			"admin" => $user->getAdmin(),
			"mcp" => $user->getMcp(),
			"reply" => $user->getAgentReply(),
		);
	}

	/**
	 * folders is the Folders block: All folders, or a box per folder of the
	 * catalogue.
	 *
	 * The catalogue is the whole of it and not the reader's own: what the form
	 * sets is what the user sees, and a folder the reader does not see is still
	 * a folder the user may be given.
	 */
	function folders($chrome, $user) {
		$name = $user->getName();
		$seen = Task\User::folders($user);
		$next = "/users/" . $name . "/edit?notice=Folders+saved#folders";
		$all = count($seen) === 0;
		$pick = array(
			"all" => $all,
			"label" => false,
			"choices" => Task\Folders::choices($chrome->all(), $seen),
		);

		return array(
			"edits" => $chrome->allowed("user/edit"),
			"next" => Task\Route::href($next),
			"pick" => $pick,
			"list" => $seen,
		);
	}

	/**
	 * permissions is the Permissions block: the user's own rules, which decide
	 * the ones the role's rules would otherwise.
	 *
	 * An admin's is fixed and the paragraph says why. Your own changes only
	 * where user/edit comes with your role: a rule of your own that granted it
	 * would let you keep granting yourself more, which the API refuses too.
	 */
	function permissions($chrome, $users, $user, $own) {
		$access = $users->access($user);
		$fixed = Task\User::admin($user);
		$can = $access->effective();
		$matrix = new Task\Matrix($access->own, $can, $fixed);

		$matrix->api = "/api/permissions/users/" . $user->getName();
		$matrix->effective = true;
		$matrix->editable = !$fixed && $chrome->allowed("user/edit") && (!$own || self::editsByRole($chrome));

		return array(
			"links" => $chrome->allowed("user/view"),
			"role" => Task\User::role($user),
			"matrix" => $matrix->data(),
		);
	}

	/**
	 * keys is the Keys block, for your own page and for a reader with
	 * key/view, and null for anybody else.
	 *
	 * A change made here returns here, so the table's $next is this page and
	 * not the profile's.
	 */
	function keys($chrome, $name, $own) {
		if (!$own && !$chrome->allowed("key/view")) {
			return null;
		}

		$folder = Task\Key::folder();
		$linked = Task\Key::linked($folder["keys"], $name);

		return Task\Key::ofUser($chrome, $name, $linked, "/users/" . $name . "/edit");
	}

	/**
	 * mcp is the MCP access block: the user's token as a page may say it, which
	 * is when it was made, when it runs out and when it last answered.
	 *
	 * Turning it on takes user/edit and the MCP agent user flag; rotating and
	 * disabling take that or the token being your own. Nothing here reads the
	 * token or its hash: the token exists once, in the answer to Enable.
	 */
	function mcp($chrome, $user, $own) {
		$name = $user->getName();
		$id = $user->getId();
		$tokens = new Task\Token($chrome->db);
		$token = $tokens->get($id, Task\Token::MCP);
		$on = $token !== null;
		$agent = $user->getMcp();
		$enable = $chrome->allowed("user/edit") && $agent;
		$created = $on ? $token->getCreatedAt() : "";
		$expires = $on ? $token->getExpiresAt() : "";
		$used = $on ? $token->getUsedAt() : "";

		return array(
			"name" => $name,
			"api" => "/api/users/" . $id . "/mcp",
			"done" => "/users/" . $name . "/edit#mcp",
			"on" => $on,
			"agent" => $agent,
			"enable" => $enable,
			"manage" => $enable || ($own && $on),
			"created" => Task\Clock::ago($created),
			"createdAt" => Task\Clock::read($created),
			"expires" => $expires !== "",
			"expiresAt" => Task\Clock::read($expires),
			"expired" => Task\Token::expired($expires),
			"tone" => Task\Token::tone($expires),
			"used" => $used !== "",
			"usedAgo" => Task\Clock::ago($used),
			"usedAt" => Task\Clock::read($used),
			"fill" => self::FILL,
			"until" => self::UNTIL,
			"add" => sprintf(self::ADD, Task\Route::site()),
		);
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
	$mod = new mod_user_edit;

	$mod->serve();
}
