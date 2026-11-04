<?php

namespace Task;

// Head is the scope the page head component reads: the heading, the sentence
// under it, the badge beside it and the actions on the right.
//
// Every page opens with one and the component reads every field, so the scope
// is written here and not as a literal on each page: a page states what it
// has and the rest stay the empty values the component draws nothing for.
//
// The lead is HTML, because a page writes a `<code>` into it. A sentence built
// out of what a reader typed goes through said(), which escapes it.
class Head {
	var $data;

	/**
	 * drawn says the div on the right is there however empty it is, which is
	 * what a folder's page and a job's do: their markup opens it before it
	 * knows whether a permission left anything in it.
	 */
	var $drawn = false;

	function __construct($title, $lead = "") {
		$this->data = array(
			"title" => $title,
			"icon" => "",
			"badge" => "",
			"tone" => "",
			"tags" => array(),
			"note" => "",
			"lead" => $lead,
			"leads" => false,
			"faint" => "",
			"live" => false,
			"run" => array(),
			"actions" => array(),
			"buttons" => array(),
		);
	}

	/**
	 * said is the sentence under the heading as text, escaped.
	 *
	 * A page that builds its lead out of a run's own values builds it out of a
	 * user name as well, which is what a reader typed.
	 */
	function said($text) {
		$this->data["lead"] = htmlspecialchars($text, ENT_QUOTES);

		return $this;
	}

	/**
	 * leads draws the paragraph under the heading whether there is anything in
	 * it or not, which is what a job's page does: its markup writes the job's
	 * description without asking whether the job has one.
	 *
	 * Every other page leaves the paragraph out where it says nothing, so this
	 * is asked for and not assumed.
	 */
	function leads() {
		$this->data["leads"] = true;

		return $this;
	}

	/**
	 * faint is a second sentence under the lead, in the faint colour: what a
	 * workflow is made of and who it runs as.
	 *
	 * It is text and not HTML, so it is escaped where the lead is not, and it
	 * is a paragraph of its own rather than the tail of the lead: the lead is
	 * what the page is for and this is what it is built out of.
	 */
	function faint($text) {
		$this->data["faint"] = $text;

		return $this;
	}

	/**
	 * live marks the actions as the ones the script redraws, which is the
	 * data-actions a workflow's page carries: a run that starts turns Run into
	 * Stop without a reload, and the script finds the div by that attribute.
	 */
	function live() {
		$this->data["live"] = true;

		return $this;
	}

	/**
	 * icon is the glyph before the heading, named without the i- prefix the
	 * sprite carries: a folder's page opens with the open folder.
	 */
	function icon($name) {
		$this->data["icon"] = $name;

		return $this;
	}

	/**
	 * badge is the label beside the heading: a run's state, or the word for the
	 * kind of document a page is.
	 *
	 * The tone is the colour, from Task\Run::tone(); no tone is the plain
	 * badge the report page draws.
	 */
	function badge($text, $tone = "") {
		$this->data["badge"] = $text;
		$this->data["tone"] = $tone;

		return $this;
	}

	/**
	 * tag is a plain badge after the first one, in the order the page adds
	 * them: a job's page states its kind, and then that it is interactive.
	 */
	function tag($text) {
		$this->data["tags"][] = $text;

		return $this;
	}

	/** note is the faint text after the badge: the title a terminal carries. */
	function note($text) {
		$this->data["note"] = $text;

		return $this;
	}

	/**
	 * run is the Run button of a job's page, from Task\Control::runButtons(),
	 * which is a split button and not one of the row: it sits before the
	 * buttons, where the job page draws it.
	 */
	function run($run) {
		$this->data["run"] = $run;

		return $this;
	}

	/** menu is the head's own [...], from Task\Menu::data(). */
	function menu($menu) {
		$this->data["actions"] = $menu;

		return $this;
	}

	/** buttons is the row of actions on the right, from Task\Buttons::data(). */
	function buttons($buttons) {
		$this->data["buttons"] = $buttons;

		return $this;
	}

	/**
	 * always draws the div on the right whether anything sits in it or not: a
	 * folder's page with no action the reader may take still holds the empty
	 * one, because that is what its markup says.
	 */
	function always() {
		$this->drawn = true;

		return $this;
	}

	/** data is the scope the component reads. */
	function data() {
		$this->data["show"] = $this->shows();

		return $this->data;
	}

	/**
	 * shows reports whether the head has anything on its right at all, which
	 * is the one div that holds both the buttons and the `[...]`.
	 *
	 * A head with a menu every item of which a permission took away has an
	 * empty menu and draws no div, the way a page with no menu draws none,
	 * unless the page called always().
	 */
	function shows() {
		if ($this->drawn) {
			return true;
		}

		if (count($this->data["run"]) > 0) {
			return true;
		}

		if (isset($this->data["buttons"]["items"]) && count($this->data["buttons"]["items"]) > 0) {
			return true;
		}

		return isset($this->data["actions"]["empty"]) && !$this->data["actions"]["empty"];
	}
}
