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

// ago reads the way every other page of the instance reads.
$now = Task\Clock::now()->unix();

foreach (array(0, 30, 300, 7200, 400000) as $since) {
	echo Task\Clock::ago(\DateTime::unix($now - $since, 0)->in(\Time\Location::utc())->format(Task\Clock::READ)), "\n";
}
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
just now
just now
5m ago
2h ago
4d ago
