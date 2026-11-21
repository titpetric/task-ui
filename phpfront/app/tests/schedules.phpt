name: schedules
description: The schedules page reads what a schedule row means, where its folder's page is, and the folders a workflow's steps reach.
---
<?php

require "modules/schedules/mod_schedules.php";

$mod = new mod_schedules;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo implode(",", $mod->tables()), "\n";
echo var_export($mod instanceof Module, true), "\n";

// A row says what it starts and whether it runs. The columns are integers on
// one driver and strings on another, which is why the readers coerce.
$job = new Task\Model\Schedule([
	"id" => "s1",
	"folder" => "infra",
	"job" => "plan",
	"expression" => "@daily",
	"enabled" => 1,
	"triage" => 0,
]);
$flow = new Task\Model\Schedule([
	"id" => "s2",
	"workflow_id" => "w1",
	"expression" => "0 3 * * *",
	"enabled" => "0",
	"triage" => "1",
]);
$own = new Task\Model\Schedule(["id" => "s3", "folder" => "internal", "job" => "@cleanup"]);

foreach ([$job, $flow, $own] as $schedule) {
	echo implode(" ", [
		var_export(Task\Schedule::enabled($schedule), true),
		var_export(Task\Schedule::triages($schedule), true),
		var_export(Task\Schedule::isWorkflow($schedule), true),
		var_export(Task\Schedule::instance($schedule), true),
	]), "\n";
}

// What a schedule is for, which its Delete confirm asks about: the folder and
// the job, the workflow's name, or the workflow's ID where it has no name.
echo $mod->what($job, null), "\n";
echo $mod->what($flow, ["name" => "Nightly", "runAs" => ""]), "\n";
echo $mod->what($flow, ["name" => "", "runAs" => ""]), "\n";

// The arguments as the line the form takes back: the words, quoted where a
// word needs it, and the command unchanged for a job of a host.
echo "[", Task\Schedule::argLine("plan", ""), "]\n";
echo "[", Task\Schedule::argLine("plan", "[]"), "]\n";
echo "[", Task\Schedule::argLine("plan", '["--full","/var/lib"]'), "]\n";
echo "[", Task\Schedule::argLine("plan", '["one two","",""]'), "]\n";
echo "[", Task\Schedule::argLine("plan", '["it'."'".'s"]'), "]\n";
echo "[", Task\Schedule::argLine("@host:web1", '["df","-h","/"]'), "]\n";

// Where a folder's page is is Task\Route's, and tests/workspace.phpt says it.
// The named workspaces are what a row's links are built against.
$workspaces = ["malta"];

// A step's kind comes from its own column, and from its fields for a row
// written before the column existed.
$plan = new Task\Model\WorkflowStep(["workflow_id" => "w1", "folder" => "infra", "job" => "plan"]);
$ref = new Task\Model\WorkflowStep(["workflow_id" => "w1", "workflow_ref" => "w2"]);
$ask = new Task\Model\WorkflowStep(["workflow_id" => "w1", "folder" => "tools", "prompt" => "ask"]);
$dump = new Task\Model\WorkflowStep(["workflow_id" => "w2", "folder" => "backups", "job" => "dump"]);
$named = new Task\Model\WorkflowStep(["workflow_id" => "w2", "workflow_ref" => "w1", "kind" => "workflow"]);
$steps = [
	"w1" => [$plan, $ref, $ask],
	"w2" => [$dump, $named],
];

$kinds = [];

foreach ($steps["w1"] as $step) {
	$kinds[] = Task\Schedules\Lookup::kind($step);
}

echo implode(" ", $kinds), "\n";

// The folders a workflow reaches, the steps of a workflow a step runs
// included, each workflow walked once so a loop ends.
echo implode(",", Task\Schedules\Lookup::folders($steps, "w1")), "\n";
echo implode(",", Task\Schedules\Lookup::folders($steps, "w2")), "\n";
echo implode(",", Task\Schedules\Lookup::folders($steps, "gone")), "\n";

// The block's line with no schedules says where to add one only where the
// reader may.
echo "[", $mod->said([], true), "]\n";
echo "[", $mod->said([], false), "]\n";
echo "[", $mod->said(["a row"], true), "]\n";

// The page of what a schedule starts. The paths are written without the
// prefix, as every path a template writes is.
echo $mod->jobHref($flow, $workspaces), "\n";
$deploy = new Task\Model\Schedule(["folder" => "malta/deploy", "job" => "plan"]);

echo $mod->jobHref($deploy, $workspaces), "\n";
echo "[", $mod->jobHref($own, $workspaces), "]\n";

// The folder a job names: the one the query named, the one the job was found
// in, or the root for a job of a host, which belongs to no folder's catalogue.
$offered = [["id" => "dump", "folder" => "fx-sched/hidden"]];

echo Task\Schedules\Catalogue::picked("", "dump", $offered), "\n";
echo Task\Schedules\Catalogue::picked("", "@host:web1", []), "\n";
echo Task\Schedules\Catalogue::picked("fx-sched/seen", "", []), "\n";
echo "[", Task\Schedules\Catalogue::picked("", "", []), "]\n";

// A row's menu: Edit, where the page drew the form that saves a schedule,
// then Delete, which asks about what the schedule is for. Whether it runs is
// a column of the table, saved with it.
function labels($menu) {
	$out = [];

	foreach ($menu["groups"] as $group) {
		foreach ($group as $item) {
			$out[] = $item["label"];
		}
	}

	return implode(",", $out);
}

echo labels($mod->menu($job, "infra / plan")), "\n";
echo labels($mod->menu($job, "infra / plan", false)), "\n";

// The rest of the page reads the chrome: who is reading decides which folders
// the form offers and which schedules the table lists. The reader is resolved
// from the user table, as a request resolves one.
$db = Task\DB::writer();

$db->replace("user", ["id" => "fx-sch-u1", "name" => "fx-sched-ann", "role" => "admin", "admin" => 1]);
$db->replace("user", ["id" => "fx-sch-u2", "name" => "fx-sched-bob", "role" => "runner", "folders" => '["fx-sched/seen"]']);

// A job the folder's own allow list keeps from the reader, and a job that
// names somebody else: neither is offered, and the folder that holds nothing
// else is offered no more than the folder that declares nothing at all.
$plan = Task\Fixture::job("plan");
$apply = Task\Fixture::job("apply", "", ["fx-sched-bob"]);
$dump = Task\Fixture::job("dump");
$lock = Task\Fixture::job("task:lock", "lock");

Task\Host::$folders = [
	Task\Fixture::folder("fx-sched/seen", [$plan, $apply]),
	Task\Fixture::folder("fx-sched/hidden", [$dump]),
	Task\Fixture::folder("fx-sched/none", []),
	Task\Fixture::folder("fx-sched/shut", [$lock], ["fx-sched-bob"]),
];
Task\Host::$whoami = "fx-sched-ann";
$_REQUEST["folder"] = "fx-sched/hidden";

$chrome = new Task\Chrome;

// The Add schedule form offers every folder that declares a job the reader may
// run, with the one a query named selected, and every job under its folder. A
// folder that declares none is in neither list. An admin may run everything
// they see, so every folder here holds one.
//
// A schedule is written with the job's ID and the picker shows its name, which
// differ on a folder that follows more than one convention.
$catalogue = $mod->catalogue($chrome);

// offers names what a catalogue holds: each folder and whether it is selected,
// then each job as its folder, its ID and the name the picker shows.
function offers($catalogue) {
	$out = [];

	foreach ($catalogue["folders"] as $folder) {
		$out[] = $folder["id"] . ($folder["selected"] ? "*" : "");
	}

	$names = [];

	foreach ($catalogue["jobs"] as $offered) {
		$names[] = $offered["folder"] . "/" . $offered["id"] . "=" . $offered["name"];
	}

	return implode(",", $out) . " | " . implode(",", $names);
}

echo offers($catalogue), "\n";

// The same catalogue to a runner, who is gated on both allow lists: a job that
// names somebody else is not offered, nor any job of a folder whose own list
// does, and a folder left with nothing is offered no more than one that
// declares nothing at all.
$cat = new Task\Model\User(["id" => "fx-sch-u3", "name" => "fx-sched-cat", "role" => "runner", "folders" => '["fx-sched/seen","fx-sched/shut"]']);
$runs = new Task\Policy(["run" => Task\Policy::ALLOW]);
$runner = new Task\Chrome;
$runner->user = $cat;
$runner->access = new Task\Access($cat, $runs);

echo offers($mod->catalogue($runner)), "\n";

// The Run as picker shows for an admin and for nobody else. Only the names are
// read here: which keys a user's runs get is Task\Key's, and tests/key.phpt
// says it, over rows this fixture does not share.
$lookup = new Task\Schedules\Lookup;
$users = $chrome->runAsUsers();
$offered = [];

foreach ($users as $option) {
	if (strpos($option["name"], "fx-sched-") === 0) {
		$offered[] = $option["name"] . "/" . $option["role"];
	}
}

echo implode(",", $offered), "\n";

// A workflow is the reader's to see where they see every folder its steps
// name, the steps of a workflow a step runs included.
$flows = [
	"fx-sch-w1" => ["name" => "fx Seen", "runAs" => "", "folders" => ["fx-sched/seen"]],
	"fx-sch-w2" => ["name" => "fx Mixed", "runAs" => "fx-sched-ann", "folders" => ["fx-sched/seen", "fx-sched/hidden"]],
];

echo implode(",", array_keys($mod->visible($chrome, $flows))), "\n";

// One schedule as the table draws it: a job's, which names its folder's page
// and the job's, and a workflow's, which names neither and whose user an admin
// may not pick, because the workflow fixes it.
$daily = new Task\Model\Schedule([
	"id" => "fx-sch-1",
	"folder" => "malta/deploy",
	"job" => "plan",
	"user_name" => "fx-sched-bob",
	"expression" => "@daily",
	"enabled" => 1,
	"triage" => 1,
	"args" => '["--full"]',
	"next_run_at" => "2026-11-03 03:00:00",
]);
$jobRow = $mod->row($chrome, $daily, null, $workspaces, $users);

echo var_export($jobRow["workflow"], true), " ", $jobRow["folder"], " ", $jobRow["folderHref"], " ", $jobRow["job"], " ", $jobRow["jobHref"], "\n";
echo $jobRow["expression"], " [", $jobRow["args"], "] [", $jobRow["last"], "] ", $jobRow["api"], "\n";
echo count($jobRow["next"]), " ", strlen($jobRow["next"][0]), " ", var_export($jobRow["on"], true), " ", var_export($jobRow["triage"], true), " ", var_export($jobRow["pick"], true), "\n";
echo $jobRow["user"]["name"], " ", $jobRow["user"]["role"], " ", $jobRow["user"]["initials"], "\n";
echo labels($jobRow["menu"]), "\n";

$nightly = new Task\Model\Schedule([
	"id" => "fx-sch-2",
	"workflow_id" => "fx-sch-w2",
	"expression" => "0 3 * * *",
	"enabled" => 0,
]);
$flowRow = $mod->row($chrome, $nightly, $flows["fx-sch-w2"], $workspaces, $users);

echo var_export($flowRow["workflow"], true), " [", $flowRow["folder"], "] ", $flowRow["job"], " ", $flowRow["jobHref"], "\n";
echo count($flowRow["next"]), " ", var_export($flowRow["on"], true), " ", var_export($flowRow["pick"], true), " ", $flowRow["user"]["name"], "\n";
echo labels($flowRow["menu"]), "\n";

// The table lists a workflow's schedule where the reader sees the workflow and
// a job's where they see its folder. bob sees one folder, so the schedule of
// the other is not his to read, nor is the workflow that reaches it.
$db->replace("schedule", ["id" => "fx-sch-s1", "folder" => "fx-sched/seen", "job" => "plan", "user_name" => "fx-sched-bob", "expression" => "@daily"]);
$db->replace("schedule", ["id" => "fx-sch-s2", "folder" => "fx-sched/hidden", "job" => "dump", "user_name" => "fx-sched-ann", "expression" => "@weekly"]);

// listed is the schedules of this fixture a reading answered, by the path of
// the API its row calls. The table is the whole instance's, so the rows of
// another fixture are read past.
function listed($rows) {
	$out = [];

	foreach ($rows as $row) {
		if (strpos($row["api"], "/api/schedules/fx-sch-") === 0) {
			$out[] = substr($row["api"], strlen("/api/schedules/"));
		}
	}

	return implode(",", $out);
}

echo listed($mod->rows($chrome, $lookup, $users)), "\n";

Task\Host::$whoami = "fx-sched-bob";
$his = new Task\Chrome;

echo listed($mod->rows($his, $lookup, $his->runAsUsers())), "\n";
echo implode(",", array_keys($mod->visible($his, $flows))), "\n";
echo count($his->runAsUsers()), "\n";
---
schedules Scheduled clock
schedule,workflow,workflow_step,workspace,user,key_info,key_role,key_user
true
true false false false
false true true false
false false false true
infra / plan
Nightly
w1
[]
[]
[--full /var/lib]
['one two' '' '']
['it'\''s']
[df -h /]
job workflow prompt
infra,tools,backups
backups,infra,tools

[No schedules yet. Add one to run a job or a workflow on a schedule.]
[No schedules yet.]
[]
/workflows/w1
/f/malta/deploy/plan
[]
fx-sched/hidden
internal
fx-sched/seen
[]
Edit,Delete
Delete
fx-sched/seen,fx-sched/hidden*,fx-sched/shut | fx-sched/seen/plan=plan,fx-sched/seen/apply=apply,fx-sched/hidden/dump=dump,fx-sched/shut/task:lock=lock
fx-sched/seen | fx-sched/seen/plan=plan
fx-sched-ann/admin,fx-sched-bob/runner
fx-sch-w1,fx-sch-w2
false malta/deploy /f/malta/deploy plan /f/malta/deploy/plan
@daily [--full] [] /api/schedules/fx-sch-1
2 10 true true true
fx-sched-bob runner FS
Edit,Delete
true [] fx Mixed /workflows/fx-sch-w2
0 false false fx-sched-ann
Edit,Delete
fx-sch-s2,fx-sch-s1
fx-sch-s1
fx-sch-w1
0
