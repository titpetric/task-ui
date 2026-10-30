<?php

// The flattened request an Apply all sends, and nothing else.
//
// A thread runs as a sequence: each correction is a run over what the
// run before it left, which drifts. By the tenth, the first correction
// is a thing nobody has read in an hour. Flattening turns the sequence
// into a statement: here is the design, here is every rule, the
// artifacts are already in the folder, make them satisfy all of it at
// once.
//
// Composed from the comment rows and never parsed back out of prose. The
// design is a comment, each rule is a comment, and an Apply all's own
// record is neither, so the second Apply all composes the way the first
// did over whatever survives.
//
// Static, and over arrays rather than over the store, so the fixtures in
// tests/ pin the text without a database.
class RefinementPrompt {
	/**
	 * SCRATCHPAD is the file a thread's runs keep for themselves, at the
	 * top of the thread's folder: the tools, the commands and the scripts
	 * that make each artifact.
	 *
	 * A scratch pad and not a specification. README.md is the report a
	 * reader reads and this is working notes, which is a thing an agent
	 * already knows how to keep.
	 */
	const SCRATCHPAD = "SCRATCHPAD.md";

	/**
	 * rules is the body of each rule, in the order they were asked for,
	 * from the comment rows of a thread's corrections.
	 */
	static function rules($comments) {
		$rules = array();

		foreach ($comments as $comment) {
			$body = trim($comment["body"]);

			if ($body !== "") {
				$rules[] = $body;
			}
		}

		return $rules;
	}

	/**
	 * flatten is the design and the rules as one request.
	 *
	 * A rule is numbered because the numbers are how a later rule says
	 * it beats an earlier one, and the run is told so: the corrections
	 * arrived in order and a person correcting twice meant the second.
	 *
	 * With no rules there is no rule section. A thread that has had no
	 * correction has nothing to flatten, and the request is the design
	 * the thread opened with.
	 */
	static function flatten($design, $rules) {
		$parts = array("## The design\n\n" . trim($design));

		if ($rules) {
			$lines = array(
				"## The rules",
				"",
				"Every correction this thread has made, as a rule the artifacts have to",
				"satisfy at once. A later rule wins where two disagree.",
				"",
			);

			$n = 0;

			foreach ($rules as $rule) {
				$n++;
				$lines[] = $n . ". " . trim($rule);
			}

			$parts[] = implode("\n", $lines);
		}

		$parts[] = implode("\n", array(
			"## Where this starts",
			"",
			"The thread's folder already holds the artifacts, and your scratch pad,",
			"`" . RefinementPrompt::SCRATCHPAD . "` at the top of it, says which tools made them and how to",
			"run them again. Read both before you change anything, run the pipelines",
			"the scratch pad names rather than starting over, and produce artifacts",
			"that satisfy every rule together rather than applying them one after",
			"another.",
		));

		return implode("\n\n", $parts);
	}
}
