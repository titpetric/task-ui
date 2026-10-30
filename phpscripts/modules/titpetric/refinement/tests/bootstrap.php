<?php

// What every fixture gets before it runs, named in phpscript.yml.
//
// One require. ext/autoload.php is composer's and the classmap over
// code/ is in it, so a fixture names a class and the loader finds it.
//
// Naming a class is not constructing one. RefinementStore wants a
// Database and RefinementManager's instance methods reach Task::, both
// of them host bindings `phpscript test` does not carry; what a fixture
// calls is the static side, which stands on its own. docs/testing.md
// says which is which.

require "ext/autoload.php";
