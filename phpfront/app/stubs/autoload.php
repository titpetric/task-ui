<?php

// What `phpscript lint --include` reads before it reads the tree: composer's
// autoloader, for the MiniTPL and Task names, and the stubs of the host
// bindings, which no declaration here holds.
//
// Nothing requires this at runtime. An endpoint requires the autoloader on its
// own and the host calls are already on the runtime.

require "ext/autoload.php";
require "stubs/host.php";
require "stubs/fixture.php";
require "stubs/database.php";
require "stubs/time.php";
