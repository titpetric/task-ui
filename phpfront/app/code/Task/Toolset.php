<?php

namespace Task;

// Toolset is the tools task-ui found in its PATH at the last check, as the
// Tooling section of the Instance page lists them: a table per category, in
// the order the page draws them, each tool with the version and the path the
// check recorded.
//
// The tool table holds a tool's name, its category, its version, its path and
// what it is used for, and no identifier, so the identifier a check of one
// tool is asked under is here. Keep the names and their identifiers in step
// with toolenv/toolenv.go, which probes them, and the titles in step with
// frontend/view/instance.go, which names the categories: a probe added there
// and not here lists with no Check button.
class Toolset {
	/** The categories a probe belongs to, as the table writes the column. */
	const RUNTIME = "runtime";

	const TOOLING = "tooling";

	const CUSTOM = "custom";

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
	}

	/**
	 * categories are the categories of the section with the titles the page
	 * gives them, in the order it lists them: the runtimes a job runs on, the
	 * tools of the image, then the instance's own.
	 */
	static function categories() {
		return array(
			array("category" => self::RUNTIME, "title" => "Runtimes"),
			array("category" => self::TOOLING, "title" => "Tooling"),
			array("category" => self::CUSTOM, "title" => "Custom tooling"),
		);
	}

	/**
	 * ids is the identifier of every probe by the name it is recorded under,
	 * which is what a check of one tool is asked under.
	 *
	 * The one that differs from its name is docker compose: it is a
	 * subcommand of docker and answers under docker-compose, because a path
	 * of the API holds no space.
	 */
	static function ids() {
		return array(
			"node" => "node",
			"phpscript" => "phpscript",
			"go" => "go",
			"docker" => "docker",
			"docker compose" => "docker-compose",
			"git" => "git",
			"make" => "make",
			"worktree" => "worktree",
			"atkins" => "atkins",
			"task" => "task",
			"semver" => "semver",
			"secretsmanager" => "secretsmanager",
			"terraform" => "terraform",
			"jq" => "jq",
		);
	}

	/** tools is every tool the last check recorded, by category and then by name. */
	function tools() {
		$found = $this->db->get_all("SELECT * FROM tool ORDER BY category, name");

		return Model\Tool::rows($found);
	}

	/**
	 * groups are the tables of the section: the categories in order, a
	 * category with no tool left out, so an instance that found nothing draws
	 * no table at all and says so instead.
	 */
	function groups() {
		$tools = $this->tools();
		$out = array();

		foreach (self::categories() as $category) {
			$rows = self::of($tools, $category["category"]);

			if (count($rows) === 0) {
				continue;
			}

			$out[] = array("title" => $category["title"], "tools" => $rows);
		}

		return $out;
	}

	/** of is the tools of one category, in the order the query answered them. */
	static function of($tools, $category) {
		$out = array();

		foreach ($tools as $tool) {
			if ($tool->getCategory() !== $category) {
				continue;
			}

			$out[] = self::row($tool);
		}

		return $out;
	}

	/** row is one tool: the cells it draws, and what its Check is asked under. */
	static function row($tool) {
		$name = $tool->getName();

		return array(
			"id" => self::id($name),
			"name" => $name,
			"version" => $tool->getVersion(),
			"path" => $tool->getPath(),
			"description" => $tool->getDescription(),
		);
	}

	/** id is what a tool is checked under, "" for a name no probe carries. */
	static function id($name) {
		$ids = self::ids();

		return isset($ids[$name]) ? $ids[$name] : "";
	}
}
