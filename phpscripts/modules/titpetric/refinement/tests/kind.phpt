name: refinement_kind
description: >
  Whether the thread page embeds an artifact or links it. A diagram is
  read on the page and everything else is a link, which is the whole of
  the decision: the extension, lower cased, against the image types.
---
<?php

$paths = array(
	"diagrams/schema.svg",
	"shot.PNG",
	"a.jpeg",
	"schema/user.up.sql",
	"README.md",
	"Makefile",
	"archive.svg.gz",
);

foreach ($paths as $path) {
	echo refinement_kind($path), " ", $path, "\n";
}
---
image diagrams/schema.svg
image shot.PNG
image a.jpeg
file schema/user.up.sql
file README.md
file Makefile
file archive.svg.gz
