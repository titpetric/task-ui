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

	/** stamp is a timestamp column, or "" where the row holds none. */
	static function stamp($row, $column) {
		if (!is_array($row) || !isset($row[$column])) {
			return "";
		}

		return (string)$row[$column];
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
