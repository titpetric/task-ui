name: report
description: Task\Report is the markdown a run stands for - the summary of a run, the answer of a run of Claude's, the batch's summary for a run of a batch, and the summary the runs page kept.
---
<?php

require "ext/autoload.php";

// One connection serves the suite and fixtures run together, so every run,
// note and summary here is named after this fixture and every query narrows to
// them. A fixture that read the whole table would read another fixture's rows.

$db = Task\DB::writer();

// The runs a report is about: one that ended, one still running, two of a
// batch, and the run of Claude's that answered a prompt.
$runs = array(
	array("id" => "fx-rep-run", "job" => "build", "state" => "ok", "at" => "10:01", "batch" => ""),
	array("id" => "fx-rep-plain", "job" => "build", "state" => "ok", "at" => "10:02", "batch" => ""),
	array("id" => "fx-rep-pending", "job" => "build", "state" => "ok", "at" => "10:03", "batch" => ""),
	array("id" => "fx-rep-failed", "job" => "build", "state" => "ok", "at" => "10:04", "batch" => ""),
	array("id" => "fx-rep-kept", "job" => "build", "state" => "ok", "at" => "10:05", "batch" => ""),
	array("id" => "fx-rep-batch-new", "job" => "@host:web2", "state" => "ok", "at" => "10:07", "batch" => "fx-rep-batch"),
	array("id" => "fx-rep-batch-old", "job" => "@host:web1", "state" => "ok", "at" => "10:06", "batch" => "fx-rep-batch"),
	array("id" => "fx-rep-claude", "job" => "@claude-prompt", "state" => "ok", "at" => "10:08", "batch" => ""),
	array("id" => "fx-rep-live", "job" => "@claude-prompt", "state" => "running", "at" => "10:09", "batch" => ""),
);

foreach ($runs as $row) {
	$db->replace("run", array(
		"id" => $row["id"],
		"folder" => "fx-rep",
		"job" => $row["job"],
		"dir" => "",
		"user_name" => "admin",
		"state" => $row["state"],
		"batch_id" => $row["batch"],
		"started_at" => "2026-11-02 " . $row["at"] . ":00.000 +0000 UTC",
	));
}

// The threads. A summary thread asks one of two questions and names the run it
// is about; its first reply names the run that answers it. A thread whose
// question is the reader's own is an answer and not a summary.
$notes = array(
	array("id" => "fx-rep-n1", "parent" => "", "body" => Task\Report::SUMMARY, "about" => "fx-rep-plain", "run" => "", "state" => "done"),
	array("id" => "fx-rep-n2", "parent" => "fx-rep-n1", "body" => "the build is clean", "about" => "fx-rep-plain", "run" => "fx-rep-sum1", "state" => "done", "by" => "claude"),
	array("id" => "fx-rep-n3", "parent" => "", "body" => Task\Report::SUMMARY, "about" => "fx-rep-pending", "run" => "fx-rep-sum3", "state" => "working"),
	array("id" => "fx-rep-n4", "parent" => "", "body" => Task\Report::SUMMARY, "about" => "fx-rep-failed", "run" => "fx-rep-sum4", "state" => "failed"),
	array("id" => "fx-rep-n5", "parent" => "", "body" => Task\Report::BATCH, "about" => "fx-rep-batch-old", "run" => "", "state" => "done"),
	array("id" => "fx-rep-n6", "parent" => "fx-rep-n5", "body" => "both hosts answered", "about" => "fx-rep-batch-old", "run" => "fx-rep-sum5", "state" => "done", "by" => "claude"),
	array("id" => "fx-rep-n7", "parent" => "", "body" => "why did this run end the way it did", "about" => "fx-rep-run", "run" => "", "state" => "done"),
	array("id" => "fx-rep-n8", "parent" => "fx-rep-n7", "body" => "it ran out of disk", "about" => "fx-rep-run", "run" => "fx-rep-claude", "state" => "done", "by" => "claude"),
);

foreach ($notes as $row) {
	$db->replace("note", array(
		"id" => $row["id"],
		"channel" => "notes",
		"parent_id" => $row["parent"],
		"user_name" => isset($row["by"]) ? $row["by"] : "admin",
		"body" => $row["body"],
		"mentions" => "",
		"state" => $row["state"],
		"run_id" => $row["run"],
		"about_run_id" => $row["about"],
		"workspace_id" => "fx-rep-ws",
		"created_at" => "2026-11-02 11:00:00.000 +0000 UTC",
		"updated_at" => "2026-11-02 11:00:00.000 +0000 UTC",
	));
}

// The summary the runs page keeps, for a run no thread asks about.
$db->replace("run_summary", array(
	"run_id" => "fx-rep-kept",
	"summary" => "kept by hand",
	"user_name" => "admin",
	"created_at" => "2026-11-02 12:00:00.000 +0000 UTC",
));

// A page reads through the reader, which refuses a write.
$reports = new Task\Report;
$rows = new Task\Run;

// told is a report as a page reads it.
function told($report) {
	if ($report === null) {
		return "none";
	}

	return $report["run"] . " about " . $report["about"] . " [" . $report["body"] . "] by " . $report["by"] . " summary " . var_export(Task\Report::isSummary($report), true);
}

// A plain run reports its summary, and the link is the report page of the run
// the report is of.
echo told($reports->of($rows->get("fx-rep-plain"))), "\n";
echo Task\Report::href($reports->of($rows->get("fx-rep-plain"))), "\n";

// A summary under way is no report, and neither is a thread that failed
// without an answer.
echo told($reports->of($rows->get("fx-rep-pending"))), "\n";
echo told($reports->of($rows->get("fx-rep-failed"))), "\n";

// A thread names the oldest run of a batch, and every run of the batch reports
// it: the report is of the run the page is on.
echo told($reports->of($rows->get("fx-rep-batch-new"))), "\n";
echo told($reports->of($rows->get("fx-rep-batch-old"))), "\n";

// A run of Claude's on a prompt reports the answer it wrote, which is no
// summary: the question is the one that was asked, and the reply before the
// answer is what it answers.
echo told($reports->of($rows->get("fx-rep-claude"))), "\n";
echo $reports->of($rows->get("fx-rep-claude"))["question"], "\n";

// A run of Claude's that has not ended has no answer yet.
echo told($reports->of($rows->get("fx-rep-live"))), "\n";

// A run no thread asks about reports the summary the runs page kept, which
// holds no thread and so reads as a summary.
echo told($reports->of($rows->get("fx-rep-kept"))), "\n";

// Nothing for a run with neither, and nothing for no run at all.
echo told($reports->of($rows->get("fx-rep-run"))), "\n";
echo told($reports->of(null)), "\n";

// The summaries of several runs at once, which is what a list page reads.
$found = $reports->summaries(array("fx-rep-plain", "fx-rep-pending", "fx-rep-kept", "fx-rep-run"));

echo implode(",", array_keys($found)), "\n";
echo var_export($found["fx-rep-pending"]["pending"], true), " ", var_export($found["fx-rep-plain"]["pending"], true), "\n";
echo count($reports->summaries(array())), "\n";
---
fx-rep-plain about fx-rep-plain [the build is clean] by admin summary true
/runs/fx-rep-plain/report
none
none
fx-rep-batch-new about fx-rep-batch-old [both hosts answered] by admin summary true
fx-rep-batch-old about fx-rep-batch-old [both hosts answered] by admin summary true
fx-rep-claude about fx-rep-run [it ran out of disk] by admin summary false
why did this run end the way it did
none
fx-rep-kept about fx-rep-kept [kept by hand] by admin summary true
none
none
fx-rep-pending,fx-rep-plain,fx-rep-kept
true false
0
