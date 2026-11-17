name: sidebar dupes
description: Task\Sidebar leaves out a module row whose title a view row the reader sees already carries.
---
<?php

$db = Task\DB::writer();

// The view and the module are the same page under two links, which is what a
// module row written beside a view row is. The module of a title no visible
// view carries still shows, and so does one whose view is hidden.
$db->replace("sidebar", array("id" => "fx-sd-1", "kind" => "view", "href" => "/fx-sd/view", "title" => "Same", "icon" => "box", "position" => 1, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sd-2", "kind" => "module", "href" => "/m/titpetric/fx-sd", "title" => "Same", "icon" => "flow", "position" => 2, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sd-3", "kind" => "module", "href" => "/m/titpetric/fx-sd2", "title" => "Other", "icon" => "flow", "position" => 3, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sd-4", "kind" => "view", "href" => "/fx-sd/hidden", "title" => "Hidden", "icon" => "box", "position" => 4, "hidden" => 1, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sd-5", "kind" => "module", "href" => "/m/titpetric/fx-sd3", "title" => "Hidden", "icon" => "flow", "position" => 5, "rule" => ""));

Task\Host::$base = "/development";

$runner = new Task\Model\User(array("id" => "fx-sd-u1", "name" => "fx-sd-ann", "role" => "runner"));
$policy = new Task\Policy();
$access = new Task\Access($runner, $policy);
$sidebar = new Task\Sidebar;

foreach ($sidebar->items($access, "/fx-sd/view") as $item) {
	if (strpos($item["href"], "/development/fx-sd") === 0 || strpos($item["href"], "/m/titpetric/fx-sd") === 0) {
		echo $item["title"], "=", $item["href"], "\n";
	}
}
---
Same=/development/fx-sd/view
Other=/m/titpetric/fx-sd2
Hidden=/m/titpetric/fx-sd3
