name: run-page
description: The page of one run lays the runner's session over the row - what the head says and offers, the live terminal, the details beside it, the files the run left, the prompt box about it, and the markup the whole scope draws.
---
<?php

require "modules/runs/mod_run.php";

$mod = new mod_run;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo var_export($mod instanceof Module, true), "\n";

// One connection serves the suite and fixtures run together, so every row here
// is named after this fixture and every read names an ID of its own.
$db = Task\DB::writer();

$db->replace("user", array("id" => "fx-rp-u1", "name" => "fx-rp-ann", "role" => "admin", "admin" => 1));

// at is a moment of the day the rows record, in the shape a write leaves.
function at($time) {
	return "2026-11-02 " . $time . " +0000 UTC";
}

// A run that ended with a recording and arguments; one in progress in a
// directory of its folder; one that waits for a slot; one that broke out of
// hours, whose recording the sweep removed; a run of Claude's; a job the
// folder declares none of; and one whose summary Claude is still writing.
$rows = array(
	array("id" => "fx-rp-1", "job" => "build", "state" => "ok", "cast" => "casts/fx-rp-1.cast", "args" => '["--verbose","./pkg"]', "from" => at("10:01:00.000"), "to" => at("10:01:02.500")),
	array("id" => "fx-rp-2", "job" => "deploy", "state" => "running", "dir" => "docker", "from" => Task\Clock::stamp()),
	array("id" => "fx-rp-3", "job" => "test", "state" => "queued"),
	array("id" => "fx-rp-4", "job" => "build", "state" => "failed", "code" => 2, "cast" => "casts/fx-rp-4.cast", "reason" => "the window of the job is closed", "forced" => 1, "scheduled" => 1, "from" => at("10:04:00.000"), "to" => at("10:04:01.000")),
	array("id" => "fx-rp-5", "job" => "@claude-prompt", "state" => "ok", "from" => at("10:05:00.000"), "to" => at("10:05:30.000")),
	array("id" => "fx-rp-6", "job" => "ghost", "state" => "ok", "from" => at("10:06:00.000"), "to" => at("10:06:01.000")),
	array("id" => "fx-rp-7", "job" => "build", "state" => "ok", "from" => at("10:09:00.000"), "to" => at("10:09:01.000")),
	array("id" => "fx-rp-b1", "job" => "build", "state" => "ok", "batch" => "fx-rp-batch", "from" => at("10:07:00.000"), "to" => at("10:07:01.000")),
	array("id" => "fx-rp-b2", "job" => "build", "state" => "ok", "batch" => "fx-rp-batch", "from" => at("10:08:00.000"), "to" => at("10:08:01.000")),
);

foreach ($rows as $row) {
	$db->replace("run", array(
		"id" => $row["id"],
		"folder" => "fx-rp",
		"job" => $row["job"],
		"dir" => isset($row["dir"]) ? $row["dir"] : "",
		"args" => isset($row["args"]) ? $row["args"] : "[]",
		"user_name" => "fx-rp-ann",
		"state" => $row["state"],
		"exit_code" => isset($row["code"]) ? $row["code"] : 0,
		"forced" => isset($row["forced"]) ? $row["forced"] : 0,
		"scheduled" => isset($row["scheduled"]) ? $row["scheduled"] : 0,
		"term_cols" => 120,
		"term_rows" => 40,
		"cast_path" => isset($row["cast"]) ? $row["cast"] : "",
		"reason" => isset($row["reason"]) ? $row["reason"] : "",
		"message" => "the row's own word",
		"batch_id" => isset($row["batch"]) ? $row["batch"] : "",
		"started_at" => isset($row["from"]) ? $row["from"] : null,
		"finished_at" => isset($row["to"]) ? $row["to"] : null,
	));
}

// The files the run left, newest first.
$db->replace("artifact", array(
	"id" => "fx-rp-a1",
	"run_id" => "fx-rp-1",
	"folder" => "fx-rp",
	"job" => "build",
	"name" => "report.html",
	"path" => "report.html",
	"byte_size" => 2048,
	"content_type" => "text/html",
	"created_at" => at("10:01:01.000"),
));
$db->replace("artifact", array(
	"id" => "fx-rp-a2",
	"run_id" => "fx-rp-1",
	"folder" => "fx-rp",
	"job" => "build",
	"name" => "build.log",
	"path" => "build.log",
	"byte_size" => 300,
	"content_type" => "text/plain",
	"created_at" => at("10:01:00.500"),
));

// The summary the runs page kept for the run that ended, which is the report
// its head links.
$db->replace("run_summary", array(
	"run_id" => "fx-rp-1",
	"summary" => "the build was clean",
	"user_name" => "claude",
	"created_at" => at("11:02:00.000"),
));

// A prompt thread about the run that ended, with Claude's answer under it. A
// prompt holds a workspace, so the board's own notes are not these.
$db->replace("note", array(
	"id" => "fx-rp-n1",
	"channel" => "notes",
	"parent_id" => "",
	"user_name" => "fx-rp-ann",
	"body" => "why did this run end the way it did",
	"mentions" => "",
	"state" => "done",
	"run_id" => "",
	"about_run_id" => "fx-rp-1",
	"workspace_id" => "fx-rp-ws",
	"created_at" => at("11:00:00.000"),
	"updated_at" => at("11:00:00.000"),
));
$db->replace("note", array(
	"id" => "fx-rp-n2",
	"channel" => "notes",
	"parent_id" => "fx-rp-n1",
	"user_name" => "claude",
	"body" => "the cache was cold",
	"mentions" => "",
	"state" => "done",
	"run_id" => "fx-rp-5",
	"about_run_id" => "fx-rp-1",
	"workspace_id" => "fx-rp-ws",
	"created_at" => at("11:01:00.000"),
	"updated_at" => at("11:01:00.000"),
));

// A summary thread Claude has not answered yet is a pending report: the head
// says so where Summarize stood, and offers no button to ask it again.
$db->replace("note", array(
	"id" => "fx-rp-n3",
	"channel" => "notes",
	"parent_id" => "",
	"user_name" => "fx-rp-ann",
	"body" => Task\Report::SUMMARY,
	"mentions" => "",
	"state" => "working",
	"run_id" => "",
	"about_run_id" => "fx-rp-7",
	"workspace_id" => "fx-rp-ws",
	"created_at" => at("11:03:00.000"),
	"updated_at" => at("11:03:00.000"),
));

// The catalogue the reader sees, the runner's own sessions, the recording the
// sweep removed and the token a prompt needs.
Task\Host::$whoami = "fx-rp-ann";
Task\Host::$claude = true;
$build = Task\Fixture::job("build");
$deploy = Task\Fixture::job("deploy");
$test = Task\Fixture::job("test");
$catalogue = Task\Fixture::folder("fx-rp", array($build, $deploy, $test));

Task\Host::$folders = array(Task\Fixture::declared($catalogue));

// The run of a directory inside the folder ran a job of that directory's own
// Taskfile, which the catalogue listing does not hold: the host reads the
// directory, so a fixture answers for it the same way.
Task\Fixture::resolves("fx-rp", "docker", $deploy);
Task\Host::$frames = array("fx-rp-2" => "<svg></svg>");
Task\Host::$expired = array("casts/fx-rp-4.cast" => true);
Task\Host::$sessions = array(
	"fx-rp-2" => array(
		"live" => true,
		"attached" => 2,
		"message" => "",
		"state" => "running",
		"exit_code" => 0,
		"title" => "pulling the image",
		"finished_at" => "",
	),
	"fx-rp-3" => array(
		"live" => true,
		"attached" => 0,
		"message" => "waits for a free slot, number 1 in line",
		"state" => "queued",
		"exit_code" => 0,
		"title" => "",
		"finished_at" => "",
	),
);

$runs = new Task\Run;
$chrome = new Task\Chrome;
$seen = new Task\Seen($chrome);

// labels names the buttons of a head, or the items of a menu, in order.
function labels($data) {
	$out = array();
	$groups = isset($data["groups"]) ? $data["groups"] : array($data["items"]);

	foreach ($groups as $group) {
		foreach ($group as $item) {
			$out[] = $item["label"];
		}
	}

	return implode(",", $out);
}

// attrs is what one button writes into the markup, escaped as the component
// draws it raw, with the link it is where it has one.
function attrs($data, $label) {
	foreach ($data["items"] as $item) {
		if ($item["label"] === $label) {
			return $item["attrs"] . " [" . $item["href"] . "] " . var_export($item["danger"], true) . " " . var_export($item["primary"], true);
		}
	}

	return "";
}

// faints are the sentences a head draws where a button would be, in order.
function faints($data) {
	$out = array();

	foreach ($data["items"] as $item) {
		if ($item["faint"] !== "") {
			$out[] = $item["faint"];
		}
	}

	return implode(",", $out);
}

// cell is one row of the details table, as the kv component reads it.
function cell($rows, $label) {
	foreach ($rows as $row) {
		if ($row["label"] === $label) {
			return $row;
		}
	}

	return array();
}

// rowLabels names the details the table draws, in the order it draws them.
function rowLabels($rows) {
	$out = array();

	foreach ($rows as $row) {
		$out[] = $row["label"];
	}

	return implode(",", $out);
}

// The session the page lays over a row is Task\Run's, because the table of
// runs lays the same one over every row of a list. A run the runner no longer
// holds is the row alone, and the row's own word is what it reports.
$one = $runs->get("fx-rp-1");

echo "[", Task\Run::overlay($one, Task\Host::session("fx-rp-1")), "] ", $one->getState(), " [", $one->getTitle(), "]\n";

// A run in progress reads its state, its title and what it reports from the
// session, which is ahead of the row.
$live = $runs->get("fx-rp-2");

echo "[", Task\Run::overlay($live, Task\Host::session("fx-rp-2")), "] ", $live->getState(), " [", $live->getTitle(), "]\n";

// A run that waits says what the runner reports, which no column holds.
$queued = $runs->get("fx-rp-3");

echo "[", Task\Run::overlay($queued, Task\Host::session("fx-rp-3")), "] ", $queued->getState(), "\n";

// The head of a run that ended: View, its job again, the report of it, and the
// replay of its recording. Stop and Kill belong to a run the runner still
// holds.
// The row is read again rather than reusing $one, which the overlay above
// wrote the session over.
$reports = new Task\Report;
$about = $runs->get("fx-rp-1");
$report = $reports->of($about);
$session = Task\Host::session("fx-rp-1");
$ended = mod_run::buttons($one, $report, false, true, $session, false, false);

echo labels($ended->data()), " ", var_export($ended->data()["small"], true), "\n";
echo attrs($ended->data(), "View"), "\n";
echo attrs($ended->data(), "Report"), "\n";
echo attrs($ended->data(), "|>"), "\n";

// Run again repeats the run as it was: its own folder, the directory it ran in,
// its job, the title the window gives it, and the arguments the column keeps.
// No args command goes with them, so nothing is asked before it starts.
echo attrs($ended->data(), "Run again"), "\n";

// Summarize asks Claude the question a summary asks, over the window the page
// is read in. A run that was given no arguments carries none.
$asks = mod_run::buttons($runs->get("fx-rp-6"), null, false, true, array(), true, false);

echo labels($asks->data()), "\n";
echo attrs($asks->data(), "Run again"), "\n";
echo attrs($asks->data(), "Summarize"), "\n";

// A summary Claude is still writing stands where Summarize would, as a faint
// sentence and no button: the row keeps its order and nothing asks twice.
$waits = mod_run::buttons($runs->get("fx-rp-7"), null, false, true, array(), false, true);

echo labels($waits->data()), " [", faints($waits->data()), "]\n";
echo "[", attrs($waits->data(), "Summarize"), "]\n";

// A run the runner holds may be stopped and killed by a reader who may run its
// job, and offers no replay until it ends. It runs again once it ended.
echo labels(mod_run::buttons($live, null, false, true, Task\Host::session("fx-rp-2"), false, false)->data()), "\n";
echo labels(mod_run::buttons($live, null, false, false, Task\Host::session("fx-rp-2"), false, false)->data()), "\n";

// A recording the sweep removed is no replay.
echo labels(mod_run::buttons($runs->get("fx-rp-4"), null, true, true, Task\Host::session("fx-rp-4"), false, false)->data()), "\n";

// The terminal is the live screen of a run the runner holds, and nothing for
// one that ended.
$screen = mod_run::terminal($live, Task\Host::session("fx-rp-2"), Task\Run::name($live));

echo $screen["id"], " ", $screen["name"], " ", $screen["watching"], " ", $screen["frame"], "\n";
echo var_export(mod_run::terminal($one, Task\Host::session("fx-rp-1"), "x"), true), "\n";

// The details of a run that ended: where it ran, what it was given, who
// started it, how long it took and how it ended. The arguments row comes of a
// run that was given some, and the directory row of one that ran in one.
$details = $mod->details($one, $chrome, "the row's own word");

echo rowLabels($details), "\n";
echo cell($details, "Folder")["link"]["href"], " ", cell($details, "Folder")["link"]["icon"], " ", cell($details, "Folder")["link"]["text"], "\n";
echo cell($details, "Job")["link"]["href"], " [", implode(" ", cell($details, "Arguments")["code"]), "] ", cell($details, "Arguments")["after"], "\n";
echo cell($details, "By")["who"]["user"]["name"], " ", var_export(cell($details, "By")["who"]["scheduled"], true), " ", var_export(cell($details, "By")["who"]["forced"], true), "\n";
echo cell($details, "Duration")["text"], " ", cell($details, "Result")["state"]["state"], " [", cell($details, "Result")["after"], "] ", cell($details, "Terminal")["text"], " ", cell($details, "ID")["text"], "\n";

// A run in progress has finished at no time, which the table says in words,
// and the job of a directory inside the folder is a page of its own.
$inside = $mod->details($live, $chrome, "");

echo rowLabels($inside), "\n";
echo cell($inside, "Directory")["link"]["href"], " ", cell($inside, "Directory")["link"]["text"], " ", cell($inside, "Directory")["after"], "\n";
echo cell($inside, "Finished")["faint"], " ", cell($inside, "Job")["link"]["href"], "\n";

// A run the scheduler started out of hours says so, and the row of the reason
// carries the badge of it.
$broke = $mod->details($runs->get("fx-rp-4"), $chrome, "");

echo rowLabels($broke), "\n";
echo cell($broke, "Reason")["text"], " ", cell($broke, "Reason")["badge"]["tone"], " ", cell($broke, "Reason")["badge"]["text"], "\n";
echo var_export(cell($broke, "By")["who"]["scheduled"], true), " ", var_export(cell($broke, "By")["who"]["forced"], true), " ", cell($broke, "Result")["state"]["code"], "\n";

// A run of Claude's has no job page, so the job is a word and no link.
$claude = $runs->get("fx-rp-5");

echo "[", cell($mod->details($claude, $chrome, ""), "Job")["link"]["href"], "] ", cell($mod->details($claude, $chrome, ""), "Job")["text"], "\n";

// The files the run left, newest first, each read from the API at the root.
$files = $mod->artifacts($chrome, $one);

echo $files["block"]["title"], " [", $files["block"]["note"], "] [", $files["block"]["empty"], "]\n";
echo $files["rows"][0]["name"], " ", $files["rows"][0]["href"], " ", $files["rows"][0]["size"], " ", $files["rows"][0]["type"], "\n";
echo $files["rows"][1]["name"], " ", $files["rows"][1]["size"], "\n";
echo labels($files["rows"][0]["menu"]), "\n";

// A run that left none says what a job does to keep a file.
$none = $mod->artifacts($chrome, $runs->get("fx-rp-6"));

echo count($none["rows"]), " [", $none["block"]["empty"], "]\n";

// The prompt box of the run: the threads asked about it, and the field to ask,
// which a reader with a token for Claude gets once the run is no longer
// running.
$box = Task\Prompt::panel($chrome, $one, $seen);

echo $box["title"], " [", $box["note"], "] ", var_export($box["hidden"], true), " ", var_export($box["watch"], true), " ", $box["api"], " ", $box["place"], "\n";
echo $box["fields"][0]["name"], " ", $box["fields"][0]["value"], " [", $box["editor"]["placeholder"], "]\n";
echo "[", $box["lead"], "] [", $box["tail"], "]\n";
echo count($box["threads"]), " ", $box["threads"][0]["id"], " ", count($box["threads"][0]["replies"]), "\n";

// The answer names a run of Claude's the reader sees, so the card links it.
echo $box["threads"][0]["replies"][0]["run"], "\n";

// A running run's box waits out of sight until a thread about it comes, and
// offers no field: the summary is asked once the run ends.
$waiting = Task\Prompt::panel($chrome, $live, $seen);

echo var_export($waiting["hidden"], true), " ", var_export($waiting["watch"], true), " ", count($waiting["threads"]), "\n";

// One of Claude's own runs has no box: the run is the answer. A run that ended
// with no thread about it and no token for Claude has none either.
echo var_export(Task\Prompt::panel($chrome, $claude, $seen), true), "\n";

Task\Host::$claude = false;

echo var_export(Task\Prompt::panel($chrome, $runs->get("fx-rp-6"), $seen), true), "\n";

Task\Host::$claude = true;

// A run of a batch asks about the batch, whose threads hang on the oldest run
// of it, so every page of the batch lists one set. The rule is the run
// table's, because a summary is written about the same run.
echo $runs->root($runs->get("fx-rp-b2")), " ", $runs->root($one), "\n";

// Stopping and killing a run is running its job, which the host resolves from
// the folder, the directory the run ran in and the job: the run of a directory
// inside the folder is the job that directory's own Taskfile declares, and a
// job the catalogue resolves none of is nobody's to stop, which is what a
// deleted job and a prompt of Claude's both are.
echo var_export($mod->can($chrome, $one), true), " ", var_export($mod->can($chrome, $live), true), "\n";
echo var_export($mod->can($chrome, $runs->get("fx-rp-6")), true), " ", var_export($mod->can($chrome, $claude), true), "\n";

// The two gates that are not an allow list: a terminal needs the shell
// permission and a job on a host needs the hosts one, whatever either allow
// list says.
$shell = new Task\Model\Run(array("id" => "fx-rp-sh", "folder" => "fx-rp", "dir" => "", "job" => "@shell", "user_name" => "fx-rp-ann", "state" => "running"));
$onhost = new Task\Model\Run(array("id" => "fx-rp-ho", "folder" => "fx-rp", "dir" => "", "job" => "@host:db1", "user_name" => "fx-rp-ann", "state" => "running"));
$term = Task\Fixture::job("@shell", "shell", array(), "shell");
$remote = Task\Fixture::job("@host:db1", "host db1", array(), "host");

Task\Fixture::resolves("fx-rp", "", $term);
Task\Fixture::resolves("fx-rp", "", $remote);

echo var_export($mod->can($chrome, $shell), true), " ", var_export($mod->can($chrome, $onhost), true), "\n";

// A job whose allow list names somebody else is theirs to stop and not the
// reader's, and so is every job of a folder whose own list does. Both lists
// come with the job, because the page holds neither the folder nor the
// directory inside it. The reader here is a runner: an admin runs everything
// they see, whatever either list says.
$runner = new Task\Policy(array("run" => Task\Policy::ALLOW));
$bob = new Task\Model\User(array("id" => "fx-rp-u2", "name" => "fx-rp-bob", "role" => "runner"));
$plain = new Task\Chrome;
$plain->user = $bob;
$plain->access = new Task\Access($bob, $runner);
$mine = Task\Fixture::job("build", "build", array("fx-rp-bob"));

Task\Fixture::resolves("fx-rp", "", $mine);

echo var_export($mod->can($plain, $one), true), " ", var_export($mod->can($chrome, $one), true), "\n";

$theirs = Task\Fixture::job("build", "build", array("fx-rp-ann"));

Task\Fixture::resolves("fx-rp", "", $theirs);

echo var_export($mod->can($plain, $one), true), " ";

$shut = Task\Fixture::job("build", "build", array("fx-rp-bob"));

Task\Fixture::resolves("fx-rp", "", $shut, array("fx-rp-ann"));

echo var_export($mod->can($plain, $one), true), "\n";

// Summarize is offered on a run that ended, is none of Claude's own, of a job
// the catalogue still resolves, with no summary read yet: the run that has one
// links the Report instead, and a run of a job that is gone summarizes
// nothing.
echo var_export($mod->summarizes($chrome, $one, $report, true), true), " ", var_export($mod->summarizes($chrome, $one, null, true), true), "\n";
echo var_export($mod->summarizes($chrome, $live, null, false), true), " ", var_export($mod->summarizes($chrome, $claude, null, true), true), "\n";
echo var_export($mod->summarizes($chrome, $runs->get("fx-rp-6"), null, true), true), "\n";

// A summary under way is no report: the gate is the summaries themselves.
// Summarize goes, the faint sentence stands in its place, and a reload asks
// nothing a second time. The reports are read under prompt/view alone.
$waiting = $runs->get("fx-rp-7");

echo var_export($mod->pending($chrome, $waiting), true), " ", var_export($mod->summarizes($chrome, $waiting, null, true), true), "\n";
echo var_export($mod->pending($chrome, $one), true), " ", var_export($mod->pending($plain, $waiting), true), "\n";

// The sentence under the heading, which is written here and not in the
// template: who started the run and how, and then what the reader may do about
// it where it stands. A run with no start reads as no time at all.
echo mod_run::lead($runs->get("fx-rp-3"), false, true), "\n";
echo mod_run::lead($runs->get("fx-rp-3"), false, false), "\n";
echo mod_run::lead($live, false, true), "\n";

// How long ago a run started counts from this moment, so a row of a fixed hour
// is read with that part of the sentence masked and the rest as it stands.
echo str_replace(Task\Clock::ago(at("10:04:00.000")), "<ago>", mod_run::lead($runs->get("fx-rp-4"), true, false)), "\n";
echo str_replace(Task\Clock::ago(at("10:01:00.000")), "<ago>", mod_run::lead($one, false, false)), "\n";

// The whole scope the template reads, which is what the page hands the view.
// The run that ended draws its recording, its details, its files and the
// threads asked about it, and no terminal.
$scope = $mod->scope($chrome, $runs->get("fx-rp-1"), $seen);

echo implode(",", array_keys($scope)), "\n";
echo implode(",", array_keys($scope["head"])), "\n";
echo $scope["head"]["title"], " [", $scope["head"]["badge"], "] ", $scope["head"]["tone"], " [", $scope["head"]["note"], "]\n";
echo labels($scope["head"]["buttons"]), "\n";
echo var_export($scope["terminal"], true), "\n";
echo implode(",", array_keys($scope["output"])), "\n";
echo var_export($scope["output"]["expired"], true), " ", var_export($scope["output"]["recording"], true), " ", $scope["output"]["id"], " ", $scope["output"]["cols"], " ", $scope["output"]["rows"], "\n";
echo count($scope["kv"]["rows"]), " ", count($scope["artifacts"]["rows"]), " ", $scope["prompt"]["title"], "\n";

// The scope of a run the runner holds: the state, the title and what it
// reports are the session's, the terminal is the screen, and a run with no
// recording draws no Output block.
$held = $mod->scope($chrome, $runs->get("fx-rp-2"), $seen);

echo $held["head"]["badge"], " ", $held["head"]["tone"], " [", $held["head"]["note"], "]\n";
echo labels($held["head"]["buttons"]), "\n";
echo $held["terminal"]["id"], " ", $held["terminal"]["watching"], " ", $held["terminal"]["frame"], "\n";
echo var_export($held["output"]["expired"], true), " ", var_export($held["output"]["recording"], true), "\n";
echo cell($held["kv"]["rows"], "Result")["after"], "\n";

// The scope of a run whose recording the sweep removed: the Output block says
// so and names the settings that decided it, and there is nothing to replay.
$sweep = $mod->scope($chrome, $runs->get("fx-rp-4"), $seen);

echo $sweep["head"]["badge"], " ", $sweep["head"]["tone"], "\n";
echo labels($sweep["head"]["buttons"]), "\n";
echo var_export($sweep["output"]["expired"], true), " ", var_export($sweep["output"]["recording"], true), "\n";
echo $sweep["output"]["text"], "\n";

// The scope of a run whose summary Claude is still writing, the page a reload
// draws: the head says summarizing where Summarize stood, and offers no button
// to ask the same question again.
$writing = $mod->scope($chrome, $runs->get("fx-rp-7"), $seen);

echo labels($writing["head"]["buttons"]), " [", faints($writing["head"]["buttons"]), "]\n";
---
runs Runs list
true
[the row's own word] ok []
[the row's own word] running [pulling the image]
[waits for a free slot, number 1 in line] queued
View,Run again,Report,|> false
data-attach="fx-rp-1" [] false true
 [/runs/fx-rp-1/report] false false
data-replay="fx-rp-1" data-replay-title="fx-rp / build" data-replay-who="fx-rp-ann" title="Replay the recording" aria-label="Replay the recording" [] false false
data-run data-folder="fx-rp" data-dir="" data-job="build" data-title="fx-rp / build" data-run-args="[&quot;--verbose&quot;,&quot;./pkg&quot;]" [] false false
View,Run again,Summarize
data-run data-folder="fx-rp" data-dir="" data-job="ghost" data-title="fx-rp / ghost" [] false false
data-summarize-run="fx-rp-6" data-question="Summarize this run." title="Claude summarizes the run&#039;s output, in this window" [] false false
View,Run again, [summarizing]
[]
Attach,Stop,Kill
Attach
View,Run again
fx-rp-2 fx-rp/docker / deploy 2 <svg></svg>
NULL
Folder,Job,Arguments,By,Started,Finished,Duration,Result,Terminal,ID
/f/internal/fx-rp folder fx-rp
/f/internal/fx-rp/build [--verbose ./pkg] after --, one argument each
fx-rp-ann false false
2.5s ok [the row's own word] 120×40 fx-rp-1
Folder,Directory,Job,By,Started,Finished,Duration,Result,Terminal,ID
/f/internal/fx-rp/docker docker/ the job ran here
in progress /f/internal/fx-rp/docker/deploy
Folder,Job,By,Started,Finished,Duration,Result,Reason,Terminal,ID
the window of the job is closed bad out of hours
true true 2
[] @claude-prompt
Artifacts [files left under $TASK_UI_ARTIFACTS] []
report.html /api/artifacts/fx-rp-a1 2.0 KB text/html
build.log 300 B
Download,Delete
0 [None: a job keeps files by writing them to the directory in $TASK_UI_ARTIFACTS.]
Prompt [ask Claude Code about this run and its output, in fx-rp/] false false /api/prompts fx-rp
run fx-rp-1 [why did this run end the way it did]
[Runs @claude-prompt as ] [ in fx-rp/, with this run's output. If Claude makes a fix, it commits to main. Claude replies under your question.]
1 fx-rp-n1 1
fx-rp-5
true true 0
NULL
NULL
fx-rp-b1 fx-rp-1
true true
false false
true true
true true
false false
false true
false false
false
true false
false false
Started by fx-rp-ann -. It waits, queued, for the run of its job before it to end or for a free slot: attach to follow it once it starts, or stop it to take it out of the queue.
Started by fx-rp-ann -. It waits, queued, for the run of its job before it to end or for a free slot: attach to follow it once it starts.
Started by fx-rp-ann just now. It's still running: attach to follow it, or stop it.
Started by fx-rp-ann <ago> by the scheduler, with Run force. The run has finished; its recording expired.
Started by fx-rp-ann <ago>. The run has finished; its full output is below.
head,terminal,output,prompt,kv,artifacts
title,icon,badge,tone,tags,note,lead,leads,faint,live,run,actions,buttons,show
fx-rp / build [ok] ok []
View,Run again,Report,|>
NULL
expired,recording,text,id,cols,rows
false true fx-rp-1 120 40
10 2 Prompt
running signal [pulling the image]
Attach,Stop,Kill
fx-rp-2 2 <svg></svg>
false false
the row's own word
failed 2 bad
View,Run again,Summarize
true false
The recording of this run expired under the Data retention settings: its output is gone. The run's details stay.
View,Run again, [summarizing]
