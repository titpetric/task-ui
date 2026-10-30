<?php

// The routing a path needs and the two pages the module draws.
//
// Every page renders through MiniTPL, from templates/, compiled into
// templates_c/ which the module ships. The module escapes what it
// renders: MiniTPL does that by default for a text node, an attribute
// and an HTML comment, and the one value that reaches the page unescaped
// is the report, which is markdown a run wrote and the module rendered.

/**
 * refinement_path is what is left of a request for the module to answer:
 * the path past its own prefix, without the query and without the
 * slashes at either end.
 */
function refinement_path($uri, $base) {
	$parts = explode("?", $uri);
	$path = $parts[0];

	if (strpos($path, $base) === 0) {
		$path = substr($path, strlen($base));
	}

	return trim($path, "/");
}

/** refinement_tpl is the module's template engine, compiled ahead of time. */
function refinement_tpl() {
	$tpl = new MiniTPL\Template("templates/");

	$tpl->set_compile_location("templates_c/", true);

	return $tpl;
}

/** refinement_go answers with a redirect, which a form that changed something should. */
function refinement_go($to) {
	header("Location: " . $to);
	http_response_code(303);
}

/** refinement_missing answers for a thread that is not there. */
function refinement_missing() {
	http_response_code(404);
	echo "<p>No thread there.</p>";
}

/** refinement_fail says what a form needs and sends the reader back to it. */
function refinement_fail($to, $why) {
	http_response_code(400);
	echo "<p class=\"bad\">", htmlspecialchars($why, ENT_QUOTES), "</p>";
	echo "<p><a href=\"", htmlspecialchars($to, ENT_QUOTES), "\">Back</a></p>";
}

/** refinement_render_threads draws the thread list and the form that starts one. */
function refinement_render_threads($base, $threads) {
	$tpl = refinement_tpl();

	$tpl->load("threads.tpl");
	$tpl->assign("base", $base);
	$tpl->assign("threads", $threads);

	echo $tpl->get();
}

/**
 * refinement_render_thread draws one thread: the index its last run
 * wrote, the artifacts behind it, the comments, and the box that asks
 * for the next correction.
 */
function refinement_render_thread($base, $thread, $comments, $artifacts) {
	$tpl = refinement_tpl();

	$tpl->load("thread.tpl");
	$tpl->assign("base", $base);
	$tpl->assign("thread", $thread);
	$tpl->assign("comments", $comments);
	$tpl->assign("artifacts", $artifacts);
	$tpl->assign("files", "/artifacts/files/" . $thread["id"] . "/");
	$tpl->assign("report", refinement_report($thread["report_markdown"], $thread["id"]));

	echo $tpl->get();
}

/**
 * refinement_report renders the README a run wrote as the HTML the page
 * shows: the index of everything the thread's comments produced.
 *
 * It is markdown the instance's own agent wrote into the artifact tree,
 * not anything a reader typed, and the page draws it rather than
 * printing it. An image in it is embedded, which is how a schema diagram
 * is read on the page instead of downloaded.
 *
 * The folder is the thread's own, and the report names its files from
 * inside it: [schema/user.up.sql](schema/user.up.sql) is the thread's
 * copy and the link has to carry the folder the run left out.
 */
function refinement_report($markdown, $folder) {
	$text = trim($markdown);

	if ($text === "") {
		return "";
	}

	return Task::markdown($text, $folder);
}
