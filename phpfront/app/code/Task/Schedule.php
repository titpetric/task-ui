<?php

namespace Task;

// Schedule answers what a row of the schedule table means: what it starts,
// whether it is on, the arguments its runs pass the job, and what a table of
// them draws.
//
// The row behaviour sits here rather than on Task\Model\Schedule because the
// runtime merges no methods into a generated class: one class owns the table
// and its rows. The rows are the schedules module's, in Task\Schedules\Schedules.
class Schedule {
	/** HOST starts the job of a schedule that runs a command on a host. */
	const HOST = "@host:";

	/** enabled reports whether a row is on. */
	static function enabled($schedule) {
		return $schedule !== null && (int)$schedule->getEnabled() !== 0;
	}

	/** triages reports whether a failed run of a row gets Claude's triage. */
	static function triages($schedule) {
		return $schedule !== null && (int)$schedule->getTriage() !== 0;
	}

	/** isWorkflow reports whether a row starts a workflow; one of a job names a folder and a job. */
	static function isWorkflow($schedule) {
		return $schedule !== null && $schedule->getWorkflowId() !== "";
	}

	/** instance reports whether a row names a job of the instance itself, which has no page to link. */
	static function instance($schedule) {
		return $schedule !== null && $schedule->getFolder() === Route::INTERNAL && strpos($schedule->getJob(), "@") === 0;
	}

	/**
	 * table is the schedules of one job as components/sched-table.tpl reads
	 * them: the cells of each, and the [...] of what may be done to it.
	 *
	 * can is the reader's permission over the whole table and not a row's,
	 * because a schedule of a job is changed by whoever may run that job, and
	 * what names what the schedules are for, which a Delete asks about.
	 */
	static function table($chrome, $schedules, $users, $can, $what) {
		$rows = array();

		foreach ($schedules as $schedule) {
			$row = self::cells($chrome, $schedule, "", $users);
			$row["menu"] = self::menu($schedule, $row["on"], $what, $can);
			$rows[] = $row;
		}

		return array("can" => $can, "rows" => $rows);
	}

	/**
	 * cells is what both tables of a schedule draw: the expression, when it
	 * runs next and when it last ran, who it runs as and whether it is on.
	 *
	 * The schedules page heads two columns more with the folder and the job
	 * and a job's page heads none, so the two tables are two components over
	 * the one set of cells.
	 *
	 * fixed is the user a schedule of a workflow runs as, which the workflow
	 * holds and no cell picks; "" is a schedule that runs as the user it was
	 * written by, and then an admin picks another.
	 */
	static function cells($chrome, $schedule, $fixed, $users) {
		$name = $fixed === "" ? $schedule->getUserName() : $fixed;
		$on = self::enabled($schedule);
		$last = $schedule->getLastRunAt();

		return array(
			"expression" => $schedule->getExpression(),
			"on" => $on,
			"next" => $on ? explode(" ", Clock::read($schedule->getNextRunAt())) : array(),
			"last" => $last === "" ? "" : Clock::ago($last),
			"user" => $chrome->who()->of($name),
			"pick" => $fixed === "" && count($users) > 0,
			"options" => array(
				"users" => $users,
				"selected" => $name,
				"short" => true,
			),
			"api" => "/api/schedules/" . $schedule->getId(),
			"triage" => self::triages($schedule),
		);
	}

	/**
	 * menu is a row's [...]: Edit, where edits says the page holds the form
	 * that saves one, then Enable or Disable, by on, whether the schedule
	 * runs, and Delete, whose confirm names what the schedule is for.
	 *
	 * Edit carries the expression rather than a page of its own: the form is
	 * already on the page and task-ui.js turns it into a save of this row.
	 * The schedules page draws that form only for a reader with a job to
	 * schedule, which is what edits answers.
	 *
	 * The endpoints are the API's at the root and carry no mount, as every
	 * data-* endpoint of a page does.
	 */
	static function menu($schedule, $on, $what, $edits = true) {
		$id = $schedule->getId();
		$expression = $schedule->getExpression();
		$menu = new Menu("Actions on the schedule " . $expression);

		$menu->when($edits, "edit", "Edit", array("data-sched-edit" => "/api/schedules/" . $id, "data-sched-expression" => $expression));
		$menu->item($on ? "stop" : "play", $on ? "Disable" : "Enable", array(
			"data-api" => "/api/schedules/" . $id,
			"data-method" => "PUT",
			"data-body" => $on ? '{"enabled":false}' : '{"enabled":true}',
			"data-next" => "reload",
			"data-notice" => $on ? "Schedule disabled" : "Schedule enabled",
		));
		$menu->danger("trash", "Delete", array(
			"data-delete" => "/api/schedules/" . $id,
			"data-confirm" => "Delete the schedule " . $expression . " for " . $what . "?",
			"data-confirm-ok" => "Delete",
			"data-next" => "reload",
			"data-notice" => "Schedule deleted",
		));

		return $menu->data();
	}

	/**
	 * argLine is the arguments a row's runs pass the job, the JSON list the
	 * column holds, as the line its form takes back: the words, quoted where a
	 * word needs it, or the command unchanged for a job of @host:<host>.
	 */
	static function argLine($job, $args) {
		$list = Text::items($args);

		if (strpos($job, self::HOST) === 0) {
			return implode(" ", $list);
		}

		$words = array();

		foreach ($list as $word) {
			$words[] = self::quotes($word) ? "'" . str_replace("'", "'\\''", $word) . "'" : $word;
		}

		return implode(" ", $words);
	}

	/**
	 * quotes reports whether a word needs quoting on the line: an empty one,
	 * or one holding a space, a tab, a quote of either kind or a backslash.
	 *
	 * The characters are walked one at a time because strpbrk is not bound.
	 */
	static function quotes($word) {
		if ($word === "") {
			return true;
		}

		foreach (array(" ", "\t", "'", '"', "\\") as $char) {
			if (str_contains($word, $char)) {
				return true;
			}
		}

		return false;
	}
}
