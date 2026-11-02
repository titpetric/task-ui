<?php

namespace Task;

// DB opens the instance's connection.
//
// A page takes reader(), which refuses every statement that is not a read, so
// a view that was meant to list something cannot write. A handler that changes
// something takes writer() and says so by the call.
class DB {
	/** NAME is the one connection this application opens. */
	const NAME = "task-ui";

	/** reader refuses every write. */
	static function reader() {
		$db = new \Database(self::NAME);

		$db->is_readonly = true;

		return $db;
	}

	/** writer is what a handler that changes something takes. */
	static function writer() {
		return new \Database(self::NAME);
	}
}
