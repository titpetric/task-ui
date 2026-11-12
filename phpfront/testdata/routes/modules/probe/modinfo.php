<?php

// @startup

// The registration every module of the tree makes, run as a startup job.

$module = new Task\Module;

$module->setName("taskui/probe");
$module->setTitle("The probe");
$module->setIcon("list");
$module->setTables(array());

Task::register($module);
