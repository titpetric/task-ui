<?php

namespace Task;

// Table is rows of cells: a heading, a body and a foot, each a list of rows.
//
// A cell is HTML, so cell() escapes what a reader typed and raw() is the opt
// out, named so a review sees it.
class Table {
	var $class;
	var $wrap;
	var $header = array();
	var $body = array();
	var $footer = array();

	function __construct($class = "", $wrap = "") {
		$this->class = $class;
		$this->wrap = $wrap;
	}

	/** head adds a row of heading cells. */
	function head($cells) {
		$this->header[] = self::line($cells);

		return $this;
	}

	/**
	 * row adds a row of the body. id anchors it and run is the run it is of,
	 * which the script keys on to update it whole.
	 */
	function row($cells, $id = "", $run = "", $class = "") {
		$this->body[] = self::line($cells, $id, $run, $class);

		return $this;
	}

	/** foot adds a row under the body. */
	function foot($cells, $class = "") {
		$this->footer[] = self::line($cells, "", "", $class);

		return $this;
	}

	/** data is the scope the component reads. */
	function data() {
		return array(
			"class" => $this->class,
			"wrap" => $this->wrap,
			"header" => $this->header,
			"body" => $this->body,
			"footer" => $this->footer,
		);
	}

	/** line is one row. */
	static function line($cells, $id = "", $run = "", $class = "") {
		return array(
			"id" => $id,
			"run" => $run,
			"class" => $class,
			"cells" => $cells,
		);
	}

	/** cell is a value as a cell, escaped. */
	static function cell($value, $class = "", $span = 0) {
		return self::raw(htmlspecialchars((string)$value, ENT_QUOTES), $class, $span);
	}

	/** raw is a cell the page built as markup, which it escaped itself. */
	static function raw($html, $class = "", $span = 0) {
		return array(
			"html" => $html,
			"class" => $class,
			"span" => $span,
		);
	}

	/** link is a cell holding one link, both halves escaped. */
	static function link($href, $text, $class = "") {
		return self::raw('<a href="' . htmlspecialchars($href, ENT_QUOTES) . '">' . htmlspecialchars($text, ENT_QUOTES) . "</a>", $class);
	}

	/** faint is a cell the page draws quietly, such as a dash for nothing. */
	static function faint($value, $class = "") {
		return self::raw('<span class="faint">' . htmlspecialchars((string)$value, ENT_QUOTES) . "</span>", $class);
	}
}
