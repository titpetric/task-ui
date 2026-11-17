<?php

namespace Task;

// Sidebar draws the top block of the sidebar: the modules whose modinfo.php
// enabled a row, in the order the table holds, and what each one links.
//
// The modules are the front's own, read from their modinfo.php. The table
// holds their order and their hidden flag and nothing else about them: a
// row's title, icon and rule are the module's own.
class Sidebar {
	/** VIEW and MODULE are the two places a link opens. */
	const VIEW = "view";

	const MODULE = "module";

	/** SETTINGS is the one rule that is not a permission path. */
	const SETTINGS = "settings";

	/** MODULE_RULE is the rule of a row that opens a module, model/sidebar.go's RuleModule. */
	const MODULE_RULE = "module";

	/** store is the settings module, which owns the table. */
	var $store;

	/** modules is the front's own modules, read once when the block is drawn. */
	var $modules;

	function __construct($db = null, $modules = null) {
		require_once dirname(__DIR__) . "/modules.php";

		$this->store = new Settings\Sidebars($db);
		$this->modules = $modules === null ? task_modules() : $modules;
	}

	/** all is every row, hidden ones included, in the order the table holds. */
	function all() {
		return $this->store->all();
	}

	/** byHref is one row by its link, or null. */
	function byHref($href) {
		return $this->store->byHref($href);
	}

	/** order writes the position and the hidden flag of each link. */
	function order($hrefs, $hidden) {
		return $this->store->order($hrefs, $hidden);
	}

	/**
	 * items is the top block as one user sees it: the modules that enabled a
	 * row, in the table's order, less the hidden rows and the rows whose rule
	 * the user lacks. A module that enabled none is served and not listed.
	 */
	function items($access, $path) {
		$settings = self::settings($access);
		$out = array();

		foreach (self::listing($this->modules, $this->all()) as $row) {
			if ($row["hidden"] || !self::sees($row["rule"], $access, $settings)) {
				continue;
			}

			$out[] = array(
				"title" => $row["title"],
				"href" => $row["link"],
				"icon" => $row["icon"],
				"active" => self::active($row["href"], $path),
				"flows" => $row["href"] === "/workflows",
				"pulse" => $row["href"] === "/threads",
			);
		}

		return $out;
	}

	/** seed writes the module rows the table does not hold yet. */
	function seed() {
		$items = $this->all();
		$known = array();

		foreach ($items as $item) {
			$known[$item->getHref()] = true;
		}

		$rows = array();

		foreach (self::listing($this->modules, $items) as $row) {
			if (isset($known[$row["href"]])) {
				continue;
			}

			$row["kind"] = self::kind($row["href"]);
			$rows[] = $row;
		}

		$this->store->ensure($rows);
	}

	/**
	 * rows are the rows of the Sidebar settings table: every module that
	 * enabled a row, the hidden ones too, because this is the page that
	 * unhides one, in the order the table holds.
	 */
	function rows($items) {
		$out = array();

		foreach (self::listing($this->modules, $items) as $row) {
			$row["kind"] = self::kind($row["href"]);
			unset($row["position"], $row["order"]);
			$out[] = $row;
		}

		return $out;
	}

	/**
	 * listing pairs the modules that enabled a row with the saved row of
	 * their link: the order and the hidden flag are the table's, the title,
	 * icon, rule and description are the module's own. A module that enabled
	 * no row, and a second module of a link already listed, are left out.
	 */
	static function listing($modules, $items) {
		$saved = array();

		foreach ($items as $item) {
			$saved[$item->getHref()] = $item;
		}

		$seen = array();
		$out = array();

		foreach ($modules as $index => $module) {
			$href = $module->getSideBarHref();

			if (!$module->getSidebar() || $href === "" || isset($seen[$href])) {
				continue;
			}

			$seen[$href] = true;
			$row = isset($saved[$href]) ? $saved[$href] : null;
			$out[] = array(
				"href" => $href,
				"link" => self::link($href),
				"title" => $module->getTitle(),
				"icon" => self::icon($module->getIcon()),
				"rule" => $module->getSideBarRule(),
				"note" => $module->getDescription(),
				"hidden" => $row !== null && $row->getHidden(),
				"position" => $row === null ? 1000 + $index : $row->getPosition(),
				"order" => $index,
			);
		}

		usort($out, function($a, $b) {
			if ($a["position"] === $b["position"]) {
				return $a["order"] - $b["order"];
			}

			return $a["position"] - $b["position"];
		});

		return $out;
	}

	/**
	 * kind says where a link opens: a module's own page under /m is another
	 * application, and any other link is a page of this one.
	 */
	static function kind($href) {
		return strpos($href, "/m/") === 0 ? self::MODULE : self::VIEW;
	}

	/**
	 * link is a row's whole link, the prefix in it where the prefix applies.
	 *
	 * A module's page is another application on the same instance, mounted at
	 * /m/<vendor>/<name> and not under this one, so the prefix is this
	 * application's own rows alone. Every other link a template writes is
	 * {base} and a path, because every other link is a page of this one.
	 */
	static function link($href) {
		if (self::kind($href) === self::MODULE) {
			return $href;
		}

		return Route::href($href);
	}

	/**
	 * active reports whether a row is the page being read: the dashboard on
	 * its own path alone, every other row on anything under it, so
	 * /settings/keys lights Settings and / does not light everything.
	 */
	static function active($href, $path) {
		if ($href === "/") {
			return $path === "/";
		}

		return $path === $href || strpos($path, $href . "/") === 0;
	}

	/**
	 * sees reports whether a user may see a row. An empty rule is a row
	 * everyone signed in sees, and SETTINGS is the one value that is not a
	 * permission path.
	 */
	static function sees($rule, $access, $settings) {
		if ($rule === "") {
			return true;
		}

		if ($rule === self::SETTINGS) {
			return $settings;
		}

		return $access !== null && $access->allowed($rule);
	}

	/** settings reports whether any step of the settings opens for a user. */
	static function settings($access) {
		if ($access === null) {
			return false;
		}

		foreach (array("workspace", "project", "key", "secret", "integration", "user") as $section) {
			if ($access->allowed($section . "/view")) {
				return true;
			}
		}

		return false;
	}

	/**
	 * icon keeps a name to the characters a sprite symbol is named with, so a
	 * value from the database cannot reach the markup as anything else. A
	 * name with nothing left falls back to the dot every list already uses.
	 */
	static function icon($icon) {
		$kept = preg_replace('/[^a-z0-9\-]/', "", strtolower(trim($icon)));

		return $kept === "" ? "box" : $kept;
	}
}
