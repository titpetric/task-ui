<?php

namespace Task;

// Matrix is one subject's policy as the permissions pages draw it: a row per
// section of config/permissions.yml with a cell per rule, then a row per
// application permission under an Application heading.
//
// A cell says what the subject's own entries say about one path and nothing
// else: inherit is the empty box, and what the role's rules then decide is the
// mark after it, which effective turns on. Reading the two apart is the whole
// point of the table, so the own entries are read raw here and not through
// Task\Policy::effect(), which folds a section's all into every rule of it.
//
// An admin's matrix is fixed: every cell allows, nothing changes, and the page
// says why instead of drawing a Save.
class Matrix {
	/** INHERIT is what a cell with no entry of its own says it is. */
	const INHERIT = "inherit";

	/** TICK and CROSS are the marks a cell draws for the two effects. */
	const TICK = "✓";

	const CROSS = "✕";

	/** api is where Save puts the cells that changed. */
	var $api = "";

	/** fixed is an admin's: every cell allows and none of them changes. */
	var $fixed = false;

	/** editable lets the cells change and draws the Save under them. */
	var $editable = false;

	/** effective draws, after each cell, whether the subject may do it at all. */
	var $effective = false;

	/** own is the subject's own entries, which decide what each cell says. */
	var $own = null;

	/** can is every checkable path and whether the subject may, by both policies. */
	var $can = array();

	function __construct($own = null, $can = array(), $fixed = false) {
		if ($own === null) {
			$own = new Policy;
		}

		$this->own = $own;
		$this->can = $can;
		$this->fixed = $fixed;
	}

	/** data is the scope the component reads. */
	function data() {
		$rules = Permission::titles("rules");
		$sections = array();

		foreach (Permission::titles("sections") as $section) {
			$sections[] = $this->section($section, $rules);
		}

		$permissions = array();

		// A rule a section added is a row of its own: no column of every
		// section holds it. frontend/view/matrix.go:51.
		foreach (Permission::extras() as $extra) {
			$permissions[] = $this->cell($extra["key"], self::extraTitle($extra, Permission::titles("sections")));
		}

		foreach (Permission::titles("permissions") as $entry) {
			$permissions[] = $this->cell($entry["key"], $entry["title"]);
		}

		$columns = count($rules) + 1;

		return array(
			"api" => $this->api,
			"fixed" => $this->fixed,
			"editable" => $this->editable,
			"effective" => $this->effective,
			"rules" => $rules,
			"sections" => $sections,
			"permissions" => $permissions,
			"columns" => $columns,
			"span" => $columns - 1,
		);
	}

	/** section is one row of the table: the section's title and a cell per rule. */
	function section($section, $rules) {
		$cells = array();

		foreach ($rules as $rule) {
			$path = $section["key"] . "/" . $rule["key"];
			$cells[] = $this->cell($path, $section["title"] . ", " . $rule["title"]);
		}

		return array("title" => $section["title"], "cells" => $cells);
	}

	/**
	 * cell is one path of the policy as the component draws it: what the
	 * subject's own entry says, and what the two policies together decide.
	 *
	 * An all rule has no result of its own, because it stands for the rules
	 * beside it rather than for an action anyone takes, so the mark after it is
	 * left out however much the table shows results.
	 */
	function cell($path, $title) {
		$effect = $this->effect($path);
		$held = isset($this->can[$path]) && $this->can[$path];
		$allowed = $this->fixed || $held;
		$shown = $this->effective && Permission::checkable($path);
		$said = $effect === "" ? self::INHERIT : $effect;

		return array(
			"path" => $path,
			"title" => $title,
			"effect" => $effect,
			"mark" => self::mark($effect),
			"label" => $title . ": " . $said,
			"locked" => !$this->editable,
			"field" => $this->editable,
			"shown" => $shown,
			"tone" => $allowed ? "ok" : "bad",
			"says" => $allowed ? "allowed" : "denied",
			"result" => $allowed ? self::TICK : self::CROSS,
		);
	}

	/** effect is what the subject's own entry says about a path, "" to inherit. */
	function effect($path) {
		if ($this->fixed) {
			return Policy::ALLOW;
		}

		return isset($this->own->entries[$path]) ? $this->own->entries[$path] : "";
	}

	/**
	 * extraTitle is what a rule a section added is called: the section's title
	 * and the rule's own, and the rule's alone where no section owns the path.
	 */
	static function extraTitle($extra, $sections) {
		$parts = explode("/", Row::text($extra, "key"));
		$title = Row::text($extra, "title");

		foreach ($sections as $section) {
			if ($section["key"] === $parts[0]) {
				return $section["title"] . ", " . $title;
			}
		}

		return $title;
	}

	/** mark is the glyph a cell draws for its effect; inherit draws none. */
	static function mark($effect) {
		if ($effect === Policy::ALLOW) {
			return self::TICK;
		}

		return $effect === Policy::DENY ? self::CROSS : "";
	}
}
