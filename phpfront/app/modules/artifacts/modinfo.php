<?php

// @startup

$module = new Task\Module;
$module->setName("taskui/artifacts");
$module->setTitle("Artifacts");
$module->setDescription("The artifact tree: every file the runs left, its versions and what they take");
$module->setIcon("box");
$module->enableSideBar();
$module->setSideBarHref("/artifacts");
$module->setSideBarRule("");
$module->setTables(["artifact"]);

return $module;
