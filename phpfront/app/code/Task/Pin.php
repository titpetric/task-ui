<?php

namespace Task;

// Pin draws a pin: what a reader asked the sidebar to list besides the folders
// it lists anyway.
//
// A pin names a folder, a directory inside one, or a job of either with the
// arguments it runs with. What the sidebar does with one is Task\Folders, and
// the rows are the folders module's, in Task\Folders\Pins.
class Pin {
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
		$title = self::says($job, false);

		if ($held !== null) {
			$label = "Unpin " . $what;
			$title = self::says($job, true);
			$attrs["data-pin-id"] = $held->getId();
		}

		$attrs["title"] = $title;

		return array("label" => $label, "attrs" => $attrs);
	}

	/**
	 * says is what a pin of a folder or of a job does, and what removing it
	 * does: a folder brings its project to the sidebar, which lists workspaces
	 * and projects, and a job is marked out on the page of the folder that
	 * runs it.
	 */
	static function says($job, $pinned) {
		if ($job !== "") {
			return $pinned ? "Stop marking it out on the folder's page" : "Mark it out on the folder's page";
		}

		if ($pinned) {
			return "Take its project out of the sidebar, unless the sidebar lists it anyway";
		}

		return "Show its project in the sidebar";
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
