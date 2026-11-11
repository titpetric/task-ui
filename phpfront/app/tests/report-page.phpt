name: report-page
description: The report page of a run is the document without the terminal - which run it names, the sentence under the heading, the buttons to the two runs and the job, and the question an answer answers.
---
<?php

require "modules/runs/mod_report.php";

$mod = new mod_report;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo var_export($mod instanceof Module, true), "\n";

// One connection serves the suite and fixtures run together, so every row here
// is named after this fixture and every read names an ID of its own.
$db = Task\DB::writer();

$db->replace("user", array("id" => "fx-doc-u1", "name" => "fx-doc-ann", "role" => "admin", "admin" => 1));

// at is a moment of the day the rows record, in the shape a write leaves.
function at($time) {
	return "2026-11-02 " . $time . " +0000 UTC";
}

// The run a report is about, the run of Claude's that answered a question
// about it, and one whose question is about a run that is gone.
$rows = array(
	array("id" => "fx-doc-1", "job" => "build", "state" => "ok", "from" => at("10:01:00.000"), "to" => at("10:01:02.000")),
	array("id" => "fx-doc-2", "job" => "@claude-prompt", "state" => "ok", "from" => at("10:02:00.000"), "to" => at("10:02:30.000")),
	array("id" => "fx-doc-3", "job" => "@claude-prompt", "state" => "ok", "from" => at("10:03:00.000"), "to" => at("10:03:30.000")),
);

foreach ($rows as $row) {
	$db->replace("run", array(
		"id" => $row["id"],
		"folder" => "fx-doc",
		"job" => $row["job"],
		"dir" => "",
		"user_name" => "fx-doc-ann",
		"state" => $row["state"],
		"started_at" => $row["from"],
		"finished_at" => $row["to"],
	));
}

// The summary the runs page kept for the run that ended, which is a report of
// its own: a summary with no thread.
$db->replace("run_summary", array(
	"run_id" => "fx-doc-1",
	"summary" => "the build was clean",
	"user_name" => "claude",
	"created_at" => Task\Clock::stamp(),
));

// The threads: a question about the run, answered by a run of Claude's, and a
// question about a run that is no longer there.
$notes = array(
	array("id" => "fx-doc-n1", "parent" => "", "by" => "fx-doc-ann", "body" => "why did this run end the way it did", "about" => "fx-doc-1", "run" => ""),
	array("id" => "fx-doc-n2", "parent" => "fx-doc-n1", "by" => "claude", "body" => "the cache was cold", "about" => "fx-doc-1", "run" => "fx-doc-2"),
	array("id" => "fx-doc-n3", "parent" => "", "by" => "fx-doc-ann", "body" => "what did this one do", "about" => "fx-doc-gone", "run" => ""),
	array("id" => "fx-doc-n4", "parent" => "fx-doc-n3", "by" => "claude", "body" => "the run is gone", "about" => "fx-doc-gone", "run" => "fx-doc-3"),
);

foreach ($notes as $row) {
	$db->replace("note", array(
		"id" => $row["id"],
		"channel" => "notes",
		"parent_id" => $row["parent"],
		"user_name" => $row["by"],
		"body" => $row["body"],
		"mentions" => "",
		"state" => "done",
		"run_id" => $row["run"],
		"about_run_id" => $row["about"],
		"workspace_id" => "fx-doc-ws",
		"created_at" => Task\Clock::stamp(),
		"updated_at" => Task\Clock::stamp(),
	));
}

Task\Host::$whoami = "fx-doc-ann";
$build = Task\Fixture::job("build");
$catalogue = Task\Fixture::folder("fx-doc", array($build));

Task\Host::$folders = array(Task\Fixture::declared($catalogue));

$runs = new Task\Runs\Runs;
$reports = new Task\Report;
$chrome = new Task\Chrome;
$seen = new Task\Seen($chrome);

// labels names the buttons of the head, in the order the head draws them.
function labels($data) {
	$out = array();

	foreach ($data["items"] as $item) {
		$out[] = $item["label"] . " " . $item["href"];
	}

	return implode(", ", $out);
}

// A stamp reads in the server's own zone and how long ago a report was written
// counts from this moment, so the sentence is read with the stamp masked.
function said($lead, $run) {
	return str_replace(Task\Clock::read($run->getStartedAt()), "<stamp>", $lead);
}

$plain = $runs->get("fx-doc-1");
$claude = $runs->get("fx-doc-2");
$orphan = $runs->get("fx-doc-3");

// The run the page names: the run the report is about, and the page's own run
// where the report is of that run.
$summary = $reports->of($plain);
$answer = $reports->of($claude);
$lost = $reports->of($orphan);

echo $mod->subject($runs, $seen, $plain, $summary)->getId(), "\n";
echo $mod->subject($runs, $seen, $claude, $answer)->getId(), "\n";

// A question about a run that is gone names the page's own run, and so does a
// report whose run the reader does not see.
echo $mod->subject($runs, $seen, $orphan, $lost)->getId(), "\n";

Task\Host::$whoami = "";

$nobody = new Task\Chrome;
$hidden = new Task\Seen($nobody);

echo $mod->subject($runs, $hidden, $claude, $answer)->getId(), "\n";

Task\Host::$whoami = "fx-doc-ann";

// The head of a summary names the one run it is of, and the job the run is of.
echo labels(mod_report::buttons($chrome, $plain, $plain)->data()), "\n";

// The head of an answer names both runs, because they differ, and the job of
// the run the answer is about.
echo labels(mod_report::buttons($chrome, $claude, $plain)->data()), "\n";

// A run of Claude's has no job page, so the head names no job.
echo labels(mod_report::buttons($chrome, $claude, $claude)->data()), "\n";

// The sentence under the heading: what the document is, of which run, who
// asked and when Claude wrote it. A kept summary names the user it was kept
// under and was written at a moment of its own.
echo said(mod_report::lead($plain, $summary, true), $plain), "\n";
echo said(mod_report::lead($plain, $answer, false), $plain), "\n";

// A run that has not ended is named without a state, and a report with neither
// a user nor a time ends after the run it is of.
$bare = array("run" => "fx-doc-1", "about" => "fx-doc-1", "thread" => "", "question" => "", "by" => "", "body" => "nothing asked", "at" => "", "pending" => false);
$live = $runs->get("fx-doc-1");

$live->setState(Task\Run::RUNNING);

echo said(mod_report::lead($live, $bare, true), $live), "\n";

// The whole scope the template reads. A summary draws one block, because the
// question it answers is the page's own, and an answer draws the question
// above it.
$scope = $mod->scope($chrome, $runs, $seen, $plain, $summary);

echo implode(",", array_keys($scope)), " ", implode(",", array_keys($scope["head"])), "\n";
echo $scope["head"]["title"], " ", $scope["head"]["badge"], " ", var_export($scope["summary"], true), " [", $scope["question"], "]\n";
echo $scope["body"], "\n";

$asked = $mod->scope($chrome, $runs, $seen, $claude, $answer);

echo $asked["head"]["title"], " ", var_export($asked["summary"], true), "\n";
echo $asked["question"], "\n";
echo $asked["body"], "\n";
echo labels($asked["head"]["buttons"]), "\n";
---
runs Runs list
true
fx-doc-1
fx-doc-1
fx-doc-3
fx-doc-2
Run /runs/fx-doc-1, Job /f/internal/fx-doc/build
Run /runs/fx-doc-1, Claude's run /runs/fx-doc-2, Job /f/internal/fx-doc/build
Run /runs/fx-doc-2
Claude's summary of the run of <stamp> by fx-doc-ann, which ended ok; asked by claude, written just now.
Claude's answer about the run of <stamp> by fx-doc-ann, which ended ok; asked by fx-doc-ann, written just now.
Claude's summary of the run of <stamp> by fx-doc-ann.
head,summary,question,body title,icon,badge,tone,tags,note,lead,leads,faint,live,run,actions,buttons,show
fx-doc / build report true [<p></p>]
<p>the build was clean</p>
fx-doc / build false
<p>why did this run end the way it did</p>
<p>the cache was cold</p>
Run /runs/fx-doc-1, Claude's run /runs/fx-doc-2, Job /f/internal/fx-doc/build
