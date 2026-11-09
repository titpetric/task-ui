<?php

// @route POST /inventory/hosts/{id}/delete

require "ext/autoload.php";

// mod_host_drop takes a machine out of the inventory.
//
// The hub is not told. A host Beszel still reports comes back on the next read
// of the page, because the hub is a source and the rows are the list; removing
// a host for good is removing it in Beszel as well.
class mod_host_drop implements Module {
	/** PAGE is the inventory, where the reader lands. */
	const PAGE = "/inventory";

	function name() {
		return "inventory";
	}

	function title() {
		return "Inventory";
	}

	function icon() {
		return "server";
	}

	function tables() {
		return array("host");
	}

	function serve() {
		$db = Task\DB::writer();
		$chrome = new Task\Chrome($db);

		if (!$chrome->allowed("inventory/delete")) {
			$view = new Task\View($this, $chrome);

			$view->refused();

			return;
		}

		$store = new Task\Inventory\Hosts($db);

		$store->drop(Task\Route::param("id"));
		Task\Route::notice(self::PAGE, "The host is out of the inventory.");
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_host_drop;

	$mod->serve();
}
