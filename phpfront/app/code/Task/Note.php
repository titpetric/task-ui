<?php

namespace Task;

// Note reads and writes the note table: the notes that start a thread, newest
// first, and the replies under them.
//
// A note of a channel is a board note. A note with a workspace is a prompt,
// which belongs to the page the prompt was asked from and not to the board.
class Note {
	/** NOTES is the channel the board's own notes carry. */
	const NOTES = "notes";

	/** LIMIT is how many threads the board lists. */
	const LIMIT = 100;

	/** PROMPTS is how many prompt threads a run page lists. */
	const PROMPTS = 20;

	/** PROFILE is how many notes and prompts of a user a profile lists. */
	const PROFILE = 10;

	/** EXCERPT is how many characters of a note an excerpt keeps. */
	const EXCERPT = 240;

	/** WORKING is the state a note holds while the agent is in it. */
	const WORKING = "working";

	/** DONE is the state a note holds once the agent's run ended ok. */
	const DONE = "done";

	/** FAILED is the state a note holds where the agent's run ended badly. */
	const FAILED = "failed";

	/** CLAUDE is the name the replies to a prompt carry. */
	const CLAUDE = "claude";

	/**
	 * AREA_WORKFLOWS marks a prompt asked from the Workflows page, which builds
	 * or changes a workflow rather than asking about a place.
	 *
	 * An area is a page and not a folder, so the threads of one are read by the
	 * area and the place the prompt ran in is the folder its select picked.
	 */
	const AREA_WORKFLOWS = "workflows";

	/**
	 * MENTION matches @name at the start of a text or after a space or an
	 * opening mark: a user name of letters, digits, dot, dash or underscore.
	 *
	 * The second group is a colon right after the name with no space after it,
	 * which makes it a job ID such as @host:dev1 and no mention. "@claude:
	 * check" stays one.
	 */
	const MENTION = '/(?:^|[\s(\[{,;])@([A-Za-z0-9][A-Za-z0-9._-]*)((?::\S)?)/';

	/** threads is the notes that start a thread of a channel, newest first. */

	/**
	 * byUser is the board notes one user wrote, newest first, which a profile
	 * lists under Notes.
	 *
	 * The replies they wrote under someone else's note are not theirs to list:
	 * a profile lists threads, as the board does.
	 */

	/**
	 * promptsByUser is the prompt threads one user asked, newest first, which a
	 * profile lists under Prompts.
	 *
	 * More are read than the profile draws, because the page then drops the
	 * ones asked in a place the reader does not see and fills the list from
	 * what is left.
	 */

	/**
	 * prompts is the prompt threads about a run, newest first.
	 *
	 * A prompt is a note with a workspace, so the board's own notes are left
	 * out by the same clause that keeps these in.
	 */

	/**
	 * inPlace is the prompt threads asked in a place, newest first: the folder
	 * of a folder's page, or the folder and the directory of a directory inside
	 * one.
	 *
	 * A thread about a run, about a workflow run or from an area's page was
	 * asked in a place as well, and the place's page lists none of them: each of
	 * those belongs to the page that asked it.
	 */

	/** ofArea is the prompt threads asked from an area's page, newest first. */

	/** ofWorkflowRun is the prompt threads of a workflow run's steps, newest first. */

	/** promptsAbout is the prompt threads about any of the runs, newest first. */

	/** get is one note by ID, or null. */

	/**
	 * byRun is the thread the agent worked on in a run, or null.
	 *
	 * The run belongs to a reply where the agent answered one, and the thread
	 * above it is what the page draws.
	 */

	/** replies is the replies to the named notes, oldest first, by note. */

	/** ids is the IDs of the notes, which replies() names. */

	/**
	 * thread is one note with its replies, as the note component reads it.
	 *
	 * The board and a run's prompt box draw the same thread, Reply and all: a
	 * box with no form to ask a new question with still takes an answer under
	 * one already asked. No permission decides the field, because none decides
	 * it in Go either — note/edit is the API's and the Edit item's, and a
	 * reader without it is refused on the post and not before it.
	 *
	 * seen is a Task\Seen, which decides the run each card links.
	 */

	/**
	 * card is one note as the component reads it: the note on top of a thread,
	 * or a reply under it, with the menu of what the reader may do to it.
	 *
	 * The thread is the note on top, which decides what the reader edits: a
	 * reply the agent wrote is the answer and changes when its question runs
	 * again, so nobody rewrites it in place.
	 *
	 * The run the agent worked in is linked for a reader who sees that run,
	 * which seen answers; a card without one links nowhere. No seen links
	 * every run, which is what a page that has already narrowed its notes to
	 * runs of its own wants.
	 */

	/**
	 * edits reports whether the reader may rewrite a note of a thread where it
	 * stands: its author or an admin, with note/edit, and the agent's answers
	 * excepted.
	 */

	/** answered reports whether a reply is the agent's answer on a thread. */

	/**
	 * deletes reports whether the reader may remove a note and the replies
	 * under it: its author or an admin, with the permission of what it is.
	 */

	/**
	 * WRITTEN is the character references the renderer writes, and PLAIN what
	 * the source held where it wrote one.
	 *
	 * The two tables are read as one pair, so a line added to the one needs a
	 * line in the other. `&amp;` is last of all: `&amp;lt;` is the source's
	 * `&lt;` and not a `<`, so the ampersand is put back once the references
	 * that carry one are gone.
	 *
	 * The curly marks are smartypants', which the renderer runs: it writes
	 * them where the source held the plain ones, so turning each back is what
	 * the source said. A source that holds the curly character itself is not a
	 * reference and nothing here touches it.
	 */
	const WRITTEN = array(
		"&ldquo;",
		"&rdquo;",
		"&lsquo;",
		"&rsquo;",
		"&hellip;",
		"&mdash;",
		"&ndash;",
		"&quot;",
		"&#34;",
		"&#39;",
		"&lt;",
		"&gt;",
		"&amp;",
	);

	const PLAIN = array('"', '"', "'", "'", "...", "---", "--", '"', '"', "'", "<", ">", "&");

	/**
	 * FRACTION matches the markup smartypants writes a fraction as, which is
	 * three tags and a reference for what the source wrote as one word.
	 *
	 * It is put back before the tags come off, because a tag becomes a space
	 * and `1/2` would read as three words.
	 */
	const FRACTION = '#<sup>([^<]*)</sup>&frasl;<sub>([^<]*)</sub>#';

	var $db;

	function __construct($db = null) {
		$this->db = $db === null ? DB::reader() : $db;
	}

	function threads($channel, $limit = self::LIMIT) {
		return Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND workspace_id = ''" . " ORDER BY rowid DESC LIMIT ?", $channel, $limit));
	}

	function byUser($channel, $name, $limit = self::PROFILE) {
		return Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND channel = ? AND user_name = ? AND workspace_id = ''" . " ORDER BY rowid DESC LIMIT ?", $channel, $name, $limit));
	}

	function promptsByUser($name, $limit = self::PROFILE) {
		return Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND user_name = ? AND workspace_id != ''" . " ORDER BY rowid DESC LIMIT ?", $name, $limit));
	}

	function prompts($runID, $limit = self::PROMPTS) {
		return Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND about_run_id = ?" . " ORDER BY rowid DESC LIMIT ?", $runID, $limit));
	}

	function inPlace($dir, $limit = self::PROMPTS) {
		return Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND dir = ?" . " AND about_run_id = '' AND about_workflow_run_id = '' AND area = '' ORDER BY rowid DESC LIMIT ?", $dir, $limit));
	}

	function ofArea($area, $limit = self::PROMPTS) {
		return Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND area = ?" . " ORDER BY rowid DESC LIMIT ?", $area, $limit));
	}

	function ofWorkflowRun($runID, $limit = 5) {
		return Model\Note::rows($this->db->get_all("SELECT * FROM note WHERE parent_id = '' AND workspace_id != '' AND about_workflow_run_id = ?" . " ORDER BY rowid DESC LIMIT ?", $runID, $limit));
	}

	function promptsAbout($runIDs) {
		if (count($runIDs) === 0) {
			return array();
		}

		$sql = "SELECT * FROM note WHERE parent_id = '' AND workspace_id != ''" . " AND about_run_id IN (" . Row::placeholders($runIDs) . ") ORDER BY rowid DESC";

		return Model\Note::rows(DB::all($this->db, $sql, $runIDs));
	}

	function get($id) {
		$row = $this->db->get("SELECT * FROM note WHERE id = ?", $id);

		if (!is_array($row)) {
			return null;
		}

		return new Model\Note($row);
	}

	function byRun($runID) {
		if ($runID === "") {
			return null;
		}

		$row = $this->db->get("SELECT * FROM note WHERE run_id = ? ORDER BY rowid LIMIT 1", $runID);

		if (!is_array($row)) {
			return null;
		}

		$note = new Model\Note($row);

		return $note->getParentId() === "" ? $note : $this->get($note->getParentId());
	}

	function replies($ids) {
		$out = array();

		if (count($ids) === 0) {
			return $out;
		}

		$rows = DB::all($this->db, "SELECT * FROM note WHERE parent_id IN (" . Row::placeholders($ids) . ") ORDER BY rowid", $ids);

		foreach (Model\Note::rows($rows) as $note) {
			$parent = $note->getParentId();

			if (!isset($out[$parent])) {
				$out[$parent] = array();
			}

			$out[$parent][] = $note;
		}

		return $out;
	}

	static function ids($notes) {
		$out = array();

		foreach ($notes as $note) {
			$out[] = $note->getId();
		}

		return $out;
	}

	static function thread($note, $replies, $chrome, $answerer, $seen = null) {
		$out = self::card($note, $chrome, $note, $answerer, $seen);

		$out["replies"] = array();

		$id = $note->getId();

		if (!isset($replies[$id])) {
			return $out;
		}

		foreach ($replies[$id] as $reply) {
			$out["replies"][] = self::card($reply, $chrome, $note, $answerer, $seen);
		}

		return $out;
	}

	static function card($note, $chrome, $thread, $answerer, $seen = null) {
		$id = $note->getId();
		$author = $note->getUserName();
		$run = $note->getRunId();
		$reply = $note->getParentId() !== "";

		if ($seen !== null && !$seen->run($run)) {
			$run = "";
		}

		$edits = self::edits($note, $chrome, $thread, $answerer);
		$what = $reply ? "reply" : "note";
		$menu = new Menu("Actions on the " . $what . " of " . $author);

		$menu->when($edits, "edit", "Edit", array("data-edit-open" => $id));
		$menu->when($run !== "", "terminal", "Open its run", array("href" => Route::href("/runs/" . $run)));
		$menu->whenDanger(self::deletes($note, $chrome), "trash", "Delete", array(
			"data-delete" => "/api/notes/" . $id,
			"data-confirm" => $reply ? "Delete this reply of " . $author . "?" : "Delete the note of " . $author . " and its replies?",
			"data-confirm-ok" => "Delete",
			"data-next" => "reload",
		));

		return array(
			"id" => $id,
			"body" => $chrome->message($note->getBody()),
			"raw" => $note->getBody(),
			"author" => $chrome->who()->of($author),
			"state" => $note->getState(),
			"working" => $note->getState() === self::WORKING,
			"run" => $run,
			"short" => substr($run, 0, 8),
			"ago" => Clock::ago($note->getCreatedAt()),
			"stamp" => Clock::read($note->getCreatedAt()),
			"edited" => $note->getEditedAt() !== "",
			"editedStamp" => Clock::read($note->getEditedAt()),
			"edits" => $edits,
			"menu" => $menu->data(),
		);
	}

	static function edits($note, $chrome, $thread, $answerer) {
		if ($chrome->user === null || !$chrome->allowed("note/edit")) {
			return false;
		}

		if ($note->getId() !== $thread->getId() && self::answered($thread, $note, $answerer)) {
			return false;
		}

		return $note->getUserName() === $chrome->name() || User::admin($chrome->user);
	}

	static function answered($thread, $reply, $answerer) {
		return $thread->getState() !== "" && $answerer !== "" && $reply->getUserName() === $answerer;
	}

	static function deletes($note, $chrome) {
		$path = self::isPrompt($note) && $note->getParentId() === "" ? "prompt/delete" : "note/delete";

		if ($chrome->user === null || !$chrome->allowed($path)) {
			return false;
		}

		return $note->getUserName() === $chrome->name() || User::admin($chrome->user);
	}

	/**
	 * excerpt is a note's body as the one line a profile's list draws: the
	 * words of its headings, its text, the labels of its links and its code,
	 * with the marks gone, cut to EXCERPT characters with an ellipsis.
	 *
	 * The marks come off by rendering the markdown and taking the text back out
	 * of the HTML, because the parser is the host's: a note reads the same in a
	 * profile's list as it does on the board. Every tag becomes a space, as two
	 * paragraphs are two words and not one, and the character references the
	 * renderer wrote are read back so the template escapes them once.
	 *
	 * What a tag carries is not text and does not survive this: the alt text of
	 * an image is dropped, which the Go page keeps. A note that spells a
	 * character reference in its source reads here as the character and there
	 * as the reference. Both are the cost of having no binding of the Go page's
	 * own excerpt; see docs/phpfront.md.
	 */
	static function excerpt($body) {
		$html = Host::markdown($body);
		$whole = preg_replace(self::FRACTION, '$1/$2', $html);
		$spaced = preg_replace('/<[^>]*>/', " ", $whole);
		$text = str_replace(self::WRITTEN, self::PLAIN, $spaced);
		$words = preg_split('/\s+/', $text, -1, PREG_SPLIT_NO_EMPTY);
		$line = implode(" ", $words);

		if (mb_strlen($line) <= self::EXCERPT) {
			return $line;
		}

		return rtrim(mb_substr($line, 0, self::EXCERPT), " ") . "…";
	}

	/** isPrompt reports whether a note is a prompt, which belongs to a page and not to the board. */
	static function isPrompt($note) {
		return $note->getWorkspaceId() !== "";
	}

	/**
	 * answerer is the name the agent's replies on a thread carry: claude on a
	 * prompt, the agent an instruction names, and "" where no agent answers
	 * it.
	 *
	 * The agents are the users whose mentions run one, which
	 * Task\User::agents() answers.
	 */
	static function answerer($thread, $replies, $agents) {
		if (self::isPrompt($thread)) {
			return self::CLAUDE;
		}

		$found = self::agentIn($thread->getUserName(), $thread->getBody(), $agents);

		if ($found !== "") {
			return $found;
		}

		foreach ($replies as $reply) {
			$found = self::agentIn($reply->getUserName(), $reply->getBody(), $agents);

			if ($found !== "") {
				return $found;
			}
		}

		return "";
	}

	/** agentIn is the first agent a text names, other than its own author, or "". */
	static function agentIn($author, $body, $agents) {
		foreach (self::names($body) as $name) {
			if ($name === $author) {
				continue;
			}

			foreach ($agents as $agent) {
				if ($agent->getName() === $name) {
					return $name;
				}
			}
		}

		return "";
	}

	/**
	 * names is the users a text names with @, once each in the order the text
	 * names them, a trailing dot or dash dropped.
	 */
	static function names($body) {
		$out = array();
		// The match is its own statement. An out parameter written by a call
		// that is an operand of && is not visible to the operand after it, so
		// `preg_match_all(..., $found) && isset($found[1])` reads $found as
		// unset. See docs/todo-correction.md.
		$count = preg_match_all(self::MENTION, $body, $found);

		if ($count < 1 || !isset($found[1])) {
			return $out;
		}

		foreach ($found[1] as $index => $name) {
			if ($found[2][$index] !== "") {
				continue;
			}

			$name = rtrim($name, ".-");

			if ($name !== "" && !in_array($name, $out, true)) {
				$out[] = $name;
			}
		}

		return $out;
	}

	/**
	 * mentions is the names a body names, as the column holds them: each one
	 * spaced, so a LIKE for "% name %" matches a whole name and not a prefix
	 * of one.
	 *
	 * The column is read off the body here and never handed in. A row whose
	 * column and body disagree is a note nobody is told about, and the only
	 * way to write one is to let a caller say.
	 */
	static function mentions($body) {
		$found = self::names($body);

		if (count($found) === 0) {
			return "";
		}

		return " " . implode(" ", $found) . " ";
	}

	/**
	 * post writes a note, or a reply when the parent is named.
	 *
	 * A reply holds the channel of the thread it sits under. The column takes
	 * the board's channel or the chat's and nothing else, and the thread list
	 * reads the parent to tell a thread from a reply.
	 */
	function post($user, $body, $parent = "") {
		$note = new Model\Note;

		$note->setId(Id::make());
		$note->setParentId($parent);
		$note->setChannel(self::NOTES);
		$note->setUserName($user);
		$note->setBody($body);
		$note->setMentions(self::mentions($body));
		$note->setCreatedAt(Clock::stamp());
		$note->setUpdatedAt($note->getCreatedAt());

		$this->db->insert(Model\Note::TABLE, $note->values());

		return $note;
	}

	/** edit rewrites a note's body, and the names it mentions with it. */
	function edit($id, $body) {
		$now = Clock::stamp();
		$mentions = self::mentions($body);

		$this->db->query("UPDATE note SET body = ?, mentions = ?, edited_at = ?, updated_at = ? WHERE id = ?", $body, $mentions, $now, $now, $id);

		return true;
	}

	/** drop removes a note and every reply under it. */
	function drop($id) {
		$this->db->query("DELETE FROM note WHERE id = ? OR parent_id = ?", $id, $id);

		return true;
	}
}
