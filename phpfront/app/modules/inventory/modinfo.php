<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/inventory");
$module->setTitle("Inventory");
$module->setDescription("The hosts the instance knows, each row expanding into its own page");
$module->setIcon("server");
$module->enableSideBar();
$module->setSideBarHref("/inventory");
$module->setSideBarRule("");
$module->setTables(["host"]);

return $module;
