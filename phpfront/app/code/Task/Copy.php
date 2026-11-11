<?php

namespace Task;

// Copy is a value with a button that copies it, as the copyable component
// reads it.
//
// The value is what reaches the clipboard and $shown is what the cell has room
// to draw; where they differ the whole value goes in the tooltip, so a reader
// without a clipboard can still read it.
class Copy {
	/** of is one value to copy. $shown defaults to the value itself. */
	static function of($value, $label, $shown = null, $fill = false) {
		return array(
			"value" => $value,
			"label" => $label,
			"shown" => $shown === null ? $value : $shown,
			"fill" => $fill,
		);
	}
}
