<?php

namespace Task;

// Workflow reads the workflow table and its steps, and answers what a page
// asks of one: who its runs run as, whether the reader may start it, which
// step stops them, and which run of it is going.
//
// Nothing here writes. Every form of the workflow pages calls /api/workflows
// from the browser, so Create, Update and Delete stay in
// storage/workflow_storage.go.
class Workflow {
	var $db;

	/** flows is the runner's runs in progress by workflow ID, read once. */
	var $flows;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
		$this->flows = null;
	}

	/** all is every workflow, by name. */
	function all() {
		return Model\Workflow::rows($this->db->get_all("SELECT * FROM workflow ORDER BY name COLLATE NOCASE, id"));
	}

	/** get is one workflow by ID, or null. */
	function get($id) {
		if ($id === "") {
			return null;
		}

		$row = $this->db->get("SELECT * FROM workflow WHERE id = ?", $id);

		if (!is_array($row)) {
			return null;
		}

		return new Model\Workflow($row);
	}

	/** names is the name of every workflow, by ID, which names a step that runs one. */
	function names() {
		$out = array();

		foreach ($this->all() as $row) {
			$out[$row->getId()] = $row->getName();
		}

		return $out;
	}

	/** steps is a workflow's step rows, in order. */
	function steps($id) {
		if ($id === "") {
			return array();
		}

		return Model\WorkflowStep::rows($this->db->get_all("SELECT * FROM workflow_step WHERE workflow_id = ? ORDER BY position", $id));
	}

	/** stepsOf is a workflow's steps as a page draws them. */
	function stepsOf($id, $names) {
		return Step::of($this->steps($id), $names);
	}

	/**
	 * expand replaces each step that runs a workflow with that workflow's
	 * steps, at any depth, dropping the step of a workflow that is gone and
	 * reading each workflow once.
	 *
	 * What it answers are the jobs a run of this workflow would really walk,
	 * which is what decides whether a reader sees it at all.
	 */
	function expand($steps) {
		$seen = array();

		return $this->flatten($steps, $seen);
	}

	/** flatten is expand, with the workflows already read. */
	function flatten($steps, &$seen) {
		$flat = array();

		foreach ($steps as $step) {
			if (Step::kind($step) !== Step::WORKFLOW) {
				$flat[] = $step;

				continue;
			}

			$ref = $step["workflow_id"];

			if (isset($seen[$ref])) {
				continue;
			}

			$seen[$ref] = true;
			$inner = $this->stepsOf($ref, array());

			foreach ($this->flatten($inner, $seen) as $one) {
				$flat[] = $one;
			}
		}

		return $flat;
	}

	/** reaches reports whether a workflow runs the target through its steps; a workflow reaches itself. */
	function reaches($from, $target) {
		$seen = array();

		return $this->follows($from, $target, $seen);
	}

	/** follows is reaches, with the workflows already walked. */
	function follows($from, $target, &$seen) {
		if ($from === $target) {
			return true;
		}

		if (isset($seen[$from])) {
			return false;
		}

		$seen[$from] = true;
		$refs = $this->db->get_all("SELECT workflow_ref FROM workflow_step WHERE workflow_id = ? AND workflow_ref != '' ORDER BY position", $from);

		if (!is_array($refs)) {
			return false;
		}

		foreach ($refs as $row) {
			if ($this->follows(Row::text($row, "workflow_ref"), $target, $seen)) {
				return true;
			}
		}

		return false;
	}

	/**
	 * running is the ID of the workflow's run in progress, "" for none.
	 *
	 * The answer is the workflow runner's memory, through Task\Host::flows(),
	 * and not the workflow_run table: a run is written running and written
	 * through as it ends, so a run whose process died stays running in the
	 * table and every page would mark the workflow for good. The runner is
	 * the memory the Go page reads too, so both front ends mark the same
	 * workflows.
	 *
	 * The sidebar asks of every workflow, so the answer is folded once and
	 * kept: one host call a page.
	 */
	function running($id) {
		if ($id === "") {
			return "";
		}

		if ($this->flows === null) {
			$out = array();

			foreach (Host::flows() as $each) {
				$out[Row::text($each, "workflow_id")] = Row::text($each, "run_id");
			}

			$this->flows = $out;
		}

		return isset($this->flows[$id]) ? $this->flows[$id] : "";
	}

	/** schedules are the schedules that start the workflow, for the hint under its name. */
	function schedules($id, $schedules) {
		$out = array();

		foreach ($schedules as $schedule) {
			if ($schedule->getWorkflowId() === $id) {
				$out[] = $schedule;
			}
		}

		return $out;
	}

	/**
	 * visible is the workflows the reader may see, by name, each decorated:
	 * every page carries them, for the list under Workflows in the sidebar.
	 */
	function visible($chrome) {
		$names = array();
		$rows = $this->all();

		foreach ($rows as $row) {
			$names[$row->getId()] = $row->getName();
		}

		$out = array();

		foreach ($rows as $row) {
			$one = $this->view($chrome, $row, $names);

			if ($one !== null) {
				$out[] = $one;
			}
		}

		return $out;
	}

	/**
	 * sidebar is the workflows as the sidebar lists them under Workflows: the
	 * hidden ones among them, because the block draws for a reader with any
	 * workflow at all and the rows inside it leave the hidden out.
	 *
	 * title is the chain of a workflow's steps, which the row carries as the
	 * hover over its name.
	 */
	static function sidebar($flows, $workspaces, $path) {
		$out = array();

		foreach ($flows as $flow) {
			$texts = array();

			foreach (Chain::of($flow["steps"], $workspaces) as $link) {
				$texts[] = $link["text"];
			}

			$out[] = array(
				"id" => $flow["id"],
				"name" => $flow["name"],
				"hidden" => $flow["sidebarHidden"],
				"active" => $path === "/workflows/" . $flow["id"],
				"running" => $flow["running"] !== "",
				"canRun" => $flow["canRun"],
				"prompted" => $flow["prompted"],
				"title" => implode(" -> ", $texts),
			);
		}

		return $out;
	}

	/**
	 * view is a workflow as a page draws it, or null where the reader may not
	 * see a folder one of its steps runs in, the steps of the workflows it
	 * runs among them.
	 *
	 * canRun says the reader may run every step, and so may the user the
	 * workflow runs as; runDenied says which step stops whichever of them it
	 * is, and is empty where the run is refused for no step: the user the
	 * workflow runs as is held by nobody.
	 */
	function view($chrome, $row, $names) {
		$id = $row->getId();
		$steps = $this->stepsOf($id, $names);
		$jobs = $this->expand($steps);

		foreach ($jobs as $step) {
			if (!$chrome->access->sees($step["folder"])) {
				return null;
			}
		}

		$refusal = $this->check($chrome, $row, $steps);
		$denied = "";

		if ($refusal["user"] !== "") {
			$denied = "Run needs " . $refusal["step"] . ", which " . $refusal["user"] . " may not run";
		} elseif ($refusal["step"] !== "") {
			$denied = "Run needs " . $refusal["step"];
		}

		$runs = new WorkflowRun($this->db);
		$last = $runs->listByWorkflow($id, array("limit" => 1));

		return array(
			"id" => $id,
			"name" => $row->getName(),
			"description" => $row->getDescription(),
			"userId" => $row->getUserId(),
			"userName" => $row->getUserName(),
			"runsOver" => $row->getRunsOver(),
			"sidebarHidden" => (int)$row->getSidebarHidden() !== 0,
			"steps" => $steps,
			"running" => $this->running($id),
			"last" => count($last) === 0 ? null : $last[0],
			"runAs" => $this->runAs($row),
			"canRun" => !$refusal["gone"] && $denied === "",
			"runDenied" => $denied,
			"canChange" => $chrome->allowed("workflow/edit") && (User::admin($chrome->user) || ($row->getUserId() !== "" && $row->getUserId() === self::readerId($chrome))),
			"prompted" => self::prompted($steps),
			"promptPlace" => self::promptPlace($steps),
			"schedules" => array(),
		);
	}

	/** readerId is the signed in user's ID, "" when nobody is. */
	static function readerId($chrome) {
		return $chrome->user === null ? "" : $chrome->user->getId();
	}

	/**
	 * runAs names the user the workflow's runs run as: "" for whoever starts
	 * it, and for a user nobody holds any more, which the page says is a run
	 * the workflow needs changing for.
	 */
	function runAs($row) {
		if ($row->getRunAsUserId() === "") {
			return "";
		}

		$users = new User($this->db);
		$user = $users->get($row->getRunAsUserId());

		return $user === null ? "" : $user->getName();
	}

	/**
	 * check is why a start by the reader would be refused: the step neither
	 * they nor the user the workflow runs as may run, and whose refusal it is.
	 *
	 * All three empty lets the start go. The reader is asked first, so a step
	 * they may not run is named as theirs however the workflow is set up.
	 *
	 * gone is the user the workflow runs as being held by nobody. It refuses
	 * the start and names no step, so the page draws no Run button and says
	 * nothing in its place: the workflow needs changing, not a permission.
	 */
	function check($chrome, $row, $steps) {
		$out = array(
			"step" => "",
			"user" => "",
			"gone" => false,
		);

		$index = self::index($chrome->folders());
		$step = $this->mayRun($chrome->access, $row, $steps, $index);

		if ($step !== "") {
			$out["step"] = $step;

			return $out;
		}

		if ($row->getRunAsUserId() === "" || $row->getRunAsUserId() === self::readerId($chrome)) {
			return $out;
		}

		$users = new User($this->db);
		$runAs = $users->get($row->getRunAsUserId());

		if ($runAs === null) {
			$out["gone"] = true;

			return $out;
		}

		$access = $users->access($runAs);
		$step = $this->mayRun($access, $row, $steps, $index);

		if ($step !== "") {
			$out["step"] = $step;
			$out["user"] = $runAs->getName();
		}

		return $out;
	}

	/**
	 * mayRun is the first step of the workflow the access may not run, "" for
	 * none.
	 *
	 * A workflow that runs over a workspace counts its steps in each project,
	 * and the steps a project skips for lack of the job not at all.
	 */
	function mayRun($access, $row, $steps, $index) {
		$over = $this->over($row, $steps);

		if ($over["error"] !== "") {
			return $over["error"];
		}

		$seen = array();

		return $this->refuse($access, $over["steps"], $over["projects"], $index, $seen);
	}

	/** refuse is mayRun over steps already spread across the projects. */
	function refuse($access, $steps, $projects, $index, &$seen) {
		foreach ($steps as $step) {
			$kind = Step::kind($step);

			if ($kind === Step::WORKFLOW) {
				$refused = $this->nested($access, $step, $index, $seen);

				if ($refused !== "") {
					return $refused;
				}

				continue;
			}

			if (Step::asks($step)) {
				if (!$access->allowed("prompt/edit")) {
					return "the prompt in " . Step::place($step);
				}

				continue;
			}

			if ($step["disabled"]) {
				continue;
			}

			if (!self::mayRunJob($access, $step, $projects, $index)) {
				return $step["folder"] . " / " . $step["job"];
			}
		}

		return "";
	}

	/**
	 * nested checks the steps of the workflow a step runs, once per workflow,
	 * as that workflow's own user where it has one.
	 */
	function nested($access, $step, $index, &$seen) {
		$ref = $step["workflow_id"];

		if (isset($seen[$ref])) {
			return "";
		}

		$seen[$ref] = true;
		$inner = $this->stepsOf($ref, array());
		$row = $this->get($ref);

		if ($row === null) {
			return "workflow " . $ref;
		}

		$set = $access;

		if ($row->getRunAsUserId() !== "") {
			$users = new User($this->db);
			$user = $users->get($row->getRunAsUserId());

			if ($user === null) {
				return "workflow " . $row->getName() . ", whose user is gone";
			}

			$set = $users->access($user);
		}

		return $this->refuse($set, $inner, null, $index, $seen);
	}

	/**
	 * mayRunJob reports whether the access runs a job step: a job the
	 * catalogue holds that is no shell, with `hosts` for a host command,
	 * `shell` for a script and the folder's and the job's allow lists
	 * otherwise.
	 *
	 * A project of a workspace run over that lacks the job skips the step,
	 * which takes no permission: projects is the projects by ID for a
	 * workflow that runs over one, and null for a workflow that does not.
	 */
	static function mayRunJob($access, $step, $projects, $index) {
		$folder = $step["folder"];
		$held = $projects !== null && isset($projects[$folder]);

		if (!$held && !isset($index[$folder])) {
			return $projects !== null;
		}

		// The script job is the folder's own, synthesised for any folder
		// there is, which is why it is answered before the catalogue is asked.
		if ($step["job"] === Step::SCRIPT_JOB) {
			return $access->allowed("shell");
		}

		$job = Host::job($folder, "", $step["job"]);

		if (!Row::flag($job, "found")) {
			return $projects !== null;
		}

		if (Row::flag($job, "shell")) {
			return false;
		}

		if (Row::flag($job, "host")) {
			return $access->allowed("hosts");
		}

		return $access->run($folder, Row::items($job, "folder_allow"), Row::items($job, "allow"));
	}

	/**
	 * over is the steps a run of the workflow walks and the projects it walks
	 * them in, by ID: the steps as they stand, and no projects, for a
	 * workflow that runs over no workspace.
	 *
	 * A workspace with no project is a workflow that cannot run at all, and
	 * error is what the page says instead of Run.
	 */
	function over($row, $steps) {
		$out = array(
			"steps" => $steps,
			"projects" => null,
			"error" => "",
		);

		if ($row === null || $row->getRunsOver() === "") {
			return $out;
		}

		$workspace = $row->getRunsOver();
		$ids = self::projectIds($workspace);

		if (count($ids) === 0) {
			$out["steps"] = array();
			$out["error"] = "the workspace " . $workspace . " has no projects to run over";

			return $out;
		}

		$projects = array();
		$names = array();

		foreach ($ids as $id) {
			$projects[$id] = true;
			$split = self::projectOf($id, $workspace);
			$names[] = $split[0];
		}

		$out["steps"] = self::spread($steps, $workspace, $names);
		$out["projects"] = $projects;

		return $out;
	}

	/**
	 * projectIds is the folder ID of every project of a workspace, as the
	 * layout names them.
	 *
	 * The host answers a workspace and a directory name joined with a slash,
	 * which is the sidebar's question. The layout drops the workspace from a
	 * project of the root, so the ID of a project of `internal` is the bare
	 * name, and that is the ID a step of one carries.
	 */
	static function projectIds($workspace) {
		$out = array();

		foreach (Host::projects($workspace) as $path) {
			$name = $path;
			$cut = strrpos($path, "/");

			if ($cut !== false) {
				$name = substr($path, $cut + 1);
			}

			$out[] = self::join($workspace, $name);
		}

		return $out;
	}

	/**
	 * spread is the steps a run over the workspace's projects walks: every
	 * step once per project, a step in a project of the workspace moved to
	 * the same path inside each, and a prompt at the workspace root asked in
	 * each project's root.
	 */
	static function spread($steps, $workspace, $projects) {
		$listed = array();

		foreach ($projects as $project) {
			$listed[$project] = true;
		}

		$out = array();

		foreach ($projects as $project) {
			foreach ($steps as $step) {
				$split = self::projectOf($step["folder"], $workspace);

				if ($split[2] && isset($listed[$split[0]]) && Step::kind($step) !== Step::WORKFLOW) {
					$step["folder"] = self::join($workspace, $project . $split[1]);
				} elseif (Step::asks($step) && self::atRoot($step["folder"], $workspace)) {
					$step["folder"] = self::join($workspace, $project);
				}

				$out[] = $step;
			}
		}

		return $out;
	}

	/**
	 * atRoot reports whether a step's place is the root of the workspace the
	 * run walks: none, which is the workspaces root, or the workspace's own
	 * folder.
	 */
	static function atRoot($id, $workspace) {
		return $id === "" || $id === $workspace;
	}

	/**
	 * projectOf splits a folder ID of the workspace into its project's name,
	 * the path below it with its leading slash, and whether it is of the
	 * workspace at all.
	 */
	static function projectOf($id, $workspace) {
		$rel = $id;

		if ($workspace !== "" && $workspace !== Route::INTERNAL) {
			$prefix = $workspace . "/";

			if (strpos($id, $prefix) !== 0) {
				return array("", "", false);
			}

			$rel = substr($id, strlen($prefix));
		}

		if ($rel === "" || $rel === Route::INTERNAL) {
			return array("", "", false);
		}

		$cut = strpos($rel, "/");

		if ($cut === false) {
			return array($rel, "", true);
		}

		$tail = substr($rel, $cut + 1);

		return array(substr($rel, 0, $cut), $tail === "" ? "" : "/" . $tail, true);
	}

	/** join is the folder ID of a path inside a workspace, the root's names dropped. */
	static function join($workspace, $rel) {
		if ($workspace === "" || $workspace === Route::INTERNAL) {
			return $rel === "" ? Route::INTERNAL : $rel;
		}

		return $rel === "" ? $workspace : $workspace . "/" . $rel;
	}

	/** index is the catalogue by folder ID, which says whether a step's folder is there at all. */
	static function index($folders) {
		$out = array();

		foreach ($folders as $folder) {
			$out[Row::text($folder, "id")] = $folder;
		}

		return $out;
	}

	/** runs reports whether the workflow has something to run: an enabled step. */
	static function runs($steps) {
		foreach ($steps as $step) {
			if (!$step["disabled"]) {
				return true;
			}
		}

		return false;
	}

	/** prompted reports whether an enabled step asks Claude, so a run of it takes input. */
	static function prompted($steps) {
		foreach ($steps as $step) {
			if (!$step["disabled"] && Step::asks($step)) {
				return true;
			}
		}

		return false;
	}

	/**
	 * promptPlace is where the workflow's first prompt or summary step asks
	 * Claude, the place its thread shows before it starts: "" for no
	 * workspace, the root for a workflow that asks nowhere.
	 */
	static function promptPlace($steps) {
		foreach ($steps as $step) {
			if (Step::asks($step)) {
				return Step::where($step);
			}
		}

		return Route::INTERNAL;
	}
}
