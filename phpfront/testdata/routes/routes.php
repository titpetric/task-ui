<?php

// @startup

// The paths a tree binds rather than annotates. The service router takes them
// while this job runs and the Bindings module mounts them beside the
// endpoints mod_*.php declare.

Runtime\Router::Bind(function (HTTP\Mux $router) {
	$router->handle("GET /bound", function ($w, $r) {
		$w->write("bound " . $r->method);
	});

	$router->handle("GET /bound/{id}", function ($w, $r) {
		$w->write("id=" . $r->path_value("id"));
	});
});
