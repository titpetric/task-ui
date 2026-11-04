name: lookup
description: Task\Schedules\Lookup reads the table the schedules page does not own - the workflows with the user they run as and the folders their steps name.
---
<?php

require "ext/autoload.php";

// Every statement here reads a whole table and every answer is keyed by an ID,
// so each is read by one: the suite shares a connection and another fixture's
// workflow is as real as these.

$db = Task\DB::writer();

// The row is written with replace, because a name is unique and another
// fixture may have written the one it is about.
$db->replace("user", array("id" => "fx-lk-u1", "name" => "fx-lookup-ann", "role" => "runner"));

// Three workflows: one whose runs run as a user, one that runs as whoever
// starts it, and one whose user is gone, which reads as the second.
$db->replace("workflow", array("id" => "fx-lk-w1", "name" => "fx Deploy", "run_as_user_id" => "fx-lk-u1"));
$db->replace("workflow", array("id" => "fx-lk-w2", "name" => "fx Backup", "run_as_user_id" => ""));
$db->replace("workflow", array("id" => "fx-lk-w3", "name" => "fx Orphan", "run_as_user_id" => "fx-lk-gone"));

// Deploy runs a job and then the whole of Backup, so the folders it reaches
// are its own and Backup's.
$db->replace("workflow_step", array("id" => "fx-lk-s1", "workflow_id" => "fx-lk-w1", "position" => 0, "folder" => "fx-lookup-infra", "job" => "plan", "kind" => "job"));
$db->replace("workflow_step", array("id" => "fx-lk-s2", "workflow_id" => "fx-lk-w1", "position" => 1, "folder" => "", "job" => "", "workflow_ref" => "fx-lk-w2", "kind" => "workflow"));
$db->replace("workflow_step", array("id" => "fx-lk-s3", "workflow_id" => "fx-lk-w2", "position" => 0, "folder" => "fx-lookup-backups", "job" => "dump", "kind" => "job"));

$lookup = new Task\Schedules\Lookup;

// A workflow by ID: its name, who its runs run as, and the folders its steps
// reach. A run as user that no row holds reads as none.
$workflows = $lookup->workflows();

foreach (array("fx-lk-w1", "fx-lk-w2", "fx-lk-w3") as $id) {
	$found = $workflows[$id];

	echo $found["name"], " [", $found["runAs"], "] [", implode(",", $found["folders"]), "]\n";
}

echo var_export(isset($workflows["fx-lk-gone"]), true), "\n";

// The steps of a workflow in position order, by ID.
$steps = $lookup->steps();

echo count($steps["fx-lk-w1"]), " ", Task\Schedules\Lookup::kind($steps["fx-lk-w1"][0]), " ", Task\Schedules\Lookup::kind($steps["fx-lk-w1"][1]), "\n";
echo $steps["fx-lk-w1"][1]->getWorkflowRef(), "\n";
---
fx Deploy [fx-lookup-ann] [fx-lookup-infra,fx-lookup-backups]
fx Backup [] [fx-lookup-backups]
fx Orphan [] []
false
2 job workflow
fx-lk-w2
