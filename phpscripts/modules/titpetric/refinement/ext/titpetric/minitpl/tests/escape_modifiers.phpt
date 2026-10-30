name: escape modifiers
description: >
  Modifiers are a chain, so more than one applies and raw can end it.
  Escaping is the outermost step whatever order the modifiers came in,
  which is why toupper still produces escaped output.
---
<?php

require 'vendor/autoload.php';

$tpl = new MiniTPL\Template("fixtures/");
$tpl->set_compile_location("compile/", true);
$tpl->load("modifiers.tpl");
$tpl->assign("value", '<b>Mixed</b>');
$tpl->render();
---
escape: &lt;b&gt;Mixed&lt;/b&gt;
toupper: &lt;B&gt;MIXED&lt;/B&gt;
tolower: &lt;b&gt;mixed&lt;/b&gt;
toupper raw: <B>MIXED</B>
raw: <b>Mixed</b>
