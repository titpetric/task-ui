<?php

namespace Task;

// Pin reads the pin table: what a reader asked the sidebar to list besides
// the folders it lists anyway.
//
// A pin names a folder, a directory inside one, or a job of either with the
// arguments it runs with. What the sidebar does with one is Task\Folders.
class Pin {
	var $db;

	function __construct($db = null) {
		if ($db === null) {
			$db = DB::reader();
		}

		$this->db = $db;
	}

	/** ofUser is every pin of one reader, in the order they were made. */
	function ofUser($name) {
		if ($name === "") {
			return array();
		}

		$rows = $this->db->get_all("SELECT * FROM pin WHERE user_name = ? ORDER BY rowid", $name);

		return Model\Pin::rows($rows);
	}

	/**
	 * of is the reader's pin of a place, or of a job of it, or null for one
	 * they have not pinned.
	 *
	 * A pin that runs with arguments is a row of its own in the sidebar and no
	 * answer here: the item it would unpin is not the one the page draws.
	 */
	static function of($pins, $folder, $job, $dir) {
		foreach ($pins as $pin) {
			if ($pin->getFolder() !== $folder || $pin->getJob() !== $job || $pin->getDir() !== $dir) {
				continue;
			}

			if ($pin->getArgs() === "" || $pin->getArgs() === "[]") {
				return $pin;
			}
		}

		return null;
	}

	/**
	 * pin is what the action that pins what it names says and carries: a
	 * folder or a job of it, or the unpinning of one the reader pinned already.
	 *
	 * what is how the action names it, which is a job's ID in the Jobs table
	 * and the place itself in the head of a page. The attributes are the pairs
	 * and not the built string, because the caller builds them: an entry of a
	 * [...] and a button of a head draw the same pin.
	 */
	static function pin($pins, $folder, $job, $dir, $what) {
		$held = self::of($pins, $folder, $job, $dir);
		$attrs = array(
			"data-pin-folder" => $folder,
			"data-pin-job" => $job,
			"data-pin-dir" => $dir,
			"data-pin-key" => $folder . "|" . $dir . "|" . $job,
		);
		$label = "Pin " . $what;
		$title = "Show it in the sidebar";

		if ($held !== null) {
			$label = "Unpin " . $what;
			$title = "Take it out of the sidebar, unless the sidebar lists it anyway";
			$attrs["data-pin-id"] = $held->getId();
		}

		$attrs["title"] = $title;

		return array("label" => $label, "attrs" => $attrs);
	}

	/**
	 * item adds to a menu the entry that pins what it names, a folder or a job
	 * of it, and the one that unpins it where the reader pinned it already.
	 */
	static function item($menu, $pins, $folder, $job, $dir, $what) {
		$pin = self::pin($pins, $folder, $job, $dir, $what);

		$menu->item("pin", $pin["label"], $pin["attrs"]);

		return $menu;
	}

	/**
	 * button adds to a row the Pin of a job's own page, which is a button of
	 * the head and not an entry of a [...].
	 *
	 * It is small in a row that is not, because it stands beside an Attach or a
	 * Run of full size.
	 */
	static function button($buttons, $pins, $folder, $job, $dir, $what) {
		$pin = self::pin($pins, $folder, $job, $dir, $what);

		$buttons->small("pin", $pin["label"], $pin["attrs"]);

		return $buttons;
	}
}
