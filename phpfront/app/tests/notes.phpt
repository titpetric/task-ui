name: notes
description: The notes module reads the names a note mentions, as the column holds them.
---
<?php

require "modules/notes/mod_notes.php";

$mod = new mod_notes;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo implode(",", $mod->tables()), "\n";
echo var_export($mod instanceof Module, true), "\n";

// A name reaches the column spaced, so a LIKE for "% name %" matches a whole
// name and not a prefix of one.
echo "[", $mod->mentions("A note for @admin and @ops"), "]\n";
echo "[", $mod->mentions("@admin @admin twice is once"), "]\n";
echo "[", $mod->mentions("nobody is named here"), "]\n";
echo "[", $mod->mentions("@a.b @c-d @e_f are all names"), "]\n";
---
notes Notes note
note,user
true
[ admin ops ]
[ admin ]
[]
[ a.b c-d e_f ]
