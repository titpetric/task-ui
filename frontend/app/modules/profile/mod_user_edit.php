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
	const LEAD = "The roles, folders, permissions, keys, MCP access and linked accounts of %s.";

	/** NOBODY is what a name no row holds answers. */
	const NOBODY = "No user by that name.";

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
			"spotify" => $this->spotify($chrome, $name, $own),
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
	 * An admin holds every rule, so the block is the sentence that says so and
	 * no table. Your own changes only where user/edit comes with your role: a
	 * rule of your own that granted it would let you keep granting yourself
	 * more, which the API refuses too.
	 */
	function permissions($chrome, $users, $user, $own) {
		$fixed = Task\User::admin($user);
		$block = array(
			"links" => $chrome->allowed("user/view"),
			"role" => Task\User::role($user),
			"fixed" => $fixed,
			"matrix" => null,
		);

		if ($fixed) {
			return $block;
		}

		$access = $users->access($user);
		$matrix = new Task\Matrix($access->own, $access->effective(), false);

		$matrix->api = "/api/permissions/users/" . $user->getName();
		$matrix->effective = true;
		$matrix->editable = $chrome->allowed("user/edit") && (!$own || self::editsByRole($chrome));
		$block["matrix"] = $matrix->data();

		return $block;
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
	 * mcp is the MCP access block: the tokens the user holds, one row each, as
	 * a page may say them, which is when each was made, when it runs out and
	 * when it last answered.
	 *
	 * Making one takes user/edit and the MCP agent user flag; making and
	 * deleting take that or the tokens being your own. A token is never
	 * rotated: one is made and one is deleted, and a user holds as many as they
	 * name. Nothing here reads a token or its hash: a token exists once, in the
	 * answer to the POST that made it.
	 */
	function mcp($chrome, $user, $own) {
		$name = $user->getName();
		$id = $user->getId();
		$api = "/api/users/" . $id . "/mcp";
		$store = new Task\Mcp\Tokens($chrome->db);
		$rows = $store->ofKind($id, Task\Mcp\Tokens::MCP);
		$agent = $user->getMcp();
		$enable = $chrome->allowed("user/edit") && $agent;

		return array(
			"name" => $name,
			"api" => $api,
			"done" => "/users/" . $name . "/edit?notice=MCP+token+made#mcp",
			"on" => count($rows) > 0,
			"agent" => $agent,
			"enable" => $enable,
			"manage" => $enable || ($own && count($rows) > 0),
			"tokens" => self::rows($rows, $api),
			"fill" => Task\Mcp\Connect::TOKEN,
			"until" => Task\Mcp\Connect::UNTIL,
			"named" => Task\Mcp\Connect::NAMED,
			"add" => Task\Mcp\Connect::add(Task\Route::site()),
			"settings" => Task\Mcp\Connect::settings(Task\Route::site()),
		);
	}

	/**
	 * spotify is the Spotify block: the Connect that links the reader's own
	 * account, on their own page and with spotify_link, and null elsewhere.
	 *
	 * A link is made in the session that starts it, so a reader links their own
	 * account and nobody else's. The block draws hidden and the page's script
	 * shows it once the API says the integration is open to the reader.
	 */
	function spotify($chrome, $name, $own) {
		if (!$own || !$chrome->allowed("spotify_link")) {
			return null;
		}

		return array("next" => Task\Route::href("/users/" . $name . "/edit"));
	}

	/** rows is one table row per token, each with the DELETE of its own. */
	static function rows($tokens, $api) {
		$out = array();

		foreach ($tokens as $token) {
			$created = $token->getCreatedAt();
			$expires = $token->getExpiresAt();
			$used = $token->getUsedAt();
			$out[] = array(
				"name" => $token->getName(),
				"delete" => $api . "/" . $token->getId(),
				"created" => Task\Clock::ago($created),
				"createdAt" => Task\Clock::read($created),
				"expires" => $expires !== "",
				"expiresAt" => Task\Clock::read($expires),
				"expired" => Task\Mcp\Tokens::expired($expires),
				"tone" => Task\Mcp\Tokens::tone($expires),
				"used" => $used !== "",
				"usedAgo" => Task\Clock::ago($used),
				"usedAt" => Task\Clock::read($used),
			);
		}

		return $out;
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
