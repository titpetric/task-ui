<?php

// @route GET /refinement/threads/{id}

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_refinement_thread is one thread: the report its last run wrote, the
// requests that got it there, and the boxes that ask for the next one.
//
// The files are not listed. They are a folder of the artifact tree and the
// instance already has the page that browses one, with the preview, the
// versions and the download beside it, so the head links that folder and the
// scratch pad the runs keep for themselves.
//
// The page reconciles before it draws. A run writes the files and rewrites the
// report and goes through nothing of this module's, so the record the tables
// keep catches up where it is read.
class mod_refinement_thread implements Module {
	/** MISSING is what the page of a thread that is not there says. */
	const MISSING = "No thread with that ID. Someone may have deleted its design, which deletes the thread.";

	function name() {
		return "refinement";
	}

	function title() {
		return "Refinement";
	}

	function icon() {
		return "flow";
	}

	function tables() {
		return array("refinement_thread", "refinement_comment", "refinement_correction");
	}

	function serve() {
		// A writer on a page that draws, which no other read page takes: the
		// artifact tree is the truth and the tables are the record of it, and
		// reconciling the two is a write. Everything else here reads.
		$db = Task\DB::writer();
		$chrome = new Task\Chrome($db);
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("module")) {
			$view->refused();

			return;
		}

		$store = new Task\Refinement\Threads($db);
		$thread = $store->find(Task\Route::param("id"));

		if (!is_array($thread)) {
			$view->missing(self::MISSING);

			return;
		}

		$thread = Task\Refinement\Work::sync($store, $thread);
		$comments = $store->comments(Task\Row::text($thread, "id"));
		$title = Task\Row::text($thread, "title");

		$chrome->heading($title, array(
			array("text" => "Refinement", "href" => "/refinement"),
			array("text" => $title, "href" => ""),
		));

		$view->page($this->scope($thread, $comments), "thread.tpl");
	}

	/**
	 * scope is what the template reads: the head and its two links, the report
	 * split into what is looked at and what is read, the requests, and which of
	 * them is the design.
	 */
	function scope($thread, $comments) {
		$folder = Task\Row::text($thread, "id");
		$report = Task\Refinement\Thread::split(Task\Row::text($thread, "report_markdown"));
		$head = new Task\Head(Task\Row::text($thread, "title"));

		$head->buttons($this->links($folder)->data());

		return array(
			"head" => $head->data(),
			"thread" => array("id" => $folder),
			"entries" => $this->entries($comments, $folder),
			"design" => Task\Refinement\Thread::designID($comments),
			"output" => Task\Refinement\Thread::report($report["images"], $folder),
			"thinking" => Task\Refinement\Thread::report($report["prose"], $folder),
			"visual" => $report["images"] !== "",
			"comment" => self::field("comment", 3, "A change to make", "use ULID as primary keys"),
			"fork" => self::field("want", 2, "What to design in the fork", "A deployment topology diagram: the hosts, what runs on each and the links between them."),
		);
	}

	/**
	 * field is one of the page's two boxes, as components/editor.tpl reads it:
	 * a request is markdown and takes the five tools of a short field.
	 *
	 * Neither box names anyone, so there is no list of users on an `@`.
	 */
	static function field($name, $rows, $label, $placeholder) {
		return array(
			"name" => $name,
			"rows" => $rows,
			"label" => $label,
			"placeholder" => $placeholder,
			"mentions" => false,
			"toolbar" => "inline",
		);
	}

	/**
	 * links is the head's two buttons: the folder of the artifact tree this
	 * thread writes in, and the scratch pad at the top of it.
	 *
	 * The scratch pad is linked where there is one. It is what the runs keep
	 * for themselves, the tools and the commands that made each artifact, and
	 * it is worth reading when a run surprises you.
	 *
	 * Serving a file is a byte stream and stays Go at the root of the site, so
	 * the file link carries no base where the page link does.
	 */
	function links($folder) {
		$buttons = new Task\Buttons;
		$files = Task\Route::href("/artifacts?path=" . Task\Route::queryEscape($folder));

		$buttons->item("box", "The files of this thread", array("href" => $files));

		$pad = $folder . "/" . Task\Refinement\Prompt::SCRATCHPAD;

		if (Task\Host::artifact($pad) !== "") {
			$href = Task\Refinement\Thread::FILES . Task\Route::escapePath($pad);

			$buttons->item("file", "The scratch pad", array("href" => $href));
		}

		return $buttons;
	}

	/**
	 * entries is the thread as the page draws it: each request that stands on
	 * its own, and under an Apply all the ones it folded.
	 */
	function entries($comments, $folder) {
		$entries = Task\Refinement\Thread::entries($comments);

		foreach ($entries as $i => $entry) {
			$entries[$i]["comment"] = $this->shown($entry["comment"], $folder);

			foreach ($entry["folded"] as $j => $under) {
				$entries[$i]["folded"][$j] = $this->shown($under, $folder);
			}
		}

		return $entries;
	}

	/**
	 * shown is one comment as the page draws it: the request as HTML, the run
	 * that answered it and the stamp as a column reads.
	 *
	 * The run is the label every page shares, a span the site's own script
	 * fills and keeps current, so a thread open while a run works reads its
	 * state without a reload. A comment whose dispatch has not answered yet has
	 * no run and draws none.
	 */
	function shown($comment, $folder) {
		$comment["html"] = Task\Refinement\Thread::request(Task\Row::text($comment, "body"), $folder);
		$comment["run"] = Task\Host::run(Task\Row::text($comment, "note_id"));
		$comment["created_at"] = Task\Row::text($comment, "created_at");

		return $comment;
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_refinement_thread;

	$mod->serve();
}
