<?php

// DateTime is a stub of the time binding, for the linter alone. Nothing
// requires this file at runtime.
//
// date() is not the one to reach for: in this runtime it answers with
// the layout it was handed, so a row written through it carries the
// layout and not a time.
class DateTime {
	/** The current time. */
	static function now() {
		return null;
	}
}
