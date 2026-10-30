name: escape script and style
description: >
  Script and style bodies are not markup, so html escaping there would
  corrupt the javascript or css instead of protecting it. The closing tag
  returns the scanner to text context, where escaping resumes.
---
<?php

require 'vendor/autoload.php';

$tpl = new MiniTPL\Template("fixtures/");
$tpl->set_compile_location("compile/", true);
$tpl->load("script.tpl");
$tpl->assign("json", '{"a":1 && 2}');
$tpl->render();
---
<script type="text/javascript">
var config = {"a":1 && 2};
</script>
<style>
body { content: "{"a":1 && 2}"; }
</style>
<p>{&quot;a&quot;:1 &amp;&amp; 2}</p>
