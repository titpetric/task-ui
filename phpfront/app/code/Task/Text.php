<?php

namespace Task;

// Text is the encoding a column carries where a column holds a list or a map:
// JSON, as model/list.go writes it.
class Text {
	/** items reads a JSON array column, an empty list for empty text. */
	static function items($text) {
		if (!is_string($text) || $text === "") {
			return array();
		}

		$found = json_decode($text, true);

		return is_array($found) ? $found : array();
	}

	/** encode writes a list back, never null, so the column holds an array. */
	static function encode($list) {
		return json_encode(is_array($list) ? array_values($list) : array());
	}

	/** map reads a JSON object of lists, such as a project's allow lists. */
	static function map($text) {
		if (!is_string($text) || $text === "") {
			return array();
		}

		$found = json_decode($text, true);

		return is_array($found) ? $found : array();
	}
}
