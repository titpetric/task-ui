<?php

// @startup
//
// Applies the module's own schema into task-ui's database.
//
// The module declares the SQL and the instance applies it, through the
// same mig its own migrations go through, under a project named for the
// module. A directory that is not there is not an error.
//
// It runs at startup of its own accord: task-ui runs modinfo.php first,
// then every other @startup file of the module in path order, so the
// tables exist before the first request and nothing has to require this.

Task::migrate("schema/");
