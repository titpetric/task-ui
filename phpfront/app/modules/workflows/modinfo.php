<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/workflows");
$module->setTitle("Workflows");
$module->setDescription("A pipeline of steps, its runs and the answers each step took");
$module->setIcon("workflow");
$module->enableSideBar();
$module->setSideBarHref("/workflows");
$module->setSideBarRule("");
$module->setTables(array(
	"workflow",
	"workflow_run",
	"workflow_run_step",
	"workflow_run_step_check",
	"workflow_step",
	"workflow_step_check",
));

return $module;
