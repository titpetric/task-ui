<?php

// Compiles every template into the templates_c beside it, which the tree
// ships, so no request compiles one.
//
//   phpscript run tools/precompile.php
//
// compile() is called for each file rather than load(): MiniTPL resolves an
// {include} at compile time and writes the component's body into the compiled
// template, and load() stats only the template's own source, so a changed
// component would leave every template that includes it compiled against the
// old body. See docs/todo-correction.md.

require "ext/autoload.php";

/** roots are the template directories, the theme's last so an include resolves into it. */
function roots() {
	$out = array();

	foreach (glob("modules/*/templates") as $dir) {
		$out[] = $dir . "/";
	}

	$out[] = "theme/templates/";

	return $out;
}

/** sources are the templates of one directory, the components among them. */
function sources($root) {
	$found = array_merge(glob($root . "*.tpl"), glob($root . "components/*.tpl"));

	sort($found);

	return $found;
}

/** target is where a template of root compiles to: the _c sibling of its own directory. */
function target($root, $source) {
	return rtrim($root, "/") . "_c/" . substr($source, strlen($root));
}

$count = 0;

foreach (roots() as $root) {
	// Every root searches itself and the theme, the way Task\View does, so an
	// include compiles against the same paths a request resolves.
	$tpl = new MiniTPL\Template(array($root, "theme/templates/"));

	foreach (sources($root) as $source) {
		$to = target($root, $source);

		if (!$tpl->compile($source, $to)) {
			echo "failed to compile ", $source, "\n";
			exit(1);
		}

		$count++;
	}
}

echo $count, " templates\n";
