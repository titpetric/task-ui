name: table
description: Table builds the plain list, and a cell escapes what a reader typed.
---
<?php

$table = new Task\Table("jobs", "fit");

$table->head(array(Task\Table::cell("Job"), Task\Table::cell("Runs", "num")));
$table->row(array(Task\Table::cell("plan"), Task\Table::cell(3)), "row-1", "01hz", "bad");
$table->foot(array(Task\Table::raw("<b>2</b>", "", 2)));

$data = $table->data();

echo $data["class"], " ", $data["wrap"], "\n";
echo count($data["header"]), " ", count($data["body"]), " ", count($data["footer"]), "\n";
echo $data["header"][0]["cells"][1]["html"], " ", $data["header"][0]["cells"][1]["class"], "\n";

$row = $data["body"][0];

echo $row["id"], " ", $row["run"], " ", $row["class"], " ", count($row["cells"]), "\n";
echo $data["footer"][0]["cells"][0]["html"], " ", $data["footer"][0]["cells"][0]["span"], "\n";

$bare = new Task\Table;
$empty = $bare->data();

echo "[", $empty["class"], "] [", $empty["wrap"], "] ", count($empty["header"]), count($empty["body"]), count($empty["footer"]), "\n";

// A cell escapes. The component draws a cell raw, so this is the whole of
// what keeps a value a reader typed out of the markup.
$cell = Task\Table::cell('<script>alert("x")</script>');

echo $cell["html"], "\n";
echo Task\Table::cell("plain", "num")["class"], "\n";

// link escapes both halves.
echo Task\Table::link('/f/a&b', "A & B")["html"], "\n";
echo Task\Table::faint("-")["html"], "\n";

// raw is the opt out, for markup the page built.
echo Task\Table::raw("<b>bold</b>")["html"], "\n";
---
jobs fit
1 1 1
Runs num
row-1 01hz bad 2
<b>2</b> 2
[] [] 000
&lt;script&gt;alert(&quot;x&quot;)&lt;/script&gt;
num
<a href="/f/a&amp;b">A &amp; B</a>
<span class="faint">-</span>
<b>bold</b>
