<?php

// @startup
//
// Artifacts. The artifact tree: every file the runs left, its versions and what they take.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/artifacts");
$module->setTitle("Artifacts");
$module->setDescription("The artifact tree: every file the runs left, its versions and what they take");
$module->setIcon("box");

$module->setTables(array("artifact"));

Task::register($module);
