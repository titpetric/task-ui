name: route
description: Route reads the request against the path the router is mounted at.
---
<?php

Task\Host::$base = "/development";
$_SERVER["REQUEST_URI"] = "/development/settings/users/01HZ?page=2";
$_SERVER["REQUEST_METHOD"] = "GET";

echo Task\Route::path(), "\n";
echo implode("|", Task\Route::segments()), "\n";
echo Task\Route::method(), " ", var_export(Task\Route::posted(), true), "\n";
echo Task\Route::href("/settings"), " ", Task\Route::href("settings"), " ", Task\Route::href("/"), "\n";

// The same request once the prefix is gone: every answer but the href is the same.
Task\Host::$base = "";
$_SERVER["REQUEST_URI"] = "/settings/users/01HZ?page=2";

echo Task\Route::path(), "\n";
echo Task\Route::href("/settings"), "\n";

// The root is no segments, not one empty one.
$_SERVER["REQUEST_URI"] = "/";

echo count(Task\Route::segments()), "\n";

$_SERVER["REQUEST_METHOD"] = "POST";
$_REQUEST = array("id" => "01HZ", "title" => "  spaced  ");

echo var_export(Task\Route::posted(), true), "\n";
echo Task\Route::param("id"), "|", Task\Route::param("none", "fallback"), "\n";
echo "[", Task\Route::input("title"), "]", Task\Route::input("none", "fallback"), "\n";
---
settings/users/01HZ
settings|users|01HZ
GET false
/development/settings /development/settings /development/
settings/users/01HZ
/settings
0
true
01HZ|fallback
[spaced]fallback
