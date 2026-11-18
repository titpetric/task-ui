<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/orb");
$module->setTitle("Meetings");
$module->setDescription("Meet in a room: live voice, a transcript attributed by speaker and a summary");
$module->setIcon("flow");
$module->enableSideBar();
$module->setSideBarHref("/orb");
$module->setSideBarRule("meeting/view");
$module->setTables(["meeting", "meeting_line", "meeting_participant"]);

return $module;
