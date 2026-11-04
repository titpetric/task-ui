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

	/**
	 * all runs a statement whose arguments are a list.
	 *
	 * The runtime has no argument unpacking, so a statement whose number of
	 * placeholders varies - a where list built from a filter, an IN of as many
	 * IDs as the caller holds - cannot be written as a call. A callback does
	 * reach the driver's own method, and that is how the call is made.
	 */
	static function all($db, $sql, $args = array()) {
		return call_user_func_array(array($db, "get_all"), array_merge(array($sql), $args));
	}

	/** row is all for a statement that answers one row. */
	static function row($db, $sql, $args = array()) {
		return call_user_func_array(array($db, "get"), array_merge(array($sql), $args));
	}
}
