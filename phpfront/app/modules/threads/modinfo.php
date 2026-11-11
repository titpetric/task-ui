<?php

// @startup
//
// Threads. The notes a reader writes, each thread a conversation with the agent.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/threads");
$module->setTitle("Threads");
$module->setDescription("The notes a reader writes, each thread a conversation with the agent");
$module->setIcon("note");

$module->setTables(array("note"));

Task::register($module);
