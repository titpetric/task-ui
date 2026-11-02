name: escape attributes
description: >
  Double quoted, single quoted and unquoted attribute values all escape.
  ENT_QUOTES covers both quote characters, so a value cannot break out of
  either attribute style and add an event handler of its own.
---
<?php

require 'vendor/autoload.php';

$tpl = new MiniTPL\Template("fixtures/");
$tpl->set_compile_location("compile/", true);
$tpl->load("attributes.tpl");
$tpl->assign("value", '" \' onclick=alert(1) &');
$tpl->render();
---
<p class="&quot; &#039; onclick=alert(1) &amp;" data-id='&quot; &#039; onclick=alert(1) &amp;' title=&quot; &#039; onclick=alert(1) &amp;>ok</p>
