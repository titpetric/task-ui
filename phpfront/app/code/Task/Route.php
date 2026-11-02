<?php

namespace Task;

// Route is the request, read against the path the router is mounted at.
//
// The request is not rewritten, so $_SERVER["REQUEST_URI"] holds the prefix
// the instance serves this application under. This is the only place that
// knows it, and it reads it from Host::base() rather than spelling it.
class Route {
	/** method is the request method. */
	static function method() {
		return isset($_SERVER["REQUEST_METHOD"]) ? $_SERVER["REQUEST_METHOD"] : "GET";
	}

	/** posted reports whether the request is a write. */
	static function posted() {
		return self::method() === "POST";
	}

	/** path is the request path past the prefix, without the query or the slashes at either end. */
	static function path() {
		$uri = isset($_SERVER["REQUEST_URI"]) ? $_SERVER["REQUEST_URI"] : "/";
		$parts = explode("?", $uri);
		$path = $parts[0];
		$base = Host::base();

		if ($base !== "" && strpos($path, $base) === 0) {
			$path = substr($path, strlen($base));
		}

		return trim($path, "/");
	}

	/** segments is path cut on the slashes, empty for the root. */
	static function segments() {
		$path = self::path();

		return $path === "" ? array() : explode("/", $path);
	}

	/** param is a path parameter the @route annotation declared. */
	static function param($name, $fallback = "") {
		return isset($_REQUEST[$name]) ? $_REQUEST[$name] : $fallback;
	}

	/** input is a query or form field, trimmed. */
	static function input($name, $fallback = "") {
		if (!isset($_REQUEST[$name])) {
			return $fallback;
		}

		return trim((string)$_REQUEST[$name]);
	}

	/** href is a link of this application, the prefix included. */
	static function href($path) {
		if ($path === "" || $path === "/") {
			return Host::base() . "/";
		}

		return Host::base() . "/" . ltrim($path, "/");
	}

	/** go answers with a redirect, which a form that changed something should. */
	static function go($path) {
		header("Location: " . self::href($path));
		http_response_code(303);
	}

	/** notice sends the reader to a path with something to report. */
	static function notice($path, $said) {
		self::go($path . "?notice=" . urlencode($said));
	}

	/** fail sends the reader back with what the form needs. */
	static function fail($path, $why) {
		self::go($path . "?error=" . urlencode($why));
	}

	/** missing answers for a page that is not there. */
	static function missing() {
		http_response_code(404);
	}

	/** refused answers for a page the user may not see. */
	static function refused() {
		http_response_code(403);
	}
}
