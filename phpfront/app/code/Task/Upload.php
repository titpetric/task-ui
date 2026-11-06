<?php

namespace Task;

// Upload is where a file dropped into a markdown field lands in the artifact
// tree.
//
// A field of a page writes under prompts/, one directory per place, so a
// picture pasted into a prompt does not land at the top of the tree beside
// what the runs left. The artifact window names its own directory instead,
// which is the one the reader is editing in.
class Upload {
	/** PROMPTS is the directory every field but the artifact window's writes under. */
	const PROMPTS = "prompts";

	/** DEPTH is how many steps of a place name the directory: the workspace and the project. */
	const DEPTH = 2;

	/**
	 * ofPlace is the directory of the artifact tree a field writes into on a
	 * page about this place, which is `prompts/<workspace>-<project>`.
	 *
	 * A place is a folder ID and the directory inside it the page is about,
	 * `titpetric/mdedit/docs`, of which the first two steps are the workspace
	 * and the project. A place with one step, which the root is, names that
	 * step alone.
	 */
	static function ofPlace($place) {
		$steps = array();

		foreach (array_slice(explode("/", $place), 0, self::DEPTH) as $step) {
			$slug = self::slug($step);

			if ($slug !== "") {
				$steps[] = $slug;
			}
		}

		if (count($steps) === 0) {
			return self::PROMPTS . "/" . Route::INTERNAL;
		}

		return self::PROMPTS . "/" . implode("-", $steps);
	}

	/**
	 * slug is one step of a place as part of a directory name: lower case,
	 * and a run of anything else as one dash.
	 *
	 * The artifact store refuses a path with a dot file or a `..` in it, so a
	 * step of dots slugs to nothing and the step is dropped.
	 */
	static function slug($step) {
		$out = preg_replace("/[^a-z0-9]+/", "-", strtolower($step));

		return trim($out, "-");
	}
}
