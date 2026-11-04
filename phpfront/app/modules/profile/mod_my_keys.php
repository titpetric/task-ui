<?php

// @route GET /keys

require "ext/autoload.php";

// The names are qualified and there is no use block: `phpscript fmt` inlines
// an alias and leaves the import behind, so a file writes what the formatter
// would leave.

// mod_my_keys is Your SSH keys: the instance SSH keys the reader's own runs
// get, by the role they act with and by their name, to copy a public half out
// of and put on a host.
//
// The page gates nothing. Every reader has their own keys and the page shows
// theirs, so there is no permission to ask for; granting one is key/edit, and
// the two buttons of the block and the Edit links item of a row come and go
// with it.
class mod_my_keys implements Module {
	/** LEAD is the sentence under the heading, with the reader's role in it. */
	const LEAD = "The keys your runs get through their ssh agent: keys made for you, and keys granted to your role%s. Copy a public key to a git host or a server that needs to trust it. The private key stays on the task-ui server.";

	/** NONE is what the block says where the reader may grant themselves a key. */
	const NONE = "No keys for you or your role yet. Generate or upload one for yourself.";

	/** ASKS is what it says where they may not, which names who can. */
	const ASKS = "No keys for you or your role yet. Someone with Allow Edits on Instance Keys can grant one under Settings, Keys.";

	/** NEXT is the page a change to a key of this table returns to. */
	const NEXT = "/keys";

	function name() {
		return "profile";
	}

	function title() {
		return "Profile";
	}

	function icon() {
		return "users";
	}

	function tables() {
		return array("user");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		$chrome->heading("Your SSH keys", array(array("text" => "Your SSH keys", "href" => "")));

		$folder = Task\Key::folder();
		$keys = Task\Key::mine($folder["keys"], $chrome->user);
		$head = new Task\Head("Your SSH keys");

		$head->said(sprintf(self::LEAD, self::role($chrome)));

		$view->page(array(
			"head" => $head->data(),
			"block" => $this->block($chrome, count($keys) === 0),
			"keys" => Task\Key::rows($chrome, $keys, self::NEXT),
		), "my_keys.tpl");
	}

	/**
	 * block is the head of the one block: what it says with no keys in it, and
	 * the two ways a reader who may grant one makes themselves a key.
	 */
	function block($chrome, $bare) {
		$name = $chrome->name();
		$grants = $name !== "" && $chrome->allowed("key/edit");
		$buttons = new Task\Buttons(true);

		if ($grants) {
			$buttons->item("key", "Generate key", array("href" => Task\Key::userHref("generate", $name)));
			$buttons->item("upload", "Upload key", array("href" => Task\Key::userHref("upload", $name)));
		}

		return array(
			"title" => "Keys",
			"note" => "",
			"empty" => self::says($bare, $grants),
			"buttons" => $buttons->data(),
		);
	}

	/** says is the empty line of the block: "" while the table has rows. */
	static function says($bare, $grants) {
		if (!$bare) {
			return "";
		}

		return $grants ? self::NONE : self::ASKS;
	}

	/** role is the lead's ", admin", which nobody signed in does not get. */
	static function role($chrome) {
		if ($chrome->user === null) {
			return "";
		}

		return ", " . Task\User::role($chrome->user);
	}
}

// The dispatch runs when a request is being served. A routed endpoint runs
// under the http SAPI and a fixture does not, so a fixture requires this file
// for the class and nothing answers.
if (PHP_SAPI === "http") {
	$mod = new mod_my_keys;

	$mod->serve();
}
