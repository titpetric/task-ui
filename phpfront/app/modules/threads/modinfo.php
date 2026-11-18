<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/threads");
$module->setTitle("Threads");
$module->setDescription("The notes a reader writes, each thread a conversation with the agent");
$module->setIcon("note");
$module->enableSideBar();
$module->setSideBarHref("/threads");
$module->setSideBarRule("");
$module->setTables(["note", "note_participant"]);

return $module;
