<?php

namespace Task;

// Clock is the time a row records: UTC, in the shape the column holds.
//
// Writing and reading are not the same shape, because two writers share the
// table and the driver converts on the way back.
//
// A write stores what the Go side stores. The Go side writes a *time.Time
// through the driver, which keeps time.Time.String(): `2026-11-02
// 10:32:12.414 +0000 UTC`, the fraction to the millisecond with its trailing
// zeros trimmed. Writing anything else leaves two formats in one column, and
// the queries that compare a timestamp compare it as text.
//
// A read gives ISO seconds with no zone: the driver parses a timestamp column
// into a time and the bridge renders that. Both formats arrive, because both
// writers are live, so parse takes either.
//
// The layouts are Go layouts, because the Time binding is Go's: the runtime
// has no gmdate, and date() reads the process timezone.
class Clock {
	/** READ is the shape a timestamp column arrives in. */
	const READ = "2006-01-02 15:04:05";

	/** STORED is the shape a write leaves in the column. */
	const STORED = "2006-01-02 15:04:05.999 -0700 MST";

	/** DAY is the layout a page shows a date in. */
	const DAY = "2006-01-02";

	/** stamp is now, as a write leaves it in the column. */
	static function stamp() {
		return self::now()->format(self::STORED);
	}

	/** now is this moment in UTC. */
	static function now() {
		return \DateTime::now()->in(\Time\Location::utc());
	}

	/** parse reads a timestamp column in either shape, or null for an empty one. */
	static function parse($stamp) {
		if (!is_string($stamp) || $stamp === "") {
			return null;
		}

		$layout = strpos($stamp, " +") === false && strpos($stamp, " -") === false ? self::READ : self::STORED;

		return \DateTime::parse_in_location($layout, $stamp, \Time\Location::utc());
	}

	/** seconds is a timestamp column as a unix time, 0 for an empty one. */
	static function seconds($stamp) {
		$at = self::parse($stamp);

		return $at === null ? 0 : $at->unix();
	}

	/** format shows a timestamp column in a layout of its own. */
	static function format($stamp, $layout) {
		$at = self::parse($stamp);

		return $at === null ? "" : $at->format($layout);
	}

	/** ago is how long since a timestamp reads, in the words every page uses. */
	static function ago($stamp) {
		$at = self::seconds($stamp);

		if ($at === 0) {
			return "-";
		}

		$since = self::now()->unix() - $at;

		if ($since < 60) {
			return "just now";
		}

		if ($since < 3600) {
			return (int)($since / 60) . "m ago";
		}

		if ($since < 172800) {
			return (int)($since / 3600) . "h ago";
		}

		return (int)($since / 86400) . "d ago";
	}
}
