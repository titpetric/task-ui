<?php

// Writes each module's README.md and schema.puml, and renders the diagram.
//
//   phpscript run tools/docs.php
//
// The pages come from the @route annotations, the data model from the tables
// the module names and schema/schema.yml, and the operations from the classes
// of storage/, read with token_get_all: `phpscript list` reports a class name
// and no methods, so the API is re-found here. See docs/todo-correction.md.
//
// A mod_ file is required for its class, which its SAPI guard makes safe.

require "ext/autoload.php";

/** SCHEMA is schema/schema.yml as tools/models.php copied it into the tree. */
const SCHEMA = "tools/schema.json";

/** docblock is the text of a /** *\/ comment, one line. */
function docblock($text) {
	$text = preg_replace('#^/\*+|\*+/$#', "", $text);
	$lines = array();

	foreach (explode("\n", $text) as $line) {
		$line = trim(preg_replace('/^\s*\*+\s?/', "", $line));

		if ($line !== "") {
			$lines[] = $line;
		}
	}

	return implode(" ", $lines);
}

/**
 * api reads the classes of a directory: each one's name, its docblock, and
 * its methods with their signatures and their docblocks.
 */
function api($dir) {
	$out = array();

	foreach (glob($dir . "/*.php") as $file) {
		foreach (classes(file_get_contents($file)) as $class) {
			$out[] = $class;
		}
	}

	return $out;
}

/** classes reads the classes a source declares. */
function classes($src) {
	$tokens = token_get_all($src);
	$out = array();
	$doc = "";
	// The class being read, by index: a reference binds by value here, so
	// the list is written through the index it was appended at.
	$at = -1;
	$depth = 0;

	for ($i = 0; $i < count($tokens); $i++) {
		$token = $tokens[$i];

		if (!is_array($token)) {
			if ($token === "{") {
				$depth++;
			}

			if ($token === "}") {
				$depth--;
			}

			continue;
		}

		if ($token[0] === T_COMMENT && substr($token[1], 0, 3) === "/**") {
			$doc = docblock($token[1]);

			continue;
		}

		if ($token[0] === T_WHITESPACE) {
			continue;
		}

		if ($token[0] === T_CLASS) {
			$out[] = array(
				"name" => name($tokens, $i),
				"doc" => $doc,
				"methods" => array(),
			);
			$at = count($out) - 1;
			$doc = "";

			continue;
		}

		// A method of the class, which is one brace deep. Anything deeper is
		// a closure and anything shallower is a free function.
		if ($token[0] === T_FUNCTION && $at >= 0 && $depth === 1) {
			$name = name($tokens, $i);

			if (substr($name, 0, 2) !== "__") {
				$out[$at]["methods"][] = array(
					"name" => $name,
					"signature" => $name . signature($tokens, $i),
					"doc" => $doc === "" ? "-" : $doc,
				);
			}

			$doc = "";

			continue;
		}

		$doc = "";
	}

	return $out;
}

/** name is the identifier after a keyword. */
function name($tokens, $at) {
	for ($i = $at + 1; $i < count($tokens); $i++) {
		if (is_array($tokens[$i]) && $tokens[$i][0] === T_WHITESPACE) {
			continue;
		}

		return is_array($tokens[$i]) ? $tokens[$i][1] : "";
	}

	return "";
}

/** signature is the parameter list of a function, as it was written. */
function signature($tokens, $at) {
	$out = "";
	$depth = 0;

	for ($i = $at + 1; $i < count($tokens); $i++) {
		$text = is_array($tokens[$i]) ? $tokens[$i][1] : $tokens[$i];

		if ($text === "(") {
			$depth++;
		}

		if ($depth > 0) {
			$out .= is_array($tokens[$i]) && $tokens[$i][0] === T_WHITESPACE ? " " : $text;
		}

		if ($text === ")") {
			$depth--;

			if ($depth === 0) {
				return preg_replace('/\s+/', " ", $out);
			}
		}
	}

	return "()";
}

/** routes are the @route annotations of a module's files. */
function routes($dir) {
	$out = array();

	foreach (glob($dir . "/*.php") as $file) {
		$found = array();
		$count = preg_match_all('/@route\s+(?:(\w+)\s+)?(\S+)/', file_get_contents($file), $found);

		if ($count < 1) {
			continue;
		}

		foreach ($found[2] as $i => $path) {
			$method = $found[1][$i] === "" ? "GET POST" : $found[1][$i];
			$out[] = array(
				"method" => $method,
				"path" => $path,
				"file" => basename($file),
			);
		}
	}

	return $out;
}

/** schema is the tables of schema.yml a module names, in the file's order. */
function schema($wanted) {
	$file = json_decode(file_get_contents(SCHEMA), true);
	$out = array();

	if (!is_array($file)) {
		return $out;
	}

	foreach ($file as $table) {
		if (!in_array($table["name"], $wanted, true)) {
			continue;
		}

		$columns = array();

		foreach ($table["columns"] as $column) {
			$key = isset($column["key"]) ? $column["key"] : "";
			$columns[] = array(
				"name" => $column["name"],
				"type" => $column["type"],
				"key" => $key === "PRI" ? "primary" : ($key === "MUL" ? "indexed" : ""),
				"mark" => $key === "PRI" ? "*" : "+",
			);
		}

		$out[] = array(
			"name" => $table["name"],
			"comment" => isset($table["comment"]) ? $table["comment"] : "",
			"columns" => $columns,
		);
	}

	return $out;
}

/**
 * links are the references between the tables shown, by the naming
 * convention: a column <table>_id points at <table>. The schema declares no
 * foreign key, so the name is what there is to read.
 */
function links($tables) {
	$names = array();

	foreach ($tables as $table) {
		$names[] = $table["name"];
	}

	$out = array();

	foreach ($tables as $table) {
		foreach ($table["columns"] as $column) {
			if (substr($column["name"], -3) !== "_id") {
				continue;
			}

			$target = substr($column["name"], 0, strlen($column["name"]) - 3);

			if ($target !== $table["name"] && in_array($target, $names, true)) {
				$out[] = array(
					"from" => $table["name"],
					"to" => $target,
					"column" => $column["name"],
				);
			}
		}
	}

	return $out;
}

/** modules are the module directories, by name. */
function modules() {
	$out = array();

	// A module is a directory holding a mod_ file, which is also what keeps
	// a stray file beside them out: the runtime has no GLOB_ONLYDIR.
	foreach (glob("modules/*/mod_*.php") as $file) {
		$name = basename(dirname($file));

		if (!in_array($name, $out, true)) {
			$out[] = $name;
		}
	}

	return $out;
}

/** load requires a module's files and answers the first Module it declared. */
function load($name) {
	$found = glob("modules/" . $name . "/mod_*.php");

	foreach ($found as $file) {
		require_once $file;
	}

	foreach ($found as $file) {
		$class = basename($file, ".php");

		if (class_exists($class)) {
			return new $class;
		}
	}

	return null;
}

/**
 * render writes a template over data, the module's own where it has one.
 *
 * {nl} is a newline, the way {ldelim} is a brace. PHP eats the newline
 * directly after a `?>`, so a template line ending in a tag loses it, and
 * generated text is a place where the line matters.
 */
function render($module, $template, $data) {
	$own = "modules/" . $module . "/";
	$tpl = new MiniTPL\Template(array($own, "tools/templates/"));

	$tpl->load($template);
	$tpl->assign("nl", "\n");

	foreach ($data as $key => $value) {
		$tpl->assign($key, $value);
	}

	return $tpl->get();
}

$written = 0;

foreach (modules() as $name) {
	$module = load($name);

	if ($module === null) {
		echo "no Module in modules/", $name, "\n";
		exit(1);
	}

	$dir = "modules/" . $name;
	$tables = schema($module->tables());

	file_put_contents($dir . "/schema.puml", render($name, "schema.puml.tpl", array("tables" => $tables, "links" => links($tables))));

	file_put_contents($dir . "/README.md", render($name, "readme.md.tpl", array(
		"title" => $module->title(),
		"module" => get_class($module),
		"description" => description($dir . "/mod_" . $name . ".php"),
		"routes" => routes($dir),
		"tables" => $tables,
		"classes" => api($dir . "/storage"),
	)));

	echo $name, ": ", count($tables), " tables, ", count(routes($dir)), " routes\n";
	$written++;
}

echo $written, " modules\n";

/** description is the comment above a module's own class. */
function description($file) {
	if (!file_exists($file)) {
		return "";
	}

	$lines = array();

	foreach (explode("\n", file_get_contents($file)) as $line) {
		$trimmed = trim($line);

		if (substr($trimmed, 0, 2) === "//" && count($lines) > 0) {
			$lines[] = trim(substr($trimmed, 2));

			continue;
		}

		if (substr($trimmed, 0, 2) === "//" && strpos($trimmed, "mod_") !== false) {
			$lines[] = trim(substr($trimmed, 2));

			continue;
		}

		if (count($lines) > 0) {
			break;
		}
	}

	return implode(" ", $lines);
}
