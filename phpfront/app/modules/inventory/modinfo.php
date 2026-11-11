<?php

// @startup
//
// Inventory. The hosts the instance knows, each row expanding into its own page.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/inventory");
$module->setTitle("Inventory");
$module->setDescription("The hosts the instance knows, each row expanding into its own page");
$module->setIcon("server");

$module->setTables(array("host"));

Task::register($module);
