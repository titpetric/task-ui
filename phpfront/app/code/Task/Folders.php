<?php

namespace Task;

// Folders is the sidebar's folder tree: which folders it lists, in which
// order, folded under the project and the workspace each is in, and what
// every row of it draws.
//
// What it lists is the catalogue the reader sees together with the checkouts
// right inside each named workspace's src, less the projects set not to show
// there and the homes, each workspace cut to its sidebar limit, with the
// reader's pins placed among them. A row reads nothing for itself: the whole
// tree is one scope and the components walk it.
class Folders {
	/** FOLD is the number of jobs above which a folder's row starts folded. */
	const FOLD = 24;

	/** HOME is the folder of a workspace's home, which the tree leaves out. */
	const HOME = "home";

	/** SRC is where a named workspace keeps the checkouts of its projects. */
	const SRC = "src";

	var $db;
	var $access;

	/** workspaces is the folder of every named workspace, in the sidebar's order. */
	var $workspaces = array();

	/** here is the folder of the page being read, and job its job's key. */
	var $here = "";
	var $job = "";

	/** running is the ID of the run in progress of a job, by the job's key. */
	var $running = array();

	/**
	 * rows is the run in progress of a job as its row, by the job's key: the
	 * Jobs table draws the title of the run its button attaches to, which the
	 * ID alone does not say, and the tree read the rows already.
	 */
	var $rows = array();

	/**
	 * held is the reader's pins as they were read, which a Jobs row's [...]
	 * reads to know whether its Pin unpins.
	 */
	var $held = array();

	/** pinned is the pinned jobs a folder's row does not list, by folder. */
	var $pinned = array();

	/** dirs is the directory of a folder, asked of the host once for each. */
	var $dirs = array();

	/** total is how many jobs the whole tree lists, which decides a fold. */
	var $total = 0;

	function __construct($db = null, $access = null) {
		if ($db === null) {
			$db = DB::reader();
		}

		if ($access === null) {
			$access = new Access;
		}

		$this->db = $db;
		$this->access = $access;
	}

	/**
	 * tree is the whole sidebar, the groups the components draw: the projects
	 * of the root first, in the catalogue's order, then each named workspace
	 * with its own projects folded under it.
	 *
	 * here is the folder the page is about and job the key of its job, which
	 * is what marks a row as the one being read. A page that names no place
	 * passes neither.
	 */
	function tree($folders, $workspaces, $here = "", $job = "") {
		$this->workspaces = $workspaces;
		$this->here = $here;
		$this->job = $job;
		$this->running = $this->runs();
		$this->pinned = array();

		// The directories are asked for in two batches, because the tree is
		// decided in two steps: the limit of a workspace orders its projects
		// by their last commit, and a pin can put a folder in the tree that no
		// default listed. One crossing each, not one per row; see warm.
		$side = $this->defaults($folders);

		$this->warm($this->places($side));

		$limited = $this->limited($side);
		$pins = $this->pins($folders);
		$placed = $this->placed($limited, $pins);

		$this->warm($this->places($placed));

		$this->total = self::jobCount($placed);

		$groups = $this->group($placed, 0, true);
		$out = array();

		foreach ($this->ordered($groups) as $group) {
			$out[] = $this->scope($group);
		}

		return $out;
	}

	/**
	 * defaults is what the sidebar lists with no pins: the folders the reader
	 * sees, then the checkouts right inside a named workspace's src that no
	 * folder of the catalogue covers, which is a checkout with no pipeline
	 * file in it. Anything at or under a hidden project or a home stays out.
	 */
	function defaults($folders) {
		$hidden = $this->hidden();
		$listed = array();
		$side = array();

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");

			if (self::under($hidden, $id)) {
				continue;
			}

			$side[] = $folder;
			$listed[$id] = true;
		}

		foreach ($this->checkouts() as $id) {
			if (isset($listed[$id]) || self::under($hidden, $id) || !$this->access->sees($id)) {
				continue;
			}

			$side[] = self::bare($id);
		}

		return $side;
	}

	/**
	 * hidden is what the tree leaves out: a project set not to show in the
	 * sidebar, and the home of the root and of every named workspace.
	 */
	function hidden() {
		$projects = new Project($this->db);
		$out = $this->homes();

		foreach ($projects->hidden() as $folder) {
			$out[$folder] = true;
		}

		return $out;
	}

	/**
	 * homes is the home folder of the root and of each named workspace, which
	 * the tree leaves out and where a pin stays out of it too.
	 */
	function homes() {
		$out = array(self::HOME => true);

		foreach ($this->workspaces as $workspace) {
			$out[$workspace . "/" . self::HOME] = true;
		}

		return $out;
	}

	/** checkouts is every directory right inside a named workspace's src. */
	function checkouts() {
		$out = array();

		foreach ($this->workspaces as $workspace) {
			foreach (Host::projects($workspace) as $id) {
				$out[] = $id;
			}
		}

		return $out;
	}

	/**
	 * limited is the folders less the projects past each workspace's sidebar
	 * limit, which lists the most recently committed first, a project with no
	 * commit last and ties by name. The folders inside a project go with it.
	 */
	function limited($side) {
		$limits = $this->limits();
		$dropped = array();

		foreach ($limits as $workspace => $limit) {
			foreach ($this->past($side, $workspace, $limit) as $project) {
				$dropped[$project] = true;
			}
		}

		if (count($dropped) === 0) {
			return $side;
		}

		$kept = array();

		foreach ($side as $folder) {
			$project = $this->projectOf(Row::text($folder, "id"));

			if (!isset($dropped[$project])) {
				$kept[] = $folder;
			}
		}

		return $kept;
	}

	/** limits is how many projects each workspace that limits them lists. */
	function limits() {
		$workspaces = new Workspace($this->db);

		return $workspaces->limits();
	}

	/** past is the projects of one workspace that its sidebar limit leaves out. */
	function past($side, $workspace, $limit) {
		if ($limit <= 0) {
			return array();
		}

		$projects = array();
		$seen = array();

		foreach ($side as $folder) {
			$id = Row::text($folder, "id");
			$project = $this->projectOf($id);

			if ($project === "" || isset($seen[$project]) || $this->workspaceOf($id) !== $workspace) {
				continue;
			}

			$seen[$project] = true;
			$projects[] = $project;
		}

		if (count($projects) <= $limit) {
			return array();
		}

		$at = array();

		foreach ($projects as $project) {
			$at[$project] = Host::commit($this->dirOf($project));
		}

		usort($projects, function($one, $other) use ($at) {
			return Folders::newer($at[$one], $at[$other], $one, $other);
		});

		return array_slice($projects, $limit);
	}

	/**
	 * newer orders two projects of a workspace the way its sidebar limit
	 * keeps them: the latest commit first, a project with no commit last, and
	 * two that committed at the same moment by name, ignoring case.
	 *
	 * A commit arrives in the shape Task\Clock calls STORED, in UTC, so the
	 * later of two is the greater as text.
	 */
	static function newer($a, $b, $one, $other) {
		if ($a === $b) {
			return strcmp(strtolower($one), strtolower($other));
		}

		if ($a === "" || $b === "") {
			return $b === "" ? -1 : 1;
		}

		return strcmp($b, $a);
	}

	/**
	 * pins is the reader's pins as the sidebar resolves them, in the order
	 * they were made.
	 *
	 * A pin resolves against the catalogue the reader sees, so a pin of a
	 * folder that is gone, or of one the reader no longer sees, is no pin.
	 */
	function pins($folders) {
		$out = array();

		if ($this->access->user === null) {
			return $out;
		}

		$store = new Pin($this->db);

		$this->held = $store->ofUser($this->access->user->getName());

		foreach ($this->held as $pin) {
			$held = $this->pin($folders, $pin);

			if ($held !== null) {
				$out[] = $held;
			}
		}

		return $out;
	}

	/**
	 * pin is one pin as the sidebar holds it, or null for one it cannot
	 * place: the folder of the catalogue the pin sits in, the job it names
	 * where it names one, and the arguments it runs with.
	 *
	 * A pin of a directory inside a folder is a folder of its own, titled
	 * after the directory; the directory has to be there. A pin of a job
	 * needs the job itself, which the catalogue resolves under its folder.
	 */
	function pin($folders, $pin) {
		$place = $pin->getFolder();
		$folder = self::resolve($folders, $place);

		if ($folder === null) {
			return null;
		}

		$id = Row::text($folder, "id");
		$rest = $place === $id ? "" : substr($place, strlen($id) + 1);

		if ($pin->getJob() !== "") {
			if ($rest !== "") {
				return null;
			}

			$job = Host::job($id, $pin->getDir(), $pin->getJob());

			if (!Row::flag($job, "found")) {
				return null;
			}

			return self::held($pin, $folder, $job);
		}

		if ($rest === "") {
			return self::held($pin, $folder, null);
		}

		if (!Host::isDir($id, $rest)) {
			return null;
		}

		$own = self::bare($place);

		return self::held($pin, $own, null);
	}

	/**
	 * resolve is the folder of the catalogue a place sits in: the folder
	 * itself for its own ID, else the longest one the place is under.
	 */
	static function resolve($folders, $place) {
		$best = null;

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");

			if ($place === $id) {
				return $folder;
			}

			if (strpos($place, $id . "/") !== 0) {
				continue;
			}

			if ($best === null || strlen($id) > strlen(Row::text($best, "id"))) {
				$best = $folder;
			}
		}

		return $best;
	}

	/** held is a resolved pin: the folder it sits in, the job it names, and what its row says and runs with. */
	static function held($pin, $folder, $job) {
		$args = Text::items($pin->getArgs());
		$written = count($args) === 0 ? "" : $pin->getArgs();

		return array(
			"folder" => $folder,
			"job" => $job,
			"named" => $pin->getJob(),
			"dir" => $pin->getDir(),
			"args" => $args,
			"argsJSON" => $written,
			"label" => self::label($pin, $args),
		);
	}

	/** label is what a pinned job's row says: its directory before it, the arguments it runs with after. */
	static function label($pin, $args) {
		$out = $pin->getJob();

		if ($pin->getDir() !== "") {
			$out = $pin->getDir() . "/ " . $out;
		}

		if (count($args) > 0) {
			$out .= " " . implode(" ", $args);
		}

		return $out;
	}

	/**
	 * placed is the default folders with the pins among them: a pinned folder
	 * in its place with the workspace and the project above it, a pinned job
	 * under its folder, and a folder listed for its pinned jobs alone listing
	 * only those. A pin in a home stays out, as the tree leaves a home out.
	 *
	 * The folders come back by ID, ignoring case, which is the order the
	 * sidebar lists them in whether a pin put one there or not.
	 */
	function placed($side, $pins) {
		$folders = $side;
		$at = array();

		foreach ($folders as $index => $folder) {
			$at[Row::text($folder, "id")] = $index;
		}

		$homes = $this->homes();
		$kept = array();

		foreach ($pins as $pin) {
			if (!self::under($homes, Row::text($pin["folder"], "id"))) {
				$kept[] = $pin;
			}
		}

		foreach ($kept as $pin) {
			$id = Row::text($pin["folder"], "id");

			if ($pin["job"] !== null || isset($at[$id])) {
				continue;
			}

			$at[$id] = count($folders);
			$folders[] = $pin["folder"];
		}

		$pinned = array();

		foreach ($kept as $pin) {
			if ($pin["job"] === null) {
				continue;
			}

			$id = Row::text($pin["folder"], "id");

			if (!isset($at[$id])) {
				$bare = $pin["folder"];
				$bare["jobs"] = array();
				$at[$id] = count($folders);
				$folders[] = $bare;
			}

			if (self::plain($pin, $folders[$at[$id]])) {
				continue;
			}

			$held = isset($pinned[$id]) ? $pinned[$id] : array();
			$held[] = $pin;
			$pinned[$id] = $held;
		}

		foreach ($kept as $pin) {
			foreach (self::above(Row::text($pin["folder"], "id"), $this->workspaces) as $id) {
				if (isset($at[$id])) {
					continue;
				}

				$at[$id] = count($folders);
				$folders[] = self::bare($id);
			}
		}

		$this->pinned = $pinned;

		usort($folders, array("Task\\Folders", "byName"));

		return $folders;
	}

	/** byName orders the tree's folders by ID, ignoring case. */
	static function byName($one, $other) {
		return strcmp(strtolower(Row::text($one, "id")), strtolower(Row::text($other, "id")));
	}

	/**
	 * plain reports whether a folder's row lists a pinned job already, which
	 * is a pin of one of the folder's own jobs with no directory and no
	 * arguments to set it apart.
	 */
	static function plain($pin, $folder) {
		if ($pin["dir"] !== "" || count($pin["args"]) > 0) {
			return false;
		}

		foreach (Row::items($folder, "jobs") as $job) {
			if (Row::text($job, "id") === $pin["named"]) {
				return true;
			}
		}

		return false;
	}

	/** above is the workspace and the project a folder is in, those above it. */
	static function above($id, $workspaces) {
		$parts = explode("/", $id);
		$levels = in_array($parts[0], $workspaces, true) ? 2 : 1;
		$out = array();

		for ($i = 1; $i <= $levels && $i < count($parts); $i++) {
			$out[] = implode("/", array_slice($parts, 0, $i));
		}

		return $out;
	}

	/** under reports whether a set holds a folder or one above it. */
	static function under($set, $id) {
		$at = $id;

		while ($at !== "") {
			if (isset($set[$at])) {
				return true;
			}

			$cut = strrpos($at, "/");
			$at = $cut === false ? "" : substr($at, 0, $cut);
		}

		return false;
	}

	/**
	 * group buckets the folders by their first depth+1 segments: a folder
	 * that deep is its group's top, a deeper one sits inside it. A group
	 * stands where its first folder does, its top or the first folder in it.
	 *
	 * named says whether a group this deep may be a workspace, which only the
	 * top level is: a workspace's projects are grouped a level down and no
	 * project of one is a workspace itself.
	 */
	function group($folders, $depth, $named) {
		$groups = array();
		$at = array();

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");
			$top = self::topOf($id, $depth);

			if (!isset($at[$top])) {
				$at[$top] = count($groups);
				$groups[] = array(
					"id" => $top,
					"workspace" => $named && in_array($top, $this->workspaces, true),
					"top" => null,
					"nested" => array(),
					"projects" => array(),
				);
			}

			$index = $at[$top];

			if (substr_count($id, "/") === $depth) {
				$groups[$index]["top"] = $folder;

				continue;
			}

			$nested = $groups[$index]["nested"];
			$nested[] = $folder;
			$groups[$index]["nested"] = $nested;
		}

		foreach ($groups as $index => $group) {
			if (!$group["workspace"]) {
				continue;
			}

			$groups[$index]["projects"] = $this->group($group["nested"], $depth + 1, false);
			$groups[$index]["nested"] = array();
		}

		return $groups;
	}

	/** topOf is the first depth+1 segments of a folder's ID, which name the group it is in. */
	static function topOf($id, $depth) {
		$parts = explode("/", $id);

		if (count($parts) <= $depth + 1) {
			return $id;
		}

		return implode("/", array_slice($parts, 0, $depth + 1));
	}

	/**
	 * ordered is the groups as the sidebar lists them: the projects of the
	 * root first, in the folders' order, then the workspaces by their rank
	 * and then by folder.
	 */
	function ordered($groups) {
		$ranks = Workspace::ranked($this->workspaces);
		$roots = array();
		$named = array();

		foreach ($groups as $group) {
			if ($group["workspace"]) {
				$named[] = $group;
			} else {
				$roots[] = $group;
			}
		}

		usort($named, function($one, $other) use ($ranks) {
			return Folders::byRank($one["id"], $other["id"], $ranks);
		});

		return array_merge($roots, $named);
	}

	/** byRank orders two workspaces by their place in the sidebar, then by folder. */
	static function byRank($one, $other, $ranks) {
		$a = isset($ranks[$one]) ? $ranks[$one] : 0;
		$b = isset($ranks[$other]) ? $ranks[$other] : 0;

		if ($a !== $b) {
			return $a < $b ? -1 : 1;
		}

		return strcmp($one, $other);
	}

	/**
	 * scope is one group as the components read it: the row of its top folder,
	 * what is folded under it, and whether it is a workspace.
	 *
	 * top says whether the folder the group is named after is listed at all.
	 * Without it the folders inside stand on their own, and the row's own
	 * fields go nowhere.
	 */
	function scope($group) {
		$top = $group["top"];

		if ($top === null) {
			$top = self::bare($group["id"]);
			$out = $this->row($top);
			$out["top"] = false;
		} else {
			$out = $this->row($top);
			$out["top"] = true;
		}

		$nested = array();
		$projects = array();

		foreach ($group["nested"] as $folder) {
			$nested[] = $this->nested($folder, $group["id"]);
		}

		foreach ($group["projects"] as $project) {
			$projects[] = $this->scope($project);
		}

		$out["workspace"] = $group["workspace"];
		$out["nested"] = $nested;
		$out["projects"] = $projects;

		if ($group["workspace"]) {
			$out["count"] = count($projects);
			$out["counted"] = true;
		}

		return $out;
	}

	/**
	 * nested is a folder inside a project: it names its path below the
	 * project, counts the pinned jobs with its own, and starts folded on a
	 * tree with more jobs than the sidebar shows at once.
	 */
	function nested($folder, $top) {
		$out = $this->row($folder);
		$here = $this->here === $out["id"];

		$out["active"] = $here;
		$out["here"] = $here;
		$out["icon"] = $here ? "folder-open" : "folder";
		$out["open"] = $here || $this->total <= self::FOLD;
		$out["dir"] = self::dirUnder($out["id"], $top);
		$out["count"] = count($out["jobs"]);
		$out["counted"] = true;

		return $out;
	}

	/**
	 * row is what every row of the tree carries: where it links, what its
	 * summary says, git's count of changed files under it, the badge a folder
	 * whose pipeline would not read draws, and the jobs it lists.
	 */
	function row($folder) {
		$id = Row::text($folder, "id");
		$dirty = Host::dirty($this->dirOf($id));
		$errors = Row::items($folder, "errors");
		$error = Row::text($folder, "error");
		$own = count(Row::items($folder, "jobs"));
		$badge = "!";
		$said = $error;

		if (count($errors) > 0) {
			$badge = "!" . count($errors);
			$said = Row::text($errors[0], "detail");
		}

		return array(
			"id" => $id,
			"name" => self::nameOf($id),
			"href" => Route::folderHref($id, $this->workspaces),
			"title" => $id . "/" . implode(", ", Row::items($folder, "sources")),
			"active" => $this->here === $id || strpos($this->here, $id . "/") === 0,
			"here" => $this->here === $id,
			"open" => true,
			"icon" => "project",
			"dirty" => $dirty,
			"changed" => self::changed($dirty),
			"dir" => "",
			"bad" => $error !== "",
			"badge" => $badge,
			"badTitle" => $said,
			"count" => $own,
			"counted" => $own > 0,
			"jobs" => $this->jobs($folder),
			"empty" => count($errors) > 0 ? "The pipeline has errors." : "Nothing to run.",
		);
	}

	/** jobs is what a folder's row lists: the jobs it declares, then the pins it lists besides them. */
	function jobs($folder) {
		$id = Row::text($folder, "id");
		$out = array();

		foreach (Row::items($folder, "jobs") as $job) {
			$out[] = $this->jobRow($folder, $job);
		}

		if (!isset($this->pinned[$id])) {
			return $out;
		}

		foreach ($this->pinned[$id] as $pin) {
			$out[] = $this->pinRow($folder, $pin);
		}

		return $out;
	}

	/**
	 * jobRow is a job of a folder's own pipeline as the sidebar draws it: its
	 * link, the convention its ID carries where its name does not say it, and
	 * the button that attaches to its run or starts one.
	 */
	function jobRow($folder, $job) {
		$id = Row::text($folder, "id");
		$name = Row::text($job, "name");
		$named = Row::text($job, "id");
		$key = self::key($id, Row::text($job, "dir"), $named);
		$attach = isset($this->running[$key]) ? $this->running[$key] : "";
		$kind = $named === $name ? "" : self::kindOf($named);

		return array(
			"key" => $key,
			"pinned" => false,
			"active" => $this->job === $key && $key !== "",
			"running" => $attach !== "",
			"attach" => $attach,
			"href" => Route::folderHref($id, $this->workspaces) . "/" . $named,
			"hint" => "",
			"kind" => $kind,
			"label" => $name,
			"can" => $this->canRun($folder, $job),
			"folder" => $id,
			"dir" => "",
			"job" => $named,
			"args" => "",
			"said" => $name,
			"name" => $name,
		);
	}

	/**
	 * pinRow is a pinned job under its folder, which the folder's row does
	 * not list: it says where it is and what it runs with, and it runs with
	 * the pin's directory and arguments.
	 */
	function pinRow($folder, $pin) {
		$id = Row::text($folder, "id");
		$job = $pin["job"];
		$named = Row::text($job, "id");
		$dir = Row::text($job, "dir");
		$label = $pin["label"];
		$key = self::key($id, $dir, $named);
		$attach = isset($this->running[$key]) ? $this->running[$key] : "";
		$href = Route::placeHref($id, $dir, $this->workspaces);

		return array(
			"key" => $key,
			"pinned" => true,
			"active" => $this->job === $key && $key !== "",
			"running" => $attach !== "",
			"attach" => $attach,
			"href" => $href . "/" . $named,
			"hint" => "Pinned: " . $id . " / " . $label,
			"kind" => "",
			"label" => $label,
			"can" => $this->canRun($folder, $job),
			"folder" => $id,
			"dir" => $pin["dir"],
			"job" => $named,
			"args" => $pin["argsJSON"],
			"said" => $label,
			"name" => Row::text($job, "name"),
		);
	}

	/** canRun reports whether the reader may start a job from its row in the tree. */
	function canRun($folder, $job) {
		return self::allows($this->access, $folder, $job);
	}

	/**
	 * allows is the rule a Run button is drawn by, which the right rail reads
	 * too: a terminal needs the shell permission, a job of a host the hosts
	 * one, and anything else the folder's and the job's own allow lists.
	 */
	static function allows($access, $folder, $job) {
		if (Row::flag($job, "shell") && !$access->allowed("shell")) {
			return false;
		}

		if (Row::flag($job, "host")) {
			return $access->allowed("hosts");
		}

		$allow = Row::items($folder, "allow");

		return $access->run(Row::text($folder, "id"), $allow, Row::items($job, "allow"));
	}

	/**
	 * runs is the ID of the run in progress of a job, by the job's key, of
	 * the runs the reader sees.
	 *
	 * The runner's memory holds every run in progress, whoever started it,
	 * and a row the reader may not see is not theirs to stop: the sidebar
	 * offers them the Run button, which is what the Go page offers, because
	 * it filters the same list through the same rule. Whether a run is seen
	 * is a permission over its row, so it is asked here and not of the host.
	 */
	function runs() {
		$out = array();

		$this->rows = $this->live();

		foreach ($this->rows as $key => $run) {
			$out[$key] = $run->getId();
		}

		return $out;
	}

	/**
	 * live is the run in progress of a job as the row holds it, by the job's
	 * key, of the runs the reader sees.
	 *
	 * The Jobs table draws the run's title on the button that attaches to it,
	 * so it reads the row and not the ID alone, and the filter is the one the
	 * sidebar reads through.
	 */
	function live() {
		$runs = new Run($this->db);
		$user = $this->access->user;
		$prompts = $this->access->allowed("prompt/view");
		$notes = $this->access->allowed("note/view");
		$out = array();

		foreach (Host::running() as $each) {
			$run = $runs->get(Row::text($each, "id"));

			if (!Run::seen($user, $run, $prompts, $notes)) {
				continue;
			}

			$out[Row::text($each, "key")] = $run;
		}

		return $out;
	}

	/** key is what a job's run is listed under: one run per job, per folder and directory. */
	static function key($folder, $dir, $job) {
		if ($dir === "") {
			return $folder . "/" . $job;
		}

		return $folder . "/" . $dir . "/" . $job;
	}

	/** kindOf is what a job's row says before its name: the part of its ID in front of the colon. */
	static function kindOf($id) {
		$cut = strpos($id, ":");

		return $cut === false ? $id : substr($id, 0, $cut);
	}

	/** nameOf is the last segment of a folder's ID, what its row shows at its depth. */
	static function nameOf($id) {
		$cut = strrpos($id, "/");

		return $cut === false ? $id : substr($id, $cut + 1);
	}

	/**
	 * ids is the ID of each folder of a list, which is what a binding takes
	 * for the folders the reader sees.
	 *
	 * Whether a directory inside a place is a folder in its own right, and so
	 * whether its row links its own page, is decided against this: a folder
	 * nobody sees is a directory of the place above it.
	 */
	static function ids($folders) {
		$out = array();

		foreach ($folders as $folder) {
			$out[] = Row::text($folder, "id");
		}

		return $out;
	}

	/**
	 * choices is a box per folder for the folder pick of the user forms, as
	 * components/folder-pick.tpl draws them: every folder of the catalogue,
	 * checked for the ones the user's row lists, then the rest of that list.
	 *
	 * A folder the user's row names that the catalogue has lost keeps a box,
	 * checked: the box is what the form puts back, so dropping it would take
	 * the folder off the user the first time anyone saved the form, which is
	 * not what the reader asked for.
	 *
	 * The folders are the whole catalogue and not the reader's own, because
	 * the form says what the user sees and not what the reader does.
	 */
	static function choices($folders, $seen) {
		$out = array();
		$known = array();

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");
			$known[$id] = true;
			$out[] = array("id" => $id, "checked" => in_array($id, $seen, true));
		}

		foreach ($seen as $id) {
			if (!isset($known[$id])) {
				$out[] = array("id" => $id, "checked" => true);
			}
		}

		return $out;
	}

	/**
	 * kept reports whether a name belongs to the layout of a workspace rather
	 * than to a project of it: its home, the src it keeps its checkouts in, or
	 * one of the names the root itself goes by.
	 *
	 * The root's page lists one row per project and none of these is one, so a
	 * folder of the catalogue named by any of them is left out of that list. It
	 * is not Route::reserved: that is the root's own names and nothing else,
	 * which is what a path of a page is read against.
	 */
	static function kept($name) {
		return $name === self::HOME || $name === self::SRC || Route::reserved($name);
	}

	/**
	 * safeID is the ID a directory of the root would be discovered under,
	 * which is its name with everything a URL path would have to escape turned
	 * into a dash and the dashes and dots at either end cut off.
	 *
	 * The root's own page links a directory by the ID it would have rather
	 * than by the folder it is, because a project the catalogue holds no
	 * folder for still has a page under the settings. A name of nothing else
	 * is the one discovery falls back to.
	 */
	static function safeID($name) {
		$id = preg_replace("/[^a-zA-Z0-9._-]+/", "-", $name);
		$id = trim($id, "-.");

		return $id === "" ? "default" : $id;
	}

	/**
	 * validID says an ID could name a folder at all: every segment is there
	 * and every one of them is already the ID discovery would give it.
	 *
	 * A page of a folder the catalogue does not hold draws the settings of a
	 * directory nobody discovered yet, which is a real page; a path that could
	 * never be a folder's is the 404 instead. folder/id.go:ValidID.
	 */
	static function validID($id) {
		if ($id === "") {
			return false;
		}

		foreach (explode("/", $id) as $segment) {
			if ($segment === "" || self::safeID($segment) !== $segment) {
				return false;
			}
		}

		return true;
	}

	/** dirUnder is the path a folder's row names below its project, "" for one right inside it. */
	static function dirUnder($id, $top) {
		$prefix = $top . "/";

		if (strpos($id, $prefix) !== 0) {
			return "";
		}

		$rest = substr($id, strlen($prefix));
		$cut = strrpos($rest, "/");

		return $cut === false ? "" : substr($rest, 0, $cut);
	}

	/** changed is what a lit icon says on a mouse-over, "" where git has nothing. */
	static function changed($dirty) {
		if ($dirty === 0) {
			return "";
		}

		$files = $dirty === 1 ? " file " : " files ";

		return $dirty . $files . "changed or untracked";
	}

	/** jobCount is how many jobs the whole tree lists, which decides whether a folder's row starts folded. */
	static function jobCount($folders) {
		$out = 0;

		foreach ($folders as $folder) {
			$out += count(Row::items($folder, "jobs"));
		}

		return $out;
	}

	/** dirOf is the directory a folder is in, asked of the host once for each. */
	function dirOf($id) {
		if (!isset($this->dirs[$id])) {
			$this->dirs[$id] = Host::folderDir($id, $this->workspaces);
		}

		return $this->dirs[$id];
	}

	/**
	 * warm asks the host for the directory of every ID it does not hold yet,
	 * in one crossing, so that dirOf answers the tree from memory.
	 *
	 * One crossing per row was 102ms of the 151ms the tree cost on every page:
	 * each handed the whole workspace list across the boundary again, for an
	 * answer that is a lookup behind it. An ID the host places nowhere is kept
	 * as the empty string dirOf answers, so a miss is not asked twice either.
	 */
	function warm($ids) {
		$ask = array();

		foreach ($ids as $id) {
			if ($id !== "" && !isset($this->dirs[$id])) {
				$ask[$id] = true;
			}
		}

		if (count($ask) === 0) {
			return;
		}

		$found = Host::folderDirs(array_keys($ask), $this->workspaces);

		foreach ($ask as $id => $_) {
			$this->dirs[$id] = Row::text($found, $id);
		}
	}

	/**
	 * places is what the rows of a tree are asked about: each folder, and the
	 * project at the top of the workspace it sits in, which the sidebar limit
	 * orders by its last commit.
	 */
	function places($folders) {
		$out = array();

		foreach ($folders as $folder) {
			$id = Row::text($folder, "id");
			$out[$id] = true;
			$project = $this->projectOf($id);

			if ($project !== "") {
				$out[$project] = true;
			}
		}

		return array_keys($out);
	}

	/**
	 * bare is a folder the catalogue does not hold: a checkout with no
	 * pipeline file in it, or one standing above a pin. It is titled after
	 * its last segment and lists nothing.
	 */
	static function bare($id) {
		return array(
			"id" => $id,
			"title" => self::nameOf($id),
			"dir" => "",
			"allow" => array(),
			"sources" => array(),
			"error" => "",
			"errors" => array(),
			"jobs" => array(),
		);
	}

	/** workspaceOf is the workspace a folder is in, the root's for anything outside a named one. */
	function workspaceOf($id) {
		return self::workspaceIn($id, $this->workspaces);
	}

	/**
	 * workspaceIn is the same against a list of the named workspaces, for a
	 * reader that has the list and not a tree.
	 */
	static function workspaceIn($id, $workspaces) {
		$parts = explode("/", $id);

		if (in_array($parts[0], $workspaces, true)) {
			return $parts[0];
		}

		return Route::INTERNAL;
	}

	/**
	 * projectOf is the checkout at the top of a workspace's src that holds a
	 * folder, "" for a workspace's own folder and for the root's.
	 */
	function projectOf($id) {
		$parts = explode("/", $id);

		if (in_array($parts[0], $this->workspaces, true)) {
			return count($parts) === 1 ? "" : $parts[0] . "/" . $parts[1];
		}

		if ($parts[0] === Route::INTERNAL) {
			return "";
		}

		return $parts[0];
	}
}
