<?php

namespace Task;

// Id makes the identifier a row carries: 26 characters, ten of the time in
// milliseconds so rows sort by creation and sixteen random so two made in the
// same instant stay apart, in Crockford base32, lower case.
//
// The same shape model/id.go makes, so a row written here sorts beside one
// written there.
class Id {
	/** ALPHABET is Crockford base32, lower case. */
	const ALPHABET = "0123456789abcdefghjkmnpqrstvwxyz";

	/** LENGTH is how long an identifier is. */
	const LENGTH = 26;

	/** make is a new identifier, as of now. */
	static function make() {
		return self::at((int)round(microtime(true) * 1000));
	}

	/** at is an identifier that sorts as of a moment in milliseconds. */
	static function at($ms) {
		return self::encode($ms, 10) . self::random(16);
	}

	/** valid reports whether an identifier is one make could have returned. */
	static function valid($id) {
		return is_string($id) && preg_match('/^[0-9a-z]{' . self::LENGTH . '}$/', $id) === 1;
	}

	/** encode writes a number in the alphabet, padded to width. */
	static function encode($number, $width) {
		$out = "";

		for ($i = 0; $i < $width; $i++) {
			$out = substr(self::ALPHABET, $number % 32, 1) . $out;
			$number = (int)($number / 32);
		}

		return $out;
	}

	/** random is width characters of the alphabet. */
	static function random($width) {
		$out = "";

		for ($i = 0; $i < $width; $i++) {
			$out .= substr(self::ALPHABET, random_int(0, 31), 1);
		}

		return $out;
	}
}
