name: menu
description: Menu groups its items, puts the danger group last and escapes every attribute.
---
<?php

$menu = new Task\Menu("Actions on deploy");

echo var_export($menu->empty(), true), "\n";

$menu->detail("Last run", "3m ago", true, "2026-11-02");
$menu->item("play", "Run", array("data-run" => "deploy"));
$menu->sep();
$menu->item("edit", "Edit", array("href" => "/f/a/deploy/edit"));
$menu->danger("trash", "Delete", array("data-delete" => "/api/x", "data-confirm" => 'Delete "deploy"?'));

$data = $menu->data();

echo var_export($data["empty"], true), " ", $data["label"], "\n";
echo count($data["details"]), " ", $data["details"][0]["label"], " ", var_export($data["details"][0]["mono"], true), "\n";

// Three groups: the first item, the second, then the danger group last.
echo count($data["groups"]), "\n";

foreach ($data["groups"] as $group) {
	foreach ($group as $item) {
		echo $item["icon"], " ", $item["label"], " [", $item["attrs"], "] ", $item["href"], " ", var_export($item["danger"], true), "\n";
	}
}

// A menu holding only a separator shows nothing.
$bare = new Task\Menu("Nothing");

$bare->sep();

echo var_export($bare->empty(), true), " ", count($bare->data()["groups"]), "\n";

// when and whenDanger add nothing where the condition does not hold.
$some = new Task\Menu("Some");

$some->when(false, "edit", "Edit");
$some->whenDanger(false, "trash", "Delete");

echo var_export($some->empty(), true), "\n";

// note and disabled set the fields rather than reaching the markup.
$off = new Task\Menu("Off");

$off->item("play", "Run", array("disabled" => "no permission"));

$item = $off->data()["groups"][0][0];

echo var_export($item["disabled"], true), " ", $item["note"], " [", $item["attrs"], "]\n";

// An attribute that is not data-*, aria-* or title is refused, so a
// misspelling fails rather than drawing nothing.
try {
	(new Task\Menu("Bad"))->item("play", "Run", array("onclick" => "x()"));

	echo "not refused\n";
} catch (Exception $e) {
	echo "refused\n";
}
---
true
false Actions on deploy
1 Last run true
3
play Run [data-run="deploy"]  false
edit Edit [] /f/a/deploy/edit false
trash Delete [data-delete="/api/x" data-confirm="Delete &quot;deploy&quot;?"]  true
true 0
true
true no permission []
refused
