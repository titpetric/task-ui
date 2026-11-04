name: folders
description: Task\Folders asks the host for the directory of every row of a tree in one crossing, keeps the answer of each, and reads a miss as the empty string it answers.
---
<?php

require "ext/autoload.php";

// The tree asks about every row it draws, which is every pin and every project
// of every workspace. The answers are the host's, so the fixture states them,
// and the number of crossings is what the batch is for.
Task\Host::$folderDirs = array(
	"infra" => "/src/infra",
	"malta/deploy" => "/src/malta/src/deploy",
	"malta/deploy/api" => "/src/malta/src/deploy/api",
	"malta" => "/src/malta",
);
Task\Host::$folderDirCalls = 0;

$folders = new Task\Folders;
$folders->workspaces = array("malta");

// places is what the rows are asked about: each folder, and the project at the
// top of the workspace it sits in, which the sidebar limit orders by its last
// commit. A workspace's own folder is in no project and names none.
$side = array(
	array("id" => "infra"),
	array("id" => "malta/deploy/api"),
	array("id" => "malta/deploy"),
	array("id" => "malta"),
	array("id" => "nowhere"),
);

echo implode(",", $folders->places($side)), "\n";

// One crossing for the lot of them, and the answers read back by ID.
$folders->warm($folders->places($side));

echo Task\Host::$folderDirCalls, "\n";
echo $folders->dirOf("infra"), " ", $folders->dirOf("malta/deploy/api"), " ", $folders->dirOf("malta"), "\n";

// An ID the host places nowhere is absent from its answer, which is the empty
// string the single reader answers, and it is kept so that a tree does not ask
// after it a second time.
echo "[", $folders->dirOf("nowhere"), "] ", Task\Host::$folderDirCalls, "\n";

// A second batch asks only about what is not held yet, and nothing at all
// where the tree holds every ID already: the tree is decided in two steps and
// the pins of the second may name a folder no default listed.
$folders->warm($folders->places($side));

echo Task\Host::$folderDirCalls, "\n";

$folders->warm(array("malta/deploy", "malta/other"));

echo Task\Host::$folderDirCalls, " [", $folders->dirOf("malta/other"), "]\n";

// An empty ID is asked of nobody, and neither is an empty list.
$bare = new Task\Folders;
$bare->workspaces = array("malta");
$bare->warm(array("", ""));
$bare->warm(array());

echo Task\Host::$folderDirCalls, "\n";

// The single reader is still there for the pages that ask about one folder,
// and it keeps its answer too.
echo $bare->dirOf("infra"), " ", Task\Host::$folderDirCalls, " ", $bare->dirOf("infra"), " ", Task\Host::$folderDirCalls, "\n";
---
infra,malta/deploy/api,malta/deploy,malta,nowhere
1
/src/infra /src/malta/src/deploy/api /src/malta
[] 1
1
2 []
2
/src/infra 3 /src/infra 3
