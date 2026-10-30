<?php

// Compiles every template into templates_c, which the module ships, so no
// request compiles one. MiniTPL has no command that does this: it compiles
// inside load(), when a source is newer than its compiled file, and the
// compiler creates the directory it writes into.
//
// The absolute form of set_compile_location prefixes the compile location to
// the template path, so templates/thread.tpl lands at
// templates_c/templates/thread.tpl. _find_path reads that same path.
//
//   atkins precompile

require "ext/autoload.php";

$tpl = new MiniTPL\Template("templates/");

$tpl->set_compile_location("templates_c/", true);

$sources = glob("templates/*.tpl");

sort($sources);

foreach ($sources as $source) {
	$target = "templates_c/" . $source;

	$tpl->compile($source, $target);
	echo "compiled ", $source, " to ", $target, "\n";
}

echo count($sources), " templates\n";
