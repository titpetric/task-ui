<?php

// @startup
//
// Workflows. A pipeline of steps, its runs and the answers each step took.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/workflows");
$module->setTitle("Workflows");
$module->setDescription("A pipeline of steps, its runs and the answers each step took");
$module->setIcon("workflow");

$module->setTables(array(
	"workflow",
	"workflow_run",
	"workflow_run_step",
	"workflow_run_step_check",
	"workflow_step",
	"workflow_step_check",
));

Task::register($module);
