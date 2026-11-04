<?php

namespace Task;

// Seen answers whether the reader may see a run, by ID, once per run.
//
// A note links the run its agent worked in, and the link is drawn only for a
// reader who sees that run; a run that is gone links nowhere. A board of a
// hundred notes names few runs, and the same run over and over, so the answer
// of each is kept.
class Seen {
	var $runs;

	var $user;

	var $prompts = false;

	var $notes = false;

	/** asked is the answer of each run looked up so far. */
	var $asked = array();

	function __construct($chrome) {
		$this->runs = new Run($chrome->db);
		$this->user = $chrome->user;
		$this->prompts = $chrome->allowed("prompt/view");
		$this->notes = $chrome->allowed("note/view");
	}

	/**
	 * of is the runs of a list the reader may see, in the order they arrived.
	 *
	 * The rows are the ones a filter read, so a list is answered without
	 * reading any run again, and the answers are kept for a page that asks
	 * about one of them by ID later.
	 */
	function of($found) {
		$out = array();

		foreach ($found as $run) {
			$id = $run->getId();
			$this->asked[$id] = Run::seen($this->user, $run, $this->prompts, $this->notes);

			if ($this->asked[$id]) {
				$out[] = $run;
			}
		}

		return $out;
	}

	/** run reports whether the reader sees the run of an ID. An empty ID is none. */
	function run($id) {
		if ($id === "") {
			return false;
		}

		if (!isset($this->asked[$id])) {
			$this->asked[$id] = Run::seen($this->user, $this->runs->get($id), $this->prompts, $this->notes);
		}

		return $this->asked[$id];
	}
}
