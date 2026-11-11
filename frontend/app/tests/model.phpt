name: model
description: A generated model reads a row, answers through its getters and gives the row back.
---
<?php

$ws = new Task\Model\Workspace(array(
	"id" => "01HZ",
	"folder" => "internal",
	"title" => "Internal",
	"shallow" => 1,
	"sidebar_limit" => "5",
));

echo $ws->getId(), "\n";
echo $ws->getFolder(), "\n";
echo $ws->getTitle(), "\n";
echo var_export($ws->getShallow(), true), "\n";
echo var_export($ws->getSidebarLimit(), true), "\n";
echo Task\Model\Workspace::TABLE, "\n";
echo implode(",", Task\Model\Workspace::$primary), "\n";
echo count(Task\Model\Workspace::$fields), "\n";

// A column the row left out reads as the type's zero, not as null.
echo var_export($ws->getDescription(), true), "\n";
echo var_export($ws->getSort(), true), "\n";

// values is the row as the database takes it, every column keyed.
$values = $ws->values();

echo count($values), " ", $values["folder"], " ", var_export($values["shallow"], true), "\n";

// A setter answers the model, so writes chain.
echo $ws->setTitle("Renamed")->getTitle(), "\n";

$rows = Task\Model\Workspace::rows(array(
	array("id" => "01", "folder" => "a"),
	array("id" => "02", "folder" => "b"),
));

echo count($rows), " ", $rows[0]->getFolder(), $rows[1]->getFolder(), "\n";
echo count(Task\Model\Workspace::rows(false)), "\n";
---
01HZ
internal
Internal
true
5
workspace
id
13
''
0
13 internal true
Renamed
2 ab
0
