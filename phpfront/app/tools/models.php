<?php

// Writes code/Task/Model from schema/schema.yml, the same file mig gen reads
// for the Go types, so the two cannot drift.
//
// Run from the repository root, because the schema is outside this tree:
//
//   phpscript run phpfront/app/tools/models.php

require "phpfront/app/ext/autoload.php";

/** TREE is this application, from the repository root. */
const TREE = "phpfront/app";

/** SCHEMA is what mig docs --yaml wrote. */
const SCHEMA = "schema/schema.yml";

/** studly turns a snake_case name into StudlyCase. */
function studly($name) {
	$out = "";

	foreach (explode("_", $name) as $part) {
		$out .= ucfirst($part);
	}

	return $out;
}

/** reader is the Task\Row call that reads a column of this type. */
function reader($type) {
	switch (strtolower($type)) {
		case "boolean":
			return "flag";
		case "integer":
			return "number";
		case "timestamp":
			return "stamp";
	}

	return "text";
}

/** zero is the value a column of this type starts at. */
function zero($type) {
	switch (strtolower($type)) {
		case "boolean":
			return "false";
		case "integer":
			return "0";
	}

	return "\"\"";
}

/** quoted is a list of names as a PHP array body. */
function quoted($names) {
	$out = array();

	foreach ($names as $name) {
		$out[] = "\"" . $name . "\"";
	}

	return implode(", ", $out);
}

/** primary is the primary key columns of a table, by the key marker. */
function primary($table) {
	$out = array();

	foreach ($table["columns"] as $column) {
		if (isset($column["key"]) && $column["key"] === "PRI") {
			$out[] = $column["name"];
		}
	}

	return $out;
}

/** stamps is the timestamp columns of a table, which are written null when empty. */
function stamps($table) {
	$out = array();

	foreach ($table["columns"] as $column) {
		if (strtolower($column["type"]) === "timestamp") {
			$out[] = $column["name"];
		}
	}

	return $out;
}

/** columns is each column as the template reads it. */
function columns($table) {
	$out = array();

	foreach ($table["columns"] as $column) {
		$comment = isset($column["comment"]) && $column["comment"] !== "" ? $column["comment"] : studly($column["name"]);

		$out[] = array(
			"name" => $column["name"],
			"studly" => studly($column["name"]),
			"comment" => $comment,
			"reader" => reader($column["type"]),
			"zero" => zero($column["type"]),
		);
	}

	return $out;
}

$schema = yaml_decode(file_get_contents(SCHEMA));

if (!is_array($schema)) {
	echo "cannot read ", SCHEMA, "\n";
	exit(1);
}

$templates = TREE . "/tools/templates/";
$tpl = new MiniTPL\Template($templates);
$names = array();

foreach ($schema as $table) {
	$class = studly($table["name"]);
	$fields = array();

	foreach ($table["columns"] as $column) {
		$fields[] = $column["name"];
	}

	$tpl->load("model.tpl");
	// The opening tag arrives as a variable: a literal one in the template
	// body would open a second tag inside the compiled template.
	$tpl->assign("open", "<?php");
	$tpl->assign("class", $class);
	$tpl->assign("table", $table["name"]);
	$tpl->assign("fields", quoted($fields));
	$tpl->assign("primary", quoted(primary($table)));
	$tpl->assign("stamps", quoted(stamps($table)));
	$tpl->assign("columns", columns($table));

	$target = TREE . "/code/Task/Model/" . $class . ".php";

	file_put_contents($target, $tpl->get());
	$names[] = $class;
}

// The schema travels into the tree, because the tree is rooted at itself: a
// script run with the app as its root cannot read a path above it, and
// tools/docs.php needs the columns to draw a module's data model.
file_put_contents(TREE . "/tools/schema.json", json_encode($schema));

echo count($names), " models: ", implode(", ", $names), "\n";
