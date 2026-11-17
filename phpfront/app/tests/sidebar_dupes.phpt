name: sidebar dupes
description: Task\Sidebar lists a link once when two modules name it, and a view row the table holds is not a module.
---
<?php

// The table holds view rows from before the menu was the module list, and two
// modules may name one link. Neither is a second row: the menu and the
// settings table draw the module list, once per link.

$db = Task\DB::writer();

$db->replace("sidebar", array("id" => "fx-sd-1", "kind" => "view", "href" => "/fx-sd/view", "title" => "Same", "icon" => "box", "position" => 1, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sd-2", "kind" => "view", "href" => "/fx-sd/only", "title" => "Only a view", "icon" => "box", "position" => 2, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sd-3", "kind" => "module", "href" => "/fx-sd/other", "title" => "Other", "icon" => "flow", "position" => 3, "hidden" => 1, "rule" => ""));

function mod($name, $title, $icon, $href) {
	$module = new Task\Module;
	$module->setName($name);
	$module->setTitle($title);
	$module->setIcon($icon);
	$module->setSideBarHref($href);
	$module->setSideBarRule("");
	$module->enableSideBar();

	return $module;
}

$modules = array(
	mod("taskui/same", "Same", "note", "/fx-sd/view"),
	mod("taskui/same2", "Same", "note", "/fx-sd/view"),
	mod("taskui/other", "Other", "flow", "/fx-sd/other"),
);

Task\Host::$base = "/development";

$runner = new Task\Model\User(array("id" => "fx-sd-u1", "name" => "fx-sd-ann", "role" => "runner"));
$policy = new Task\Policy();
$access = new Task\Access($runner, $policy);
$sidebar = new Task\Sidebar(null, $modules);

foreach ($sidebar->items($access, "/fx-sd/view") as $item) {
	if (strpos($item["href"], "/development/fx-sd") === 0) {
		echo $item["title"], "=", $item["href"], "\n";
	}
}

foreach ($sidebar->rows($sidebar->all()) as $row) {
	echo $row["title"], "=", $row["hidden"] ? "hidden" : "shown", "\n";
}
---
Same=/development/fx-sd/view
Same=shown
Other=hidden
