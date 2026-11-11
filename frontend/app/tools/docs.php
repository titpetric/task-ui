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

/**
 * MODIFIERS are the words that may stand between a docblock and the function
 * it documents.
 *
 * They are matched by text and not by token: the runtime's tokenizer answers
 * T_STRING for each one and declares no T_STATIC, T_PUBLIC or T_ABSTRACT. See
 * docs/todo-correction.md.
 */
const MODIFIERS = array("static", "public", "private", "protected", "final", "abstract");

/**
 * classes reads the classes a source declares.
 *
 * The prose of a class is a run of // lines, the way a Go type's is, and the
 * prose of a member is a /** block: both are read, and the block wins where a
 * declaration carries one. The run counts only where nothing but whitespace
 * stands between it and the declaration, so a comment at the end of a method
 * body does not become the next method's.
 */
function classes($src) {
	$tokens = token_get_all($src);
	$out = array();
	$doc = "";
	$note = "";
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

			$note = "";

			continue;
		}

		if ($token[0] === T_COMMENT && substr($token[1], 0, 3) === "/**") {
			$doc = docblock($token[1]);
			$note = "";

			continue;
		}

		if ($token[0] === T_COMMENT && substr($token[1], 0, 2) === "//") {
			$line = trim(substr($token[1], 2));
			$note = $note === "" ? $line : trim($note . " " . $line);

			continue;
		}

		if ($token[0] === T_WHITESPACE) {
			continue;
		}

		// A modifier keeps the prose: `static function` is two tokens and the
		// comment belongs to the function, not to the word before it.
		if ($token[0] === T_STRING && in_array($token[1], MODIFIERS, true)) {
			continue;
		}

		if ($token[0] === T_CLASS) {
			$out[] = array(
				"name" => name($tokens, $i),
				"doc" => $doc === "" ? $note : $doc,
				"methods" => array(),
			);
			$at = count($out) - 1;
			$doc = "";
			$note = "";

			continue;
		}

		// A method of the class, which is one brace deep. Anything deeper is
		// a closure and anything shallower is a free function.
		if ($token[0] === T_FUNCTION && $at >= 0 && $depth === 1) {
			$name = name($tokens, $i);
			$prose = $doc === "" ? $note : $doc;

			if (substr($name, 0, 2) !== "__") {
				$out[$at]["methods"][] = array(
					"name" => $name,
					"signature" => $name . signature($tokens, $i),
					"doc" => $prose === "" ? "-" : $prose,
				);
			}

			$doc = "";
			$note = "";

			continue;
		}

		$doc = "";
		$note = "";
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

/**
 * load requires a module's files and answers every Module they declared.
 *
 * All of them, because a module directory holds one class per page and each
 * names the tables its own page reads: the data model of the directory is what
 * they name together, and the first file glob happens to answer is not it.
 */
function load($name) {
	$found = glob("modules/" . $name . "/mod_*.php");
	$out = array();

	foreach ($found as $file) {
		require_once $file;
	}

	foreach ($found as $file) {
		$class = basename($file, ".php");

		if (class_exists($class)) {
			$out[] = new $class;
		}
	}

	return $out;
}

/**
 * own is the module's own class of the classes a directory declares, the one
 * named after it, which is where the directory's title and prose are.
 *
 * A directory that declares no eponymous class is answered by its first, so a
 * module is documented either way.
 */
function own($modules, $name) {
	foreach ($modules as $module) {
		if (get_class($module) === "mod_" . $name) {
			return $module;
		}
	}

	return $modules[0];
}

/** prune drops the diagram a module used to have and no longer earns. */
function prune($dir) {
	$found = array($dir . "/schema.puml", $dir . "/schema.svg");

	foreach ($found as $file) {
		if (file_exists($file)) {
			unlink($file);
		}
	}
}

/** declares is the tables a directory's classes name, each one once. */
function declares($modules) {
	$out = array();

	foreach ($modules as $module) {
		foreach ($module->tables() as $table) {
			if (!in_array($table, $out, true)) {
				$out[] = $table;
			}
		}
	}

	return $out;
}

/** names is the classes of a directory, as the footer spells them. */
function names($modules) {
	$out = array();

	foreach ($modules as $module) {
		$out[] = "`" . get_class($module) . "`";
	}

	return implode(", ", $out);
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
	$paths = array($own, "tools/templates/");
	$tpl = new MiniTPL\Template($paths);

	$tpl->load($template);
	$tpl->assign("nl", "\n");

	foreach ($data as $key => $value) {
		$tpl->assign($key, $value);
	}

	return $tpl->get();
}

$written = 0;

foreach (modules() as $name) {
	$declared = load($name);

	if (count($declared) === 0) {
		echo "no Module in modules/", $name, "\n";
		exit(1);
	}

	$module = own($declared, $name);
	$dir = "modules/" . $name;
	$tables = schema(declares($declared));

	// The operations are both of a module's own directories: storage/ runs the
	// statements and code/ is the logic over the rows that is neither a
	// statement nor a page, which a module large enough to have one puts there
	// because a mod_ file is outside the classmap and cannot be called from
	// another. The store comes first: it says what the rows are, and the rest
	// is read in terms of them.
	$stored = api($dir . "/storage");
	$logic = api($dir . "/code");
	$classes = array_merge($stored, $logic);

	// A module whose pages read no table of their own gets no diagram: an
	// empty one renders to a blank SVG, which reads as a drawing that failed
	// rather than as a module with nothing to draw. The README says so in
	// words instead.
	if (count($tables) === 0) {
		prune($dir);
	} else {
		file_put_contents($dir . "/schema.puml", render($name, "schema.puml.tpl", array("tables" => $tables, "links" => links($tables))));
	}

	file_put_contents($dir . "/README.md", render($name, "readme.md.tpl", array(
		"title" => $module->title(),
		"module" => names($declared),
		"description" => description($dir . "/mod_" . $name . ".php"),
		"routes" => routes($dir),
		"tables" => $tables,
		"classes" => $classes,
	)));

	echo $name, ": ", count($tables), " tables, ", count(routes($dir)), " routes\n";
	$written++;
}

echo $written, " modules\n";

/**
 * description is the comment above a module's own class, as markdown: an empty
 * comment line is a paragraph break, the way it is in the source, and not two
 * spaces in the middle of a sentence.
 *
 * The comment is the one that names the class, so the notes every mod_ file
 * carries about the formatter are not the module's prose.
 */
function description($file) {
	if (!file_exists($file)) {
		return "";
	}

	$paragraphs = array();
	$lines = array();
	$started = false;

	foreach (explode("\n", file_get_contents($file)) as $line) {
		$trimmed = trim($line);
		$comment = substr($trimmed, 0, 2) === "//";

		if ($comment && !$started && strpos($trimmed, "mod_") === false) {
			continue;
		}

		if ($comment) {
			$started = true;
			$text = trim(substr($trimmed, 2));

			if ($text !== "") {
				$lines[] = $text;

				continue;
			}

			if (count($lines) > 0) {
				$paragraphs[] = implode(" ", $lines);
				$lines = array();
			}

			continue;
		}

		if ($started) {
			break;
		}
	}

	if (count($lines) > 0) {
		$paragraphs[] = implode(" ", $lines);
	}

	return implode("\n\n", $paragraphs);
}
