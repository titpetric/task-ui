name: artifact
description: Task\Artifacts\Artifacts reads the artifact table - the files of one run newest first, and the versions of a path in the tree.
---
<?php

require "ext/autoload.php";

// One connection serves the suite and fixtures run together, so every query
// here names this fixture's own run, folder or path. A fixture that read the
// whole table would read another fixture's rows.

$run = "fx-art-run";
$folder = "fx-art";

$db = Task\DB::writer();

// The files of one run, a minute apart, and two versions of a path of the
// artifact tree, which are the rows that hold a commit.
$rows = array(
	array("id" => "fx-art-1", "run" => $run, "name" => "build.log", "path" => "", "size" => 512, "at" => "10:01"),
	array("id" => "fx-art-2", "run" => $run, "name" => "coverage.html", "path" => "", "size" => 2048, "at" => "10:02"),
	array("id" => "fx-art-3", "run" => "fx-art-other-run", "name" => "elsewhere.log", "path" => "", "size" => 10, "at" => "10:03"),
	array("id" => "fx-art-4", "run" => "", "name" => "notes.md", "path" => "fx-art/notes.md", "size" => 30, "at" => "10:04", "commit" => "fx-art-c1"),
	array("id" => "fx-art-5", "run" => "", "name" => "notes.md", "path" => "fx-art/notes.md", "size" => 40, "at" => "10:05", "commit" => "fx-art-c2"),
);

foreach ($rows as $row) {
	$db->replace("artifact", array(
		"id" => $row["id"],
		"run_id" => $row["run"],
		"folder" => $folder,
		"job" => $row["run"] === "" ? "" : "build",
		"name" => $row["name"],
		"path" => $row["path"],
		"byte_size" => $row["size"],
		"content_type" => "text/plain",
		"commit_id" => isset($row["commit"]) ? $row["commit"] : "",
		"user_name" => "admin",
		"created_at" => "2026-11-02 " . $row["at"] . ":00.000 +0000 UTC",
	));
}

// A page reads through the reader, which refuses a write.
$artifacts = new Task\Artifacts\Artifacts;

// named is what a query answered, in the order it answered.
function named($found) {
	$out = array();

	foreach ($found as $artifact) {
		$out[] = $artifact->getId();
	}

	return implode(",", $out);
}

// The files of one run, newest first, which is what a run page lists.
echo named($artifacts->find(array("run" => $run))), "\n";
echo "[", named($artifacts->find(array("run" => "fx-art-no-run"))), "]\n";
echo named($artifacts->find(array("run" => $run, "limit" => 1))), "\n";

// The rest of the conditions: a folder, a job, and a path of the tree, whose
// rows are the versions of that file.
echo named($artifacts->find(array("folder" => $folder))), "\n";
echo named($artifacts->find(array("folder" => $folder, "job" => "build"))), "\n";
echo named($artifacts->find(array("path" => "fx-art/notes.md"))), "\n";

// One artifact by ID, and null for an ID that names none.
$one = $artifacts->get("fx-art-2");

echo $one->getName(), " ", $one->getByteSize(), " ", $one->getContentType(), "\n";
echo var_export($artifacts->get(""), true), " ", var_export($artifacts->get("fx-art-nothing"), true), "\n";
---
fx-art-2,fx-art-1
[]
fx-art-2
fx-art-5,fx-art-4,fx-art-3,fx-art-2,fx-art-1
fx-art-3,fx-art-2,fx-art-1
fx-art-5,fx-art-4
coverage.html 2048 text/plain
NULL NULL
