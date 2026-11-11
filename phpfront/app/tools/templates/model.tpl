{*noescape*}{open}

// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// {class} is a row of the {table} table.
class {class} {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "{table}";

	/** fields is every column, in the order the schema declares them. */
	static $fields = array({fields});

	/** primary is the primary key columns. */
	static $primary = array({primary});

	/** stamps is the timestamp columns, which are written null when empty. */
	static $stamps = array({stamps});
{foreach $columns as $column}
	/** {column.comment} */
	var ${column.name} = {column.zero};
{/foreach}

	function __construct($row = array()) {
{foreach $columns as $column}		$this->{column.name} = \Task\Row::{column.reader}($row, "{column.name}");
{/foreach}	}
{foreach $columns as $column}
	/** get{column.studly} is {column.comment}. */
	function get{column.studly}() {
		return $this->{column.name};
	}

	/** set{column.studly} sets {column.comment}. */
	function set{column.studly}($value) {
		$this->{column.name} = $value;

		return $this;
	}
{/foreach}

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
			$out[] = new {class}($row);
		}

		return $out;
	}
}
