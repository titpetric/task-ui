<?php

namespace Task;

// HostJobs is the jobs that run a command on a host over ssh, as a job picker
// offers them: one per host task-ui has an ssh login to, largest capacity
// first, and one over all of them at the top.
//
// They belong to no folder's catalogue. The root declares them, which is why
// they carry the root's ID and sit in a group of their own, and why a page
// that offers them offers the root as a folder even where it holds no other
// job.
//
// A reader without the hosts permission is offered none: the page asks, and
// the host answers nothing rather than probing for a reader who may not see
// the answer.
class HostJobs {
	/** ALL is the host of the job that runs on every host with an ssh login. */
	const ALL = "*";

	/** PREFIX starts the ID of a job of the root that runs a command on a host. */
	const PREFIX = "@host:";

	/**
	 * of is the jobs as components/host-jobs.tpl reads them, an empty list
	 * where there are none, which is what closes the group.
	 */
	static function of($chrome, $selected = "") {
		$jobs = array();

		if (!$chrome->allowed("hosts")) {
			return $jobs;
		}

		$hosts = Host::reachable();

		if (count($hosts) === 0) {
			return $jobs;
		}

		$jobs[] = self::job(self::ALL, $selected);

		foreach ($hosts as $host) {
			$jobs[] = self::job($host, $selected);
		}

		return $jobs;
	}

	/** job is the one job that runs a command on a host, which it is named by. */
	static function job($host, $selected = "") {
		$id = self::PREFIX . $host;
		$said = "A command on " . $host . " over ssh";

		if ($host === self::ALL) {
			$said = "A command on every host with an ssh login, a run a host";
		}

		return array(
			"id" => $id,
			"name" => $id,
			"folder" => Route::INTERNAL,
			"title" => $said,
			"selected" => $id === $selected,
		);
	}
}
