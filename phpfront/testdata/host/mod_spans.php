<?php

// @route GET /spans

// Page is a frame of PHP, which is what the interpreter measures: a new and a
// method call are spans while the runtime carries an observer and nothing at
// all without one. mod_host.php is a flat script and would pass either way.
class Page {
	/** draw is the frame the trace holds. */
	function draw() {
		return "drawn=" . Task\Host::whoami() . "\n";
	}
}

$page = new Page;
echo $page->draw();
