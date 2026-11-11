<?php

// @route GET /dispatch

// Dispatch is its own endpoint because it is the one host call that refuses:
// an instance that starts no runs is an error and not an empty string, so a
// page cannot record a note that was never opened. The rest of the calls
// answer empty under a nil host and share mod_host.php.

echo "dispatched=", Task\Host::dispatch("use ULID keys", "modules/refinement/action/refinement.md", "01THREAD"), "\n";
