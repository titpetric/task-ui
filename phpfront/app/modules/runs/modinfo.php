<?php

// @startup
//
// Runs. Every run of a job: its terminal, its output and the report it leaves.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/runs");
$module->setTitle("Runs");
$module->setDescription("Every run of a job: its terminal, its output and the report it leaves");
$module->setIcon("list");

$module->setTables(array("run", "run_queued", "run_summary"));

Task::register($module);
