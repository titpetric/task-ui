name: table
description: Table builds the plain list, and a cell escapes what a reader typed.
---
<?php

$table = new Task\Table;

$table->column("Job", "c-name")->column("Runs", "", true)->none("No jobs yet.");

$data = $table->data();

echo count($data["columns"]), " ", $data["width"], " ", $data["empty"], "\n";
echo $data["columns"][0]["title"], " ", $data["columns"][0]["class"], " ", var_export($data["columns"][1]["numeric"], true), "\n";
echo count($data["rows"]), "\n";

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

$table->row(array(Task\Table::cell("plan"), Task\Table::cell(3)), "row-1", "01hz", "bad");

$row = $table->data()["rows"][0];

echo $row["id"], " ", $row["run"], " ", $row["class"], " ", count($row["cells"]), "\n";
echo $table->data()["width"], "\n";

// A table with no columns still spans one, so the empty row draws.
echo (new Task\Table)->data()["width"], "\n";
---
2 2 No jobs yet.
Job c-name true
0
&lt;script&gt;alert(&quot;x&quot;)&lt;/script&gt;
num
<a href="/f/a&amp;b">A &amp; B</a>
<span class="faint">-</span>
<b>bold</b>
row-1 01hz bad 2
2
1
