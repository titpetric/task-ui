<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/runs");
$module->setTitle("Runs");
$module->setDescription("Every run of a job: its terminal, its output and the report it leaves");
$module->setIcon("list");
$module->enableSideBar();
$module->setSideBarHref("/runs");
$module->setSideBarRule("");
$module->setTables(["run", "run_queued", "run_summary"]);

return $module;
