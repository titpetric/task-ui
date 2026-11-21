<?php

// @startup

$module = new Task\Module;
$module->setName("taskui/schedules");
$module->setTitle("Scheduled");
$module->setDescription("The jobs a clock starts, and what each one ran as");
$module->setIcon("clock");
$module->enableSideBar();
$module->setSideBarHref("/schedules");
$module->setSideBarRule("");
$module->setTables(["schedule"]);

return $module;
