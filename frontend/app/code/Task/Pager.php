<?php

namespace Task;

// Pager links a page of a list to the pages either side of it: Previous to
// the newer rows, Next to the older ones.
//
// The query is kept whole, less the two cursors, which is what view.NewPager
// keeps. A page is narrowed by more than the form it draws: ?dir= names the
// directory inside a folder a run ran in and no field offers it, so a link
// that wrote the fields it knows about would answer a different list from the
// page it sits on.
//
// $_GET and not $_REQUEST: the request fields are the path parameters and the
// cookies as well, and a link carries neither.
class Pager {
	/**
	 * SIZES are the rows a page of a list may hold, the first by default.
	 *
	 * They are text, because the select draws them and ?n= arrives as text;
	 * size() is the number behind the pick. Go holds one pageSizes for every
	 * list that pages, so this is one list and not one per module.
	 */
	const SIZES = array("15", "50", "100");

	/**
	 * size is the rows a page holds: ?n= when it names one, else the first.
	 *
	 * The pick is written back into the query, because the pager keeps the
	 * query whole and a page of fifty rows links to another fifty: Go's
	 * pageSize sets n on the values it was handed, so a link carries it even
	 * where the request did not, and a size the list does not offer is the
	 * default in the link as it is in the form.
	 */
	static function size() {
		$asked = Route::input("n");
		$size = in_array($asked, self::SIZES, true) ? $asked : self::SIZES[0];
		$_GET["n"] = $size;

		return (int)$size;
	}

	/**
	 * paged trims a page read with one row more than it holds and names the
	 * cursors of the pages either side: prev, the newest row, where newer
	 * rows exist, and next, the oldest, where older ones do.
	 *
	 * A page read after a cursor has the extra row on top, any other at the
	 * bottom. A row answers its own ID, which every model a list pages over
	 * does.
	 */
	static function paged($rows, $size, $before, $after) {
		$more = count($rows) > $size;
		$newer = $before !== "";
		$older = $more;

		if ($after !== "") {
			if ($more) {
				$rows = array_slice($rows, 1);
			}

			$newer = $more;
			$older = true;
		} elseif ($more) {
			$rows = array_slice($rows, 0, $size);
		}

		if (count($rows) === 0) {
			return array(
				"rows" => $rows,
				"prev" => "",
				"next" => "",
			);
		}

		return array(
			"rows" => $rows,
			"prev" => $newer ? $rows[0]->getId() : "",
			"next" => $older ? $rows[count($rows) - 1]->getId() : "",
		);
	}

	/**
	 * of is the pages either side of a path: prev names the newest row of this
	 * page and reads after it, next the oldest and reads before it.
	 *
	 * Neither cursor draws no nav at all, which the component answers, so a
	 * list that fits on one page needs no test of its own here.
	 */
	static function of($path, $prev, $next) {
		return array(
			"prev" => self::link($path, "after", $prev),
			"next" => self::link($path, "before", $next),
		);
	}

	/**
	 * link is one page's link: the query as it arrived, the two cursors
	 * dropped and this one's written.
	 *
	 * The path is the page's own, without the mount prefix, which the template
	 * writes as {base} the way it does for every other link of a scope.
	 *
	 * The fields are written by name and escaped the way a Go query encodes
	 * them, so the two front ends answer one URL. An empty field is kept: a
	 * form submits every field it draws, and the Go page keeps what it was
	 * given.
	 */
	static function link($path, $key, $id) {
		if ($id === "") {
			return "";
		}

		$fields = self::query();
		$fields[$key] = $id;

		ksort($fields);

		$query = array();

		foreach ($fields as $name => $value) {
			$query[] = Route::queryEscape($name) . "=" . Route::queryEscape($value);
		}

		return $path . "?" . implode("&", $query);
	}

	/** query is the request's query, less the cursor of either direction. */
	static function query() {
		$out = array();

		if (!isset($_GET) || !is_array($_GET)) {
			return $out;
		}

		foreach ($_GET as $name => $value) {
			if ($name === "before" || $name === "after" || !is_string($value)) {
				continue;
			}

			$out[$name] = $value;
		}

		return $out;
	}
}
