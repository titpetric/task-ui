<?php

// @startup

$module = new Task\Module;
$module->setName("taskui/settings");
$module->setTitle("Settings");
$module->setDescription("What the server runs with, section by section");
$module->setIcon("settings");
$module->enableSideBar();
$module->setSideBarHref("/settings");
$module->setSideBarRule("settings");
$module->setTables([
	"env_file",
	"permission_policy",
	"project",
	"secret",
	"secret_read",
	"setting",
	"sidebar",
	"tool",
	"workspace",
]);

return $module;
