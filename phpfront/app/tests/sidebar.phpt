name: sidebar
description: Task\Sidebar is the top block of the sidebar as one reader sees it - the table's order, less what is hidden and what their rules deny, each row with the whole link and the flags the component keys on.
---
<?php

// items reads the whole sidebar table, because the block is the instance's
// own. So the answer is narrowed to the rows this fixture wrote: one
// connection serves the suite and another fixture's row is as real as these.
// The href carries a unique index, so the rows are written with replace.

$db = Task\DB::writer();

// The position runs against the titles on purpose, so the order read is the
// column's and not the alphabet's.
$db->replace("sidebar", array("id" => "fx-sb-1", "kind" => "view", "href" => "/fx-sb/runs", "title" => "Runs", "icon" => "play", "position" => 3, "rule" => "run"));
$db->replace("sidebar", array("id" => "fx-sb-2", "kind" => "view", "href" => "/fx-sb/notes", "title" => "Notes", "icon" => "note", "position" => 2, "rule" => "note/view"));
$db->replace("sidebar", array("id" => "fx-sb-3", "kind" => "view", "href" => "/fx-sb/open", "title" => "Open", "icon" => "", "position" => 1, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sb-4", "kind" => "view", "href" => "/fx-sb/off", "title" => "Hidden", "icon" => "box", "position" => 0, "hidden" => 1, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sb-5", "kind" => "view", "href" => "/fx-sb/settings", "title" => "Settings", "icon" => "cog", "position" => 4, "rule" => "settings"));
$db->replace("sidebar", array("id" => "fx-sb-6", "kind" => "module", "href" => "/m/titpetric/fx-sb", "title" => "A module", "icon" => "flow", "position" => 5, "rule" => ""));

Task\Host::$rules = array(
	"paths" => array("note/view", "run", "user/view"),
	"checks" => array("note/view", "run", "user/view"),
	"extras" => array(),
);
Task\Host::$base = "/development";

$sidebar = new Task\Sidebar;

// mine is the rows this fixture wrote, in the order they came back, each as
// its title and the link the component draws.
function mine($items) {
	$out = array();

	foreach ($items as $item) {
		if (strpos($item["href"], "/development/fx-sb/") === 0 || strpos($item["href"], "/m/titpetric/fx-sb") === 0) {
			$out[] = $item["title"] . "=" . $item["href"];
		}
	}

	return $out;
}

// The reader runs jobs and reads nobody's settings, so the rows they may see
// are the open one, Runs and the module. The hidden row is nobody's.
$runner = new Task\Model\User(array("id" => "fx-sb-u1", "name" => "fx-sb-ann", "role" => "runner"));
$role = new Task\Policy(array("run" => Task\Policy::ALLOW));
$access = new Task\Access($runner, $role);
$items = $sidebar->items($access, "/fx-sb/runs");

echo implode(",", mine($items)), "\n";

// A reader who may read a section of the settings sees the settings row,
// which is the one rule that is not a permission path.
$reads = new Task\Policy(array("run" => Task\Policy::ALLOW, "user/view" => Task\Policy::ALLOW));
$keeper = new Task\Access($runner, $reads);

echo implode(",", mine($sidebar->items($keeper, "/fx-sb/runs"))), "\n";

// A reader with no policy at all sees the rows that carry no rule.
$bare = new Task\Policy();
$nobody = new Task\Access($runner, $bare);

echo implode(",", mine($sidebar->items($nobody, "/fx-sb/runs"))), "\n";

// The link is the whole one: this application's rows carry the prefix, and a
// module's row does not, because a module is another application on the
// instance and sits outside this one's mount.
$links = array();

foreach ($items as $item) {
	$links[$item["title"]] = $item["href"];
}

echo $links["Runs"], " ", $links["A module"], "\n";

// The page being read is the row it sits under: the row's own path, and
// anything below it. The dashboard is its own path alone, so a page does not
// light every row.
echo var_export(Task\Sidebar::active("/fx-sb/runs", "/fx-sb/runs"), true), " ", var_export(Task\Sidebar::active("/fx-sb/runs", "/fx-sb/runs/01hz"), true), "\n";
echo var_export(Task\Sidebar::active("/fx-sb/runs", "/fx-sb/runsaway"), true), " ", var_export(Task\Sidebar::active("/", "/fx-sb/runs"), true), " ", var_export(Task\Sidebar::active("/", "/"), true), "\n";

// An icon from the database reaches the markup as a sprite name and nothing
// else: what is not a sprite name's character is dropped, so a value that
// tries to close the attribute arrives as a name that names no symbol, and a
// name with nothing left falls back to the one every list uses.
echo Task\Sidebar::icon("Play"), " ", Task\Sidebar::icon("note-o"), " ", Task\Sidebar::icon('"/><script>'), " ", Task\Sidebar::icon(""), "\n";

// One row by its link, and nothing for a link that names none.
$row = $sidebar->byHref("/m/titpetric/fx-sb");

echo $row->getKind(), " ", $row->getTitle(), " ", var_export($sidebar->byHref("/fx-sb/nope"), true), "\n";
---
Open=/development/fx-sb/open,Runs=/development/fx-sb/runs,A module=/m/titpetric/fx-sb
Open=/development/fx-sb/open,Runs=/development/fx-sb/runs,Settings=/development/fx-sb/settings,A module=/m/titpetric/fx-sb
Open=/development/fx-sb/open,A module=/m/titpetric/fx-sb
/development/fx-sb/runs /m/titpetric/fx-sb
true true
false false true
play note-o script box
module A module NULL
