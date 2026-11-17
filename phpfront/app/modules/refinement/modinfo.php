<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/refinement");
$module->setTitle("Refinement");
$module->setDescription("Refine a design through a thread of comments, each one a Claude run");
$module->setIcon("flow");
$module->enableSideBar();
$module->setSideBarHref("/refinement");
$module->setSideBarRule("module");
$module->setTables(array("refinement_artifact", "refinement_comment", "refinement_correction", "refinement_thread"));

return $module;
