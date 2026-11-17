<?php

namespace Task\Schedules;

// Catalogue is what a schedule's form offers: the folders and jobs the reader
// may run, and the workflows every folder of whose steps the reader sees.
//
// A job is offered where the reader may run it, which takes the allow list of
// the job and of its folder together, and a folder where it holds one such
// job. A schedule is written with the job's ID and the picker shows its name,
// which differ on a folder that follows more than one convention.
//
// The jobs that run a command on a host belong to no catalogue, so they come
// last and in a group of their own, and the root comes with them: a reader
// with the hosts permission is offered it even where it holds no other job.
class Catalogue {
	/**
	 * of is the folders, the jobs and the hosts a schedule may name, with the
	 * folder and the job a query named selected.
	 *
	 * A job query with no folder names a folder by being in it, so the two are
	 * read together: the folder is the one the job was found in, and the host
	 * jobs are the root's.
	 */
	static function of($chrome, $pickFolder = "", $pickJob = "") {
		$folders = array();
		$jobs = array();
		$hosts = \Task\HostJobs::of($chrome, $pickJob);

		foreach ($chrome->folders() as $folder) {
			$id = \Task\Row::text($folder, "id");
			$allow = \Task\Row::items($folder, "allow");
			$runnable = array();

			foreach (\Task\Row::items($folder, "jobs") as $job) {
				if ($chrome->access->run($id, $allow, \Task\Row::items($job, "allow"))) {
					$runnable[] = array(
						"id" => \Task\Row::text($job, "id"),
						"name" => \Task\Row::text($job, "name"),
						"folder" => $id,
					);
				}
			}

			if (count($runnable) === 0) {
				continue;
			}

			$folders[] = array("id" => $id);
			$jobs = array_merge($jobs, $runnable);
		}

		if (count($hosts) > 0) {
			$internal = \Task\Route::INTERNAL;
			$folders[] = array("id" => $internal);
		}

		$pick = self::picked($pickFolder, $pickJob, $jobs);

		foreach ($folders as $at => $folder) {
			$folders[$at]["selected"] = $folder["id"] === $pick;
		}

		foreach ($jobs as $at => $job) {
			$jobs[$at]["selected"] = $job["folder"] === $pick && $job["id"] === $pickJob;
		}

		return array(
			"folders" => $folders,
			"jobs" => $jobs,
			"hosts" => $hosts,
		);
	}

	/**
	 * picked is the folder a job names: the folder the query named, or the one
	 * the job was found in where the query named only the job.
	 */
	static function picked($folder, $job, $jobs) {
		if ($folder !== "" || $job === "") {
			return $folder;
		}

		if (strpos($job, \Task\HostJobs::PREFIX) === 0) {
			return \Task\Route::INTERNAL;
		}

		foreach ($jobs as $offered) {
			if ($offered["id"] === $job) {
				return $offered["folder"];
			}
		}

		return "";
	}

	/**
	 * workflows is the workflows the reader may schedule, each named as its
	 * own name or its ID where it has none.
	 */
	static function workflows($chrome) {
		$lookup = new Lookup;
		$out = array();

		foreach (self::visible($chrome, $lookup->workflows()) as $id => $workflow) {
			$out[] = array("id" => $id, "name" => $workflow["name"] === "" ? $id : $workflow["name"]);
		}

		return $out;
	}

	/**
	 * visible is the workflows the reader may see: those every folder of whose
	 * steps they see, the steps of a workflow a step runs included.
	 */
	static function visible($chrome, $workflows) {
		$out = array();

		foreach ($workflows as $id => $workflow) {
			$seen = true;

			foreach ($workflow["folders"] as $folder) {
				if (!\Task\User::sees($chrome->user, $folder)) {
					$seen = false;
				}
			}

			if ($seen) {
				$out[$id] = $workflow;
			}
		}

		return $out;
	}
}
