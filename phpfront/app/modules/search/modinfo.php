<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/search");
$module->setTitle("Search");
$module->setDescription("Ask a question over the documentation and job targets of every project");
$module->setIcon("magnifier");
$module->enableSideBar();
$module->setSideBarHref("/search");
$module->setSideBarRule("prompt/edit");
$module->setTables(array("search_document", "search_section", "search_term"));

return $module;
