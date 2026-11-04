<?php

// Loaded ahead of every fixture, so a fixture names the class it is about.
//
// The host calls arrive as stubs/host.php, whose answers are statics a fixture
// writes: the real ones are Go and no fixture runs under them.
//
// stubs/fixture.php builds the shapes those statics hold, so a fixture names
// the keys its assertions turn on and no more.

require "ext/autoload.php";
require "stubs/host.php";
require "stubs/fixture.php";
