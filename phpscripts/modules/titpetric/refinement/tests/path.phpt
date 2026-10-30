name: refinement_path
description: >
  What is left of a request for the module to answer: the path past its
  own prefix, without the query and without the slashes at either end.
  task-ui mounts the module at that prefix and REQUEST_URI carries it
  whole, so taking it off is the first thing every request does.
---
<?php

$base = "/m/titpetric/refinement";

echo "[", refinement_path($base, $base), "]\n";
echo "[", refinement_path($base . "/", $base), "]\n";
echo "[", refinement_path($base . "/threads", $base), "]\n";
echo "[", refinement_path($base . "/threads/01K8/comments", $base), "]\n";
echo "[", refinement_path($base . "/threads?page=2", $base), "]\n";
echo "[", refinement_path("/elsewhere/threads", $base), "]\n";
---
[]
[]
[threads]
[threads/01K8/comments]
[threads]
[elsewhere/threads]
