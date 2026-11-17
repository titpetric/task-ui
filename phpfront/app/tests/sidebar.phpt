name: sidebar
description: Task\Sidebar is the top block of the sidebar as one reader sees it - the modules that enabled a row, in the table's order, less what is hidden and what their rules deny, each row named by its own modinfo.php.
---
<?php

// The menu is the module list, so the fixture writes the modules and the
// table holds only their order and visibility. A row's title and icon are the
// module's own: the saved row spells both differently, and neither is read.

$db = Task\DB::writer();

$db->replace("sidebar", array("id" => "fx-sb-1", "kind" => "view", "href" => "/fx-sb/runs", "title" => "Stale", "icon" => "box", "position" => 3, "rule" => "run"));
$db->replace("sidebar", array("id" => "fx-sb-2", "kind" => "view", "href" => "/fx-sb/notes", "title" => "Notes", "icon" => "note", "position" => 2, "rule" => "note/view"));
$db->replace("sidebar", array("id" => "fx-sb-3", "kind" => "view", "href" => "/fx-sb/open", "title" => "Open", "icon" => "", "position" => 1, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sb-4", "kind" => "view", "href" => "/fx-sb/off", "title" => "Hidden", "icon" => "box", "position" => 0, "hidden" => 1, "rule" => ""));
$db->replace("sidebar", array("id" => "fx-sb-5", "kind" => "view", "href" => "/fx-sb/settings", "title" => "Settings", "icon" => "cog", "position" => 4, "rule" => "settings"));
$db->replace("sidebar", array("id" => "fx-sb-6", "kind" => "module", "href" => "/m/titpetric/fx-sb", "title" => "A module", "icon" => "flow", "position" => 5, "rule" => ""));

function mod($name, $title, $icon, $href, $sidebar, $rule) {
	$module = new Task\Module;
	$module->setName($name);
	$module->setTitle($title);
	$module->setIcon($icon);
	$module->setSideBarHref($href);
	$module->setSideBarRule($rule);

	if ($sidebar) {
		$module->enableSideBar();
	}

	return $module;
}

$modules = array(
	mod("taskui/runs", "Runs", "play", "/fx-sb/runs", true, "run"),
	mod("taskui/notes", "Notes", "note", "/fx-sb/notes", true, "note/view"),
	mod("taskui/open", "Open", "", "/fx-sb/open", true, ""),
	mod("taskui/off", "Hidden", "box", "/fx-sb/off", true, ""),
	mod("taskui/settings", "Settings", "settings", "/fx-sb/settings", true, "settings"),
	mod("taskui/module", "A module", "flow", "/m/titpetric/fx-sb", true, ""),
	mod("taskui/new", "New", "grid", "/fx-sb/new", true, ""),
	mod("taskui/quiet", "Quiet", "box", "/fx-sb/quiet", false, ""),
);

$modules[0]->setDescription("Every run.");

Task\Host::$rules = array(
	"paths" => array("note/view", "run", "user/view"),
	"checks" => array("note/view", "run", "user/view"),
	"extras" => array(),
);
Task\Host::$base = "/development";

$sidebar = new Task\Sidebar(null, $modules);

// mine is the rows this fixture's modules answer, each as its title and the
// link the component draws.
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
// are the open one, Runs and the module. The hidden row is nobody's, and a
// module that never enabled a row is never listed.
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

// The settings table is every enabled module, hidden ones too, in the table's
// order: the one no row saved goes last. Each carries the module's own title,
// icon and description, not the saved row's.
foreach ($sidebar->rows($sidebar->all()) as $row) {
	echo $row["title"], "=", $row["hidden"] ? "hidden" : "shown", "=", $row["kind"], "=", $row["icon"], "=", $row["note"], "\n";
}

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

// The real module tree answers the list the page draws: every module, and
// the ones whose modinfo.php enabled a row among them.
$real = task_modules();
$shown = 0;

foreach ($real as $module) {
	if ($module->getSidebar()) {
		$shown++;
	}
}

echo count($real), " ", $shown, " ", $real[0]->getSideBarHref(), "\n";
---
Open=/development/fx-sb/open,Runs=/development/fx-sb/runs,A module=/m/titpetric/fx-sb,New=/development/fx-sb/new
Open=/development/fx-sb/open,Runs=/development/fx-sb/runs,Settings=/development/fx-sb/settings,A module=/m/titpetric/fx-sb,New=/development/fx-sb/new
Open=/development/fx-sb/open,A module=/m/titpetric/fx-sb,New=/development/fx-sb/new
/development/fx-sb/runs /m/titpetric/fx-sb
Hidden=hidden=view=box=
Open=shown=view=box=
Notes=shown=view=note=
Runs=shown=view=play=Every run.
Settings=shown=view=settings=
A module=shown=module=flow=
New=shown=view=grid=
true true
false false true
play note-o script box
module A module NULL
11 9 /artifacts
