name: recent
description: Task\Recent builds the right rail - the merge of two lists that are each newest first, the heading of when an entry started, and what one run of a job reads as.
---
<?php

require "ext/autoload.php";

// The rail reads the whole run table, because that is what a rail of the
// latest runs by anyone is. So nothing here calls entries(): the merge, the
// grouping and the shape of one entry are each reached directly, and the one
// row this fixture reads it reads by its own ID.

$folder = "fx-recent";

$db = Task\DB::writer();

$db->replace("run", array(
	"id" => "fx-recent-1",
	"folder" => $folder,
	"job" => "build",
	"dir" => "",
	"user_name" => "ops",
	"state" => "running",
	"started_at" => "2026-11-02 10:01:00.000 +0000 UTC",
	"term_cols" => 120,
	"term_rows" => 30,
));

// The merge walks two lists that each arrive newest first. A run and a
// workflow run of the same instant keep the order the Go side's stable sort
// keeps, which is the run first.
$runs = array(
	array("at" => "2026-11-02 10:05:00.000 +0000 UTC", "href" => "/runs/a"),
	array("at" => "2026-11-02 09:03:00.000 +0000 UTC", "href" => "/runs/b"),
	array("at" => "2026-11-02 08:01:00.000 +0000 UTC", "href" => "/runs/c"),
);
$flows = array(
	array("at" => "2026-11-02 10:04:00.000 +0000 UTC", "href" => "/workflows/x"),
	array("at" => "2026-11-02 09:03:00.000 +0000 UTC", "href" => "/workflows/y"),
);

// hrefs names what a merge answered, in the order it answered.
function hrefs($entries) {
	$out = array();

	foreach ($entries as $entry) {
		$out[] = $entry["href"];
	}

	return implode(",", $out);
}

echo hrefs(Task\Recent::merge($runs, $flows, 50)), "\n";
echo hrefs(Task\Recent::merge($runs, $flows, 2)), "\n";
echo hrefs(Task\Recent::merge($runs, array(), 50)), "\n";
echo hrefs(Task\Recent::merge(array(), $flows, 50)), "\n";
echo "[", hrefs(Task\Recent::merge(array(), array(), 50)), "]\n";

// Two entries of the same second are ordered by the fraction the column
// holds, as the Go page's compare of two times is. Whole seconds would read
// them as equal and answer the run first whichever started first, so the pair
// is merged both ways round and the newer leads either way.
$earlier = array(array("at" => "2026-11-02 10:05:00.100 +0000 UTC", "href" => "/runs/d"));
$later = array(array("at" => "2026-11-02 10:05:00.500 +0000 UTC", "href" => "/workflows/z"));

echo hrefs(Task\Recent::merge($earlier, $later, 50)), "\n";
echo hrefs(Task\Recent::merge($later, $earlier, 50)), "\n";

// The heading of an entry: whole hours, then whole days, and the hour it
// started in for anything under one. The since is nanoseconds, as a
// time.Duration is, so a second short of an hour is still the last hour.
$second = Task\Clock::NANO;

foreach (array(0, 3599, 3600, 7200, 86399, 86400, 172800) as $since) {
	echo Task\Recent::hoursAgo($since * $second), "\n";
}

// Entries under their headings, newest first: the ones that share a heading
// sit under one, and a heading that changes starts another. The moment is
// given, so the headings are the same on every run of the suite.
$now = Task\Clock::nanos("2026-11-02 12:30:00.000 +0000 UTC");
$groups = Task\Recent::group(Task\Recent::merge($runs, $flows, 50), $now);

foreach ($groups as $group) {
	echo $group["label"], ": ", hrefs($group["entries"]), "\n";
}

// One run of a job as the component reads it. The icon is an eye and not a
// play: the catalogue here holds no folder, so nothing says the reader may
// start the job again.
$ops = new Task\Model\User(array("id" => "fx-recent-u", "name" => "ops", "role" => "runner", "admin" => true));
$access = new Task\Access($ops);
$roles = array("ops" => "runner");
$rail = new Task\Recent;
$table = new Task\Run;
$first = $table->get("fx-recent-1");
$entry = $rail->ofRun($first, $access, $roles, array());

echo $entry["href"], " ", $entry["name"], " ", $entry["place"], " ", $entry["state"], "\n";
echo var_export($entry["ended"], true), " ", var_export($entry["flow"], true), " ", $entry["act"], "\n";
echo $entry["title"], "\n";
echo $entry["avatar"]["name"], " ", $entry["avatar"]["role"], " ", $entry["avatar"]["initials"], "\n";

// runnable is the icon an entry carries: the reader may start the job again
// where the catalogue lists the run's folder and the host resolves the run's
// job out of it. The host is asked and no catalogue is walked for the job
// itself, because the jobs of a directory inside a folder are in no catalogue
// and neither are the four a folder has without declaring them.
$build = Task\Fixture::job("build");
$test = Task\Fixture::job("test");
$run = $table->get("fx-recent-1");

// A folder that declares no such job: the host answers not found, so nothing
// says the reader may start it again. Nobody signed in may either.
$undeclared = array(Task\Fixture::folder($folder, array($test)));

echo var_export(Task\Recent::runnable($run, $access, $undeclared), true), "\n";
echo var_export(Task\Recent::runnable($run, null, $undeclared), true), "\n";

// The folder declares the job, so the host answers it and the rule over it is
// the one the sidebar's own rows are drawn by. A catalogue of other folders
// alone answers no, because the rule reads the allow list of the run's folder
// and the page does not hold it.
$known = Task\Fixture::folder($folder, array($build, $test));
$catalogue = array(Task\Fixture::declared($known));
$elsewhere = array(Task\Fixture::folder("fx-recent-other", array($build)));

echo var_export(Task\Recent::runnable($run, $access, $catalogue), true), "\n";
echo var_export(Task\Recent::runnable($run, $access, $elsewhere), true), "\n";

// The role of a name no row holds is "", which draws the avatar without one.
echo "[", Task\Recent::role($roles, "nobody"), "]\n";

// Where the rail's own link goes, and how many entries it lists.
$data = $rail->data($access, $roles, array());

echo $data["href"], " ", Task\Recent::COUNT, "\n";
---
/runs/a,/workflows/x,/runs/b,/workflows/y,/runs/c
/runs/a,/workflows/x
/runs/a,/runs/b,/runs/c
/workflows/x,/workflows/y
[]
/workflows/z,/runs/d
/workflows/z,/runs/d
In the last hour
In the last hour
1 hour ago
2 hours ago
23 hours ago
1 day ago
2 days ago
2 hours ago: /runs/a,/workflows/x
3 hours ago: /runs/b,/workflows/y
4 hours ago: /runs/c
/runs/fx-recent-1 build fx-recent running
false false eye
Open fx-recent / build, ops, running
ops runner OP
false
false
true
false
[]
/runs 50
