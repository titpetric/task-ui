name: prompt
description: Task\Prompt is where the header's prompt asks on a page - the place of a folder's page, the run of a run's page, and the root anywhere else.
---
<?php

require "ext/autoload.php";

// Nothing here reads a table: a page says where it is and this answers the
// rest, over the folders the reader sees and the named workspaces.

$folders = array(
	array("id" => "fx-prompt-malta"),
	array("id" => "fx-prompt-malta/vhosts"),
);
$workspaces = array("fx-prompt-malta");

// said is a prompt as the header carries it: the place, the run, the line the
// panel writes and the page it opens.
function said($prompt) {
	return implode(" ", array(
		"[" . $prompt["folder"] . "]",
		"[" . $prompt["dir"] . "]",
		"[" . $prompt["run"] . "]",
		"[" . $prompt["label"] . "]",
		$prompt["href"],
	));
}

// The page of a place is the deepest folder the reader sees that holds it,
// with the rest of the place after it. The root is its own page, and a place
// no folder of theirs holds has none.
echo Task\Route::heldHref("internal", $folders, $workspaces), "\n";
echo Task\Route::heldHref("fx-prompt-malta", $folders, $workspaces), "\n";
echo Task\Route::heldHref("fx-prompt-malta/vhosts/www", $folders, $workspaces), "\n";
echo Task\Route::heldHref("fx-prompt-malta/docs", $folders, $workspaces), "\n";
echo "[", Task\Route::heldHref("fx-prompt-private/x", $folders, $workspaces), "]\n";

// A page of no folder asks in the root, and so does the root itself, with or
// without a workspace of its own.
echo said(Task\Prompt::root()), "\n";
echo said(Task\Prompt::at("", "", $folders, $workspaces)), "\n";
echo said(Task\Prompt::at(Task\Route::INTERNAL, "", $folders, $workspaces)), "\n";

// A folder's page asks in the folder, a directory's and a job's of one in the
// directory inside it, and a workspace's settings in its folder.
echo said(Task\Prompt::at("fx-prompt-malta", "", $folders, $workspaces)), "\n";
echo said(Task\Prompt::at("fx-prompt-malta", "vhosts", $folders, $workspaces)), "\n";

// A place the reader's folders do not hold keeps its own line and links to the
// root, because the thread is still somewhere even with no page to open.
echo said(Task\Prompt::at("fx-prompt-gone", "deploy", $folders, $workspaces)), "\n";

// A run's page asks in the place the run ran in and about the run itself,
// which the line names by the start of its ID. An ID of four characters or
// fewer is written whole.
$run = new Task\Model\Run(array("id" => "06gp9t19", "folder" => "fx-prompt-malta", "dir" => "vhosts", "job" => "build"));

echo said(Task\Prompt::ofRun($run, $folders, $workspaces)), "\n";
echo Task\Prompt::short("06gp9t19"), " ", Task\Prompt::short("06gp"), " [", Task\Prompt::short(""), "]\n";

// A run of Claude's own is the thread and not something to ask about: a
// prompt's run asks in its place, and one of the root asks in the root.
$prompt = new Task\Model\Run(array("id" => "06gp9t19", "folder" => "fx-prompt-malta", "job" => Task\Run::PROMPT_JOB));
$claude = new Task\Model\Run(array("id" => "06gp9t19", "folder" => Task\Route::INTERNAL, "job" => "claude"));

echo said(Task\Prompt::ofRun($prompt, $folders, $workspaces)), "\n";
echo said(Task\Prompt::ofRun($claude, $folders, $workspaces)), "\n";
---
/f/internal
/f/fx-prompt-malta
/f/fx-prompt-malta/vhosts/www
/f/fx-prompt-malta/docs
[]
[internal] [] [] [internal] /f/internal
[internal] [] [] [internal] /f/internal
[internal] [] [] [internal] /f/internal
[fx-prompt-malta] [] [] [fx-prompt-malta] /f/fx-prompt-malta
[fx-prompt-malta] [vhosts] [] [fx-prompt-malta/vhosts] /f/fx-prompt-malta/vhosts
[fx-prompt-gone] [deploy] [] [fx-prompt-gone/deploy] /f/internal
[fx-prompt-malta] [vhosts] [06gp9t19] [run 06gp…] /runs/06gp9t19
06gp… 06gp []
[fx-prompt-malta] [] [] [fx-prompt-malta] /f/fx-prompt-malta
[internal] [] [] [internal] /f/internal
