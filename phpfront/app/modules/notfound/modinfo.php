<?php

// @startup
//
// Not found. The catch-all: what answers a path no other module routes.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/notfound");
$module->setTitle("Not found");
$module->setDescription("The catch-all: what answers a path no other module routes");
$module->setIcon("file");

$module->setTables(array());

Task::register($module);
