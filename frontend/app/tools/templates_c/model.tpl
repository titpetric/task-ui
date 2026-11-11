<?php $_v=&$this->vars; echo $_v['open'];?>


// Code generated from schema/schema.yml by tools/models.php. DO NOT EDIT.

namespace Task\Model;

// <?php echo $_v['class'];?>
 is a row of the <?php echo $_v['table'];?>
 table.
class <?php echo $_v['class'];?>
 {
	/** TABLE is the table a row of this model comes from. */
	const TABLE = "<?php echo $_v['table'];?>
";

	/** fields is every column, in the order the schema declares them. */
	static $fields = array(<?php echo $_v['fields'];?>
);

	/** primary is the primary key columns. */
	static $primary = array(<?php echo $_v['primary'];?>
);

	/** stamps is the timestamp columns, which are written null when empty. */
	static $stamps = array(<?php echo $_v['stamps'];?>
);
<?php if(!empty($_v['columns']))foreach($_v['columns'] as $_v['column']){?>
	/** <?php echo $_v['column']['comment'];?>
 */
	var $<?php echo $_v['column']['name'];?>
 = <?php echo $_v['column']['zero'];?>
;
<?php }?>

	function __construct($row = array()) {
<?php if(!empty($_v['columns']))foreach($_v['columns'] as $_v['column']){?>		$this-><?php echo $_v['column']['name'];?>
 = \Task\Row::<?php echo $_v['column']['reader'];?>
($row, "<?php echo $_v['column']['name'];?>
");
<?php }?>	}
<?php if(!empty($_v['columns']))foreach($_v['columns'] as $_v['column']){?>
	/** get<?php echo $_v['column']['studly'];?>
 is <?php echo $_v['column']['comment'];?>
. */
	function get<?php echo $_v['column']['studly'];?>
() {
		return $this-><?php echo $_v['column']['name'];?>
;
	}

	/** set<?php echo $_v['column']['studly'];?>
 sets <?php echo $_v['column']['comment'];?>
. */
	function set<?php echo $_v['column']['studly'];?>
($value) {
		$this-><?php echo $_v['column']['name'];?>
 = $value;

		return $this;
	}
<?php }?>

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
			$out[] = new <?php echo $_v['class'];?>
($row);
		}

		return $out;
	}
}
