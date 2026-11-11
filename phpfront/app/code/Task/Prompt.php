<?php

namespace Task;

// Prompt is where the header's prompt asks on a page: the folder and the
// directory of a folder's page, of a directory, of a job and of a workspace's
// settings; the run of a run's page, in its folder and directory; the root
// folder anywhere else.
//
// A page says where it is with one call on the chrome and this answers the
// rest. The header carries the answer in data- attributes and the script reads
// them, so a page that says nothing asks in the root rather than nowhere.
//
// href carries the mount, because the script both links it and fetches the
// thread from it: a panel opened on a page of this front end stays on it.
class Prompt {
	/** API is where a prompt box posts, at the root with the rest of the API. */
	const API = "/api/prompts";

	/** root is where the prompt asks on a page of no folder, which is the root's. */
	static function root() {
		return array(
			"folder" => Route::INTERNAL,
			"dir" => "",
			"run" => "",
			"label" => Route::INTERNAL,
			"href" => Route::href("/f/" . Route::INTERNAL),
		);
	}

	/**
	 * at is where the prompt asks on the page of a folder and a directory
	 * inside it, which is the place itself.
	 *
	 * The root with no directory is the root, and so is a page that names no
	 * folder. A place the reader's folders do not hold keeps its own name and
	 * links to the root, because the panel's line still says where the thread
	 * is even where there is no page to open.
	 */
	static function at($folder, $dir, $folders, $workspaces) {
		if ($folder === Route::INTERNAL && $dir === "" || $folder === "") {
			return self::root();
		}

		$where = Route::place($folder, $dir);
		$held = Route::heldHref($where, $folders, $workspaces);
		$href = $held === "" ? "/f/" . Route::INTERNAL : $held;

		return array(
			"folder" => $folder,
			"dir" => $dir,
			"run" => "",
			"label" => $where,
			"href" => Route::href($href),
		);
	}

	/**
	 * ofRun is where the prompt asks on a run's page: in the place the run ran
	 * in, and about the run itself, which the panel names by the start of its
	 * ID and links to.
	 *
	 * A run of Claude's own is not a run to ask about: a prompt's run, the
	 * Notes agent's and Claude in a terminal are the thread, so the page of
	 * one asks in its place and about nothing.
	 */
	static function ofRun($run, $folders, $workspaces) {
		$place = self::at($run->getFolder(), $run->getDir(), $folders, $workspaces);

		if (Run::isClaude($run)) {
			return $place;
		}

		$id = $run->getId();
		$place["run"] = $id;
		$place["label"] = "run " . self::short($id);
		$place["href"] = Route::href("/runs/" . $id);

		return $place;
	}

	/**
	 * field is the box's own markdown field, as components/editor.tpl reads
	 * it: three rows, and the question the empty one asks.
	 *
	 * A prompt names nobody, so there is no list of users on an `@`: the
	 * mentions are the board's, where a note is written to someone.
	 */
	static function field($placeholder) {
		return array(
			"name" => "prompt",
			"rows" => 3,
			"label" => "The prompt",
			"placeholder" => $placeholder,
			"mentions" => false,
			"toolbar" => "full",
		);
	}

	/** short is the start of a run's ID, which is what the panel's line names it by. */
	static function short($id) {
		return strlen($id) > 4 ? substr($id, 0, 4) . "…" : $id;
	}

	/**
	 * inPlace is the prompt box a folder's page and a directory inside one draw
	 * under the Jobs table, with the threads asked there. Null where the page
	 * draws none.
	 *
	 * Only a reader who may ask sees it: there is no box that watches a place,
	 * because a place has no run of its own to wait on. The box opens itself
	 * where a thread exists and stays folded where none does, and a named
	 * workspace's page opens it either way, which is what $open says.
	 *
	 * Every thread here is a prompt, so Claude answers it and no agent is
	 * looked up.
	 */
	static function inPlace($chrome, $folder, $dir, $open, $seen) {
		if (!$chrome->allowed("prompt/edit") || !Host::claude()) {
			return null;
		}

		$place = Route::place($folder, $dir);
		$fields = array(
			array("name" => "folder", "value" => $folder),
			array("name" => "dir", "value" => $dir),
		);

		return self::asked($chrome, self::API, $place, $place . "/", $fields, $open, $seen);
	}

	/**
	 * ofRoot is the prompt box the root's own page draws, which asks in the
	 * root itself. Null where the page draws none.
	 *
	 * It names no folder, so the form carries neither it nor a directory and
	 * the workspace's own endpoint takes the question; the line names the
	 * place in words, because the root has no path to print. The box is open,
	 * as a workspace's page opens its own.
	 */
	static function ofRoot($chrome, $seen) {
		if (!$chrome->allowed("prompt/edit") || !Host::claude()) {
			return null;
		}

		$api = "/api/workspaces/" . Route::INTERNAL . "/prompt";

		return self::asked($chrome, $api, Route::INTERNAL, "the root", array(), true, $seen);
	}

	/**
	 * asked is a place's box with the threads asked there under it, as
	 * components/prompt.tpl reads them.
	 *
	 * The two kinds of place differ in three things and nothing else: where
	 * the form posts, how the line names the place, and whether the form
	 * carries the place at all.
	 */
	static function asked($chrome, $api, $place, $where, $fields, $open, $seen) {
		$notes = new Note($chrome->db);
		$roots = $notes->inPlace($place);
		$replies = $notes->replies(Note::ids($roots));
		$threads = array();

		foreach ($roots as $note) {
			$threads[] = Note::thread($note, $replies, $chrome, $seen);
		}

		return array(
			"title" => "Prompt",
			"note" => "ask Claude Code in " . $where . "; it reads the CLAUDE.md and .claude/ there, in the folders above and in the workspace's home",
			"open" => $open || count($threads) > 0,
			"hidden" => false,
			"watch" => false,
			"api" => $api,
			"fields" => $fields,
			"editor" => self::field("what does this directory build, and how"),
			"lead" => "Runs @claude-prompt as ",
			"tail" => " in " . $where . ", one prompt at a time. Claude replies under your question.",
			"picks" => array(),
			"after" => "",
			"options" => $chrome->runAs(),
			"place" => $place,
			"area" => "",
			"threads" => $threads,
		);
	}

	/**
	 * panel is the prompt box a run's page draws under the output, and the
	 * threads already asked about that run, as components/prompt.tpl reads
	 * them. Null where the page draws none.
	 *
	 * The box asks about the root of the run's chain, so a question about the
	 * third attempt is a question about the job: a reader following a chain
	 * reads one thread, not one per attempt. A run of Claude's own is the
	 * answer to a prompt and not something to prompt about.
	 *
	 * Three things close the panel: no prompt/view, a reader who may not ask
	 * and a run nobody has asked about, and a run that is still going with
	 * nothing asked. A reader who may not ask still reads the threads.
	 *
	 * Both the page of a run and the page of a batch of runs draw it, which is
	 * why it is here and not on either.
	 */
	static function panel($chrome, $run, $seen) {
		if (Run::isClaude($run) || !$chrome->allowed("prompt/view")) {
			return null;
		}

		$notes = new Note;
		$runs = new Runs\Runs;
		$about = $runs->root($run);
		$where = Run::where($run);
		$running = $run->getState() === Run::RUNNING;
		$roots = $notes->prompts($about);
		$asks = !$running && $chrome->allowed("prompt/edit") && Host::claude();

		if (!$asks && !$running && count($roots) === 0) {
			return null;
		}

		$users = new User;
		$replies = $notes->replies(Note::ids($roots));
		$agents = $users->agents();
		$threads = array();

		foreach ($roots as $note) {
			$threads[] = Note::thread($note, $replies, $chrome, $seen);
		}

		return array(
			"title" => "Prompt",
			"note" => "ask Claude Code about this run and its output, in " . $where . "/",
			"open" => true,
			"hidden" => !$asks && count($threads) === 0,
			"watch" => !$asks,
			"api" => self::API,
			"fields" => array(array("name" => "run", "value" => $about)),
			"editor" => self::field("why did this run end the way it did"),
			"lead" => "Runs @claude-prompt as ",
			"tail" => " in " . $where . "/, with this run's output. If Claude makes a fix, it commits to main. Claude replies under your question.",
			"picks" => array(),
			"after" => "",
			"options" => $chrome->runAs(),
			"place" => $where,
			"area" => "",
			"threads" => $threads,
		);
	}

	/**
	 * ofWorkflowRun is the box a workflow run's page draws under the steps: the
	 * threads of the run's prompt and summary steps, each with the output of
	 * the steps before it. Null where the page draws none.
	 *
	 * There is no form to ask a new one with, because what asked these is the
	 * run itself; a reader still replies under a thread, as they do on the
	 * board. It opens while the run goes and folds once it ended, when the
	 * thread over it holds what Claude replied, and a run that asks nothing
	 * and that nobody asked draws no box.
	 *
	 * The place is the workflow's own until a thread exists, and then the
	 * thread's: what the steps asked may have run somewhere else since.
	 */
	static function ofWorkflowRun($chrome, $flow, $run) {
		if (!$chrome->allowed("prompt/view")) {
			return null;
		}

		$notes = new Note($chrome->db);
		$roots = $notes->ofWorkflowRun($run->getId());
		$running = $run->getState() === Run::RUNNING;

		if (count($roots) === 0 && !($flow["prompted"] && $running)) {
			return null;
		}

		$users = new User($chrome->db);
		$replies = $notes->replies(Note::ids($roots));
		$agents = $users->agents();
		$place = count($roots) === 0 ? $flow["promptPlace"] : $roots[0]->getDir();
		$threads = array();

		foreach ($roots as $note) {
			$threads[] = Note::thread($note, $replies, $chrome, null);
		}

		return array(
			"title" => "Prompts",
			"note" => "the run's prompt and summary steps, each with the output of the steps before it",
			"open" => $running,
			"hidden" => count($threads) === 0,
			"watch" => true,
			"api" => "",
			"fields" => array(),
			"editor" => self::field(""),
			"lead" => "",
			"tail" => "",
			"picks" => array(),
			"after" => "",
			"options" => array("users" => array()),
			"place" => $place,
			"area" => "",
			"threads" => $threads,
		);
	}
}
