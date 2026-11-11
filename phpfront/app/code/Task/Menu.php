<?php

namespace Task;

// Menu is the [...] menu of a row, or of a block: an icon button that opens a
// list on a click, its details on top, then its items in groups, the danger
// group last.
//
// A page builds one and hands it to the theme's menu component. Nothing but
// this escapes an attribute value, so the component draws the attributes raw.
class Menu {
	/** ATTRS is the attributes an item may carry besides its own fields. */
	const ATTRS = '/^(data-[a-z0-9-]+|aria-[a-z]+|title)$/';

	var $label = "";
	var $details = array();
	var $items = array();

	function __construct($label) {
		$this->label = $label;
	}

	/**
	 * detail adds a line to the block on top: what the row holds that its
	 * columns leave out.
	 */
	function detail($label, $value, $mono = false, $title = "") {
		$this->details[] = array(
			"label" => $label,
			"value" => $value,
			"mono" => $mono,
			"title" => $title,
		);

		return $this;
	}

	/**
	 * item adds an entry: an icon, a label, and pairs of an attribute and its
	 * value. href makes it a link, download sets the attribute, note is the
	 * faint text after the label, disabled turns it off with the note; a
	 * data-*, aria-* or title pair reaches the markup, and any other name is
	 * refused, so a misspelling fails rather than drawing nothing.
	 *
	 * A null value is the attribute on its own, with no `=""` after it: what
	 * the script reads of `data-run` is whether it is there.
	 */
	function item($icon, $label, $attrs = array()) {
		$this->items[] = $this->build($icon, $label, $attrs, false);

		return $this;
	}

	/** danger adds an entry of the last group, in the bad colour. */
	function danger($icon, $label, $attrs = array()) {
		$this->items[] = $this->build($icon, $label, $attrs, true);

		return $this;
	}

	/** sep ends a group: the component draws a line before the next item. */
	function sep() {
		$this->items[] = array("sep" => true);

		return $this;
	}

	/** when adds an entry only where the condition holds. */
	function when($condition, $icon, $label, $attrs = array()) {
		if ($condition) {
			$this->item($icon, $label, $attrs);
		}

		return $this;
	}

	/** whenDanger adds a danger entry only where the condition holds. */
	function whenDanger($condition, $icon, $label, $attrs = array()) {
		if ($condition) {
			$this->danger($icon, $label, $attrs);
		}

		return $this;
	}

	/** empty reports whether the menu has nothing to show. */
	function empty() {
		if (count($this->details) > 0) {
			return false;
		}

		foreach ($this->items as $item) {
			if (!isset($item["sep"])) {
				return false;
			}
		}

		return true;
	}

	/** data is the scope the component reads. */
	function data() {
		return array(
			"label" => $this->label,
			"empty" => $this->empty(),
			"details" => $this->details,
			"groups" => $this->groups(),
		);
	}

	/**
	 * groups is the items between separators, empty groups left out, then the
	 * danger items as the last group.
	 */
	function groups() {
		$groups = array();
		$group = array();
		$danger = array();

		foreach ($this->items as $item) {
			if (isset($item["sep"])) {
				if (count($group) > 0) {
					$groups[] = $group;
				}

				$group = array();

				continue;
			}

			if ($item["danger"]) {
				$danger[] = $item;

				continue;
			}

			$group[] = $item;
		}

		if (count($group) > 0) {
			$groups[] = $group;
		}

		if (count($danger) > 0) {
			$groups[] = $danger;
		}

		return $groups;
	}

	/**
	 * build is one item, its attribute values escaped.
	 *
	 * A null value is the attribute alone, with no `=""`: `data-run` is read
	 * by its presence and the Go template writes it bare, so a button that
	 * carries it is one attribute and not a pair. An empty string is a pair
	 * with an empty value, which is a different thing and still written.
	 */
	function build($icon, $label, $attrs, $danger) {
		$item = array(
			"icon" => $icon,
			"label" => $label,
			"note" => "",
			"href" => "",
			"download" => false,
			"disabled" => false,
			"danger" => $danger,
			"attrs" => "",
		);
		$written = array();

		foreach ($attrs as $name => $value) {
			switch ($name) {
				case "href":
					$item["href"] = $value;
					break;
				case "download":
					$item["download"] = true;
					break;
				case "note":
					$item["note"] = $value;
					break;
				case "disabled":
					$item["disabled"] = true;

					if ($item["note"] === "") {
						$item["note"] = $value;
					}

					break;
				default:
					if (preg_match(self::ATTRS, $name) !== 1) {
						throw new \Exception("menu item " . $label . ": attribute " . $name . " is not data-*, aria-* or title");
					}

					if ($value === null) {
						$written[] = $name;

						break;
					}

					$written[] = $name . '="' . htmlspecialchars($value, ENT_QUOTES) . '"';
			}
		}

		$item["attrs"] = implode(" ", $written);

		return $item;
	}
}
