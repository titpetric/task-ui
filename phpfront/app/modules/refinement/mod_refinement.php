<?php

// @route GET /refinement

require "ext/autoload.php";

// mod_refinement is the refinement board: the threads newest first, and the
// paragraph that starts one.
//
// The page is the front end's own and draws in the theme's head and table,
// where the plugin it came from drew in a frame that named the module and
// timed its render.
//
// The gate is the one the plugin page answered behind, which is the capability
// to open a module at all.
class mod_refinement implements Module {
	/** LEAD is the sentence under the heading, which modinfo.php described the plugin with. */
	const LEAD = "Refine a design through a thread of comments";

	/** HELP is the paragraph above the box, as HTML: it emphasises a word. */
	const HELP = 'One paragraph, and a <strong>new</strong> thread starts from it, on an empty folder of its own. To correct a design that already exists, open its thread below and use the box at the bottom.';

	/** WANT is what the box asks for when it is empty. */
	const WANT = "A database schema with strict separation of user, user_login_method and user_login_method_password.";

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
		return array("refinement_thread", "refinement_comment");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("module")) {
			$view->refused();

			return;
		}

		$chrome->heading("Refinement", array(array("text" => "Refinement", "href" => "")));

		$head = new Task\Head("Refinement", self::LEAD);
		$store = new Task\Refinement\Threads;
		$threads = $store->all();
		$table = self::table($threads);

		$view->page(array(
			"head" => $head->data(),
			"table" => $table->data(),
			"help" => self::HELP,
			"editor" => self::field(),
		));
	}

	/**
	 * field is the box that opens a thread, as components/editor.tpl reads it:
	 * a design is a document and takes every tool.
	 *
	 * A thread names nobody, so there is no list of users on an `@`.
	 */
	static function field() {
		return array(
			"name" => "want",
			"rows" => 5,
			"label" => "What to design",
			"placeholder" => self::WANT,
			"mentions" => false,
			"toolbar" => "full",
		);
	}

	/** table is the threads: the title, how many comments it holds, when it started, and the way in. */
	static function table($threads) {
		$table = new Task\Table;

		$table->head(array(
			Task\Table::cell("Thread"),
			Task\Table::cell("Comments", "num"),
			Task\Table::cell("Started"),
			Task\Table::raw(""),
		));

		foreach ($threads as $thread) {
			$id = Task\Row::text($thread, "id");
			$href = Task\Route::href("/refinement/threads/" . Task\Route::pathEscape($id));
			$open = '<a class="btn btn-sm" href="' . htmlspecialchars($href, ENT_QUOTES) . '">Open</a>';

			$table->row(array(
				Task\Table::link($href, Task\Row::text($thread, "title"), "key"),
				Task\Table::cell(Task\Row::number($thread, "comments"), "num mono"),
				Task\Table::cell(Task\Row::text($thread, "created_at"), "nowrap faint"),
				Task\Table::raw($open, "actions"),
			));
		}

		return $table;
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_refinement;

	$mod->serve();
}
