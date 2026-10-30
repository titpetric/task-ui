<?php

// @startup
//
// The refinement module. A thread holds the comments a user writes; each one
// starts a Claude run over the thread's own folder of the artifact tree, and
// the run rewrites what is there and the README.md that reports on it.
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
