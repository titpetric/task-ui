<?php

// Loaded ahead of every fixture, so a fixture names the class it is about.
//
// The host calls arrive as stubs/host.php, whose answers are statics a fixture
// writes: the real ones are Go and no fixture runs under them.

require "ext/autoload.php";
require "stubs/host.php";
