<?php

// @startup
//
// Folders. Everything under /f: a workspace, a project, a folder and its jobs.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/folders");
$module->setTitle("Folders");
$module->setDescription("Everything under /f: a workspace, a project, a folder and its jobs");
$module->setIcon("folder");

$module->setTables(array("pin"));

Task::register($module);
