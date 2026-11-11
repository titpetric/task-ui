<?php

// @startup
//
// MCP tokens. The tokens an agent signs the MCP API with.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/mcp");
$module->setTitle("MCP tokens");
$module->setDescription("The tokens an agent signs the MCP API with");
$module->setIcon("plug");

$module->setTables(array());

Task::register($module);
