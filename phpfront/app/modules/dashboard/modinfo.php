<?php

// @startup
//
// Dashboard. The page the instance opens on: what runs now and what ran last.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/dashboard");
$module->setTitle("Dashboard");
$module->setDescription("The page the instance opens on: what runs now and what ran last");
$module->setIcon("grid");

$module->setTables(array());

Task::register($module);
