<?php

// @startup
//
// Scheduled. The jobs a clock starts, and what each one ran as.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/schedules");
$module->setTitle("Scheduled");
$module->setDescription("The jobs a clock starts, and what each one ran as");
$module->setIcon("clock");

$module->setTables(array("schedule"));

Task::register($module);
