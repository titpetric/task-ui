name: schedule form
description: The schedule form offers a workflow or a job on add and names the target on edit.
---
<?php

require "ext/autoload.php";

// render is the form on its own, over the data the add and edit pages hand it:
// the page head and the shell are the callers'.
function render($data) {
	$paths = ["modules/schedules/templates/", "theme/templates/"];
	$tpl = new MiniTPL\Template($paths);
	$tpl->load("_edit.tpl");
	$tpl->assign("base", "");
	$tpl->assign("chrome", []);
	$tpl->assign("data", $data);

	return $tpl->get();
}

$users = [];

$add = render([
	"edit" => false,
	"target" => "",
	"api" => "/api/schedules",
	"method" => "POST",
	"next" => "/development/schedules?notice=Schedule+added",
	"back" => "/development/schedules",
	"label" => "Add schedule",
	"folders" => [["id" => "fx", "selected" => true]],
	"jobs" => [["id" => "plan", "name" => "Plan", "folder" => "fx", "selected" => true]],
	"hosts" => [],
	"workflows" => [["id" => "w1", "name" => "Nightly"]],
	"workflow" => "",
	"folder" => "fx",
	"job" => "plan",
	"expression" => "",
	"args" => "",
	"triage" => false,
	"enabled" => true,
	"pickUser" => false,
	"runAs" => ["id" => "sch-user", "options" => ["users" => $users, "selected" => "", "short" => false]],
	"workflowName" => "",
	"workflowHref" => "",
	"targetHref" => "",
]);

echo var_export(strpos($add, "data-sched-workflow") !== false, true), "\n";
echo var_export(strpos($add, "data-sched-job") !== false, true), "\n";
echo var_export(strpos($add, 'data-folder="fx"') !== false, true), "\n";
echo var_export(strpos($add, ">Nightly</option>") !== false, true), "\n";
echo var_export(strpos($add, ">Add schedule</button>") !== false, true), "\n";

$edit = render([
	"edit" => true,
	"target" => "workflow",
	"api" => "/api/schedules/s1",
	"method" => "PUT",
	"next" => "/development/schedules?notice=Schedule+saved",
	"back" => "/development/schedules",
	"label" => "Save schedule",
	"folders" => [],
	"jobs" => [],
	"hosts" => [],
	"workflows" => [],
	"workflow" => "w1",
	"folder" => "",
	"job" => "",
	"expression" => "0 3 * * *",
	"args" => "",
	"triage" => false,
	"enabled" => true,
	"pickUser" => false,
	"runAs" => ["id" => "sch-user", "options" => ["users" => $users, "selected" => "fx-sch-ann", "short" => false]],
	"workflowName" => "Nightly",
	"workflowHref" => "/workflows/w1",
	"targetHref" => "",
]);

echo var_export(strpos($edit, "data-sched-workflow") === false, true), "\n";
echo var_export(strpos($edit, "data-sched-job") === false, true), "\n";
echo var_export(strpos($edit, 'name="workflow" value="w1"') !== false, true), "\n";
echo var_export(strpos($edit, 'value="0 3 * * *"') !== false, true), "\n";
echo var_export(strpos($edit, ">Save schedule</button>") !== false, true), "\n";
---
true
true
true
true
true
true
true
true
true
true
