<?php

namespace Task;

// Buttons is the row of actions on the right of a head: a link or a button
// each, the strong one in the primary colour and a destructive one in the bad
// colour, small in a block head and full size in a page head.
//
// A page builds one and hands it to the theme's buttons component. An item is
// the same shape as a menu's, so one place validates an attribute name and
// escapes its value, and the component draws the attributes raw.
class Buttons {
	/** small says the buttons are a block head's, which are the small ones. */
	var $small = false;

	var $items = array();

	/** pairs validates and escapes an item's attributes, as a menu's are. */
	var $pairs = null;

	function __construct($small = false) {
		$this->small = $small;
		$this->pairs = new Menu("");
	}

	/**
	 * item adds a button: an icon, a label, and pairs of an attribute and its
	 * value. href makes it a link, download sets the attribute, disabled turns
	 * the button off with the note as its hover; a data-*, aria-* or title pair
	 * reaches the markup, and any other name is refused.
	 */
	function item($icon, $label, $attrs = array()) {
		return $this->add($icon, $label, $attrs, false, false);
	}

	/** primary adds the button of the page's own action, in the strong colour. */
	function primary($icon, $label, $attrs = array()) {
		return $this->add($icon, $label, $attrs, true, false);
	}

	/** danger adds a button in the bad colour, which stops or removes something. */
	function danger($icon, $label, $attrs = array()) {
		return $this->add($icon, $label, $attrs, false, true);
	}

	/**
	 * small adds a button that is small in a row that is not, which is how a
	 * job's page draws its Pin beside an Attach of full size.
	 */
	function small($icon, $label, $attrs = array()) {
		return $this->add($icon, $label, $attrs, false, false, false, true);
	}

	/**
	 * icon adds a button that is the icon alone, which the theme draws without
	 * the chrome a button with a label carries.
	 *
	 * It takes no label, so the title and the aria-label say what it does: the
	 * icon says it to nobody who cannot see it.
	 */
	function icon($icon, $attrs = array()) {
		return $this->add($icon, "", $attrs, false, false, true);
	}

	/**
	 * faint adds a sentence where a button would be, which is what a reader
	 * who may not run something is told in place of the Run button.
	 *
	 * It is an item of the row and not a line beside it, because the row keeps
	 * its order: the note stands where the button would, and the buttons after
	 * it stay after it.
	 */
	function faint($text) {
		$item = $this->pairs->build("", "", array(), false);
		$item["primary"] = false;
		$item["bare"] = false;
		$item["small"] = false;
		$item["faint"] = $text;
		$this->items[] = $item;

		return $this;
	}

	/** when adds a button where the condition holds, and nothing where it does not. */
	function when($condition, $icon, $label, $attrs = array()) {
		if ($condition) {
			$this->item($icon, $label, $attrs);
		}

		return $this;
	}

	/** whenDanger adds a danger button where the condition holds. */
	function whenDanger($condition, $icon, $label, $attrs = array()) {
		if ($condition) {
			$this->danger($icon, $label, $attrs);
		}

		return $this;
	}

	/** empty reports whether the row holds no button. */
	function empty() {
		return count($this->items) === 0;
	}

	/** data is the scope the component reads. */
	function data() {
		return array("small" => $this->small, "items" => $this->items);
	}

	/** add is one button, its attribute values escaped. */
	function add($icon, $label, $attrs, $primary, $danger, $bare = false, $small = false) {
		$item = $this->pairs->build($icon, $label, $attrs, $danger);
		$item["primary"] = $primary;
		$item["bare"] = $bare;
		$item["small"] = $small;
		$item["faint"] = "";
		$this->items[] = $item;

		return $this;
	}
}
