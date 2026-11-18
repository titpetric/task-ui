<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/dashboard");
$module->setTitle("Dashboard");
$module->setDescription("The page the instance opens on: what runs now and what ran last");
$module->setIcon("grid");
$module->enableSideBar();
$module->setSideBarHref("/");
$module->setSideBarRule("");
$module->setTables([]);

return $module;
