<?php

// @route GET /probe

// The shape every module follows: the endpoint file carries its routes,
// declares its class and runs its dispatch. A pooled runtime drops what an
// endpoint declared between requests, so the declaration is safe to repeat.

class mod_probe {
	static $served = 0;

	function name() {
		return "probe";
	}

	function serve() {
		mod_probe::$served++;
		echo "probe served ", mod_probe::$served, " base ", Task\Host::base();
	}
}

$mod = new mod_probe;

$mod->serve();
