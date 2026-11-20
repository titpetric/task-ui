name: Dashboard processes
description: >
  A run in progress is one row of the dashboard's process table: its PID, who
  started it, its state and start, what it runs, and where its work is read -
  the thread of a prompt, a host's page, or the page of its job.
---
<?php

require "modules/dashboard/mod_dashboard.php";

$mod = new mod_dashboard;

// at is a moment the rows record, in the shape a write leaves.
function at($time) {
	return "2026-11-02 " . $time . " +0000 UTC";
}

// run makes a run row: the table holds the state, the runner names the title.
function run($id, $folder, $job, $state, $title = "", $from = "", $to = "") {
	return new Task\Model\Run([
		"id" => $id,
		"folder" => $folder,
		"dir" => "",
		"job" => $job,
		"args" => "[]",
		"user_name" => "fx-db-ann",
		"state" => $state,
		"title" => $title,
		"started_at" => $from === "" ? "" : $from,
		"finished_at" => $to === "" ? "" : $to,
	]);
}

// notes is the thread each run answers, as the store's byRun reads it.
class DashNotes {
	var $threads = [];

	function byRun($id) {
		return isset($this->threads[$id]) ? $this->threads[$id] : null;
	}
}

function note($id, $body) {
	return new Task\Model\Note(["id" => $id, "body" => $body, "parent_id" => ""]);
}

$threads = new DashNotes;
$threads->threads["fx-db-prompt-0001"] = note("fx-db-note-0001", "why is the build red");
$threads->threads["fx-db-prompt-0002"] = note("fx-db-note-0001", "why is the build red");
$threads->threads["fx-db-note-run-0001"] = note("fx-db-note-0003", "noted");

// line is a row as the template reads it, minus the length of a live run,
// which counts from the process clock.
function line($row) {
	return implode("|", [$row["pid"], $row["user"], $row["state"], $row["start"], $row["task"], $row["href"], $row["command"]]);
}

// A job's row opens the page of the job, and its length reads to the tenth.
$job = $mod->process(run("fx-db-job-0001", "fx-db", "build", "running", "building the image", at("10:01:00.000"), at("10:01:02.500")), $threads, true, []);

echo line($job), " ", $job["time"], "\n";

// A prompt's row opens its thread, and the command is the question.
echo line($mod->process(run("fx-db-prompt-0001", "internal", "@claude-prompt", "running", "", at("10:02:00.000")), $threads, true, [])), "\n";

// A reply's run opens the thread it answers, whose question is the command.
echo line($mod->process(run("fx-db-prompt-0002", "internal", "@claude-prompt", "running", "", at("10:03:00.000")), $threads, true, [])), "\n";

// The Notes agent's run answers the note it was asked.
echo line($mod->process(run("fx-db-note-run-0001", "internal", "@claude-note", "running", "", at("10:04:00.000")), $threads, true, [])), "\n";

// A command on a host opens the host's page.
echo line($mod->process(run("fx-db-host-0001", "internal", "@host:db1", "running", "uptime", at("10:05:00.000")), $threads, true, [])), "\n";

// A reader who may read no thread gets no link to one.
echo line($mod->process(run("fx-db-prompt-0001", "internal", "@claude-prompt", "running", "", at("10:02:00.000")), $threads, false, [])), "\n";

// A queued run has no start and no length.
$queued = $mod->process(run("fx-db-queued-0001", "fx-db", "test", "queued"), $threads, true, []);

echo "[", $queued["start"], "] [", $queued["started"], "] [", $queued["time"], "]\n";
---
fx-db-jo|fx-db-ann|running|10:01|fx-db / build|/f/internal/fx-db/build|building the image 2.5s
fx-db-pr|fx-db-ann|running|10:02|@claude-prompt|/threads/fx-db-note-0001|why is the build red
fx-db-pr|fx-db-ann|running|10:03|@claude-prompt|/threads/fx-db-note-0001|why is the build red
fx-db-no|fx-db-ann|running|10:04|@claude-note|/threads/fx-db-note-0003|noted
fx-db-ho|fx-db-ann|running|10:05|@host:db1|/inventory/hosts/db1|uptime
fx-db-pr|fx-db-ann|running|10:02|@claude-prompt||why is the build red
[] [] [0.0s]
