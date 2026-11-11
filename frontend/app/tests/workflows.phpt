name: workflows
description: The workflows page draws a workflow's description as markdown on the line under its name, and nothing at all for a workflow without one.
---
<?php

require "modules/workflows/mod_workflows.php";

$mod = new mod_workflows;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo var_export($mod instanceof Module, true), "\n";

// The row is built out of what Task\Workflow::view answers, so the fixture
// writes that shape rather than a workflow table: what is held here is the
// description, and a row reads the rest of it as it is given.
function flow($id, $description) {
	return array(
		"id" => $id,
		"name" => "fx-workflows-" . $id,
		"description" => $description,
		"userName" => "fx-workflows-ann",
		"runAs" => "",
		"running" => "",
		"last" => null,
		"canRun" => true,
		"canChange" => false,
		"prompted" => false,
		"runDenied" => "",
	);
}

$chrome = new Task\Chrome;

// The description is markdown, so a link to the runbook is a link and the name
// of a thing in code is code. The renderer is the host's; what the page does is
// draw the answer as markup, which is why the markup writes it raw.
$row = $mod->row($chrome, flow("w1", "Deploys **prod**, see [runbook](docs/deploy.md)"), array());

echo "[", $row["description"], "]\n";

// Several blocks are a document and not a line, and are drawn as what they are.
$long = $mod->row($chrome, flow("w2", "what it does\n\nand when"), array());

echo "[", str_replace("\n", "|", $long["description"]), "]\n";

// A workflow with no description draws no line: the markup asks whether there
// is one, and the empty string is no.
$bare = $mod->row($chrome, flow("w3", ""), array());

echo "[", $bare["description"], "]\n";
---
workflows Workflows workflow
true
[Deploys **prod**, see [runbook](docs/deploy.md)]
[<p>what it does</p>||<p>and when</p>]
[]
