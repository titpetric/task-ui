name: escape default
description: >
  A variable printed in a text node or a comment is escaped without the
  template asking for it. The value is what a form submits, so an
  unescaped echo would close the paragraph and open a script tag.
---
<?php

require 'vendor/autoload.php';

$tpl = new MiniTPL\Template("fixtures/");
$tpl->set_compile_location("compile/", true);
$tpl->load("text.tpl");
$tpl->assign("text", '</p><script>alert("xss")</script>');
$tpl->render();
---
<p>&lt;/p&gt;&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;</p>
<!-- &lt;/p&gt;&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt; -->
