<?php

function task_modules($root = null) {
	$root = $root === null ? dirname(__DIR__) . "/modules" : $root;
	$out = array();

	foreach (glob($root . "/*/modinfo.php") as $file) {
		$module = require $file;

		if (!is_object($module)) {
			continue;
		}

		$out[$module->name()] = $module;
	}

	ksort($out);

	return array_values($out);
}
