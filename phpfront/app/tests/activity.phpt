name: activity
description: Task\Activity charts when runs ran - the days, the hours and the week, the heights of their bars, the durations as lines, and the duration line of the latest runs.
---
<?php

require "ext/autoload.php";

// A zone east of UTC, so a run early in the morning is on the day before in
// UTC and on its own day here. The clock is a parameter of the chart, so every
// slot below is stated and none of it moves with the machine's own time.
$zone = Time\Location::fixed("CEST", 7200);

// moment is a wall clock reading of October 2026 in that zone.
function moment($zone, $day, $hour, $minute) {
	return DateTime::parse_in_location("2006-01-02 15:04", sprintf("2026-10-%02d %02d:%02d", $day, $hour, $minute), $zone);
}

// one is a run as the charts read it, its duration in milliseconds. The two
// readings of the clock are what Activity::of projects: whole seconds place a
// run in its slot, the nanoseconds order the duration line.
function one($at, $elapsed, $state) {
	return array("at" => $at, "unix" => $at->unix(), "nano" => $at->unixnano(), "elapsed" => $elapsed, "state" => $state);
}

$now = moment($zone, 19, 22, 0);
// The last two are outside the window: a day before it starts, and tomorrow.
$runs = array(
	one(moment($zone, 19, 7, 0), 2000, "ok"),
	one(moment($zone, 19, 7, 30), 4000, "failed"),
	one(moment($zone, 19, 1, 30), 6000, "killed"),
	one(moment($zone, 17, 12, 0), 8000, "stopped"),
	one(moment($zone, 19, 21, 0), 0, "running"),
	one(moment($zone, 19, 7, 0)->addDate(0, 0, -30), 0, "ok"),
	one(moment($zone, 20, 7, 0), 0, "ok"),
);

$a = new Task\Activity($runs, $now, Task\Activity::DAYS);
$data = $a->data();

// The columns of the three charts, and the runs of the window.
echo count($a->bars[0]), " ", count($a->bars[1]), " ", count($a->bars[2]), "\n";
echo $a->total, " ", $a->failed, " ", $a->median, "\n";
echo $data["sum"], " | ", $data["failed"], " | ", $data["median"], "\n";

// A day tick a week, today the last, each in the middle of its column; the
// hours every third clock hour, labelled every sixth.
echo count($a->ticks[0]), " ", $a->ticks[0][0]["label"], " ", $a->ticks[0][0]["at"], " ", $a->ticks[0][4]["label"], " ", $a->ticks[0][4]["at"], "\n";
echo count($a->ticks[1]), " [", $a->ticks[1][2]["label"], "] ", $a->ticks[1][2]["at"], " ", $a->ticks[1][3]["label"], " ", $a->ticks[1][3]["at"], "\n";

// Each window's finished runs, in the order count, fastest, mean, slowest and
// spent: the thirty days hold the stopped run of the 17th too, the last 24
// hours only today's three.
echo implode(" ", $a->stats[0]), "\n";
echo implode(" ", $a->stats[1]), "\n";

// Today, split by how its runs ended: the killed run is 01:30 here and the day
// before in UTC. The busiest day is a full bar and its parts share it.
$today = $a->bars[0][29];
$before = $a->bars[0][28];

echo $today["ok"], " ", $today["failed"], " ", $today["other"], " ", $today["edge"], "\n";
echo $today["title"], "\n";
echo $today["okStyle"], " ", $today["failedStyle"], " ", $today["otherStyle"], "\n";
echo "[", $before["okStyle"], "] ", $before["title"], " | ", $before["hitTitle"], "\n";
echo $a->bars[0][17]["title"], "\n";

// The mean duration of a day's finished runs, the fastest and the slowest of
// them: a running run is in none of it.
echo $today["average"], " ", $today["finished"], " ", $today["min"], " ", $today["max"], "\n";
echo $today["hitTitle"], "\n";
echo $a->bars[0][27]["otherStyle"], " ", $a->bars[0][27]["average"], " ", $a->bars[0][27]["hitTitle"], "\n";

// The hours start at 23:00 on the 18th, so 07:00 is the ninth of them.
echo $a->bars[1][8]["ok"], " ", $a->bars[1][8]["failed"], " ", $a->bars[1][8]["average"], " ", $a->bars[1][2]["failed"], "\n";
echo $a->bars[1][8]["title"], "\n";
echo $a->bars[1][8]["hitTitle"], "\n";
echo $a->bars[1][22]["hitTitle"], "\n";

// The durations of the days as lines against the slowest run, 8 seconds on the
// 17th: a point in the middle of each column, a day with no finished run on
// the baseline, and a curve through them that stays inside the plot.
echo $a->lines[0]["label"], "\n";

foreach (array("avg", "min", "max") as $name) {
	$path = $a->lines[0][$name];

	echo $name, " ", substr($path, 0, 16), "| ", substr_count($path, " C "), " ", var_export(strpos($path, "91.67 10.00") !== false, true), " ", substr($path, -11), "\n";
}

// The week is three hour slots and its line goes straight from point to point,
// since a curve through a ragged hour reads as a trend that is not there. Its
// last slot ends at midnight.
echo substr_count($a->lines[2]["avg"], " L "), " ", $a->bars[2][6]["edge"], " [", $a->bars[2][7]["edge"], "] ", $a->bars[2][55]["edge"], "\n";
echo $a->bars[2][55]["title"], "\n";

// What the component reads: a figure per chart, its caption, and a bar whose
// every number is already in the shape its attribute takes.
$charts = $data["charts"];

echo $charts[0]["class"], " | ", $charts[0]["label"], " | ", $charts[0]["n"], "\n";
echo $charts[1]["class"], " | ", $charts[1]["label"], "\n";
echo $charts[2]["class"], " | ", $charts[2]["label"], "\n";
echo $charts[0]["stats"]["text"], "\n";
echo $charts[0]["stats"]["title"], "\n";
echo $charts[0]["bars"][29]["class"], " | ", $charts[0]["bars"][29]["ok"], " ", $charts[0]["bars"][29]["failed"], " ", $charts[0]["bars"][29]["other"], "\n";
echo "[", $charts[0]["bars"][28]["ok"], "] ", $charts[0]["bars"][28]["class"], "\n";

// The duration line: the finished runs oldest first, the longest of them the
// top of the scale, and the plot under the line as the wash.
$line = $data["durations"];

echo count($line["points"]), " ", $line["longest"], " | ", $line["label"], " | ", $line["caption"], "\n";
echo $line["line"], "\n";
echo $line["area"], "\n";
echo $line["points"][0]["class"], " | ", $line["points"][0]["style"], " | ", $line["points"][0]["at"], " | ", $line["points"][0]["title"], " | ", var_export($line["points"][0]["failed"], true), "\n";
echo $line["points"][3]["style"], " | ", $line["points"][3]["title"], " | ", var_export($line["points"][3]["failed"], true), "\n";

// No run in the window draws nothing at all. The duration line is one job's
// own, so a page of many jobs leaves it out and keeps the rest.
$none = new Task\Activity(array(), $now, Task\Activity::DAYS);
$tail = array_slice($runs, 5);
$some = new Task\Activity($tail, $now, Task\Activity::DAYS);
$jobs = new Task\Activity($runs, $now, Task\Activity::DAYS, false);

echo count($none->data()), "\n";
echo count($some->data()), "\n";

$mixed = $jobs->data();

echo count($mixed["durations"]), " ", $mixed["sum"], " ", count($mixed["charts"]), "\n";

// One connection serves the suite and fixtures run together, so the rows are
// this fixture's own and the filter names its folder.
$db = Task\DB::writer();

$db->replace("run", array(
	"id" => "fx-activity-1",
	"folder" => "fx-activity",
	"job" => "build",
	"dir" => "",
	"user_name" => "admin",
	"state" => "ok",
	"exit_code" => 0,
	"started_at" => "2026-10-19 07:00:00.000 +0000 UTC",
	"finished_at" => "2026-10-19 07:00:02.500 +0000 UTC",
	"term_cols" => 120,
	"term_rows" => 30,
));
$db->replace("run", array(
	"id" => "fx-activity-2",
	"folder" => "fx-activity",
	"job" => "build",
	"dir" => "",
	"user_name" => "admin",
	"state" => "queued",
	"exit_code" => 0,
	"term_cols" => 120,
	"term_rows" => 30,
));

// of() reads the rows a filter listed: when each run started, how long it took
// and how it ended. A row that never started is on no chart.
$table = new Task\Runs\Runs;
$found = $table->find(array("folder" => "fx-activity", "limit" => Task\Activity::LIMIT));
$read = Task\Activity::of($found);

echo count($found), " ", count($read), " ", $read[0]["elapsed"], " ", $read[0]["state"], " ", $read[0]["at"]->format(Task\Clock::READ), "\n";

// The rows of a folder nobody ran in are no chart at all.
$empty = $table->find(array("folder" => "fx-activity-none"));

echo count(Task\Activity::of($empty)), "\n";

// rows() is the same five values of the rows themselves, which is what a page
// reads: charted() asks for the columns the charts and the permission filter
// want and hydrates nothing, so a window of five thousand runs costs no model
// at all. The two paths answer the same chart, which is the whole point.
$charted = $table->charted(array("folder" => "fx-activity"));
$rows = Task\Activity::rows($charted);

echo count($charted), " ", count($rows), " ", $rows[0]["elapsed"], " ", $rows[0]["state"], " ", $rows[0]["at"]->format(Task\Clock::READ), "\n";
echo var_export($rows[0]["unix"] === $read[0]["unix"], true), " ", var_export($rows[0]["nano"] === $read[0]["nano"], true), "\n";

$one = new Task\Activity($read, $now, Task\Activity::DAYS);
$other = new Task\Activity($rows, $now, Task\Activity::DAYS);

echo var_export($one->data() === $other->data(), true), "\n";

// took is elapsed of two stamps already read: a run still going is counted to
// now, and one that never started took nothing.
$start = Task\Clock::parse("2026-10-19 07:00:00.000 +0000 UTC");

echo Task\Run::took($start, Task\Clock::parse("2026-10-19 07:00:02.500 +0000 UTC")), " ", Task\Run::took(null, null), " ", var_export(Task\Run::took($start, null) > 0, true), "\n";

// A folder nobody ran in is no rows and no chart, as above.
echo count($table->charted(array("folder" => "fx-activity-none"))), "\n";
---
30 24 56
5 2 5000
5 runs in 30 days | 2 failed | median 5.0s
5 Sep 21 5.00 Oct 19 98.33
8 [] 31.25 09:00 43.75
4 2000 5000 8000 20000
3 2000 4000 6000 12000
1 2 1 tip-end
4 runs, Mon 2026-10-19 · 2 failed · 1 running or stopped
height:25.00% height:50.00% height:25.00%
[] no runs, Sun 2026-10-18 | no finished runs, Sun 2026-10-18
no runs, Wed 2026-10-07
4000 3 2000 6000
avg 4.0s · min 2.0s · max 6.0s over 3 runs, Mon 2026-10-19
height:25.00% 8000 8.0s, 1 run, Sat 2026-10-17
1 1 3000 1
2 runs, Mon 07:00–08:00 · 1 failed
avg 3.0s · min 2.0s · max 4.0s over 2 runs, Mon 07:00–08:00
no finished runs, Mon 21:00–22:00
Average, slowest and fastest run per column, the slowest 8.0s
avg M 1.67 100.00 C | 29 true 98.33 55.00
min M 1.67 100.00 C | 29 true 98.33 77.50
max M 1.67 100.00 C | 29 true 98.33 32.50
55 tip-start [] tip-end
1 run, Mon 21:00–00:00 · 1 running or stopped
activity-chart | Runs per day, last 30 days | 30
activity-chart hours | Runs per hour, last 24 hours
activity-chart wide | Runs per 3 hours, last 7 days
min/avg/max/total: 2.0s 5.0s 8.0s 20.0s
4 finished runs: the fastest, the mean, the slowest and their time together
tip tip-end | height:25.00% height:50.00% height:25.00%
[] tip tip-end
4 8.0s | Duration of the latest 4 runs, the longest 8.0s | duration of the latest 4 runs
M 0.00 10.00 L 33.33 32.50 L 66.67 77.50 L 100.00 55.00
M 0 100 L 0.00 10.00 L 33.33 32.50 L 66.67 77.50 L 100.00 55.00 L 100 100 Z
hit tip tip-start | left:0.00%;width:16.67% | 0.00 | 8.0s, Sat 2026-10-17 12:00 · stopped | false
left:83.33%;width:16.67% | 4.0s, Mon 2026-10-19 07:30 · failed | true
0
0
0 5 runs in 30 days 3
2 1 2500 ok 2026-10-19 07:00:00
0
2 1 2500 ok 2026-10-19 07:00:00
true true
true
2500 0 true
0
