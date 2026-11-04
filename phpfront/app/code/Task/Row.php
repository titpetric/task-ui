<?php

namespace Task;

// Row reads the columns of a database row.
//
// A row is a map and carries no column order, so a reader names its column.
// The readers coerce, because a driver answers the same column as a string on
// one and an integer on another.
class Row {
	/** text is a column as a string. */
	static function text($row, $column, $fallback = "") {
		if (!is_array($row) || !isset($row[$column])) {
			return $fallback;
		}

		return (string)$row[$column];
	}

	/** number is a column as an integer. */
	static function number($row, $column, $fallback = 0) {
		if (!is_array($row) || !isset($row[$column])) {
			return $fallback;
		}

		return (int)$row[$column];
	}

	/** flag is a column as a boolean, which sqlite answers as 0 or 1. */
	static function flag($row, $column) {
		if (!is_array($row) || !isset($row[$column])) {
			return false;
		}

		$value = $row[$column];

		if (is_string($value)) {
			return $value !== "" && $value !== "0" && $value !== "false";
		}

		return (bool)$value;
	}

	/**
	 * items is a key that holds a list, as a list.
	 *
	 * No column holds one; what does is the host's answers, which arrive as
	 * arrays all the way down and are read with the same readers as a row. A
	 * key that is absent or holds something else is an empty list, so a
	 * foreach over the answer is safe without a test of its own.
	 */
	static function items($row, $column) {
		if (!is_array($row) || !isset($row[$column]) || !is_array($row[$column])) {
			return array();
		}

		return $row[$column];
	}

	/**
	 * stamp is a timestamp column, or "" where the row holds none.
	 *
	 * The driver answers such a column as a time and not as a string, and a
	 * cast of one renders whole seconds: the fraction the column holds is
	 * dropped, which is what orders two runs that started in the same second.
	 * So a time is formatted in the shape the column holds, in UTC, which is
	 * what Clock says a row records.
	 *
	 * The shape matters a second time: a model read from a row and written
	 * back carries this value, and a column that holds two formats is
	 * compared as text by every statement that orders by it.
	 */
	static function stamp($row, $column) {
		if (!is_array($row) || !isset($row[$column])) {
			return "";
		}

		$value = $row[$column];

		if (is_object($value) && method_exists($value, "format")) {
			return $value->in(\Time\Location::utc())->format(Clock::STORED);
		}

		return (string)$value;
	}

	/**
	 * values is a model as the database takes it, keyed by column.
	 *
	 * A timestamp column that holds nothing is written as null and not as an
	 * empty string: the column is nullable, the Go side stores a nil
	 * *time.Time as null, and a "" in it is a row the Go scanner cannot read
	 * back. The model names its timestamp columns, because get_object_vars
	 * answers values and not types.
	 */
	static function values($model, $stamps = array()) {
		$values = get_object_vars($model);

		foreach ($stamps as $column) {
			if (isset($values[$column]) && $values[$column] === "") {
				$values[$column] = null;
			}
		}

		return $values;
	}

	/** only is values cut to the named columns. */
	static function only($model, $columns) {
		$values = get_object_vars($model);
		$out = array();

		foreach ($columns as $column) {
			if (isset($values[$column])) {
				$out[$column] = $values[$column];
			}
		}

		return $out;
	}

	/** placeholders is a comma separated ? for each value, for an IN clause. */
	static function placeholders($values) {
		if (count($values) === 0) {
			return "";
		}

		return implode(",", array_fill(0, count($values), "?"));
	}
}
