<?php

// @route GET /host

// Echoes every host call, so the test sees what a Go value looks like once it
// has crossed into the runtime.

echo "base=", Task\Host::base(), "\n";
echo "version=", Task\Host::version(), "\n";
echo "whoami=", Task\Host::whoami(), "\n";
echo "dirty=", Task\Host::dirty("/src/infra"), "\n";
echo "folders=", json_encode(Task\Host::folders()), "\n";
echo "running=", json_encode(Task\Host::running()), "\n";
echo "rules=", json_encode(Task\Host::rules()), "\n";

// Indexing by key is the whole contract: a Go struct crosses as a host object
// whose fields no key reaches, so the boundary converts and every value that
// arrives here is an array all the way down.
$folders = Task\Host::folders();
$rules = Task\Host::rules();

echo "indexed=", var_export(is_array($folders) && is_array($rules), true), "\n";
echo "keys=", json_encode(array_keys($rules)), "\n";

foreach ($folders as $folder) {
	echo "folder=", $folder["id"], " jobs=", implode(",", $folder["jobs"]), "\n";
}

foreach ($rules["sections"] as $section) {
	echo "section=", $section["key"], ":", $section["title"], "\n";
}

echo "checks=", implode(",", $rules["checks"]), "\n";
echo "markdown=", Task\Host::markdown("# Title", ""), "\n";
echo "published=", json_encode(Task\Host::publish("runs", array("id" => "01hz"))), "\n";
