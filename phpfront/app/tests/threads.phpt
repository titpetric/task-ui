name: threads
description: The threads module names itself, its icon and the tables its docs read, implements the whole Module contract, and the address the board held redirects to it.
---
<?php

require "modules/threads/mod_threads.php";
require "modules/threads/mod_notes.php";

$mod = new mod_threads;

echo $mod->name(), " ", $mod->title(), " ", $mod->icon(), "\n";
echo implode(",", $mod->tables()), "\n";
echo var_export($mod instanceof Module, true), "\n";

// The sentence under the heading is HTML the module writes, because it names a
// tag. An instance with no agent says the first part alone; one with an agent
// says who to name, and the name is escaped where the prose around it is not,
// because a user name is what a reader typed.
echo var_export(mod_threads::lead("") === mod_threads::LEAD, true), "\n";
echo substr(mod_threads::lead("claude"), strlen(mod_threads::LEAD)), "\n";
echo substr(mod_threads::lead("a<b"), strlen(mod_threads::LEAD)), "\n";

// The board's old address answers the same module, and a link written before
// the move lands on the timeline.
$old = new mod_notes;

echo $old->name(), " ", $old->title(), "\n";
---
threads Threads note
note,run,user
true
true
 Name <span class="mention">@claude</span> to start the agent on a thread, and reply to follow up: a reply goes to it without naming it again.
 Name <span class="mention">@a&lt;b</span> to start the agent on a thread, and reply to follow up: a reply goes to it without naming it again.
threads Threads
