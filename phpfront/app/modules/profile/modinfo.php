<?php

// @startup
//
// Profile. A user's own page: their runs, their keys and what they may do.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/profile");
$module->setTitle("Profile");
$module->setDescription("A user's own page: their runs, their keys and what they may do");
$module->setIcon("users");

$module->setTables(array());

Task::register($module);
