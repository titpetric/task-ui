<?php

// @route GET /{rest...}

// The catch-all. chi matches a literal and a parameter ahead of a wildcard,
// so this answers only what no other route claimed.

http_response_code(404);

echo "nothing at ", $_SERVER["REQUEST_URI"];
