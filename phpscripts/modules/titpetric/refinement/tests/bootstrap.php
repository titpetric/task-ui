<?php

// What every fixture gets before it runs, named in phpscript.yml.
//
// The paths are the module's own, from its root: phpscript.yml puts the
// suite root there, so a test reads the same tree index.php reads.
//
// Loading a file is not calling it. lib/thread.php declares functions that
// reach Task:: and the repository, and declaring them needs neither; a
// fixture calls the ones that stand on their own. storage/Refinement.php
// stays out: it is a class over a Database, which is a host binding
// `phpscript test` does not carry. docs/testing.md says which is which.

require "ext/autoload.php";
require "lib/render.php";
require "lib/thread.php";
