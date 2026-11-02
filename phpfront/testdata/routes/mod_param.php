<?php

// @route GET /probe/{id}
// @route POST /probe/{id}/comments
// @route: /probe/both

// Path parameters arrive in $_REQUEST under the name the annotation
// declared, which is the one thing a handler needs of the route it matched.

$id = isset($_REQUEST["id"]) ? $_REQUEST["id"] : "";

echo $_SERVER["REQUEST_METHOD"], " ", $_SERVER["REQUEST_URI"], " id=", $id;
