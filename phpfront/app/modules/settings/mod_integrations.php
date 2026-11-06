<?php

// @route GET /settings/integrations

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_integrations is the services task-ui talks to: a card each with what the
// service is for, the settings it takes, and the steps to the credential among
// them.
//
// The page reads and does not write. Save puts to the API, Clear in a card's
// menu deletes through it, and the sync buttons start a run.
//
// Neither half of a card is a row of this database. The registry is Go data,
// down to every field's label and the steps to its token, and the settings are
// secrets the instance key encrypts, a key that does not enter PHP; both
// arrive from Task\Host::integrations(), where a secret crosses as set or empty
// and never as its value.
class mod_integrations implements Module {
	/** LEAD opens the sentence under the heading and KEPT closes it, the key between them. */
	const LEAD = "Tokens are encrypted with the instance key ";

	const KEPT = "; leave a secret field empty to keep it.";

	/** NONE stands where the instance has no encryption key yet. */
	const NONE = '<span class="badge bad">none yet</span>';

	/**
	 * GATEWAY draws its variables as a table and not a field each: the reader
	 * fills in the one key and the rest are what a run gets.
	 *
	 * FORGEJO is the one card with a sync behind it: the hook its token
	 * installs, the buttons that run the job and the line on its last run.
	 *
	 * Both are integration/kind.go's IDs, as the Go page's own conditions are;
	 * keep them in step with that file.
	 */
	const GATEWAY = "openrouter";

	const FORGEJO = "forgejo";

	/** TOKEN is the one setting of the gateway the reader types. */
	const TOKEN = "token";

	/** SYNC is the job that syncs the git host, in the default workspace. */
	const SYNC = "@forgejo-sync";

	/** The kinds of field, which the field component draws one branch each for. */
	const WHO = "who";

	const ENV = "env";

	const SECRET = "secret";

	const SELECT = "select";

	const TEXT = "text";

	/** KEEPS is the placeholder of a secret that has a value. */
	const KEEPS = "leave empty to keep it";

	/** EVERYONE is what the Who field says where nothing is checked. */
	const EVERYONE = "every run gets the token";

	/** HOOK says where a sync installs the push hook, and NOHOOK that it installs none. */
	const HOOK = "Sync installs the push hook <code>";

	const HOOKED = "/api/webhooks/forgejo</code> on every repository it syncs.";

	const NOHOOK = "The instance has no public URL (<code>TASK_UI_URL</code>), so sync installs no push hook.";

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
		return array("user", "run");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("integration/view")) {
			$view->refused();

			return;
		}

		$chrome->heading(Task\Setup::INTEGRATIONS, array(
			array("text" => "Settings", "href" => "/settings"),
			array("text" => Task\Setup::INTEGRATIONS, "href" => ""),
		));

		$users = new Task\User($chrome->db);
		$everyone = $users->all();
		$instance = Task\Host::instance(false);
		$url = Task\Row::text($instance, "url");
		$admin = Task\User::admin($chrome->user);
		$last = $this->sync($chrome);
		$steps = Task\Setup::steps($chrome, Task\Setup::INTEGRATIONS);
		$cards = array();

		foreach (Task\Host::integrations() as $at => $service) {
			$cards[] = $this->card($service, $at + 1, $everyone, $url, $admin, $last);
		}

		$view->page(array(
			"head" => $this->head(),
			"cards" => $cards,
			"trail" => Task\Setup::trail($steps),
		), "integrations.tpl");
	}

	/** head is the page head: the heading and the key the settings are stored on. */
	function head() {
		$folder = Task\Key::folder();
		$head = new Task\Head(Task\Setup::INTEGRATIONS, self::lead($folder));

		return $head->data();
	}

	/**
	 * lead is the sentence under the heading, as HTML: the fingerprint of the
	 * instance's encryption key, which every setting here is stored on.
	 *
	 * The key folder names its encryption key, so the key is the row of that
	 * name and the page reads no scope of its own to find it.
	 */
	static function lead($folder) {
		$named = Task\Row::text($folder, "instance");
		$key = self::NONE;

		foreach ($folder["keys"] as $row) {
			if ($row["name"] === $named) {
				$key = '<b class="mono-break">' . htmlspecialchars($row["fingerprint"], ENT_QUOTES) . "</b>";
			}
		}

		return self::LEAD . $key . self::KEPT;
	}

	/**
	 * card is one service: its state, its menu, what it is for, the form of
	 * its settings, the steps to its credential and, for the git host, the
	 * sync behind it.
	 */
	function card($service, $number, $everyone, $url, $admin, $last) {
		$id = $service["id"];
		$title = $service["title"];
		$menu = $this->menu($service);
		$sync = $id === self::FORGEJO;
		$gateway = $id === self::GATEWAY;

		return array(
			"id" => $id,
			"number" => $number,
			"title" => $title,
			"configured" => $service["configured"],
			"menu" => $menu->data(),
			"desc" => Task\Host::message($service["description"]),
			"api" => "/api/integrations/" . $id,
			"next" => "/settings/integrations?notice=" . $title . "+saved",
			"fields" => $this->fields($service, $everyone),
			"env" => $gateway ? $this->env($service) : array(),
			"credential" => $service["credential"],
			"steps" => self::steps($service),
			"hook" => $sync ? self::hook($url) : "",
			"runs" => $sync && $service["configured"] && $admin,
			"sync" => $sync,
			"last" => $last,
		);
	}

	/** menu is the card's own actions: the Clear that takes every setting out. */
	function menu($service) {
		$title = $service["title"];
		$menu = new Task\Menu("Actions on " . $title);

		$menu->danger("trash", "Clear", array(
			"data-delete" => "/api/integrations/" . $service["id"],
			"data-confirm" => "Clear every setting of " . $title . "?",
			"data-confirm-ok" => "Clear",
			"data-next" => "reload",
		));

		return $menu;
	}

	/**
	 * fields are the settings the form draws a field for.
	 *
	 * The gateway draws the key alone: its other settings are the variables a
	 * run gets, which the table under the form holds.
	 */
	function fields($service, $everyone) {
		$gateway = $service["id"] === self::GATEWAY;
		$out = array();

		foreach ($service["fields"] as $field) {
			if ($gateway && $field["name"] !== self::TOKEN) {
				continue;
			}

			$out[] = $this->field($service, $field, $everyone);
		}

		return $out;
	}

	/**
	 * field is one setting as its branch of the form draws it: the Who boxes,
	 * a switch alone, a secret, a list to pick from or a line of text.
	 *
	 * A secret holds no value: the field says whether one is set, and the
	 * placeholder says that an empty field keeps it.
	 *
	 * spaced is the space that separates the help from the variable under the
	 * input, which the field writes whether a variable follows it or not. The
	 * gateway's key is the one field whose help is the whole line, because its
	 * variables are the table under the form and not the field's own.
	 */
	function field($service, $field, $everyone) {
		$name = $field["name"];
		$kind = self::kind($field);
		$set = isset($service["set"][$name]) && $service["set"][$name];
		$value = self::value($service, $field);
		$secret = $kind === self::SECRET;

		return array(
			"kind" => $kind,
			"id" => $service["id"] . "-" . $name,
			"name" => "values." . $name,
			"label" => $field["label"],
			"wide" => $kind === self::WHO || $kind === self::ENV || $secret,
			"badge" => $secret ? ($set ? "set" : "empty") : "",
			"tone" => $secret && $set ? "ok" : "",
			"help" => $field["help"],
			"spaced" => $service["id"] !== self::GATEWAY,
			"env" => $field["env"],
			"placeholder" => $secret && $set ? self::KEEPS : $field["placeholder"],
			"value" => $value,
			"options" => self::options($field, $value),
			"toggle" => $field["toggle"],
			"switch" => $this->flag($service, $field),
			"who" => $kind === self::WHO ? $this->whose($service, $field, $everyone) : array(),
		);
	}

	/** kind is which branch of the form draws the field. */
	static function kind($field) {
		if ($field["who"]) {
			return self::WHO;
		}

		if ($field["empty"] || $field["source"] !== "") {
			return self::ENV;
		}

		if ($field["secret"]) {
			return self::SECRET;
		}

		if (count($field["options"]) > 0) {
			return self::SELECT;
		}

		return self::TEXT;
	}

	/**
	 * value is what the field holds: the setting as it stands, and the default
	 * where the setting is empty. A secret holds none.
	 */
	static function value($service, $field) {
		$name = $field["name"];
		$value = isset($service["values"][$name]) ? $service["values"][$name] : "";

		if ($value !== "") {
			return $value;
		}

		if (count($field["options"]) > 0) {
			return $field["options"][0];
		}

		return $field["default"];
	}

	/** options are the values the setting takes, the one it holds selected. */
	static function options($field, $value) {
		$out = array();

		foreach ($field["options"] as $option) {
			$out[] = array("value" => $option, "selected" => $option === $value);
		}

		return $out;
	}

	/**
	 * flag is the switch that sends the setting's variable in a run or leaves
	 * it out; the empty array for a setting without one.
	 */
	function flag($service, $field) {
		$name = $field["name"];
		$kind = self::kind($field);

		if ($kind !== self::ENV && !$field["toggle"]) {
			return array();
		}

		return array(
			"name" => "values." . $name . ".enabled",
			"env" => $field["env"],
			"on" => isset($service["enabled"][$name]) && $service["enabled"][$name],
		);
	}

	/** whose is the Who boxes of the setting: the roles and users its runs go to. */
	function whose($service, $field, $everyone) {
		$checked = $service["checked"];
		$empty = isset($field["who_empty"]) ? $field["who_empty"] : self::EVERYONE;
		$who = new Task\WhoField("values." . $field["name"], "joined", $checked, $checked, $everyone, $empty);

		return $who->data();
	}

	/**
	 * env are the variables of the gateway as the table under its form draws
	 * them: the two the key itself fills, the one sent empty, and the rest as
	 * the fields they are.
	 */
	function env($service) {
		$out = array();

		foreach ($service["fields"] as $field) {
			if ($field["name"] === self::TOKEN || $field["env"] === "") {
				continue;
			}

			$out[] = array(
				"env" => $field["env"],
				"help" => $field["help"],
				"source" => $field["source"] !== "",
				"empty" => $field["empty"],
				"id" => $service["id"] . "-" . $field["name"],
				"name" => "values." . $field["name"],
				"value" => self::value($service, $field),
			);
		}

		return $out;
	}

	/**
	 * steps are how to get the service's credential, each rendered from the
	 * markdown it is written in.
	 *
	 * The renderer is the one a note's text goes through and not the one a
	 * file goes through: the prose of the registry is written as a note is,
	 * with `code` in it and no smart quotes, and no name in it is a mention.
	 */
	static function steps($service) {
		$out = array();

		foreach ($service["steps"] as $step) {
			$out[] = Task\Host::message($step);
		}

		return $out;
	}

	/**
	 * hook is the line under the git host's steps: the push hook a sync
	 * installs, which is the instance's public URL and the webhook path, and
	 * what an instance with no public URL installs instead, which is nothing.
	 */
	static function hook($url) {
		if ($url === "") {
			return self::NOHOOK;
		}

		return self::HOOK . htmlspecialchars($url, ENT_QUOTES) . self::HOOKED;
	}

	/**
	 * sync is the last run of the sync job, for the line the git host's card
	 * ends with: how it went, when it ran, who ran it and its report.
	 *
	 * The runner is ahead of the table, so the session is laid over the row
	 * before it is read, as every page that draws a run reads one. The empty
	 * array is no sync yet.
	 */
	function sync($chrome) {
		$runs = new Task\Run($chrome->db);
		$found = $runs->find(array(
			"folder" => Task\Route::INTERNAL,
			"job" => self::SYNC,
			"limit" => 1,
		));

		if (count($found) === 0) {
			return array();
		}

		$run = $found[0];
		$session = Task\Host::session($run->getId());

		Task\Run::overlay($run, $session);

		return array(
			"state" => Task\Run::label($run),
			"ago" => Task\Clock::ago($run->getStartedAt()),
			"who" => $run->getUserName(),
			"href" => Task\Route::href("/runs/" . $run->getId()),
		);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_integrations;

	$mod->serve();
}
