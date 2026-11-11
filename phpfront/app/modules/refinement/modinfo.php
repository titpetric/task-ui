<?php

// @startup
//
// Refinement. Refine a design through a thread of comments, each one a Claude run.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/refinement");
$module->setTitle("Refinement");
$module->setDescription("Refine a design through a thread of comments, each one a Claude run");
$module->setIcon("flow");

$module->setTables(array("refinement_artifact", "refinement_comment", "refinement_correction", "refinement_thread"));

Task::register($module);
