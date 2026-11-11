name: tooling
description: Task\Tooling is the panel of what atkins recorded - a chart per window over the samples, the slots and their sums, the legend's colours kept across the windows and the scopes, the newest commands, and the daily rollups where an atkins lists no samples.
---
<?php

require "ext/autoload.php";

// A zone east of UTC, so a command early in the morning is on the day before
// in UTC and on its own day here. The clock is a parameter of the panel, so
// every slot below is stated and none of it moves with the machine's own time.
$zone = Time\Location::fixed("CEST", 7200);

// moment is a wall clock reading of October 2026 in that zone.
function moment($zone, $day, $hour, $minute) {
	return DateTime::parse_in_location("2006-01-02 15:04", sprintf("2026-10-%02d %02d:%02d", $day, $hour, $minute), $zone);
}

// sample is one command of the report, as Task\Host::tooling() answers it: the
// moment both as the text the list prints and as the number the charts bucket
// by, which is what the report carries them both for.
function sample($zone, $day, $hour, $minute, $command, $tool, $job, $ms, $exit) {
	$at = moment($zone, $day, $hour, $minute);

	return array(
		"at" => $at->in(Time\Location::utc())->format(Task\Clock::STORED),
		"at_unix" => $at->unix(),
		"command" => $command,
		"tool" => $tool,
		"job" => $job,
		"duration_ms" => $ms,
		"exit_code" => $exit,
		"failed" => false,
	);
}

$now = moment($zone, 19, 22, 0);

// The samples are oldest first, which is the order the report lists them in.
// The last two are not on a chart: one whose text the charts never read is
// placed by its number alone, and one the report gives no moment for at all.
$samples = array(
	sample($zone, 17, 12, 0, "npm ci", "npm", "", 4000, 0),
	sample($zone, 19, 7, 0, "go test ./...", "go", "test", 2500, 0),
	sample($zone, 19, 7, 30, "go build ./...", "go", "build", 500, 1),
	sample($zone, 19, 21, 0, "go test ./...", "go", "test", 1000, 0),
	sample($zone, 19, 21, 0, "docker build", "docker", "", 1000, 0),
);

// A chart buckets by the number and nothing else: the text of this one is not
// a stamp, and the slot it lands in is the one its number names.
$samples[3]["at"] = "whenever";
$samples[4]["at"] = "";
$samples[4]["at_unix"] = 0;

// Two sections of one report: the project's samples, and a tree whose atkins
// recorded daily rollups alone. A command keeps its colour across both, so the
// second section's colours carry on where the first left off.
//
// A command key reports its last run, which Task\Clock::ago words against the
// process clock; the report here names none, because a fixture that stated an
// answer would state a different one tomorrow. clock.phpt holds that reading.
$report = array(array(
	"name" => "task-ui",
	"repository" => "titpetric/task-ui",
	"database" => ".atkins/atkins.db",
	"days" => 30,
	"from" => "2026-09-20",
	"to" => "2026-10-19",
	"runs" => 5,
	"total_ms" => 9000,
	"samples" => $samples,
	"usage" => array(),
	"command_usage" => array(),
	"commands" => array(array(
		"command" => "npm ci",
		"tool" => "npm",
		"runs" => 1,
		"total_ms" => 4000,
		"average_ms" => 4000,
		"max_ms" => 4000,
		"failures" => 0,
		"last_run_at" => "",
	), array(
		"command" => "go test ./...",
		"tool" => "go",
		"runs" => 2,
		"total_ms" => 3500,
		"average_ms" => 1750,
		"max_ms" => 2500,
		"failures" => 1,
		"last_run_at" => "",
	)),
), array(
	"name" => "deploy",
	"repository" => "malta/deploy",
	"database" => ".atkins/deploy.db",
	"days" => 7,
	"from" => "2026-10-13",
	"to" => "2026-10-19",
	"runs" => 4,
	"total_ms" => 6000,
	"samples" => array(),
	"command_usage" => array(),
	"usage" => array(
		array("date" => "2026-10-19", "command" => "", "tool" => "go", "runs" => 2, "total_ms" => 3000),
		array("date" => "2026-10-19", "command" => "", "tool" => "npm", "runs" => 1, "total_ms" => 1000),
		array("date" => "2026-10-15", "command" => "", "tool" => "go", "runs" => 1, "total_ms" => 2000),
		array("date" => "2026-10-01", "command" => "", "tool" => "go", "runs" => 9, "total_ms" => 9000),
	),
	"commands" => array(),
));

$panel = new Task\Tooling($report, $now);
$data = $panel->data();
$scope = $data["scopes"][0];

// The section names its database and sums the whole report, which is every
// command atkins recorded and not only the window's.
echo count($data["scopes"]), " | ", $scope["database"], " | ", $scope["name"], " | ", $scope["repository"], "\n";
echo $scope["sum"], " | ", $scope["toggle"], "\n";

// The three windows the samples draw, the activity charts' own spans. The
// legend is each window's own, so the month lists a command the day does not.
echo count($scope["charts"]), "\n";

foreach ($scope["charts"] as $chart) {
	echo $chart["class"], " | ", $chart["title"], " | ", $chart["stats"], " | ", $chart["bars"]["style"], "\n";
}

// counted names the slots of a chart that anything ran in, by their index, so
// a window is read as what it holds and not as what it draws.
function counted($chart) {
	$out = array();

	foreach ($chart["bars"]["slots"] as $i => $slot) {
		if (count($slot["segments"]) > 0) {
			$out[] = $i . "=" . $slot["title"];
		}
	}

	return implode(" · ", $out);
}

// The days: the 17th and the 19th, the 19th three commands of the 4 seconds
// the 17th took in one. A slot nothing ran in says so and draws no bar.
echo counted($scope["charts"][0]), "\n";
echo $scope["charts"][0]["bars"]["slots"][28]["title"], " | ", count($scope["charts"][0]["bars"]["slots"][28]["segments"]), "\n";

// The hours start at 23:00 on the 18th, so 07:00 is the ninth of them and
// 21:00 the twenty third. The 17th is outside this window.
echo counted($scope["charts"][1]), "\n";

// The week is three hour slots and its last ends at midnight: 06:00 on the
// 19th is the fifty first, 21:00 the fifty sixth, noon of the 17th the thirty
// seventh.
echo counted($scope["charts"][2]), "\n";

// Every bar is against the tallest one command took in a slot, so the group
// reads as the commands in it: the slowest of the day chart is the npm ci of
// the 17th, and the 19th's go test is seven eighths of it. The command that
// did not run that day holds its place in the group and says nothing.
$slot = $scope["charts"][0]["bars"]["slots"][29];

echo $slot["class"], "\n";

foreach ($slot["segments"] as $segment) {
	echo "[", $segment["class"], "] [", $segment["style"], "] [", $segment["title"], "]\n";
}

// The legend is the window's top commands by their total, each in its colour,
// and a command keeps that colour in every window of the panel.
function named($chart) {
	$out = array();

	foreach ($chart["legend"] as $entry) {
		$out[] = $entry["name"] . "=" . $entry["color"] . "/" . $entry["total"] . "/" . $entry["runs"];
	}

	return implode(" ", $out);
}

echo named($scope["charts"][0]), "\n";
echo named($scope["charts"][1]), "\n";
echo named($scope["charts"][2]), "\n";

// The newest commands, newest first, at their clock time in the zone of the
// panel's clock: one the report gives no moment for and one whose text is not
// a stamp name none, and a command on no chart draws in the colour of the
// rest. A stamp reads in the server's own zone, so its shape is what is
// stated and not the reading, as in clock.phpt.
foreach ($scope["latest"] as $line) {
	echo strlen($line["stamp"]), " | ", $line["when"], " | ", $line["color"], " | ", $line["title"], " | ", $line["took"], " | ", var_export($line["failed"], true), "\n";
}

// The table behind the toggle: the command keys as atkins totals them, and a
// key it reports no last run for reads as none.
foreach ($scope["commands"] as $line) {
	echo $line["command"], " | ", $line["tool"], " | ", $line["runs"], " | ", $line["total"], " | ", $line["average"], " | ", $line["max"], " | ", $line["failures"], " | ", $line["last"], "\n";
}

// An atkins that lists no samples leaves the charts to the daily rollups,
// which are days wide, and names no newest commands. A week of history draws
// the week and not the month: a window the report is too short to fill is no
// chart at all.
$rolled = $data["scopes"][1];
$chart = $rolled["charts"][0];

echo count($rolled["charts"]), " ", count($rolled["latest"]), " | ", $chart["class"], " | ", $chart["title"], " | ", $chart["stats"], "\n";
echo $chart["bars"]["class"], " | ", $chart["bars"]["label"], " | ", $chart["bars"]["style"], "\n";
echo counted($chart), "\n";

// The rollups report no command series, so the bars are of the tools, and the
// two of them take the colours the first section's commands left.
echo named($chart), "\n";
echo count($chart["ticks"]), " ", $chart["ticks"][0]["label"], " ", $chart["ticks"][0]["at"], " ", $chart["ticks"][6]["label"], " ", $chart["ticks"][6]["at"], "\n";

// A tree nobody has run a job in through atkins draws no panel at all, which
// is the usual answer.
$empty = new Task\Tooling(array(), $now);

echo count($panel->data()), " ", count($empty->data()), "\n";

// stamps reads the number the report carries beside the text, by the index the
// sample holds: nothing parses a stamp, so a chart of five thousand samples
// crosses into the host for none of them.
echo implode(",", Task\Tooling::stamps($samples)), "\n";
---
2 | .atkins/atkins.db | task-ui | titpetric/task-ui
5 commands in 30 days · 9.0s of tooling | 2 commands by the time they have taken
3
tooling-chart | Last 30 days | 4 commands · 8.0s · busiest day 4.0s | --n: 30; --k: 3
tooling-chart | Last 24 hours | 3 commands · 4.0s · busiest hour 3.0s | --n: 24; --k: 2
tooling-chart wide | Last 7 days | 4 commands · 8.0s · busiest 3 hours 4.0s | --n: 56; --k: 3
27=Sat 2026-10-17: 1 command in 4.0s · 29=Mon 2026-10-19: 3 commands in 4.0s
Sun 2026-10-18: nothing ran | 0
8=Mon 07:00–08:00: 2 commands in 3.0s · 22=Mon 21:00–22:00: 1 command in 1.0s
36=Sat 12:00–15:00: 1 command in 4.0s · 50=Mon 06:00–09:00: 2 commands in 3.0s · 55=Mon 21:00–00:00: 1 command in 1.0s
tip tip-end
[] [] []
[tip tip-end s1] [height: 87.5%] [Mon 2026-10-19 · go test ./...: 3.5s, 2 runs]
[tip tip-end s2] [height: 12.5%] [Mon 2026-10-19 · go build ./...: 0.5s, 1 run]
npm ci=s0/4.0s/1 run go test ./...=s1/3.5s/2 runs go build ./...=s2/0.5s/1 run
go test ./...=s1/3.5s/2 runs go build ./...=s2/0.5s/1 run
npm ci=s0/4.0s/1 run go test ./...=s1/3.5s/2 runs go build ./...=s2/0.5s/1 run
1 |  | s-other | docker build | 1.0s | false
1 |  | s1 | go test ./..., job test | 1.0s | false
19 | 07:30 | s2 | go build ./..., job build | 0.5s | true
19 | 07:00 | s1 | go test ./..., job test | 2.5s | false
19 | Sat 12:00 | s0 | npm ci | 4.0s | false
npm ci | npm | 1 | 4.0s | 4.0s | 4.0s | 0 | -
go test ./... | go | 2 | 3.5s | 1.8s | 2.5s | 1 | -
1 0 | tooling-chart | Last 7 days | 4 commands · 6.0s · busiest day 4.0s
tooling-bars sparse | Time spent per day by tool, last 7 days | --n: 7; --k: 2
2=2026-10-15: 1 command in 2.0s · 6=2026-10-19: 3 commands in 4.0s
go=s3/5.0s/3 runs npm=s4/1.0s/1 run
7 Tue 7.14 Mon 92.86
1 0
1792231200,1792386000,1792387800,1792436400,0
