<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/folders");
$module->setTitle("Folders");
$module->setDescription("Everything under /f: a workspace, a project, a folder and its jobs");
$module->setIcon("folder");
$module->setTables(array("pin"));

return $module;
