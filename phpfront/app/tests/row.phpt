name: row
description: Row reads a column of a database row, coercing what a driver answered.
---
<?php

$row = array(
	"id" => "01HZ",
	"count" => "7",
	"on" => 1,
	"off" => 0,
	"word" => "false",
	"at" => "2026-11-02 10:00:00",
	"blank" => "",
);

echo Task\Row::text($row, "id"), "\n";
echo Task\Row::text($row, "count"), "\n";
echo Task\Row::text($row, "missing"), "|", Task\Row::text($row, "missing", "fallback"), "\n";
echo var_export(Task\Row::number($row, "count"), true), "\n";
echo var_export(Task\Row::number($row, "missing", 3), true), "\n";
echo var_export(Task\Row::flag($row, "on"), true), var_export(Task\Row::flag($row, "off"), true), "\n";
echo var_export(Task\Row::flag($row, "word"), true), var_export(Task\Row::flag($row, "blank"), true), "\n";
echo var_export(Task\Row::flag($row, "missing"), true), "\n";
echo Task\Row::stamp($row, "at"), "|", Task\Row::stamp($row, "missing"), "\n";

// A driver answers a timestamp column as a time and not as a string, and a
// cast of one renders whole seconds. A read keeps the shape the column holds,
// so the fraction that orders two rows of the same second survives it. The
// row is read by its own ID, because the table is the whole instance's.
$db = Task\DB::writer();

$db->replace("run", array("id" => "fx-row-1", "folder" => "fx-row", "job" => "plan", "started_at" => "2026-11-02 10:00:00.522 +0000 UTC"));

$read = $db->get("SELECT started_at FROM run WHERE id = ?", "fx-row-1");

echo var_export(is_object($read["started_at"]), true), " ", Task\Row::stamp($read, "started_at"), "\n";
echo Task\Clock::nanos(Task\Row::stamp($read, "started_at")), " ", Task\Clock::seconds(Task\Row::stamp($read, "started_at")), "\n";

// Nothing is a row either, and a reader answers its fallback rather than throwing.
echo var_export(Task\Row::text(false, "id"), true), var_export(Task\Row::number(false, "id"), true), "\n";

$ws = new Task\Model\Workspace(array("id" => "01", "folder" => "internal"));

echo count(Task\Row::values($ws)), "\n";
echo json_encode(Task\Row::only($ws, array("id", "folder"))), "\n";
echo Task\Row::placeholders(array("a", "b", "c")), "|", Task\Row::placeholders(array()), "|\n";
---
01HZ
7
|fallback
7
3
truefalse
falsefalse
false
2026-11-02 10:00:00|
true 2026-11-02 10:00:00.522 +0000 UTC
1793613600522000000 1793613600
''0
13
{"id":"01","folder":"internal"}
?,?,?||
