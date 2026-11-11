<?php

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// Run is a row of the run table.
class Run {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "run";

	/** fields is every column, in the order the schema declares them. */
	public static $fields = array(
		"id",
		"folder",
		"job",
		"user_id",
		"user_name",
		"state",
		"exit_code",
		"forced",
		"scheduled",
		"term_cols",
		"term_rows",
		"cast_path",
		"reason",
		"started_at",
		"finished_at",
		"created_at",
		"updated_at",
		"dir",
		"args",
		"target",
		"batch_id",
		"message",
		"title",
		"parent_id",
		"parent_kind",
	);

	/** primary is the primary key columns. */
	public static $primary = array("id");

	/** stamps is the timestamp columns, which are written null when empty. */
	public static $stamps = array("started_at", "finished_at", "created_at", "updated_at");
	/** ID */
	var $id = "";
	/** Folder */
	var $folder = "";
	/** Job */
	var $job = "";
	/** User ID */
	var $user_id = "";
	/** User Name */
	var $user_name = "";
	/** State */
	var $state = "";
	/** Exit Code */
	var $exit_code = 0;
	/** Forced */
	var $forced = 0;
	/** Scheduled */
	var $scheduled = 0;
	/** Term Cols */
	var $term_cols = 0;
	/** Term Rows */
	var $term_rows = 0;
	/** Cast Path */
	var $cast_path = "";
	/** Reason */
	var $reason = "";
	/** Started At */
	var $started_at = "";
	/** Finished At */
	var $finished_at = "";
	/** Created At */
	var $created_at = "";
	/** Updated At */
	var $updated_at = "";
	/** Dir */
	var $dir = "";
	/** Args */
	var $args = "";
	/** Target */
	var $target = "";
	/** Batch ID */
	var $batch_id = "";
	/** Message */
	var $message = "";
	/** Title */
	var $title = "";
	/** Parent ID */
	var $parent_id = "";
	/** Parent Kind */
	var $parent_kind = "";

	function __construct($row = array()) {
		$this->id = \Task\Row::text($row, "id");
		$this->folder = \Task\Row::text($row, "folder");
		$this->job = \Task\Row::text($row, "job");
		$this->user_id = \Task\Row::text($row, "user_id");
		$this->user_name = \Task\Row::text($row, "user_name");
		$this->state = \Task\Row::text($row, "state");
		$this->exit_code = \Task\Row::number($row, "exit_code");
		$this->forced = \Task\Row::number($row, "forced");
		$this->scheduled = \Task\Row::number($row, "scheduled");
		$this->term_cols = \Task\Row::number($row, "term_cols");
		$this->term_rows = \Task\Row::number($row, "term_rows");
		$this->cast_path = \Task\Row::text($row, "cast_path");
		$this->reason = \Task\Row::text($row, "reason");
		$this->started_at = \Task\Row::stamp($row, "started_at");
		$this->finished_at = \Task\Row::stamp($row, "finished_at");
		$this->created_at = \Task\Row::stamp($row, "created_at");
		$this->updated_at = \Task\Row::stamp($row, "updated_at");
		$this->dir = \Task\Row::text($row, "dir");
		$this->args = \Task\Row::text($row, "args");
		$this->target = \Task\Row::text($row, "target");
		$this->batch_id = \Task\Row::text($row, "batch_id");
		$this->message = \Task\Row::text($row, "message");
		$this->title = \Task\Row::text($row, "title");
		$this->parent_id = \Task\Row::text($row, "parent_id");
		$this->parent_kind = \Task\Row::text($row, "parent_kind");
	}
	/** getId is ID. */
	function getId() {
		return $this->id;
	}

	/** setId sets ID. */
	function setId($value) {
		$this->id = $value;

		return $this;
	}
	/** getFolder is Folder. */
	function getFolder() {
		return $this->folder;
	}

	/** setFolder sets Folder. */
	function setFolder($value) {
		$this->folder = $value;

		return $this;
	}
	/** getJob is Job. */
	function getJob() {
		return $this->job;
	}

	/** setJob sets Job. */
	function setJob($value) {
		$this->job = $value;

		return $this;
	}
	/** getUserId is User ID. */
	function getUserId() {
		return $this->user_id;
	}

	/** setUserId sets User ID. */
	function setUserId($value) {
		$this->user_id = $value;

		return $this;
	}
	/** getUserName is User Name. */
	function getUserName() {
		return $this->user_name;
	}

	/** setUserName sets User Name. */
	function setUserName($value) {
		$this->user_name = $value;

		return $this;
	}
	/** getState is State. */
	function getState() {
		return $this->state;
	}

	/** setState sets State. */
	function setState($value) {
		$this->state = $value;

		return $this;
	}
	/** getExitCode is Exit Code. */
	function getExitCode() {
		return $this->exit_code;
	}

	/** setExitCode sets Exit Code. */
	function setExitCode($value) {
		$this->exit_code = $value;

		return $this;
	}
	/** getForced is Forced. */
	function getForced() {
		return $this->forced;
	}

	/** setForced sets Forced. */
	function setForced($value) {
		$this->forced = $value;

		return $this;
	}
	/** getScheduled is Scheduled. */
	function getScheduled() {
		return $this->scheduled;
	}

	/** setScheduled sets Scheduled. */
	function setScheduled($value) {
		$this->scheduled = $value;

		return $this;
	}
	/** getTermCols is Term Cols. */
	function getTermCols() {
		return $this->term_cols;
	}

	/** setTermCols sets Term Cols. */
	function setTermCols($value) {
		$this->term_cols = $value;

		return $this;
	}
	/** getTermRows is Term Rows. */
	function getTermRows() {
		return $this->term_rows;
	}

	/** setTermRows sets Term Rows. */
	function setTermRows($value) {
		$this->term_rows = $value;

		return $this;
	}
	/** getCastPath is Cast Path. */
	function getCastPath() {
		return $this->cast_path;
	}

	/** setCastPath sets Cast Path. */
	function setCastPath($value) {
		$this->cast_path = $value;

		return $this;
	}
	/** getReason is Reason. */
	function getReason() {
		return $this->reason;
	}

	/** setReason sets Reason. */
	function setReason($value) {
		$this->reason = $value;

		return $this;
	}
	/** getStartedAt is Started At. */
	function getStartedAt() {
		return $this->started_at;
	}

	/** setStartedAt sets Started At. */
	function setStartedAt($value) {
		$this->started_at = $value;

		return $this;
	}
	/** getFinishedAt is Finished At. */
	function getFinishedAt() {
		return $this->finished_at;
	}

	/** setFinishedAt sets Finished At. */
	function setFinishedAt($value) {
		$this->finished_at = $value;

		return $this;
	}
	/** getCreatedAt is Created At. */
	function getCreatedAt() {
		return $this->created_at;
	}

	/** setCreatedAt sets Created At. */
	function setCreatedAt($value) {
		$this->created_at = $value;

		return $this;
	}
	/** getUpdatedAt is Updated At. */
	function getUpdatedAt() {
		return $this->updated_at;
	}

	/** setUpdatedAt sets Updated At. */
	function setUpdatedAt($value) {
		$this->updated_at = $value;

		return $this;
	}
	/** getDir is Dir. */
	function getDir() {
		return $this->dir;
	}

	/** setDir sets Dir. */
	function setDir($value) {
		$this->dir = $value;

		return $this;
	}
	/** getArgs is Args. */
	function getArgs() {
		return $this->args;
	}

	/** setArgs sets Args. */
	function setArgs($value) {
		$this->args = $value;

		return $this;
	}
	/** getTarget is Target. */
	function getTarget() {
		return $this->target;
	}

	/** setTarget sets Target. */
	function setTarget($value) {
		$this->target = $value;

		return $this;
	}
	/** getBatchId is Batch ID. */
	function getBatchId() {
		return $this->batch_id;
	}

	/** setBatchId sets Batch ID. */
	function setBatchId($value) {
		$this->batch_id = $value;

		return $this;
	}
	/** getMessage is Message. */
	function getMessage() {
		return $this->message;
	}

	/** setMessage sets Message. */
	function setMessage($value) {
		$this->message = $value;

		return $this;
	}
	/** getTitle is Title. */
	function getTitle() {
		return $this->title;
	}

	/** setTitle sets Title. */
	function setTitle($value) {
		$this->title = $value;

		return $this;
	}
	/** getParentId is Parent ID. */
	function getParentId() {
		return $this->parent_id;
	}

	/** setParentId sets Parent ID. */
	function setParentId($value) {
		$this->parent_id = $value;

		return $this;
	}
	/** getParentKind is Parent Kind. */
	function getParentKind() {
		return $this->parent_kind;
	}

	/** setParentKind sets Parent Kind. */
	function setParentKind($value) {
		$this->parent_kind = $value;

		return $this;
	}

	/** values is this row as the database takes it, keyed by column. */
	function values() {
		return \Task\Row::values($this, self::$stamps);
	}

	/** rows turns what the database answered into models. */
	static function rows($found) {
		$out = array();

		if (!is_array($found)) {
			return $out;
		}

		foreach ($found as $row) {
			$out[] = new Run($row);
		}

		return $out;
	}
}
