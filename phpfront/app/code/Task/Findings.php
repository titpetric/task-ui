<?php

namespace Task;

// Findings is what atkins' linter said about the pipelines of a place, which a
// page raises as one alert: atkins lists no job of a pipeline its linter
// refuses, so a folder with a broken file shows no jobs and this says why.
//
// A place is a folder and every folder inside it, and the root is every folder
// of no named workspace. The catalogue carries the findings, so nothing here
// reads a file or a database.
class Findings {
	/** ROWS is how many findings the alert shows before it folds the rest. */
	const ROWS = 5;

	/**
	 * alert is the findings of the folders in a place, null where there are
	 * none.
	 *
	 * A pipeline is counted once however many findings it has, because the
	 * sentence counts the files atkins refused and the list the reasons.
	 */
	static function alert($folders, $place, $workspaces) {
		$pipelines = array();
		$shown = array();
		$more = array();

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");
			$errors = Row::items($folder, "errors");

			if (count($errors) === 0 || !self::inPlace($place, $id, $workspaces)) {
				continue;
			}

			foreach ($errors as $finding) {
				$pipelines[$id . "\x00" . Row::text($finding, "path")] = true;
				$row = self::row($id, $finding, $workspaces);

				if (count($shown) < self::ROWS) {
					$shown[] = $row;

					continue;
				}

				$more[] = $row;
			}
		}

		$count = count($pipelines);

		if ($count === 0) {
			return null;
		}

		return array(
			"pipelines" => $count,
			"says" => $count === 1 ? "pipeline has" : "pipelines have",
			"shown" => $shown,
			"more" => $more,
			"rest" => count($more),
		);
	}

	/**
	 * note is the dashboard's one line about the same findings, null where
	 * there are none: how many pipelines the linter refused and the workspace
	 * of each, linked to the page that lists them.
	 *
	 * The dashboard names no place, so every folder the reader sees counts.
	 * It lists no finding, so it reads the workspaces the alert does not
	 * carry and none of the rows the alert is built out of.
	 */
	static function note($folders, $workspaces) {
		$pipelines = array();
		$named = array();
		$out = array();

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");
			$errors = Row::items($folder, "errors");

			if (count($errors) === 0) {
				continue;
			}

			$workspace = Folders::workspaceIn($id, $workspaces);

			if (!isset($named[$workspace])) {
				$named[$workspace] = true;
				$out[] = array(
					"name" => $workspace,
					"href" => Route::href(Route::folderHref($workspace, $workspaces)),
				);
			}

			foreach ($errors as $finding) {
				$pipelines[$id . "\x00" . Row::text($finding, "path")] = true;
			}
		}

		$count = count($pipelines);

		if ($count === 0) {
			return null;
		}

		return array(
			"pipelines" => $count,
			"says" => $count === 1 ? "pipeline has" : "pipelines have",
			"lists" => count($out) === 1 ? "Its page lists" : "Their pages list",
			"workspaces" => $out,
		);
	}

	/** row is one finding as the alert lists it: its folder, its file and its job. */
	static function row($id, $finding, $workspaces) {
		return array(
			"folder" => $id,
			"href" => Route::href(Route::folderHref($id, $workspaces)),
			"path" => Row::text($finding, "path"),
			"job" => Row::text($finding, "job"),
			"issue" => Row::text($finding, "issue"),
			"detail" => Row::text($finding, "detail"),
		);
	}

	/**
	 * inPlace reports whether a folder is the place or inside it: every folder
	 * for the empty place, every folder of no named workspace for the root,
	 * and the folder itself and its own below it otherwise.
	 */
	static function inPlace($place, $id, $workspaces) {
		if ($place === "") {
			return true;
		}

		if ($place === Route::INTERNAL) {
			return Folders::workspaceIn($id, $workspaces) === Route::INTERNAL;
		}

		return $id === $place || strpos($id, $place . "/") === 0;
	}
}
