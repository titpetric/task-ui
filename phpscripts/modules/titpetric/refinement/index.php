<?php

// The catch-all. Every request the public/ tree does not answer arrives
// here, with $_SERVER["REQUEST_URI"] holding the path the browser
// requested, the /m/titpetric/refinement prefix included.
//
// What this echoes becomes the body of a task-ui page: the heading, the
// description and the badges are the ones modinfo.php registered, and
// the layout, the sidebar and the footer are the instance's.
//
// Four routes, read off the path past the prefix:
//
//   ""                          the threads, and the form that starts one
//   "threads"            POST   starts a thread and dispatches its run
//   "threads/<id>"              one thread: its index, artifacts, comments
//   "threads/<id>/comments" POST a correction, and the run that applies it
//
// Nothing here writes SQL. storage/Refinement.php holds every statement
// the module runs, lib/thread.php decides what to ask it for, and this
// routes.

// A require is a path from the module's root, the directory modinfo.php
// sits in, whatever file it is written in: the runtime roots the module
// there and nothing above it resolves.
//
// ext/ is composer's install directory, `composer install` writes it and
// it ships with the module. It is not named vendor: the binary is
// installed from the Go module proxy, whose zip drops every file more
// than one level under a directory named vendor, so a composer tree
// there arrives with its autoloader and nothing for it to load.
require "ext/autoload.php";
require "storage/Refinement.php";
require "lib/thread.php";
require "lib/render.php";

$base = $_SERVER["TASK_UI_MODULE_BASE"];
$user = isset($_SERVER["TASK_UI_USER"]) ? $_SERVER["TASK_UI_USER"] : "";
$store = new Refinement(new Database("task-ui"));

$path = refinement_path($_SERVER["REQUEST_URI"], $base);
$parts = $path === "" ? array() : explode("/", $path);
$posted = $_SERVER["REQUEST_METHOD"] === "POST";

// POST /threads: the first comment of a new thread, and the run that
// answers it. The answer is a redirect, so a reload does not start a
// second run.
if ($posted && count($parts) === 1 && $parts[0] === "threads") {
	$want = isset($_POST["want"]) ? trim($_POST["want"]) : "";
	if ($want === "") {
		refinement_fail($base, "A thread starts from one paragraph of what to design.");
		return;
	}

	$thread = refinement_start($store, $user, $want);

	refinement_dispatch($store, $thread["id"], $thread["comment_id"], $want);
	refinement_go($base . "/threads/" . $thread["id"]);
	return;
}

// POST /threads/<id>/comments: a correction over the artifacts as they
// stand, and the run that applies it.
if ($posted && count($parts) === 3 && $parts[0] === "threads" && $parts[2] === "comments") {
	$body = isset($_POST["comment"]) ? trim($_POST["comment"]) : "";
	$thread = $store->getThread($parts[1]);
	if (!is_array($thread)) {
		refinement_missing();
		return;
	}

	if ($body === "") {
		refinement_fail($base . "/threads/" . $thread["id"], "A correction says one thing to change.");
		return;
	}

	$comment = refinement_correct($store, $thread["id"], $user, $body);

	refinement_dispatch($store, $thread["id"], $comment, $body);
	refinement_go($base . "/threads/" . $thread["id"]);
	return;
}

// GET /threads/<id>: one thread.
if (count($parts) === 2 && $parts[0] === "threads") {
	$thread = $store->getThread($parts[1]);
	if (!is_array($thread)) {
		refinement_missing();
		return;
	}

	$thread = refinement_sync($store, $thread);

	refinement_render_thread($base, $thread, $store->getComments($thread["id"]));
	return;
}

// GET /: every thread, newest first, and the form that starts one.
refinement_render_threads($base, $store->getThreads());
