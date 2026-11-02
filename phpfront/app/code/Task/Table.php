<?php

namespace Task;

// Table builds the plain list the theme's table component draws: a heading,
// rows and a line when there are none.
//
// A cell is HTML, so the page composing it escapes what a reader typed.
// cell() does that and raw() is the opt out, named so a review sees it.
class Table {
	var $columns = array();
	var $rows = array();
	var $empty = "Nothing here yet.";

	/** column adds a heading cell. */
	function column($title, $class = "", $numeric = false) {
		$this->columns[] = array(
			"title" => $title,
			"class" => $class,
			"numeric" => $numeric,
		);

		return $this;
	}

	/** none sets what an empty list says. */
	function none($said) {
		$this->empty = $said;

		return $this;
	}

	/**
	 * row adds a row of cells, each one HTML. id anchors it and run is the
	 * run it is of, which the script keys on to update it whole.
	 */
	function row($cells, $id = "", $run = "", $class = "") {
		$this->rows[] = array(
			"id" => $id,
			"run" => $run,
			"class" => $class,
			"cells" => $cells,
		);

		return $this;
	}

	/** data is the scope the component reads. */
	function data() {
		return array(
			"columns" => $this->columns,
			"rows" => $this->rows,
			"width" => max(1, count($this->columns)),
			"empty" => $this->empty,
		);
	}

	/** cell is a value as a cell, escaped. */
	static function cell($value, $class = "") {
		return array("html" => htmlspecialchars((string)$value, ENT_QUOTES), "class" => $class);
	}

	/** raw is a cell the page built as markup, which it escaped itself. */
	static function raw($html, $class = "") {
		return array("html" => $html, "class" => $class);
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
