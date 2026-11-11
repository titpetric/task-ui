name: clock
description: Clock writes a timestamp the way the Go side writes it, and reads either shape back.
---
<?php

// A read gives ISO seconds with no zone, because the driver parses the column
// and the bridge renders the time it got.
echo Task\Clock::seconds("2026-11-02 10:32:12"), "\n";

// A write leaves time.Time.String(), which is what the Go side leaves, so one
// format lives in the column. Both arrive while both writers are live.
foreach (array(
	"2026-11-02 10:32:12.414 +0000 UTC",
	"2026-11-02 10:32:12.42 +0000 UTC",
	"2026-11-02 10:32:12 +0000 UTC",
) as $stamp) {
	echo Task\Clock::seconds($stamp), " ", Task\Clock::format($stamp, Task\Clock::READ), "\n";
}

// stamp writes the stored shape and reads back as the moment it names.
$now = Task\Clock::stamp();

echo var_export(strpos($now, " +0000 UTC") !== false, true), "\n";
echo var_export(Task\Clock::seconds($now) > 1700000000, true), "\n";

// A whole second and a fraction of the same second sort the way they compare
// as numbers, which is what a query comparing the column as text relies on.
echo var_export("2026-11-02 10:32:12 +0000 UTC" < "2026-11-02 10:32:12.414 +0000 UTC", true), "\n";
echo var_export("2026-11-02 10:32:12.414 +0000 UTC" < "2026-11-02 10:32:12.42 +0000 UTC", true), "\n";

// An empty column is no time, not the epoch.
echo "[", Task\Clock::format("", Task\Clock::DAY), "]", Task\Clock::seconds(""), "\n";
echo "[", Task\Clock::ago(""), "]\n";
echo Task\Clock::format("2026-11-02 10:32:12.414 +0000 UTC", Task\Clock::DAY), "\n";

// read is the cell every page draws a timestamp in: the server's own zone to
// the second, and a dash for an empty column. The zone is the machine's, so
// the shape is what is stated and not the reading.
echo "[", Task\Clock::read(""), "]\n";
echo strlen(Task\Clock::read("2026-11-02 23:00:00")), "\n";

// ago reads against the process clock, which is the one view.Ago reads.
$now = Task\Clock::real()->unix();

foreach (array(0, 30, 300, 7200, 400000) as $since) {
	echo Task\Clock::ago(\DateTime::unix($now - $since, 0)->in(\Time\Location::utc())->format(Task\Clock::READ)), "\n";
}

// The clock is the instance's and not the runtime's: the host answers this
// moment, and a pinned one is what every page of a screenshot run reads.
Task\Host::$now = "2025-06-01T12:00:00+02:00";

echo Task\Clock::now()->format(Task\Clock::READ), "\n";

// The zone arrives with it and is not converted away, because Task\Activity
// counts a chart's days in the zone of the clock it was handed.
echo Task\Clock::now()->format(Task\Clock::CLOCK), "\n";

// A write still stores UTC, whatever zone the clock reads in, because one
// column holds one format.
echo Task\Clock::stamp(), "\n";

// A pin does not move ago: view.Ago is time.Since, so a stamp two hours before
// the pinned moment does not read as two hours ago.
echo var_export(Task\Clock::ago("2025-06-01 10:00:00") !== "2h ago", true), "\n";

// A host with no clock to pin is a fixture, and then the runtime's own reads.
Task\Host::$now = "";

echo var_export(Task\Clock::now()->unix() > 1793615532, true), "\n";
---
1793615532
1793615532 2026-11-02 10:32:12
1793615532 2026-11-02 10:32:12
1793615532 2026-11-02 10:32:12
true
true
true
true
[]0
[-]
2026-11-02
[-]
19
just now
just now
5m ago
2h ago
4d ago
2025-06-01 12:00:00
2025-06-01T12:00:00+02:00
2025-06-01 10:00:00 +0000 UTC
true
true
