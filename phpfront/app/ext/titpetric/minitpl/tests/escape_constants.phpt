name: escape constants
description: >
  A {_CONSTANT} carries whatever define() was given it, so it escapes by
  context like any other value. A quote in a constant would otherwise close
  the attribute it sits in.
---
<?php

require 'vendor/autoload.php';

define("_SITE_NAME", 'Tom & "Jerry"');

$tpl = new MiniTPL\Template("fixtures/");
$tpl->set_compile_location("compile/", true);
$tpl->load("constants.tpl");
$tpl->render();
---
<p title="Tom &amp; &quot;Jerry&quot;">Tom &amp; &quot;Jerry&quot;</p>
<script>var name = "Tom & "Jerry"";</script>
<div>Tom & "Jerry"</div>
