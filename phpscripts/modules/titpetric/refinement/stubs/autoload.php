<?php

// What `phpscript lint --include` reads before it reads the module: MiniTPL,
// for its names, and the stubs of the host bindings, which no declaration in
// this tree holds.
//
// Nothing requires this file at runtime. index.php names every file it needs
// on its own, and the host bindings are already on the runtime.

require "lib/MiniTPL/Compiler.php";
require "lib/MiniTPL/Hook.php";
require "lib/MiniTPL/Template.php";
require "stubs/module.php";
require "stubs/task.php";
require "stubs/database.php";
require "stubs/time.php";
