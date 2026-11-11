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

	/** NANO is how many nanoseconds a second holds, which ago counts in. */
	const NANO = 1000000000;

	/**
	 * CLOCK is the shape the host answers this moment in: RFC 3339, to the
	 * nanosecond, which is the one layout a Go time round trips through text
	 * with its zone intact.
	 */
	const CLOCK = "2006-01-02T15:04:05.999999999Z07:00";

	/**
	 * stamp is now, as a write leaves it in the column: UTC, whatever zone the
	 * clock reads in, because one column holds one format and the Go writer
	 * writes UTC.
	 */
	static function stamp() {
		$at = self::now();

		return $at->in(\Time\Location::utc())->format(self::STORED);
	}

	/**
	 * now is this moment, in the zone the instance's clock reads in.
	 *
	 * The clock is the host's and not the runtime's, because the instance has
	 * one that is pinned: TASK_UI_NOW is what the hours gate and every Go page
	 * read, and a chart of the last thirty days ends where it says. A page
	 * reading the process clock would draw another window over the same rows.
	 *
	 * The zone comes with it, because a chart counts its days in the zone of
	 * the clock: a reading converted to UTC on the way here would put a run of
	 * the late evening in tomorrow and the Go page it is compared against in
	 * today. Task\Activity reads the zone off this.
	 *
	 * A host that answers no clock is a fixture, which has no instance to pin.
	 * The shape is checked rather than trusted, as in parse: a stamp the layout
	 * cannot read is a fatal, and this one is read on every page.
	 */
	static function now() {
		$stamp = Host::now();
		$shape = '/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$/';

		if (!is_string($stamp) || !preg_match($shape, $stamp)) {
			return self::real();
		}

		return \DateTime::parse(self::CLOCK, $stamp);
	}

	/**
	 * real is this moment on the process clock, which is not the instance's.
	 *
	 * Three things a page draws read it in Go and nothing makes them one with
	 * the rest: view.Ago is time.Since, model.Run.Elapsed is time.Since, and
	 * the sidebar's recents are grouped on time.Now() while the page they sit
	 * on carries the instance's Now. So a pinned clock moves the charts, the
	 * hours gate and the expiry badges and leaves those three where they are,
	 * and the port reads what each of them reads: the alternative is a port
	 * that differs from the page it is compared against.
	 */
	static function real() {
		return \DateTime::now();
	}

	/**
	 * parse reads a timestamp column in either shape, or null for an empty one
	 * and for one in neither shape.
	 *
	 * The shape is checked before the layout reads it, because a stamp the
	 * layout cannot read is a fatal and not an error: one bad timestamp from
	 * one host binding would take the whole page down. A column holds what the
	 * two writers write, but a binding answers with whatever the host told it.
	 */
	static function parse($stamp) {
		if (!is_string($stamp) || $stamp === "") {
			return null;
		}

		$layout = self::READ;
		$shape = '/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/';

		if (strpos($stamp, " +") !== false || strpos($stamp, " -") !== false) {
			$layout = self::STORED;
			$shape = '/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(\.\d+)? [+-]\d{4} \S+$/';
		}

		if (!preg_match($shape, $stamp)) {
			return null;
		}

		return \DateTime::parse_in_location($layout, $stamp, \Time\Location::utc());
	}

	/** seconds is a timestamp column as a unix time, 0 for an empty one. */
	static function seconds($stamp) {
		$at = self::parse($stamp);

		return $at === null ? 0 : $at->unix();
	}

	/**
	 * nanos is a timestamp column as a unix time in nanoseconds, 0 for an
	 * empty one.
	 *
	 * The fraction of a second a column holds is what orders two rows that
	 * started in the same second, which whole seconds read as equal. A column
	 * holds the fraction to the millisecond; Row::stamp keeps it.
	 */
	static function nanos($stamp) {
		$at = self::parse($stamp);

		return $at === null ? 0 : $at->unixnano();
	}

	/**
	 * read shows a timestamp column the way a cell shows it: the server's own
	 * zone, to the second, and a dash for an empty column.
	 *
	 * A column holds UTC, which parse reads it in; the cell is local, as
	 * view.Stamp is, so a reading matches the clock on the wall of whoever
	 * runs the instance.
	 */
	static function read($stamp) {
		$at = self::parse($stamp);

		return $at === null ? "-" : $at->in(\Time\Location::local())->format(self::READ);
	}

	/** format shows a timestamp column in a layout of its own. */
	static function format($stamp, $layout) {
		$at = self::parse($stamp);

		return $at === null ? "" : $at->format($layout);
	}

	/**
	 * ago is how long since a timestamp reads, in the words every page uses.
	 *
	 * The since is the process clock's, because view.Ago is time.Since and no
	 * page hands it the instance's: a pinned clock moves the charts beside this
	 * cell and leaves this cell where it is. See real.
	 *
	 * It counts in nanoseconds because view.Ago counts a time.Duration, and
	 * two readings truncated to the second are a second apart from it once in
	 * every sixty: a run of 179.6 seconds ago reads as 2m there and as 3m here,
	 * on every reload, on whichever page holds the cell.
	 */
	static function ago($stamp) {
		$at = self::nanos($stamp);

		if ($at === 0) {
			return "-";
		}

		$minute = 60 * self::NANO;
		$hour = 60 * $minute;
		$since = self::real()->unixnano() - $at;

		if ($since < $minute) {
			return "just now";
		}

		if ($since < $hour) {
			return intdiv($since, $minute) . "m ago";
		}

		if ($since < 48 * $hour) {
			return intdiv($since, $hour) . "h ago";
		}

		return intdiv($since, 24 * $hour) . "d ago";
	}
}
