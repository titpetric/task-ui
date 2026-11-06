name: notes
description: The notes module names itself, its icon and the tables its docs read, and implements the whole Module contract.
---
<?php

require "modules/notes/mod_notes.php";

$mod = new mod_notes;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo implode(",", $mod->tables()), "\n";
echo var_export($mod instanceof Module, true), "\n";

// The sentence under the heading is HTML the module writes, because it names a
// tag. An instance with no agent says the first part alone; one with an agent
// says who to name, and the name is escaped where the prose around it is not,
// because a user name is what a reader typed.
echo var_export(mod_notes::lead("") === mod_notes::LEAD, true), "\n";
echo substr(mod_notes::lead("claude"), strlen(mod_notes::LEAD)), "\n";
echo substr(mod_notes::lead("a<b"), strlen(mod_notes::LEAD)), "\n";
---
notes Notes note
note,run,user
true
true
 Name <span class="mention">@claude</span> in a note to start the agent on the thread, and reply to follow up: a reply goes to it without naming it again. It works on the request and replies here when it finishes, or explains why it can't.
 Name <span class="mention">@a&lt;b</span> in a note to start the agent on the thread, and reply to follow up: a reply goes to it without naming it again. It works on the request and replies here when it finishes, or explains why it can't.
