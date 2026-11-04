<?php

namespace Task;

// Size words a quantity of bytes the way every page words one.
//
// The three are the filters frontend/view registers, by the names the Go
// templates call them under, and all three are the one formatter: a size
// reads the same whether it came off a file, off Beszel in megabytes, or off
// a second of traffic.
class Size {
	/**
	 * bytes words a byte count: whole bytes under a kilobyte, and a tenth of
	 * the larger unit above it.
	 */
	static function bytes($bytes) {
		$bytes = (int)$bytes;

		if ($bytes < 1024) {
			return $bytes . " B";
		}

		$value = $bytes / 1024;
		$unit = "KB";

		if ($value >= 1024) {
			$value = $value / 1024;
			$unit = "MB";
		}

		if ($value >= 1024) {
			$value = $value / 1024;
			$unit = "GB";
		}

		if ($value >= 1024) {
			$value = $value / 1024;
			$unit = "TB";
		}

		return sprintf("%.1f %s", $value, $unit);
	}

	/**
	 * space is what a directory takes on disk as components/space.tpl reads
	 * it: the files, the git history behind them and what git ignores.
	 *
	 * Nothing was measured on an instance with no meter, and then the answer
	 * is empty and the page says nothing about space rather than saying the
	 * directory takes none.
	 *
	 * The total leaves out what git ignores, which the legend names apart, so
	 * the two shares of the bar add up to the whole of it.
	 */
	static function space($space) {
		if (!Row::flag($space, "measured")) {
			return array();
		}

		$files = Row::number($space, "files");
		$history = Row::number($space, "history");
		$ignored = Row::number($space, "ignored");
		$total = $files + $history;
		$share = $total === 0 ? 0 : $files * 100 / $total;

		return array(
			"total" => self::bytes($total),
			"files" => self::bytes($files),
			"history" => self::bytes($history),
			"ignored" => $ignored === 0 ? "" : self::bytes($ignored),
			"filesPercent" => sprintf("%.2f", $share),
			"historyPercent" => sprintf("%.2f", $total === 0 ? 0 : 100 - $share),
		);
	}

	/**
	 * mini is the same measure in a table cell, as components/space-mini.tpl
	 * reads it: the two shares of a short bar, the total beside it, and the
	 * whole measure as the one title a cell has room for.
	 *
	 * Nothing was measured on an instance with no meter, and then the answer
	 * is empty and the cell says nothing about space rather than nothing at
	 * all, which is what the column itself does.
	 */
	static function mini($space) {
		$read = self::space($space);

		if (count($read) === 0) {
			return array();
		}

		$title = $read["files"] . " of files, " . $read["history"] . " of history";

		if ($read["ignored"] !== "") {
			$title = $title . "; " . $read["ignored"] . " more that git ignores";
		}

		return array(
			"title" => $title,
			"total" => $read["total"],
			"filesPercent" => $read["filesPercent"],
			"historyPercent" => $read["historyPercent"],
		);
	}

	/** megabytes words a size Beszel reports in megabytes. */
	static function megabytes($value) {
		return self::bytes((int)((float)$value * 1048576));
	}

	/** rate words a traffic in bytes a second. */
	static function rate($value) {
		return self::bytes((int)$value) . "/s";
	}
}
