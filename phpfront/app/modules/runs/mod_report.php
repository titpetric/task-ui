<?php

// @route GET /runs/{id}/report

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_report is the report of a run as a document, without the terminal: the
// answer Claude wrote to a prompt, or its summary of another run, for a reader
// who sees the run and may read prompts.
//
// The page names two runs where they differ: the run the report is about, which
// the heading and the details name, and the run of Claude's that wrote it.
class mod_report implements Module {
	function name() {
		return "runs";
	}

	function title() {
		return "Runs";
	}

	function icon() {
		return "list";
	}

	function tables() {
		return array("run", "note", "run_summary", "user", "workspace");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);
		$runs = new Task\Runs\Runs;
		$seen = new Task\Seen($chrome);
		$run = $runs->get(Task\Route::param("id"));

		if ($run === null || !$seen->run($run->getId())) {
			$view->missing("No run with that ID.");

			return;
		}

		$report = null;

		if ($chrome->allowed("prompt/view")) {
			$reports = new Task\Report;
			$report = $reports->of($run);
		}

		if ($report === null) {
			$view->missing("The run has no report: no summary, nor an answer of Claude's.");

			return;
		}

		$name = Task\Run::name($run);

		$chrome->heading($name . " report", array(
			array("text" => "Runs", "href" => "/runs"),
			array("text" => $name, "href" => "/runs/" . $run->getId()),
			array("text" => "Report", "href" => ""),
		));
		$chrome->asksOfRun($run);

		$scope = $this->scope($chrome, $runs, $seen, $run, $report);

		$view->page($scope, "report.tpl");
	}

	/**
	 * scope is what the template reads: the head, whether the document is a
	 * summary, the question it answers and the answer itself.
	 */
	function scope($chrome, $runs, $seen, $run, $report) {
		$about = $this->subject($runs, $seen, $run, $report);
		$summary = Task\Report::isSummary($report);
		$buttons = self::buttons($chrome, $run, $about);
		$head = new Task\Head(Task\Run::name($about));

		$head->said(self::lead($about, $report, $summary));
		$head->badge("report");
		$head->buttons($buttons->data());

		return array(
			"head" => $head->data(),
			"summary" => $summary,
			"question" => Task\Host::markdown($report["question"]),
			"body" => Task\Host::markdown($report["body"]),
		);
	}

	/**
	 * subject is the run the report is about, which the page names: the run
	 * Claude summarized or asked about, and the page's own run where the
	 * report is of that run, where the other run is gone, or where the reader
	 * does not see it.
	 *
	 * The name is not about(): the report's own field is $report["about"] and
	 * this answers the run it names, which a page reads as a row.
	 */
	function subject($runs, $seen, $run, $report) {
		$id = $report["about"];

		if ($id === "" || $id === $run->getId() || !$seen->run($id)) {
			return $run;
		}

		$found = $runs->get($id);

		return $found === null ? $run : $found;
	}

	/**
	 * buttons are the actions of the page head: the run the report is about,
	 * the run of Claude's that wrote it where that is another run, and the job
	 * the run is of.
	 */
	static function buttons($chrome, $run, $about) {
		$id = $about->getId();
		$jobHref = Task\Run::runJobHref($about, $chrome->workspaces());
		$buttons = new Task\Buttons;

		$buttons->item("", "Run", array("href" => Task\Route::href("/runs/" . $id)));
		$buttons->when($id !== $run->getId(), "", "Claude's run", array("href" => Task\Route::href("/runs/" . $run->getId())));
		$buttons->when($jobHref !== "", "", "Job", array("href" => Task\Route::href($jobHref)));

		return $buttons;
	}

	/**
	 * lead is the sentence under the heading: what the document is, of which
	 * run, who asked and when Claude wrote it.
	 *
	 * The whole sentence is built here and not in the template. A template
	 * writes the newline that sits before a tag, so a sentence cut across
	 * branches arrives cut. See docs/phpfront.md.
	 */
	static function lead($about, $report, $summary) {
		$said = $summary ? "Claude's summary of the run" : "Claude's answer about the run";
		$said .= " of " . Task\Clock::read($about->getStartedAt()) . " by " . $about->getUserName();

		if (Task\Run::ended($about->getState())) {
			$said .= ", which ended " . $about->getState();
		}

		if ($report["by"] !== "") {
			$said .= "; asked by " . $report["by"];
		}

		if ($report["at"] !== "") {
			$said .= ", written " . Task\Clock::ago($report["at"]);
		}

		return $said . ".";
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_report;

	$mod->serve();
}
