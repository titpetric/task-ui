<?php

namespace Task;

// Fixture builds the shapes the host answers, for a fixture that drives one.
//
// A fixture writes what its assertions turn on and leaves the rest of a shape
// at its empty value, so a key added to a host answer is added here and in
// phpfront/convert.go and no fixture is touched.
//
// Nothing requires this at runtime: the real answers are Go. Keep it in step
// with phpfront/host.go.
class Fixture {
	/** folder is one entry of the catalogue, with the jobs it declares. */
	static function folder($id, $jobs = array(), $allow = array()) {
		return array(
			"id" => $id,
			"title" => $id,
			"dir" => "",
			"allow" => $allow,
			"jobs" => $jobs,
		);
	}

	/**
	 * job is one job of a folder: the ID a schedule is written with, the name
	 * a picker shows, the allow list that narrows who may run it, and the
	 * kind where it is one of the two a permission gates rather than an allow
	 * list - "shell" for a terminal, "host" for a job on a host.
	 */
	static function job($id, $name = "", $allow = array(), $kind = "") {
		$called = $name === "" ? $id : $name;

		return array(
			"id" => $id,
			"name" => $called,
			"dir" => "",
			"shell" => $kind === "shell",
			"host" => $kind === "host",
			"allow" => $allow,
		);
	}

	/**
	 * resolves writes the job a folder answers at a directory inside it, for
	 * Task\Host::job to find.
	 *
	 * The host walks a Taskfile to answer it, so a fixture states the answer
	 * rather than a catalogue to resolve: the folder, the directory the run
	 * ran in, the job, and the allow list of the folder it was found in.
	 */
	static function resolves($folder, $dir, $job, $folderAllow = array()) {
		$key = $folder . "|" . $dir . "|" . Row::text($job, "id");

		Host::$jobs[$key] = array(
			"found" => true,
			"id" => Row::text($job, "id"),
			"name" => Row::text($job, "name"),
			"dir" => $dir,
			"shell" => Row::flag($job, "shell"),
			"host" => Row::flag($job, "host"),
			"allow" => Row::items($job, "allow"),
			"folder_allow" => $folderAllow,
		);
	}

	/**
	 * declared writes every job of a catalogue entry as one the folder itself
	 * resolves, which is what a page gating a run of a declared job reads.
	 *
	 * A fixture that writes a catalogue and then gates a run of it wants both
	 * answers to agree, and the host's two calls do agree: the listing and
	 * the lookup read the same Taskfile.
	 */
	static function declared($folder) {
		$id = Row::text($folder, "id");
		$allow = Row::items($folder, "allow");

		foreach (Row::items($folder, "jobs") as $job) {
			self::resolves($id, "", $job, $allow);
		}

		return $folder;
	}
}
