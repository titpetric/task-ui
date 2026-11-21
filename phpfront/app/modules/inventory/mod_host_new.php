<?php

// @route GET /inventory/hosts/new

require "ext/autoload.php";

// mod_host_new is where a host is entered: the form the inventory's head
// action opens, which posts to mod_host_add.
//
// The inventory is the view and this is the edit: a list does not carry the
// form that changes it.
class mod_host_new implements Module {
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
		return ["host"];
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("inventory/edit")) {
			$view->refused();

			return;
		}

		$chrome->heading("Inventory", [
			["text" => "Inventory", "href" => Task\Route::href("/inventory")],
			["text" => "Add a host", "href" => ""],
		]);
		$head = new Task\Head("Add a host", "task-ui signs in over ssh before it keeps the host, which takes a moment.");
		$view->page(["head" => $head->data(), "action" => Task\Route::href("/inventory/hosts")], "host_new.tpl");
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_host_new;
	$mod->serve();
}
