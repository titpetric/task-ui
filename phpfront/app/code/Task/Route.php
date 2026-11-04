<?php

namespace Task;

// Route is the request, read against the path the router is mounted at.
//
// The request is not rewritten, so $_SERVER["REQUEST_URI"] holds the prefix
// the instance serves this application under. This is the only place that
// knows it, and it reads it from Host::base() rather than spelling it.
class Route {
	/**
	 * INTERNAL is the root folder: the page of a place outside a named
	 * workspace sits under it, and the instance's own jobs run in it.
	 */
	const INTERNAL = "internal";

	/**
	 * LEGACY is the name the root folder had. Every path of it moved, for
	 * good, to the one it has, so a link that spells it redirects.
	 */
	const LEGACY = "_";

	/** UNRESERVED is the bytes RFC 3986 says a URL producer leaves literal. */
	const UNRESERVED = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~";

	/**
	 * SAFE is the bytes a whole link keeps literal: the unreserved ones and
	 * the delimiters a URL is built out of, which linked() leaves alone.
	 */
	const SAFE = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~!#$&*+,/:;=?@[]";

	/** HEX is the digits an escape in a link already spells itself with. */
	const HEX = "0123456789abcdefABCDEF";

	/**
	 * canonical is a folder ID with the name the root had rewritten to the one
	 * it has, and any other ID as it was written.
	 */
	static function canonical($id) {
		if ($id === self::LEGACY) {
			return self::INTERNAL;
		}

		if (strpos($id, self::LEGACY . "/") !== 0) {
			return $id;
		}

		return self::INTERNAL . substr($id, strlen(self::LEGACY));
	}

	/** reserved reports whether an ID starts with a name task-ui keeps for the root. */
	static function reserved($id) {
		$parts = explode("/", $id);

		return $parts[0] === self::INTERNAL || $parts[0] === self::LEGACY;
	}

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

	/** input is a query or form field, verbatim: Go's Query().Get() does not trim either. */
	static function input($name, $fallback = "") {
		if (!isset($_REQUEST[$name])) {
			return $fallback;
		}

		return (string)$_REQUEST[$name];
	}

	/** href is a link of this application, the prefix included. */
	static function href($path) {
		if ($path === "" || $path === "/") {
			return Host::base() . "/";
		}

		return Host::base() . "/" . ltrim($path, "/");
	}

	/**
	 * site is the public base link of the instance, with no slash at the end:
	 * the URL it is configured with, and the scheme and host the request came
	 * to where it is configured with none.
	 *
	 * It is a whole link and not a path, because what reads it is a line a
	 * reader copies into a client of their own: the MCP endpoint, and the
	 * `claude mcp add` that names it.
	 *
	 * A connection that is TLS with no proxy in front of it reads as http,
	 * because the request carries nothing that says otherwise: the SAPI
	 * answers the headers and the host and not the state of the socket. An
	 * instance served over TLS directly sets TASK_UI_URL.
	 */
	static function site() {
		$instance = Host::instance(false);
		$url = Row::text($instance, "url");

		if ($url !== "") {
			return $url;
		}

		return self::scheme() . "://" . self::host();
	}

	/** scheme is what the proxy in front says the reader asked over. */
	static function scheme() {
		$proto = isset($_SERVER["HTTP_X_FORWARDED_PROTO"]) ? $_SERVER["HTTP_X_FORWARDED_PROTO"] : "";

		return strcasecmp($proto, "https") === 0 ? "https" : "http";
	}

	/**
	 * host is the host the reader asked for: the first name the proxy
	 * forwards, and the one the request carries where it forwards none.
	 */
	static function host() {
		$forwarded = isset($_SERVER["HTTP_X_FORWARDED_HOST"]) ? $_SERVER["HTTP_X_FORWARDED_HOST"] : "";
		$names = explode(",", $forwarded);
		$first = trim($names[0]);

		if ($first !== "") {
			return $first;
		}

		return isset($_SERVER["HTTP_HOST"]) ? $_SERVER["HTTP_HOST"] : "";
	}

	/**
	 * folderHref is the page of a folder: its own under a named workspace,
	 * and under the root otherwise. The workspaces are the folders of the
	 * named ones, which Task\Workspace::folders() answers.
	 */
	static function folderHref($id, $workspaces) {
		if ($id === self::INTERNAL) {
			return "/f/" . self::INTERNAL;
		}

		$parts = explode("/", $id);

		if (in_array($parts[0], $workspaces, true)) {
			return "/f/" . $id;
		}

		return "/f/" . self::INTERNAL . "/" . $id;
	}

	/** placeHref is the page of a folder, or of a directory inside it. */
	static function placeHref($id, $dir, $workspaces) {
		$href = self::folderHref($id, $workspaces);

		return $dir === "" ? $href : $href . "/" . $dir;
	}

	/**
	 * crumbs is the trail above the page of a place: a step per segment of the
	 * folder's ID, each linking the page of the path down to it, and a last
	 * step of its own where the page names one.
	 *
	 * A folder in no named workspace starts at the root, because its page sits
	 * under the root's. The last step of a trail that names none loses its
	 * link, as the page it would open is the one being read.
	 */
	static function crumbs($id, $last, $workspaces) {
		$out = array();

		if (strpos(self::folderHref($id, $workspaces), "/f/" . self::INTERNAL . "/") === 0) {
			$out[] = array(
				"text" => self::crumbName(self::INTERNAL),
				"href" => self::folderHref(self::INTERNAL, $workspaces),
			);
		}

		$path = "";

		foreach (explode("/", $id) as $segment) {
			if ($path !== "") {
				$path .= "/";
			}

			$path .= $segment;
			$out[] = array(
				"text" => self::crumbName($segment),
				"href" => self::folderHref($path, $workspaces),
			);
		}

		if ($last !== "") {
			$out[] = array("text" => $last, "href" => "");

			return $out;
		}

		$count = count($out);

		if ($count > 0) {
			$out[$count - 1]["href"] = "";
		}

		return $out;
	}

	/** crumbName is what a step of the trail says, the root being Internal. */
	static function crumbName($segment) {
		return $segment === self::INTERNAL ? "Internal" : $segment;
	}

	/** place is what a title names a folder and a directory inside it by. */
	static function place($id, $dir) {
		return $dir === "" ? $id : $id . "/" . $dir;
	}

	/**
	 * heldHref is the page of a place a run or a note names, which is a folder
	 * and whatever the place holds under it: the page of the longest folder
	 * the reader sees that holds it, with the rest of the place after it.
	 *
	 * The root is its own page. A place no folder of theirs holds has no page
	 * and answers "", which is a checkout they do not see or one that is gone.
	 *
	 * The folders are the catalogue as Task\Host::folders() answers it,
	 * narrowed to what the reader sees, and the workspaces the named ones.
	 */
	static function heldHref($place, $folders, $workspaces) {
		if ($place === self::INTERNAL) {
			return "/f/" . self::INTERNAL;
		}

		$best = "";

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");

			if ($place !== $id && strpos($place, $id . "/") !== 0) {
				continue;
			}

			if (strlen($id) > strlen($best)) {
				$best = $id;
			}
		}

		if ($best === "") {
			return "";
		}

		return self::folderHref($best, $workspaces) . substr($place, strlen($best));
	}

	/**
	 * pathEscape is one step of a URL path, escaped as Go's url.PathEscape
	 * escapes it.
	 *
	 * rawurlencode is RFC 3986 and escapes more: Go leaves the six characters
	 * a path segment may carry as themselves, so a file named `a+b` links to
	 * `a+b` and not to `a%2Bb`. Both resolve to the same file, and the port is
	 * compared against the Go page by its HTML, so they are spelled the same.
	 *
	 * A literal `%24` in the name is `%2524` after rawurlencode and no step
	 * puts it back, so the pairs replaced are the ones this escaped.
	 */
	static function pathEscape($step) {
		$escaped = array("%24", "%26", "%2B", "%3A", "%3D", "%40");
		$plain = array("$", "&", "+", ":", "=", "@");

		return str_replace($escaped, $plain, rawurlencode($step));
	}

	/**
	 * escapePath is a whole path as the steps of a URL path: every step
	 * escaped, the slashes between them left as they are.
	 */
	static function escapePath($path) {
		$out = array();

		foreach (explode("/", $path) as $step) {
			$out[] = self::pathEscape($step);
		}

		return implode("/", $out);
	}

	/**
	 * queryEscape is a query value, escaped as Go's url.QueryEscape escapes it.
	 *
	 * urlencode is the same but for a tilde, which Go leaves; rawurlencode is
	 * the same but for a space, which Go writes as a plus in a query where it
	 * writes it as %20 in a path.
	 */
	static function queryEscape($value) {
		return str_replace("%20", "+", rawurlencode($value));
	}

	/**
	 * urlEscape is a value the markup of a link writes into its query itself,
	 * escaped as Go's html/template escapes one written there.
	 *
	 * It escapes the bytes rawurlencode escapes and spells the two hex digits
	 * in lower case, which is the one way the two differ; a space is %20 and
	 * not a plus. The case matters because the pages are compared byte for
	 * byte against the Go front end's.
	 *
	 * A link built in PHP and handed to a component carries queryEscape()
	 * instead, as the Go template's own urlquery does.
	 */
	static function urlEscape($value) {
		$out = "";
		$length = strlen($value);

		for ($at = 0; $at < $length; $at++) {
			$char = $value[$at];

			if (strpos(self::UNRESERVED, $char) !== false) {
				$out .= $char;

				continue;
			}

			$out .= sprintf("%%%02x", ord($char));
		}

		return $out;
	}

	/**
	 * linked is a whole link as the markup of an href carries it: a byte a URL
	 * may not hold spelled as an escape, an escape already there left alone.
	 *
	 * Go's html/template normalises every value it writes into an href, so a
	 * link built out of something a reader typed - the workspace the add
	 * project form was opened for - reaches the markup escaped, and a space in
	 * it is %20 and not a space. The attributes the script reads beside it,
	 * data-next and data-api, are no links to the escaper and carry the value
	 * as it is, so a page asks for this per link rather than href() doing it.
	 *
	 * It is html/template's urlNormalizer: the unreserved bytes and the
	 * delimiters stay, a per cent that already spells an escape stays, and the
	 * hex is lower case, which is what Go writes.
	 */
	static function linked($url) {
		$out = "";
		$length = strlen($url);

		for ($at = 0; $at < $length; $at++) {
			$char = $url[$at];

			if (strpos(self::SAFE, $char) !== false) {
				$out .= $char;

				continue;
			}

			if ($char === "%" && self::escaped($url, $at)) {
				$out .= $char;

				continue;
			}

			$out .= sprintf("%%%02x", ord($char));
		}

		return $out;
	}

	/**
	 * escaped reports whether the per cent at a place in a link starts an
	 * escape, which normalising leaves as it is rather than escaping again.
	 *
	 * The two digits have to be there and a byte has to follow them, which is
	 * the test html/template makes.
	 */
	static function escaped($url, $at) {
		if ($at + 2 >= strlen($url)) {
			return false;
		}

		if (strpos(self::HEX, $url[$at + 1]) === false) {
			return false;
		}

		return strpos(self::HEX, $url[$at + 2]) !== false;
	}

	/**
	 * safeBack is where the sign in form sends the reader once they are in: a
	 * path of this instance, and the root for anything else.
	 *
	 * It is Go's access.SafeBack, which reads the value this writes.
	 */
	static function safeBack($back) {
		if ($back === "" || strpos($back, "/") !== 0) {
			return "/";
		}

		if (strpos($back, "//") === 0 || strpos($back, "/\\") === 0) {
			return "/";
		}

		return $back;
	}

	/** go answers with a redirect, which a form that changed something should. */
	static function go($path) {
		header("Location: " . self::href($path));
		http_response_code(303);
	}

	/**
	 * seeOther answers a GET with a redirect, which a page a reader has no
	 * business on should: the sign in page of somebody already signed in.
	 *
	 * The body is the one Go's http.Redirect writes for a GET, to the byte, so
	 * a client that follows nothing reads the same from either front end.
	 */
	static function seeOther($path) {
		$target = self::href($path);

		header("Location: " . $target);
		http_response_code(303);

		$link = htmlspecialchars($target, ENT_QUOTES);

		echo '<a href="' . $link . '">See Other</a>.' . "\n\n";
	}

	/**
	 * found answers a GET with a redirect to the page the reader asked for
	 * under the name it has, which a link by ID does: /settings/users/<id>
	 * opens the edit page of the user that ID names.
	 *
	 * The body is the one Go's http.Redirect writes for a GET, to the byte, as
	 * seeOther's is. The status is 302 and not 303 because the link is a GET
	 * that found its target, not the answer to something a form changed.
	 */
	static function found($path) {
		$target = self::href($path);

		header("Location: " . $target);
		http_response_code(302);

		$link = htmlspecialchars($target, ENT_QUOTES);

		echo '<a href="' . $link . '">Found</a>.' . "\n\n";
	}

	/**
	 * moved answers with a redirect that is permanent, which a path renamed
	 * for good should: a link to the name the root had opens the name it has,
	 * and the query it carried goes with it.
	 *
	 * The body is the one a client that follows nothing is left with, and it
	 * is written to the byte: a link to the target and the status in words,
	 * the line the Go handler writes, so the two answers read as one.
	 */
	static function moved($path) {
		$query = isset($_SERVER["QUERY_STRING"]) ? $_SERVER["QUERY_STRING"] : "";
		$target = self::href($path);

		if ($query !== "") {
			$target .= "?" . $query;
		}

		header("Location: " . $target);
		http_response_code(301);

		$link = htmlspecialchars($target, ENT_QUOTES);

		echo '<a href="' . $link . '">Moved Permanently</a>.' . "\n\n";
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
