<?php

// @startup
//
// Settings. What the server runs with, section by section.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/settings");
$module->setTitle("Settings");
$module->setDescription("What the server runs with, section by section");
$module->setIcon("settings");

$module->setTables(array(
	"env_file",
	"permission_policy",
	"project",
	"secret",
	"secret_read",
	"setting",
	"sidebar",
	"tool",
	"user",
	"workspace",
));

Task::register($module);
