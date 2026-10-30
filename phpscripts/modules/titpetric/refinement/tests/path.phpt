name: RefinementPage::path
description: >
  What is left of a request for the module to answer: the path past its
  own prefix, without the query and without the slashes at either end.
  task-ui mounts the module at that prefix and REQUEST_URI carries it
  whole, so taking it off is the first thing every request does.
---
<?php

$page = new RefinementPage("/m/titpetric/refinement");
$base = "/m/titpetric/refinement";

echo "[", $page->path($base), "]\n";
echo "[", $page->path($base . "/"), "]\n";
echo "[", $page->path($base . "/threads"), "]\n";
echo "[", $page->path($base . "/threads/01K8/comments"), "]\n";
echo "[", $page->path($base . "/threads/01K8/comments/01K9/delete"), "]\n";
echo "[", $page->path($base . "/threads?page=2"), "]\n";
echo "[", $page->path("/elsewhere/threads"), "]\n";
---
[]
[]
[threads]
[threads/01K8/comments]
[threads/01K8/comments/01K9/delete]
[threads]
[elsewhere/threads]
