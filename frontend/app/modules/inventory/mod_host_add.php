<?php

// @route POST /inventory/hosts

require "ext/autoload.php";

// mod_host_add enters a machine in the inventory: the address ssh signs in to,
// and a name for the pages.
//
// task-ui signs in before the row is written, and an address no login reaches
// is refused with what it tried. A host of the inventory is a host task-ui can
// read over ssh and run a command on, and a row for a machine nobody reaches
// would be a line that never fills in.
//
// Every path here redirects to the page, so a reload enters nothing twice.
class mod_host_add implements Module {
	/** PAGE is the inventory, where the form posted from and where a reader lands. */
	const PAGE = "/inventory";

	/** WANTED is the question for a post with no address on it. */
	const WANTED = "A host goes in by the address ssh signs in to: a name or an IP.";

	/** HELD is what an address the inventory already holds is told. */
	const HELD = "The inventory holds that address already.";

	/** CLOSED is what an address no login reaches is told. */
	const CLOSED = "No ssh login of the instance signs in to that address. task-ui keeps a host it reaches, so check the keys under Settings, Keys and that the host accepts them for debian or root.";

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

		if (!$chrome->allowed("inventory/edit")) {
			$view = new Task\View($this, $chrome);

			$view->refused();

			return;
		}

		$address = trim(Task\Route::input("address"));
		$name = trim(Task\Route::input("name"));

		if ($address === "") {
			Task\Route::fail(self::PAGE, self::WANTED);

			return;
		}

		$store = new Task\Inventory\Hosts($db);

		if (is_array($store->find($address))) {
			Task\Route::fail(self::PAGE, self::HELD);

			return;
		}

		$login = Task\Host::reach($address);

		if ($login === "") {
			Task\Route::fail(self::PAGE, self::CLOSED);

			return;
		}

		$store->add($name, $address, Task\Inventory\Hosts::MANUAL, $login, $chrome->name());
		Task\Route::notice(self::PAGE, "Added " . ($name === "" ? $address : $name) . ", which answers as " . $login . ".");
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_host_add;

	$mod->serve();
}
