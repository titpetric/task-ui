name: pager
description: Pager keeps the whole query of the page it links from, less the cursors, and encodes it the way a Go query is encoded.
---
<?php

// The query the page was read with, which the links keep: a field the form
// draws, a field no form draws, an empty field, and the cursor this page
// arrived on.
$_GET = array(
	"n" => "15",
	"folder" => "fx-pager",
	"dir" => "a/b",
	"batch" => "x y",
	"state" => "",
	"before" => "fx-pager-9",
);

$pages = Task\Pager::of("/runs", "fx-pager-1", "fx-pager-4");

// By name, which is the order a Go query encodes: the cursor of this link is
// written, the one the page arrived on is dropped, a slash and a space are
// escaped, and an empty field is kept because the Go page keeps it.
echo $pages["prev"], "\n";
echo $pages["next"], "\n";

// Neither cursor draws no nav at all, which is a list that fits on one page.
$one = Task\Pager::of("/runs", "", "");

echo "[", $one["prev"], "] [", $one["next"], "]\n";

// A page read with no query at all links by the cursor alone.
$_GET = array();
$bare = Task\Pager::of("/notes", "fx-pager-1", "");

echo $bare["prev"], "\n";

// size is the rows a page holds, and the pick is in the query the links keep:
// a page of fifty rows links to another fifty, and a page that asked for no
// size links to the default rather than to nothing, which is what Go's
// pageSize leaves on the values the pager is handed.
//
// $_REQUEST is what asks, $_GET what the link keeps, as in a request.
$_REQUEST = array();
$_GET = array();
$size = Task\Pager::size();
$links = Task\Pager::of("/runs", "", "fx-pager-4");

echo $size, " ", $links["next"], "\n";

// A size the list does not offer is the default, in the link as in the form.
$_REQUEST = array("n" => "7");
$_GET = array("n" => "7", "user" => "admin");
$size = Task\Pager::size();
$links = Task\Pager::of("/runs", "", "fx-pager-4");

echo $size, " ", $links["next"], "\n";

// One it offers is kept.
$_REQUEST = array("n" => "50");
$_GET = array("n" => "50");
$size = Task\Pager::size();
$links = Task\Pager::of("/runs", "", "fx-pager-4");

echo $size, " ", $links["next"], "\n";
---
/runs?after=fx-pager-1&batch=x+y&dir=a%2Fb&folder=fx-pager&n=15&state=
/runs?batch=x+y&before=fx-pager-4&dir=a%2Fb&folder=fx-pager&n=15&state=
[] []
/notes?after=fx-pager-1
15 /runs?before=fx-pager-4&n=15
15 /runs?before=fx-pager-4&n=15&user=admin
50 /runs?before=fx-pager-4&n=50
