<?php

// @route GET /host

// Echoes every host call, so the test sees what a Go value looks like once it
// has crossed into the runtime.

echo "base=", Task\Host::base(), "\n";
echo "version=", Task\Host::version(), "\n";
echo "whoami=", Task\Host::whoami(), "\n";
echo "dirty=", Task\Host::dirty("/src/infra"), "\n";
echo "folders=", json_encode(Task\Host::folders()), "\n";
echo "job=", json_encode(Task\Host::job("infra", "api", "@shell")), "\n";
echo "nojob=", json_encode(Task\Host::job("infra", "", "gone")), "\n";
echo "running=", json_encode(Task\Host::running()), "\n";
echo "flows=", json_encode(Task\Host::flows()), "\n";
echo "rules=", json_encode(Task\Host::rules()), "\n";
echo "session=", json_encode(Task\Host::session("01hz")), "\n";
echo "nosession=", json_encode(Task\Host::session("gone")), "\n";
echo "frame=", Task\Host::frame("01hz"), "\n";
echo "expired=", json_encode(Task\Host::expired("gone.cast", "ok")), "\n";
echo "kept=", json_encode(Task\Host::expired("there.cast", "ok")), "\n";
echo "claude=", json_encode(Task\Host::claude()), "\n";
echo "tree=", json_encode(Task\Host::tree("lint")), "\n";
echo "notree=", json_encode(Task\Host::tree("gone")), "\n";
echo "space=", json_encode(Task\Host::space()), "\n";
// A directory nothing measured is left out of the answer, so a page reads a
// size for the projects that have one and says nothing about the rest.
echo "spaces=", json_encode(Task\Host::spaces(array("/root/src/infra", "/root/src/gone"))), "\n";
// The directory of a folder, one at a time for the pages that ask about one
// and keyed for a tree that asks about every row it draws. An ID the layout
// places nowhere is left out of the keyed answer, as a size is.
// The answer is read by key and never walked, so its order is nobody's, which
// is why this prints the lookups and not the array.
echo "folderDir=", Task\Host::folderDir("infra", array("malta")), "\n";

$dirs = Task\Host::folderDirs(array("infra", "deploy", ""), array("malta"));

function held($dirs, $id) {
	return isset($dirs[$id]) ? $dirs[$id] : "-";
}

echo "folderDirs=", count($dirs), " ", held($dirs, "infra"), " ", held($dirs, "deploy"), " ", held($dirs, ""), "\n";
// A service's settings cross as the form draws them: what is set, never what
// it is set to.
echo "integrations=", json_encode(Task\Host::integrations()), "\n";
// The hosts permission is the reader's, which PHP decides and passes, and it
// is what makes task-ui read a host rather than answer out of what it kept.
echo "inventory=", json_encode(Task\Host::inventory(true)), "\n";
echo "watching=", json_encode(Task\Host::inventory(false)["logins"]), "\n";
echo "host=", json_encode(Task\Host::host("web1", true)), "\n";
echo "nohost=", json_encode(Task\Host::host("gone", true)), "\n";

// A login and a host's records are read by the address and the ID of the line
// that wants them, so both arrive keyed and not as a list to search.
$inventory = Task\Host::inventory(true);
$logins = $inventory["logins"];
$records = $inventory["history"];

echo "login=", isset($logins["10.0.0.1"]) ? $logins["10.0.0.1"] : "-", "\n";
echo "records=", isset($records["sys1"]) ? count($records["sys1"]) : 0, "\n";

// Indexing by key is the whole contract: a Go struct crosses as a host object
// whose fields no key reaches, so the boundary converts and every value that
// arrives here is an array all the way down.
$folders = Task\Host::folders();
$rules = Task\Host::rules();

echo "indexed=", var_export(is_array($folders) && is_array($rules), true), "\n";
echo "keys=", json_encode(array_keys($rules)), "\n";

foreach ($folders as $folder) {
	$names = array();

	foreach ($folder["jobs"] as $job) {
		$names[] = $job["id"] . "/" . $job["name"];
	}

	echo "folder=", $folder["id"], " allow=", implode(",", $folder["allow"]), " jobs=", implode(",", $names), "\n";
}

foreach ($rules["sections"] as $section) {
	echo "section=", $section["key"], ":", $section["title"], "\n";
}

echo "checks=", implode(",", $rules["checks"]), "\n";
echo "markdown=", Task\Host::markdown("# Title", ""), "\n";
// What a relative link resolves against is the caller's: no prefix leaves it
// as text, which is what a run's report wants, and a prefix links it under the
// directory the text was written in.
echo "plain=", Task\Host::markdown("see [notes](lint/out.txt)", ""), "\n";
echo "linked=", Task\Host::markdown("see [notes](out.txt)", "/artifacts/files/", "lint"), "\n";
echo "published=", json_encode(Task\Host::publish("runs", array("id" => "01hz"))), "\n";
// The run behind a dispatch arrives after the dispatch answered, so asking
// for the run of no note is the label of a run that has not started.
echo "run=", Task\Host::run("01NOTE"), "|", Task\Host::run(""), "\n";
