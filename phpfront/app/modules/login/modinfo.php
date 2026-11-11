<?php

// @startup
//
// Sign in. The sign in page; reading the credentials stays in Go.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/login");
$module->setTitle("Sign in");
$module->setDescription("The sign in page; reading the credentials stays in Go");
$module->setIcon("logout");

$module->setTables(array());

Task::register($module);
