name: head
description: The scope every page head hands the component is one shape - the heading and the sentence under it, the badge beside it with its tone, the terminal's title, and the actions on either side.
---
<?php

// A page states what it has and the rest stay the empty values the component
// draws nothing for, so the shape is the same on every page.
$list = new Task\Head("Runs", "Every run, <code>newest</code> first.");
$scope = $list->data();

echo implode(",", array_keys($scope)), "\n";
echo $scope["title"], " [", $scope["badge"], "] [", $scope["tone"], "] [", $scope["note"], "]\n";
echo $scope["lead"], "\n";

// A lead the page wrote is HTML and arrives as it stands. One built out of a
// run's own values goes through said, which escapes it: the sentence holds a
// user name, which is what a reader typed.
$run = new Task\Head("fx / build");

$run->said("Started by a<b just now. It's still running.");
$run->badge("failed 2", "bad");
$run->note("pulling the image");

$held = $run->data();

echo $held["lead"], "\n";
echo $held["badge"], " ", $held["tone"], " [", $held["note"], "]\n";

// A badge with no tone is the plain one a document page draws. The row of
// buttons is the page's own and the menu is the list page's; a head carries
// either, and the component draws what it was given.
$doc = new Task\Head("fx / build");
$buttons = new Task\Buttons;
$menu = new Task\Menu("Runs");

$buttons->primary("", "Run", array("href" => "/runs/fx-1"));
$menu->item("list", "Everything", array("href" => "/runs"));
$doc->badge("report");
$doc->buttons($buttons->data());
$doc->menu($menu->data());

$read = $doc->data();

echo $read["badge"], " [", $read["tone"], "] ", count($read["buttons"]["items"]), " ", count($read["actions"]["groups"]), "\n";
---
title,icon,badge,tone,tags,note,lead,leads,faint,live,run,actions,buttons,show
Runs [] [] []
Every run, <code>newest</code> first.
Started by a&lt;b just now. It&#039;s still running.
failed 2 bad [pulling the image]
report [] 1 1
