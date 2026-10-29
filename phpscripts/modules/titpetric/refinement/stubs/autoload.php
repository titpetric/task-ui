<?php

// What `phpscript lint --include` reads before it reads the module: composer's
// autoloader, for the MiniTPL names, and the stubs of the host bindings, which
// no declaration in this tree holds.
//
// Nothing requires this file at runtime. index.php requires composer's
// autoloader on its own, and the host bindings are already on the runtime.

require "vendor/autoload.php";
require "stubs/module.php";
require "stubs/task.php";
