<?php

// The catch-all. Every request the public/ tree does not answer arrives here,
// with $_SERVER["REQUEST_URI"] holding the path the browser requested, the
// /m/titpetric/refinement prefix included.
//
// What this echoes becomes the body of a task-ui page: the layout, the sidebar
// and the footer are the instance's.

require "vendor/autoload.php";

$base = $_SERVER["TASK_UI_MODULE_BASE"];

$tpl = new MiniTPL\Template("templates/");

$tpl->set_compile_location("templates_c/", true);

// The page's heading, its description and its badges are task-ui's: the
// module registered them in modinfo.php and the layout draws them. What
// this renders is the body under them.
$tpl->load("thread.tpl");
$tpl->assign("base", $base);

echo $tpl->get();
