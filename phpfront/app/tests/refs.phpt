name: refs
description: Task\Project::refs is what a move of a project leaves where it was - the schedules, the variables, the env files and the workflow steps that name the folder, each module answering for its own table.
---
<?php

require "ext/autoload.php";

// One connection serves the suite and fixtures run together, so this fixture
// owns the folder fx-refs and the rows that name it, and reads no others.

$db = Task\DB::writer();

$db->query("DELETE FROM schedule WHERE id LIKE 'fx-refs-%'");
$db->query("DELETE FROM secret WHERE id LIKE 'fx-refs-%'");
$db->query("DELETE FROM env_file WHERE id LIKE 'fx-refs-%'");
$db->query("DELETE FROM workflow WHERE id LIKE 'fx-refs-%'");
$db->query("DELETE FROM workflow_step WHERE id LIKE 'fx-refs-%'");

$now = Task\Clock::stamp();

// A row of each kind on the project itself, one on a folder inside it, and
// one on a folder whose name starts the same way, which no move touches.
$db->replace("schedule", array("id" => "fx-refs-s1", "folder" => "fx-refs", "job" => "build", "expression" => "@daily", "created_at" => $now, "updated_at" => $now));
$db->replace("schedule", array("id" => "fx-refs-s2", "folder" => "fx-refs/api", "job" => "test", "expression" => "@daily", "created_at" => $now, "updated_at" => $now));
$db->replace("schedule", array("id" => "fx-refs-s3", "folder" => "fx-refs-other", "job" => "build", "expression" => "@daily", "created_at" => $now, "updated_at" => $now));

$db->replace("secret", array("id" => "fx-refs-v1", "scope" => "fx-refs", "name" => "TOKEN", "created_at" => $now, "updated_at" => $now));
$db->replace("env_file", array("id" => "fx-refs-e1", "folder" => "fx-refs/api", "path" => ".env", "created_at" => $now, "updated_at" => $now));

$db->replace("workflow", array("id" => "fx-refs-w1", "name" => "Nightly", "created_at" => $now, "updated_at" => $now));
$db->replace("workflow_step", array("id" => "fx-refs-p1", "workflow_id" => "fx-refs-w1", "position" => 0, "folder" => "fx-refs", "job" => "deploy"));

$projects = new Task\Project($db);

// The kinds read in order, the folder inside the project kept, the folder
// that merely starts the same left out.
foreach ($projects->refs("fx-refs") as $line) {
	echo $line, "\n";
}

echo "[", implode(",", $projects->refs("fx-refs-nothing")), "]\n";

// A folder inside the project answers its own rows and none of the parent's.
foreach ($projects->refs("fx-refs/api") as $line) {
	echo $line, "\n";
}
---
env file fx-refs/api · .env
schedule fx-refs · build
schedule fx-refs/api · test
secret fx-refs · TOKEN
workflow step fx-refs · Nightly: deploy
[]
env file fx-refs/api · .env
schedule fx-refs/api · test
