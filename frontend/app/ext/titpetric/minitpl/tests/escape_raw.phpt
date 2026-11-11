name: escape raw
description: >
  A value that is already markup opts out of escaping with the raw
  modifier, or its unescape alias. The opt-out is per tag, so the same
  variable is escaped and raw in the same template.
---
<?php

require 'vendor/autoload.php';

$tpl = new MiniTPL\Template("fixtures/");
$tpl->set_compile_location("compile/", true);
$tpl->load("raw.tpl");
$tpl->assign("content", '<b>bold</b>');
$tpl->render();
---
<div>&lt;b&gt;bold&lt;/b&gt;</div>
<div><b>bold</b></div>
<div><b>bold</b></div>
<div title="<b>bold</b>">attribute</div>
