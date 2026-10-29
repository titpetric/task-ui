<?php

// @startup
//
// The refinement module. A thread holds the comments a user writes; each one
// starts a Claude run over the artifacts as they stand, and the run rewrites
// them and the README.md that indexes them.
//
// task-ui runs this file once, before it serves a request, and the module
// serves only if Task::register takes what it built.

$module = new Task\Module;

$module->setName("titpetric/refinement");
$module->setTitle("Refinement");
$module->setDescription("Refine a design through a thread of comments");
$module->setIcon("flow");
$module->setVersion("0.1.0");

Task::register($module);
