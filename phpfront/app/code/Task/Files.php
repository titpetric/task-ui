<?php

namespace Task;

// Files is the Files panel of a place: the directory the page is about as a
// table of what it holds, each name linked to whatever opens it.
//
// A directory inside the place opens as a page of the same shape, and one that
// is a folder of the catalogue in its own right opens as that folder's page
// instead. A file opens in the preview window, which reads the bytes from
// /files/ at the root of the instance: that is a byte stream and no page of
// this front end, so its link carries no mount.
//
// The host answers the listing, so nothing here reads a directory. A folder of
// the catalogue inside the place that the listing does not hold is added as a
// row after the last directory it does hold, which is where the listing sorts
// a directory anyway.
class Files {
	/** FILES is where a file of a checkout answers, at the root and outside the mount. */
	const FILES = "/files/";

	/**
	 * panel is the table of a place: a folder's own directory, or a directory
	 * inside one, with the folders of the catalogue inside it linked to their
	 * own pages.
	 *
	 * up is the href of the row that goes a level up, "" for a page that draws
	 * none, and subs the folders of the catalogue the place holds.
	 */
	static function panel($entries, $id, $workspaces, $up, $subs) {
		$read = self::listed($entries, $id, $workspaces, true);
		$rows = self::folders($read["rows"], $read["held"], $subs, $workspaces);

		return array("up" => $up, "rows" => $rows);
	}

	/**
	 * home is the table of a workspace's home, the $HOME of every run in it:
	 * the same rows, each path written under ~/ because that is the name a run
	 * knows it by.
	 *
	 * The link reads the real path, because /files/ knows the folder and not
	 * the shell's name for it.
	 */
	static function home($entries, $id, $workspaces) {
		$rows = array();

		foreach ($entries as $entry) {
			$rows[] = self::row($entry, $id, $workspaces, true, false, "~/");
		}

		return array("up" => "", "rows" => $rows);
	}

	/**
	 * root is the table of the root's own directories, which is the listing and
	 * nothing more: a directory that is a folder opens as its page, and
	 * everything else is a name.
	 *
	 * The root is no checkout of its own, so there is no path to preview a file
	 * at and no directory to browse that is not a folder already.
	 *
	 * A directory of the root links by the ID it would be discovered under as
	 * well as by the folder it is: $named is every ID the page of a project or
	 * of a named workspace answers for, and a directory whose name makes one
	 * of them opens that page even where the catalogue holds no folder it is.
	 * The folder it is wins, because that page is the one the row is about.
	 */
	static function root($entries, $workspaces, $subs, $named) {
		$rows = array();
		$held = array();

		foreach ($entries as $entry) {
			$row = self::row($entry, "", $workspaces, false, true, "");
			$id = Folders::safeID($row["name"]);

			if ($row["dir"] && isset($named[$id]) && $row["href"] === "") {
				$row["href"] = Route::href(Route::folderHref($id, $workspaces));
			}

			$rows[] = $row;
			$folder = Row::text($entry, "folder");

			if ($folder !== "") {
				$held[$folder] = true;
			}
		}

		$rows = self::folders($rows, $held, $subs, $workspaces);

		return array("up" => "", "rows" => $rows);
	}

	/**
	 * listed is a listing as its rows, and the folders of the catalogue those
	 * rows link already, which is what decides the rows to add after them.
	 */
	static function listed($entries, $id, $workspaces, $linked) {
		$rows = array();
		$held = array();

		foreach ($entries as $entry) {
			$rows[] = self::row($entry, $id, $workspaces, $linked, true, "");
			$folder = Row::text($entry, "folder");

			if ($folder !== "") {
				$held[$folder] = true;
			}
		}

		return array("rows" => $rows, "held" => $held);
	}

	/**
	 * row is one entry: what opens it, how big it is and when it last changed.
	 *
	 * A sealed entry is listed and never opened, and so is one the folder
	 * refuses a path for, which is a symlink leading out of it. linked says the
	 * place has paths to open at all, which the root has not; prefix is the
	 * name a run knows the path by, which is ~/ in a workspace's home.
	 *
	 * catalogued says a directory that is a folder of the catalogue opens as
	 * that folder's page rather than as a directory of this place. A
	 * workspace's home does not: it is listed as the $HOME of its runs, and a
	 * checkout somewhere inside it is still a path under the home.
	 */
	static function row($entry, $id, $workspaces, $linked, $catalogued, $prefix) {
		$dir = Row::flag($entry, "dir");
		$folder = $catalogued ? Row::text($entry, "folder") : "";
		$path = $linked ? Row::text($entry, "path") : "";
		$stamp = Row::text($entry, "mod_time");
		$href = "";

		if ($folder !== "") {
			$href = Route::href(Route::folderHref($folder, $workspaces));
		} elseif ($path !== "" && $dir) {
			$href = Route::href(Route::folderHref($id, $workspaces)) . "/" . Route::escapePath($path);
		} elseif ($path !== "") {
			$href = self::FILES . Route::escapePath($id . "/" . $path);
		}

		$row = array(
			"name" => Row::text($entry, "name"),
			"dir" => $dir,
			"sealed" => Row::flag($entry, "sealed"),
			"exec" => Row::flag($entry, "exec"),
			"href" => $href,
			"path" => $path === "" ? "" : $prefix . $path,
			"size" => $dir ? "" : Size::bytes(Row::number($entry, "size")),
			"stamp" => Clock::read($stamp),
			"ago" => Clock::ago($stamp),
			"menu" => array(),
		);

		if (!$dir && $href !== "") {
			$row["menu"] = self::menu($row);
		}

		return $row;
	}

	/** menu is the [...] of a file: when it changed, and the two ways to open it. */
	static function menu($row) {
		$menu = new Menu("Actions on " . $row["name"]);

		$menu->detail("Changed", $row["stamp"]);
		$menu->item("eye", "Preview", array("href" => $row["href"], "data-file" => $row["path"]));
		$menu->item("download", "Download", array("href" => $row["href"] . "?download"));

		return $menu->data();
	}

	/**
	 * folders adds a row per folder of the catalogue inside the place that the
	 * listing does not hold, after the last directory it does: the listing
	 * sorts its directories first, so that is where one belongs.
	 *
	 * A folder the listing holds is linked by its own row already, which is
	 * what held says.
	 */
	static function folders($rows, $held, $subs, $workspaces) {
		$dirs = 0;

		foreach ($rows as $index => $row) {
			if ($row["dir"]) {
				$dirs = $index + 1;
			}
		}

		$missing = array();

		foreach ($subs as $sub) {
			$id = Row::text($sub, "id");

			if (isset($held[$id])) {
				continue;
			}

			$missing[] = self::folderRow($id, Row::text($sub, "dir"), $workspaces);
		}

		if (count($missing) === 0) {
			return $rows;
		}

		$head = array_slice($rows, 0, $dirs);
		$tail = array_slice($rows, $dirs);

		return array_merge($head, $missing, $tail);
	}

	/**
	 * folderRow is the row of a folder the listing did not hold: a directory
	 * that opens as its page, of no size, as old as the directory itself.
	 */
	static function folderRow($id, $dir, $workspaces) {
		$stamp = Host::modTime($dir);

		return array(
			"name" => Folders::nameOf($id),
			"dir" => true,
			"sealed" => false,
			"exec" => false,
			"href" => Route::href(Route::folderHref($id, $workspaces)),
			"path" => "",
			"size" => "",
			"stamp" => Clock::read($stamp),
			"ago" => Clock::ago($stamp),
			"menu" => array(),
		);
	}
}
