<?php

// @route GET /settings/users

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_users is who may sign in: a row per user with the role they act with,
// their flags, the folders they see and when they were added.
//
// The page reads and does not write. Add user opens the form, and Delete in a
// row's menu puts to the API, which gates the change again.
class mod_users implements Module {
	/** LEAD is the sentence under the heading, as HTML: it emphasises a menu item. */
	const LEAD = "Who can sign in, their role and folders. A name opens the user's profile; <b>Edit</b> in the row's menu opens their roles, folders, permissions, keys and MCP access.";

	/** DROPS is what deleting a user costs, which the confirmation says. */
	const DROPS = "? This removes them from the password file and ends their sessions and tokens right away. Their runs, notes and history keep the name.";

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
		return array("user");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("user/view")) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::USERS, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::USERS, "href" => ""),
		));

		$instance = Task\Host::instance(false);
		$adds = Task\Row::text($instance, "htpasswd") !== "";
		$users = new Task\User($chrome->db);
		$steps = Task\Setup::steps($chrome, Task\Setup::USERS);
		$rows = array();

		foreach ($users->all() as $user) {
			$rows[] = $this->row($chrome, $user, $adds);
		}

		$view->page(array(
			"head" => $this->head($adds),
			"rows" => $rows,
			"mcp" => $this->mcp($chrome),
			"trail" => Task\Setup::trail($steps),
		), "users.tpl");
	}

	/** head is the page head: the heading, the lead and Add user. */
	function head($adds) {
		$buttons = new Task\Buttons;
		$head = new Task\Head(Task\Setup::USERS, self::LEAD);

		if ($adds) {
			$buttons->primary("plus", "Add user", array("href" => Task\Route::href("/settings/users/new")));
		}

		$head->buttons($buttons->data());

		return $head->data();
	}

	/** row is one user: the cells the table draws and the menu beside them. */
	function row($chrome, $user, $adds) {
		$name = $user->getName();
		$role = Task\User::role($user);
		$folders = Task\User::folders($user);
		$created = $user->getCreatedAt();
		$menu = $this->menu($chrome, $user, $adds);

		return array(
			"name" => $name,
			"href" => Task\Route::href("/users/" . $name),
			"avatar" => Task\Who::avatar($name, $role),
			"role" => $role,
			"admin" => Task\User::admin($user),
			"roleHref" => Task\Route::href("/settings/permissions#role-" . $role),
			"mcp" => $user->getMcp(),
			"reply" => $user->getAgentReply(),
			"folders" => implode(", ", $folders),
			"ago" => Task\Clock::ago($created),
			"at" => Task\Clock::read($created),
			"menu" => $menu->data(),
		);
	}

	/**
	 * menu is a row's actions: the pages of the user, the runs they started,
	 * and the Delete that takes them out of the password file.
	 *
	 * Delete is there for a reader who may remove a user, on an instance with a
	 * password file to remove them from, and never on an admin's row or the
	 * reader's own.
	 *
	 * New MCP token is one click and draws nothing: the script posts it, takes
	 * the token out of the answer and puts the `claude mcp add` line that
	 * carries it on the clipboard. The token is never shown, so the list stays
	 * a list.
	 */
	function menu($chrome, $user, $adds) {
		$name = $user->getName();
		$id = $user->getId();
		$reader = $chrome->user;
		$drops = $adds && $chrome->allowed("user/delete") && !Task\User::admin($user);
		$menu = new Task\Menu("Actions on " . $name);

		$menu->item("edit", "Edit", array("href" => Task\Route::href("/users/" . $name . "/edit")));
		$menu->item("users", "Profile", array("href" => Task\Route::href("/users/" . $name)));
		$menu->item("list", "Runs by this user", array("href" => Task\Route::href("/runs?user=" . Task\Route::queryEscape($name))));
		$menu->when($chrome->allowed("user/edit") && $user->getMcp(), "copy", "New MCP token", array(
			"data-api" => "/api/users/" . $id . "/mcp",
			"data-method" => "POST",
			"data-copy-fill" => Task\User\Connect::add(Task\Route::site()),
			"data-notice" => "MCP token for " . $name . " copied as the claude mcp add command",
		));
		$menu->whenDanger($drops && $reader !== null && $reader->getId() !== $id, "trash", "Delete", array(
			"data-delete" => "/api/users/" . $id,
			"data-confirm" => "Delete the user " . $name . self::DROPS,
			"data-confirm-ok" => "Delete",
			"data-next" => "/settings/users?notice=User+deleted",
		));

		return $menu;
	}

	/**
	 * mcp is the Connect an MCP client block: the base link and the endpoint to
	 * copy, and the line that adds this instance to Claude Code.
	 *
	 * The token is nobody's to read here. The block says where to turn one on,
	 * which is the reader's own MCP access.
	 */
	function mcp($chrome) {
		$base = Task\Route::site();
		$name = $chrome->name();
		$add = 'claude mcp add --transport http --scope user task ' . $base . '/api/mcp --header "Authorization: Bearer <token>"';

		return array(
			"name" => $name,
			"edit" => Task\Route::href("/users/" . $name . "/edit#mcp"),
			"base" => Task\Copy::of($base, "the base link"),
			"endpoint" => Task\Copy::of($base . "/api/mcp", "the MCP endpoint"),
			"add" => Task\Copy::of($add, "the claude mcp add command"),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_users;

	$mod->serve();
}
