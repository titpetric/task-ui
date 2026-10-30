<?php

// The catch-all. Every request the public/ tree does not answer arrives
// here, with $_SERVER["REQUEST_URI"] holding the path the browser
// requested, the /m/titpetric/refinement prefix included.
//
// What this echoes becomes the body of a task-ui page: the heading, the
// description and the badges are the ones modinfo.php registered, and
// the layout, the sidebar and the footer are the instance's.
//
// Six routes, read off the path past the prefix:
//
//   ""                                   the threads, and the form that starts one
//   "threads"                     POST   starts a thread and dispatches its run
//   "threads/<id>"                       one thread: its report and its requests
//   "threads/<id>/comments"       POST   a correction, and the run that applies it
//   "threads/<id>/apply"          POST   folds every rule into one run
//   "threads/<id>/comments/<c>/delete" POST retires a rule, or the whole thread
//
// Nothing here holds logic. RefinementStore has every SQL statement the
// module runs, RefinementManager decides what to ask it for,
// RefinementPage draws, and this routes.

// One require, and the classes of code/ arrive as they are named: ext/
// is composer's install directory, `composer install` writes it and it
// ships with the module. It is not named vendor, because the binary is
// installed from the Go module proxy, whose zip drops every file more
// than one level under a directory named vendor, so a composer tree
// there arrives with its autoloader and nothing for it to load.
//
// A path in a require is read from the module's root, the directory
// modinfo.php sits in, whatever file it is written in: the runtime roots
// the module there and nothing above it resolves.
require "ext/autoload.php";

$base = $_SERVER["TASK_UI_MODULE_BASE"];
$user = isset($_SERVER["TASK_UI_USER"]) ? $_SERVER["TASK_UI_USER"] : "";

$page = new RefinementPage($base);
$store = new RefinementStore(new Database("task-ui"));
$manager = new RefinementManager($store);

$path = $page->path($_SERVER["REQUEST_URI"]);
$parts = $path === "" ? array() : explode("/", $path);
$posted = $_SERVER["REQUEST_METHOD"] === "POST";

// POST /threads: the first comment of a new thread, and the run that
// answers it. The answer is a redirect, so a reload does not start a
// second run.
if ($posted && count($parts) === 1 && $parts[0] === "threads") {
	// A correction that arrives here is a correction that lost its
	// thread: the comment form posts to threads/<id>/comments and only
	// the form that opens a thread posts here. Opening a thread from it
	// would bury the correction in a thread of its own, with none of the
	// work it was correcting, so this says what happened instead.
	if (isset($_POST["comment"])) {
		RefinementPage::fail($base, "That is a correction and it needs its thread. Open the thread and send it from the box at the bottom.");
		return;
	}

	$want = isset($_POST["want"]) ? trim($_POST["want"]) : "";
	if ($want === "") {
		RefinementPage::fail($base, "A thread starts from one paragraph of what to design.");
		return;
	}

	$opened = $manager->start($user, $want);

	$manager->dispatch($opened["id"], $opened["comment_id"], $want);
	RefinementPage::go($base . "/threads/" . $opened["id"]);
	return;
}

// POST /threads/<id>/comments: a correction over the artifacts as they
// stand, and the run that applies it.
if ($posted && count($parts) === 3 && $parts[0] === "threads" && $parts[2] === "comments") {
	$body = isset($_POST["comment"]) ? trim($_POST["comment"]) : "";
	$thread = $store->getThread($parts[1]);
	if (!is_array($thread)) {
		RefinementPage::missing();
		return;
	}

	if ($body === "") {
		RefinementPage::fail($base . "/threads/" . $thread["id"], "A correction says one thing to change.");
		return;
	}

	$comment = $manager->correct($thread["id"], $user, $body);

	$manager->dispatch($thread["id"], $comment, $body);
	RefinementPage::go($base . "/threads/" . $thread["id"]);
	return;
}

// POST /threads/<id>/apply: every rule the thread still holds, as one
// request, over the artifacts that are already there.
if ($posted && count($parts) === 3 && $parts[0] === "threads" && $parts[2] === "apply") {
	$thread = $store->getThread($parts[1]);
	if (!is_array($thread)) {
		RefinementPage::missing();
		return;
	}

	$to = $base . "/threads/" . $thread["id"];
	$comment = $manager->apply($thread["id"], $user);

	// A thread whose design was deleted is a thread with no design, and
	// a rule corrects something: there is nothing here to apply it to.
	if (!$comment) {
		RefinementPage::fail($to, "This thread has no design left to apply its rules to.");
		return;
	}

	$manager->dispatch($thread["id"], $comment["id"], $comment["body"]);
	RefinementPage::go($to);
	return;
}

// POST /threads/<id>/comments/<comment>/delete: one rule, or the whole
// thread when the comment is the design the rules correct.
if ($posted && count($parts) === 5 && $parts[0] === "threads" && $parts[2] === "comments" && $parts[4] === "delete") {
	$thread = $store->getThread($parts[1]);
	if (!is_array($thread)) {
		RefinementPage::missing();
		return;
	}

	$design = $store->getDesignComment($thread["id"]);

	// The design is what the rules correct, so removing it removes the
	// thread: nothing is left for a rule to be a rule about.
	if (is_array($design) && $design["id"] === $parts[3]) {
		$manager->drop($thread["id"]);
		RefinementPage::go($base);
		return;
	}

	$manager->remove($parts[3]);
	RefinementPage::go($base . "/threads/" . $thread["id"]);
	return;
}

// GET /threads/<id>: one thread.
if (count($parts) === 2 && $parts[0] === "threads") {
	$thread = $store->getThread($parts[1]);
	if (!is_array($thread)) {
		RefinementPage::missing();
		return;
	}

	$thread = $manager->sync($thread);

	$page->thread($thread, $store->getComments($thread["id"]));
	return;
}

// GET /: every thread, newest first, and the form that starts one.
$page->threads($store->getThreads());
