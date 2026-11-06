# The front end

task-ui's pages are being replaced by a [phpscript](https://github.com/titpetric/phpscript) application: a theme of components, a module per view, and the task-ui data model as PHP classes over the instance's database. The tree is `phpfront/app`, the Go that mounts it is `phpfront/`.

**The ported pages hold the root: they are the front end the instance serves.** `app.FrontBase` is their mount and it is the empty string, so a template writing `{base}/settings` writes `/settings`. The Go pages answer under `app.GoBase`, `/legacy`, and both of those lines are the whole of the arrangement.

The Go pages are kept to be read against the ported ones and not to be browsed. The Go view layer has no notion of a mount, so a page under the prefix draws its links without one and a link followed from it leaves for the ported page at the root. Reading one page and diffing it is what they are for, and `scripts/compare.sh` is what fetches them. A view proven equivalent leaves a `TODO: delete, replaced with phpscript, <hash>` on the Go handler, its template and the view helpers only it used; the cutover drops `GoBase` and every marked file in one commit.

Five things answer at the root because they are pages of neither front end, and `frontend.Handlers.Root` is where they are registered: the session POSTs, `/healthz`, the theme's files, the four byte streams, and a plugin's pages under `/m`.

Equivalent means the whole page, chrome included, so no page is marked until the sidebar's folder tree arrives with the `folders` module: every page draws it, and a marked Go handler would be deleted in favour of a page whose sidebar is empty. A helper is marked only where no unported page calls it, and a partial only where no unported template includes it, which today leaves `_thread.tpl`, `_tooling.tpl`, `_chart.tpl` and `_schedule.tpl` unmarked whatever their pages do. `scripts/compare.sh` is what proves a page, and it takes several in one run.

**The whole site is the gate, and the crawl is the page list.** `scripts/compare-site.sh` walks every link from `/` with `scripts/crawl.sh` and hands each page it reached to `scripts/compare.sh`, which is 224 pages of the demo data against 46 a list kept by hand named — every job page of every folder, every filter a query writes, every row's own page. A list goes stale the moment a page is added and reads as a clean sweep; the site cannot. `scripts/demo.sh` runs it over the instance it builds, as the admin and again as the viewer, so the comparison is [a suite](testing.md#the-front-end) rather than a ritual: `TASK_UI_SHOT_FILTER='^$' scripts/demo.sh` is the crawl and the comparison with no screenshot taken.

**Both front ends are crawled and the two page lists are read together**, because a page list is one front end's own links and each crawl is blind to a page only the other draws: a sidebar row one of them has and the other does not would never be fetched, and a page never fetched reads as a page compared. The second crawl puts `GoBase` on every path, with `TASK_UI_CRAWL_PREFIX`. Its status is discarded and only its paths are kept: a Go page links the root as well as itself — the telemetry dashboard, the byte streams, a module — so prefixing those fabricates a 404 that is not a break. The broken-link check is the crawl of the root alone, which is the front end a reader meets.

Two kinds of path are not compared, and the run names each one it dropped. A byte stream is not a page, which is `/files/*`, `/artifacts/files/*` and the archives. A page under `/m` belongs to a module and not to this front end: the refinement board is the one that moved here, and the move changed the page — the module chrome around it is gone, its badges and its render footer with it, and the thread page became a page with a heading, a crumb trail and the thread's files in its actions. Its `.phpt` fixtures are what cover it, as for any module's PHP.

[Corrections](todo-correction.md), sections 9 to 18, records every place the runtime lacks what this needs, with the workaround taken.

## The tree

| Path                          | What it is                                                                        |
|-------------------------------|-----------------------------------------------------------------------------------|
| `composer.json`               | `vendor-dir: ext`, psr-4 over `code/Task`, a classmap over `code/` and `modules/` |
| `ext/`                        | what `composer install` wrote, the autoloader among it                            |
| `code/Module.php`             | the interface every `mod_` class declares                                         |
| `code/Task/`                  | the task-ui domain: a DAO per table, and the helpers every page shares            |
| `code/Task/Model/`            | generated from `schema/schema.yml`; do not edit                                   |
| `theme/templates/index.tpl`   | the page shell, which places the components                                       |
| `theme/templates/components/` | a component and its `.md` per piece of shared markup                              |
| `theme/assets/`               | the stylesheet, the script, the fonts and the icon, served at `/assets/theme/`    |
| `modules/<name>/`             | one directory per sidebar entry                                                   |
| `stubs/`                      | the host calls, for the linter and the fixtures; nothing requires them at runtime |
| `tools/`                      | the generators and the precompiler                                                |
| `tests/`                      | the `.phpt` fixtures, and `bootstrap.php`                                         |

A module directory holds:

| Path                                     | What it is                                                       |
|------------------------------------------|------------------------------------------------------------------|
| `mod_<name>.php`                         | one per view: its `@route` lines, its class, its dispatch        |
| `schema/`                                | the `*.up.sql` of each table this module's pages own             |
| `storage/`                               | the SQL of a data model the module owns                          |
| `templates/index.tpl`                    | the module's body; more templates beside it                      |
| `templates_c/`                           | the same compiled, written by `tools/precompile.php` and shipped |
| `assets/`                                | served at `/assets/<name>/`                                      |
| `api/`                                   | new JSON endpoints, `@route` annotated                           |
| `docs/`                                  | what a module owning `storage/` adds to its generated README     |
| `README.md`, `schema.puml`, `schema.svg` | generated; the last two only where `tables()` names one          |

`schema/` is where a table's migration lives, with the pages that own it, and `//go:embed all:app` is what ships it. `phpfront.Migrations()` answers the module tree and `app.Migrations` merges it with the other three owners — `keystore/schema/`, `access/schema/`, `phpscript/schema/` — into one flat filesystem keyed by base name, which the instance applies at start. Flat is load-bearing: the `migrations` table records the name of a file relative to the root it was applied from, so a file that gains a directory is a file nobody ran, and the second `ALTER TABLE` of a column that is already there takes the database with it. [schema/schema.go](../schema/schema.go) carries the whole argument, and `atkins schema` is what regenerates the data model after adding one.

Ownership is a decision and not a read set. `user` and `run` are each read by seven modules; what decides is which pages' writes the table exists for. The 12 tables no page queries sit with the Go package that owns them.

Twelve modules in, none has a `docs/`. The generated `README.md` already lists the tables and the operations over them, and a module that reads the instance's rows through `code/Task` owns no data model to write up; the directory is for one that owns a model through `storage/` and needs prose the generator cannot write. An empty `docs/` restating its README is worse than no directory.

## A module

`mod_<name>.php` holds its own routes, declares its class and runs its dispatch. The pooled runtime drops what an endpoint declared between requests, so declaring the class in the endpoint file is safe. The dispatch is guarded, so a fixture requires the file for the class alone.

```php
<?php

// @route GET /notes
// @route POST /notes

require "ext/autoload.php";

class mod_notes implements Module {
	function name() {
		return "notes";
	}

	function title() {
		return "Notes";
	}

	function icon() {
		return "note";
	}

	function tables() {
		return array("note", "user");
	}

	function serve() {
		$chrome = new Task\Chrome;
		$view = new Task\View($this, $chrome);

		if (!$chrome->allowed("note/view")) {
			Task\Route::refused();
			$chrome->heading("Not allowed");
			$view->page(array(), "forbidden.tpl");

			return;
		}

		$chrome->heading("Notes", array(array("text" => "Notes", "href" => "")));
		$view->page(array("threads" => $threads));
	}
}

if (PHP_SAPI === "http") {
	$mod = new mod_notes;

	$mod->serve();
}
```

`Module` names five methods and an interface here inherits nothing, so each one is written out. `tables()` is what the generated docs read.

**`tables()` is the module body's own read set, not the chrome's.** Every page draws the chrome, and `Task\Chrome` reads `user`, `workspace`, `sidebar`, `pin`, `workflow`, `workflow_step`, `key_info`, `key_role`, `key_user` and the recents' `run` and `note` — eleven tables on every page in the tree. Naming them in each `tables()` would draw the same eleven-table diagram twelve times and the per-module diagram would stop telling modules apart, which is the only thing it is for. So a module names what its own body queries, the directory's data model is what its pages name together, and a module that queries nothing names nothing: `tools/docs.php` then writes no `schema.puml`, deletes a stale one, and the README says in words that the chrome does the reading. `modules/notfound/` is the one that does.

A path is written without the mount. `Task\Host::base()` answers it and the templates write it as `{base}`, which is the empty string now that the ported pages hold the root; nothing in the tree spells a prefix of its own, and nothing spells the Go pages' `/legacy` either. `rules/` refuses both literals.

A file holds no `use` block. `phpscript fmt` inlines an alias and leaves the import behind, so a file writes the qualified name the formatter would leave: `Task\Chrome` from the global namespace of a `mod_` file, `\Task\Chrome` from inside a namespace.

### Statements

**An allocation is a statement of its own, bound to a name.** `(new Task\View($this, $chrome))->page($scope)` is three things on one line: it builds a view, calls it and names neither. It is written as two statements, and the name says what was built:

```php
$view = new Task\View($this, $chrome);

$view->page($scope, "run.tpl");
```

The same holds for an allocation reached as an argument, as an array value or inside a ternary: `Task\User::options($users, (new Task\Key)->ofUsers($users))` reads the key table in the middle of an argument list, where nothing names the store it read it from. `rules/` refuses all four shapes.

An argument is a name or a literal wherever the call it would otherwise hold does work worth naming. A read, a build and a call that each do something are each a line:

```php
$store = new Task\Key($this->db);
$keys = $store->ofUsers($users);

return Task\User::options($users, $keys);
```

This is the same reason a `get` answers through a guard rather than a ternary. `is_array($row) ? new Model\Note($row) : null` hides the allocation in a branch; the guard puts it on its own line and leaves one way out:

```php
if (!is_array($row)) {
	return null;
}

return new Model\Note($row);
```

### Routes

`// @route GET /path` and `// @route POST /path`. A line naming no method routes both. A path parameter arrives in `$_REQUEST` under the name the annotation declared, so `/runs/{id}` reads `Task\Route::param("id")`.

The matched pattern reaches no superglobal, so a file holding several routes matches on the method and the segments: `Task\Route::posted()` and `Task\Route::segments()`.

An `@route` never names a file. The static files are Go's, at `/assets/theme/` and `/assets/<module>/`, and `rules/` refuses a route path with an extension.

**A byte stream is not a page and is not ported.** `/artifacts/files/*`, `/artifacts/*.tgz` and `/artifacts.tgz` answer with the bytes of a file of the artifact tree or a tar of a directory of it, which the tree opens outside the directory the PHP may read; they stay Go, at the root of the instance, in the same class as `/login` POST and `/logout`. `/files/*` is one of them, which the plan had down as a page: `frontend/file.go`'s `checkoutFile` resolves the folder, joins `f.Path(rel)`, demands a regular file and hands it to `serveFile`, and the only markup it can answer with is the shared 404. So the plan's `mod_file` row drops, and a folder's listing links `/files/...` unmounted. So a link to a stream is written **without** `Task\Host::base()` — `mod_artifacts::FILES` and `mod_artifacts::archive()` are the two places that spell them — and every link to a page of this application carries the mount. The same holds for `/api/*`, which the browser reads from `task-ui.js` and which has no mount either.

**The sign in page answers before a sign in, and it is the only one that does.** The Go router puts the Guard on a chi group with `/login`, `/healthz` and the assets outside it (`frontend/frontend.go:43-49`), because a sign in page behind a sign in redirects to the page it is already being asked for. Everything under a mount is one handler, so `phpfront.Public` is that list for this tree — `/login`, and nothing else — and `Front.Guarded` is the shape: `Guard.Attach` for a public path, `Guard.Require` for every other. `Attach` names the reader of a request and keeps nobody out, which the page needs both ways round: it draws the form for a reader with no session and redirects a reader who has one home, the same two answers `frontend/login.go:12` gives. Without it the ported page is unreachable in a browser and only a patched binary compares it.

The redirect is one string on both sides: Go sends a signed in reader to `/`, and `{base}/` is `/` with the base empty. A prefixed mount is what made the two differ, and the page is the only one that has a redirect of its own. `scripts/compare.sh` reads the body and not the headers, so it measures neither.

**A path no route claimed is a page too.** The Go router draws the shared 404 from `r.NotFound` behind the same guard as everything else (`frontend/frontend.go:115`), so the ported tree needs the same catch-all or an unrouted path answers chi's own plain `404 page not found`, which is no page at all. `modules/notfound/mod_notfound.php` is it: `@route GET /{rest...}` and the same for `POST`, because chi matches a literal and a parameter ahead of a wildcard and a methodless line routes only those two. It is the one module with no sidebar row and no templates of its own — `Task\View::missing()` draws the theme's `notfound.tpl`, and MiniTPL skips a search path that is not there. The reason reads `There is nothing at /` + `Task\Route::path()`, which drops the mount, so the string the reader sees is the one Go writes and the cutover changes nothing.

**One wildcard, one module.** `/f/*` answers three pages — a folder, a directory inside it and a job — and the path alone says which, so one route resolves all three and `mod_folder` owns it. The plan's second row for the subtree, `mod_job`, is therefore not a module: the job page is `Task\Job`, which is where every view model that is not one page's own already lives (`Task\Chain`, `Task\Control`, `Task\Pipeline`, `Task\HostJobs`). A module-local class is autoloadable — `Task\Schedules\Lookup` is in the committed classmap — but a class under `modules/` is a module's own data model, which a page view model is not.

**A link the script writes reads the mount off the page.** `task-ui.js` is one file both front ends serve, so a page path it composes from a live event — the recents, a run's tile, a step of a workflow, a folder's page — has to land on the front end that drew the page. `mount()` answers the prefix, read once off the brand's `href`, and nothing new is written into the markup to carry it. Both front ends write `/` there now that the ported pages hold the root, so it answers the empty string on either side and one script serves both; it is kept because it is what made the prefixed arrangement work and what any future mount would read. The `/api` and byte-stream paths do not ask, as above.

Where the element the script reads is server-drawn, the server names the URL and the script composes nothing: a host row of `/inventory` carries `data-host-url`, the page the expanded row is fetched from, written by both front ends and compared like any other link.

### Writes

**A page of the instance writes nothing, and a module's page is the one that does.** Every form on a page of the instance posts to `/api`, which is Go and stays Go, so the twenty-eight routes ported so far are all `GET` and none of them writes. A module is the exception the Go router already makes for the same reason — `frontend/frontend.go:85` wraps a module's pages in `access.SameOrigin` because a module's forms post to itself — and `modules/refinement` is the only module, so its five `POST` routes are every write this tree will hold until a page of the instance grows one of its own.

A handler validates and the storage class writes. One statement where that does, and a transaction inside the storage class where it does not. A handler that changed something answers with a redirect, `Task\Route::go("/notes")`, so a reload does not repeat the work.

A write publishes. `Task\Host::publish($topic, $payload)` sends the event every live label on every open page reads; a write that does not publish leaves the interface stale until a reload.

## The domain

`code/Task` holds a class per table and the helpers every page shares. A page reads rows itself: no binding answers a workspace, a run or a user.

| Class             | What it is                                                                             |
|-------------------|----------------------------------------------------------------------------------------|
| `Task\DB`         | `reader()`, which refuses every write, and `writer()`                                  |
| `Task\Row`        | the readers over a database row, and `values()` over a model                           |
| `Task\Model\*`    | a row: its columns, its getters and setters, `TABLE`, `$fields`, `$primary`, `$stamps` |
| `Task\Route`      | the request past the prefix, and the links back                                        |
| `Task\Chrome`     | who is reading, what they may do, the sidebar, the crumbs                              |
| `Task\View`       | a module's body inside the theme's shell                                               |
| `Task\Menu`       | the `[...]` menu of a row or a block                                                   |
| `Task\Who`        | a user as every page draws them: avatar, name, menu                                    |
| `Task\Access`     | what one user may do                                                                   |
| `Task\Policy`     | a subject's entries, an effect per path                                                |
| `Task\Permission` | `config/permissions.yml`, over `Task\Host::rules()`                                    |
| `Task\Clock`      | the two shapes a timestamp column takes                                                |
| `Task\Id`         | the 26 character identifier of a row                                                   |
| `Task\Text`       | the JSON a column holds where it holds a list or a map                                 |

A class owns a table and the behaviour of its rows both, because the runtime merges no methods into a generated class: `Task\User` holds the statements and `Task\User::admin($row)` answers what a row means.

**Every SQL statement lives in `code/Task/*.php` or `modules/*/storage/*.php`.** `rules/` refuses one anywhere else. A read page takes `Task\DB::reader()`, which refuses a statement that does not start with a read-only keyword, so a view that only lists cannot write.

A nullable timestamp column holds null and not `""`. The generated model names its timestamp columns and `Task\Row::values()` writes null for an empty one; writing `""` makes the Go scanner fail and the Go page lists nothing, silently.

A number a page prints is Go's arithmetic and not an equivalent of it. `Task\Activity` plots a duration in nanoseconds because the Go page plots a `time.Duration`: the mean of one truncates, and a level that sits within a millionth of a rounding boundary then rounds the other way from a mean of exact milliseconds. A scan of 224 000 `(spent, runs, slowest)` combinations prints a different second decimal on 1891 of them. Scaling both sides of a quotient by the same whole number costs nothing, since IEEE division is correctly rounded and `a/b` is `ka/kb` to the bit; a truncation in another unit is not free, and neither is a rounding. The sweep finds one of these a run, where the data lands on a boundary and nowhere else.

## The host

`Task\Host::x()` is Go. Every other `Task\` name is PHP. That spelling is the whole of the boundary.

All fifty-two of them, in the order `phpfront/bind.go` registers them. Calls sharing a reason share a row.

| Call                                            | Why it is not PHP                                                                    |
|-------------------------------------------------|--------------------------------------------------------------------------------------|
| `base()`                                        | where the router is mounted                                                          |
| `version()`                                     | the instance's version                                                               |
| `now()`                                         | the instance's clock, which `TASK_UI_NOW` pins                                       |
| `login()`                                       | whether a sign in is required, which the Guard decides                               |
| `assets()`, `assetVersion()`                    | where the static files answer, and their content hash                                |
| `whoami()`                                      | the name the Guard resolved; the row behind it is `Task\User`                        |
| `folders()`                                     | `folder.Catalogue` parses Taskfiles to find jobs                                     |
| `job($folder, $dir, $id)`                       | the job a run names, read off that directory's own Taskfile                          |
| `projects($workspace)`                          | the directories under a workspace's `src`, a walk of the filesystem                  |
| `folderDir($id, $workspaces)`                   | the directory a folder ID names, the catalogue's or the layout's                     |
| `isDir($folder, $rel)`                          | whether a directory a pin names is there, outside the PHP root                       |
| `commit($dir)`                                  | git's last commit time, through the cache both front ends read                       |
| `running()`                                     | the runner's memory, where a run lives until it ends                                 |
| `flows()`                                       | the workflow runner's memory, by the workflow each run is of                         |
| `session($id)`                                  | one run as the runner holds it, which is ahead of its row                            |
| `frame($id)`                                    | the last screen of a live run, as the SVG the page embeds                            |
| `screen($id, $cast, $cols)`                     | the last screen of a run that ended, replayed from its recording                     |
| `expired($cast, $state)`                        | whether the recording a run names is still on disk                                   |
| `claude()`                                      | whether the reader has a token for Claude, which the integration holds               |
| `dirty($dir)`                                   | git's answer, not a row                                                              |
| `rules()`                                       | `config/permissions.yml`, embedded in the binary                                     |
| `publish($topic, $payload)`                     | the stream every live label reads                                                    |
| `dispatch($comment, $action, $folder)`          | starting a run is the runner's, and the prompt reads a file outside the PHP root     |
| `run($note)`                                    | which run a dispatch turned into, asked once the board has started one               |
| `markdown($src, $prefix, $dir)`                 | the renderer that owns the content security policy                                   |
| `message($src, $prefix, $link)`                 | the same, with a line break kept and `@name` marked                                  |
| `artifacts($folder)`, `artifact($path)`         | the artifact tree, outside the directory the PHP may open                            |
| `tree($dir)`                                    | one directory of that tree, listed a level at a time                                 |
| `space()`, `spaces($dirs)`                      | `diskusage`, which walks the folders                                                 |
| `inventory($hosts)`, `host($n, $hosts)`         | `beszel`, `hostinfo` and the ssh prober, one instance of each                        |
| `reachable()`                                   | the same prober, for the hosts a command can run on                                  |
| `tooling()`                                     | atkins' report, which task-ui reads by running atkins                                |
| `place($folder, $rel)`                          | a folder of the catalogue, or a directory inside one                                 |
| `groups($folder, $rel, $seen, $shell, $claude)` | the Jobs table of a place: a row per pipeline file the walk found                    |
| `listing($folder, $rel, $seen)`                 | a directory as the Files panel draws it                                              |
| `rootTree($named, $seen)`                       | the root's own directories, as the root's page draws them                            |
| `modTime($dir)`                                 | when a directory last changed, which no row records                                  |
| `status($dir)`                                  | git's branch, ahead, behind and remote of a checkout                                 |
| `checkout($dir)`                                | the commit a checkout is at and the space it takes                                   |
| `hours($text)`                                  | an hours rule read, by the parser the scheduler obeys                                |
| `keys()`                                        | `keystore` holds private key material, which does not enter PHP                      |
| `keyFolders()`                                  | the `.ssh` of a checkout that keeps keys of its own, which no run is given           |
| `instance($admin)`                              | what the server runs with: its flags, its environment, its paths                     |
| `usage()`                                       | what each area of the retention table holds now, measured off disk                   |
| `telemetry()`                                   | the options the tracer runs on, which the tracer owns                                |
| `integrations()`                                | the services' registry, Go data, and their settings, which the instance key encrypts |
| `modules()`                                     | the plugins the instance serves under `/m`, which the module host loaded             |
| `entries($workspaces)`                          | the project directories under each workspace's `src`, a walk                         |

A Go slice crosses as a PHP array and a Go struct crosses as a host object whose fields no key reaches, so `phpfront/convert.go` converts: everything that arrives is an array all the way down, keyed by the names `phpfront/host.go` declares as json tags. A binding that answers a shape adds a converter there and a case to `bind_test`.

`stubs/host.php` declares the same calls for the linter and the fixtures, each answering a static a fixture writes. Keep it in step with `phpfront/bind.go`. `stubs/fixture.php` builds the shapes those statics hold, so a fixture names the keys its assertions turn on and a key added to an answer is added in two places and in no fixture.

**A state column is not a live fact.** A run and a workflow run are written `running` and written through as they end, so a process that died leaves the state behind: the row says running for good. Go reads neither off the table — `runner.Manager` holds the runs and `workflow.Runner.Active` the workflow runs, both in the memory of the process that started them — and a page that read the column instead would mark a run nobody is watching, differ from the Go page, and keep differing. Both are bindings for that reason, `running()` and `flows()`, and PHP folds each into a map once a page rather than asking per row. `session($id)` is the one that is asked per run: it answers the whole of one run as the runner holds it, which the run's own page draws and a list has no use for.

**A gate carries what the Go gate reads, or it is a different gate.** `view.Page.CanRun` takes the allow list of a job and of its folder and the job's kind, and resolves a run's job through `f.Sub(dir).Job(id)`, which reads the Taskfile of the directory the run ran in and answers the four jobs a folder declares nowhere. A binding that answered a folder's job *names* let through a job whose allow list named somebody else and every job of a directory, on a page whose HTML matched because the instance it was compared on signs in as an admin. Any gate a page draws is transcribed from the Go one and the binding answers what that transcription reads, which for a run is `Task\Host::job()`: whether the job is there at all, the two kinds a permission gates, and both allow lists in one answer.

**The clock is a host fact too.** `TASK_UI_NOW` pins what the hours gate and every Go page read, which is what makes a screenshot of an activity chart the same picture twice, so `Task\Clock::now()` asks the host for this moment and a runtime clock is the fallback, for a fixture, which has no instance to pin. Reading `\DateTime::now()` instead drew a window of thirty days ending today against a Go window ending on the pinned Friday: seventeen pages differed in the window's dates, in every bar's tooltip, in all three chart paths and in the axis labels under them — a divergence a whole-site comparison finds in one run and a page-at-a-time sweep of an unpinned instance never finds at all.

The shape is RFC 3339 to the nanosecond and not the shape a timestamp column holds, because the zone has to survive the crossing: `view.NewActivity` takes `loc := now.Location()` and counts a chart's days in it, so a reading converted to UTC on the way puts a run of the late evening in tomorrow on one front end and in today on the other. The demo instance pins `+02:00` on a machine that reads `+01:00`, which is the case that proves it.

**Two clocks cross, because two clocks are read in Go.** `Task\Clock::real()` is the process clock and `now()` the instance's, and which one a page reads is transcribed and not chosen: the charts, the tooling panel, the hours gate and a token's expiry badge read `p.Now`, while `view.Ago` is `time.Since`, `model.Run.Elapsed` is `time.Since`, and `frontend.Page` groups the sidebar's recents on `time.Now()` under a page carrying `p.Now`. Reading the pinned clock for all of them put every recent of the demo instance under one heading three days wide, and the sidebar is on all 263 pages: the sweep differed on every one of them.

A binding answers the report and not the view over it. `tooling()` answers what `atkins --stats --json --all` recorded, scope by scope, and the whole of `frontend/view/tooling.go` is `Task\Tooling`: the windows, the palette, the legends, the heights. The prober is one instance both front ends hold, so the probe is paid for once however a reader came to the page and the two show the same figures. The same shape of split is what every page takes — a host fact crosses, a computation over it does not.

`inventory($hosts)` answers beszel's systems with the ssh logins and the hub's records in one array, and `host($name, $hosts)` the facts of one. The `$hosts` argument is the reader's `hosts` permission, which PHP decides and Go obeys: with it the prober reads the host now, without it the binding answers from what it kept. That is the shape of every gated binding — the page decides, the host does the work.

**A binding answers what the page asks for and not the nearest thing that contains it.** A job picker offers one job per host task-ui has an ssh login to, which is beszel's systems in capacity order with the ones the prober could not reach dropped: `reachable()`, two calls. The inventory contains that list, and reusing it would make `/schedules` and a workflow's edit page read the hub's containers, its month of history and — for a reader with the permission — every host over ssh, none of which the Go page they are compared against reads. A page that costs more than the one it replaces is not equivalent to it, whatever its HTML says.

`Task\HostJobs` builds the jobs off that list, with the IDs and the descriptions of `folder/host.go`, and `components/host-jobs.tpl` draws the `Hosts` group of a `<select>` and nothing around it. Three pages include it, which is why it is a component: `/schedules` and both pages of a workflow that pick a job. A page that offers host jobs offers `internal` as a folder too, because the root declares them and nothing else does, and host jobs alone are enough to draw the form.

## The theme

`theme/templates/index.tpl` is the shell. It places each component and assigns its own slice of the chrome as `data` immediately before including it:

```smarty
{eval $data = $chrome.sidebar}
{include components/sidebar-menu.tpl}
```

Every component has a sibling `.md` stating its input scope and the schema of that scope, and nothing else. A component that draws a row inside another names its scope apart, `data_row`, `data_menu`, `data_user`, so neither overwrites the other.

A module's `templates/index.tpl` is the body. `Task\View` renders it with the module's templates and the theme's as the search paths, in that order, so `{include components/table.tpl}` resolves into the theme, then renders the shell around it.

**A component cannot include itself.** `{include}` is a compile-time pre-pass: `Compiler.php:62` runs `while (preg_match_all("/\{include (.*?)\}/s", …))` and `str_replace`s each tag with the file's text, so a template that names itself reintroduces its own tag on every turn of the loop and the compile never finishes. There is no depth cap and no error. `{load}` is no way round it either — `Template.php:65` pushes `$this->filename` onto the stack and leaves `$this->vars` alone, so a nested load renders against the caller's scope and overwrites it. A recursive Go partial is therefore unrolled into one component per level, each fed a scope the PHP flattened for it; the class that builds the scope may recurse as deep as it likes, because that is PHP and not the compiler.

No `set_compile_location` is called. MiniTPL's default writes beside each search path, which is `modules/<name>/templates_c/` and `theme/templates_c/`.

**A module repeats no markup the theme has.** A piece of markup two views share is a component with a `.md`, and the component set starts from the `{{ define }}` list of `frontend/view/templates/_layout.tpl` and the shared parts table of [the sitemap](sitemap.md).

**A component draws the branch a page reaches, not every branch the Go partial holds.** A Go define is written for the callers it once had, and a caller that went away leaves its branch behind: `_control.tpl`'s `controltable` draws a head row, a cell per repeated job name, an `other` column and `controlmore` for a Control with columns, and the only thing that builds one is `view.NewControl`, which `control_test.go` calls and no page does. Every live table comes from `NewJobControl`, whose rows are all `Named` and whose `Columns` is empty, so the port is the named row and the job button, and `components/control-table.tpl` has no head. The test that keeps the dead branch alive is not a caller; check `grep` for the constructor before transcribing a branch, because a branch no page reaches cannot be compared, and porting it writes code the gate can never exercise.

**A page gates on a permission where Go gates, and nowhere else.** Go draws the post form of the notes board and the Reply field of every thread with no check at all: `note/edit` is the API's, which refuses the write, and the Edit item's, which `page.go:482` builds from it. The port had both forms behind `note/edit`, which is a better interface and the wrong port — the markup differs for any reader without the rule, and an autologin admin makes the difference invisible to `compare.sh`. So read the Go template for the check rather than reasoning about who ought to see a form, and run the round's pages a second time as a reader whose role is not admin, which is what `TASK_UI_AUTOLOGIN=miguel` against the demo data does. A gate Go does not have is a divergence the gate cannot see, so it survives to the cutover and changes what the instance does.

**A branch no demo row reaches is compared by nobody, so write the row.** The demo data held runs of `@cleanup` and `@claude-prompt` and nothing else, both built-in jobs of the root, whose page draws no Jobs table: so `jobSummaries` answered nothing on either side and no run page was ever about a job of the sidebar. One run of `tools/lint` and two summary threads about it, written into `.live/data`, drew both branches for the first time and both diverged at once — `frontend/job_table.go` ranged a map to pick one of two threads about one run where PHP walks the list newest first and lets the oldest stand, and the run page marked the sidebar's project and job in Go and nothing in PHP. A page that compares clean on the data at hand is compared on the data at hand, so before a round closes, read what each page draws only when a row exists, and make the row.

**The sidebar's marks are the page's folder and job.** `_layout.tpl` reads `$.Folder` and `$.Job` in the folder tree and the job rows and nowhere else, and `Chrome::at($folder, $job)` is the whole of the port's side of it. Every page that is about a place calls it: a folder's, a directory's, a job's — and a run's, which is the one that reads like a list page and is not. The run page reads both off the catalogue and not off the row, as `frontend/run.go:98` does, so a run of a folder that is gone marks nothing and a run of a job that is gone marks the folder alone.

**A `time.Duration` is nanoseconds and the ported row is milliseconds, so Go's integer division truncates where ours cannot.** The activity sparkline summed durations and divided by the count on both sides; Go's remainder is sub-nanosecond and invisible, `intdiv()` over milliseconds threw away up to a whole one, and `Activity::level()` scaled it into the SVG path — 19.52 against 19.60 on one point of three charts, with the min and max lines byte-identical because neither divides. The hover title agreed throughout, since `Run::duration()` rounds to a tenth of a second. So wherever Go divides a `Duration`, the port divides as a float: the unit cancels in a ratio and the truncation does not. This is the one place the held `intdiv` sweep must not reach.

A tag adds no whitespace of its own. The compiler closes every tag as `?>` and a newline, and PHP eats the newline after `?>`, so `A{data.x}B{if $data.x}C{/if}D` renders `AXBCD`: what a template writes between its tags is what the page gets, byte for byte. A component transcribed from a Go partial keeps that partial's spacing, and `scripts/compare.sh` only forgives the whitespace between two tags, never the whitespace inside an attribute.

## Running it

**A template edit takes effect on the next request; a PHP edit needs a restart.** `TASK_UI_PHP_FRONT` serves the checkout, but the runtime's include cache keys on the path alone and never stats the file (`runner/include_cache.go`), so a parsed `.php` lives as long as the process. MiniTPL compiles a template against its mtime, which is why the two differ. This is worth knowing because of how it fails: the fix is on disk, the page still draws the old answer, and the diagnosis reads as wrong when it was right. Restart before concluding anything about a `code/` or `mod_*` edit, and restart before a sweep — a comparison run against a long-lived process compares pages the process parsed hours ago.

```sh
# the instance, with the tree served from the checkout so a template edit
# takes effect on the next request
TASK_UI_WORKSPACE=$PWD/testdata/demo TASK_UI_DATA=.live/data \
  TASK_UI_AUTOLOGIN=admin TASK_UI_PHP_FRONT=$PWD/phpfront/app go run .

# the pipeline
cd phpfront/app && atkins

# one page, both ways
scripts/compare.sh /notes

# every page of the instance, both ways: the crawl is the page list
scripts/compare-site.sh

# the same pages as a reader who is not an admin, which is what finds a
# permission gate one front end has and the other does not
TASK_UI_AUTOLOGIN=miguel …

# both of those over the demo data, with the instances built for it
TASK_UI_SHOT_FILTER='^$' scripts/demo.sh

# a branch the demo data does not reach: start the job the page needs, write
# the rows the branch reads into .live/data, and compare that page again
curl -X POST localhost:3000/api/runs -H 'Origin: http://localhost:3000' \
  -d '{"folder":"tools","job":"lint"}'
sqlite3 .live/data/task-ui.db "INSERT INTO note …"
```

**The entrypoint of a request is re-read; everything it includes is not.** `mod_*.php` is parsed on every request, because `runner.LoadFile` goes past the include cache when `Precompile` is off, which is how `phpfront/front.go` builds the runtime. Every `include` under it — `ext/autoload.php`, each `code/Task/*.php` the autoloader pulls and each compiled template MiniTPL includes — goes through `runner.resolveInclude`, which consults the cache unconditionally, keyed by cleaned path with no mtime; `SessionPool.Put` resets the session between requests and never clears it. One cache serves the whole source tree for the life of the process.

So an edit to `mod_*.php` shows on the next request, and an edit to a class under `code/Task`, to a template, or to `composer`'s autoload map needs the instance restarted. A method added to `Task\Buttons` is `call to undefined method` until then — on the page, not in the log of the editor, which is why a 500 after an edit under `code/` is this and not the edit. A template edit also needs `php tools/precompile.php` first, and the restart after it. A new `@route` needs a restart too: the tree is scanned once, when the front end is built.

**A page's trace is the interpreter's, filtered by the recorder.** `phpfront/front.go` runs the runtime at `trace_level: info`: the interpreter reports a span per PHP frame and marks each as detail, and reports the files a page included as one `includes` span at the end of the request rather than one span each. What a reader then sees is the Shortest debug span setting of `/settings/telemetry`, a millisecond on this instance: a frame under it is dropped and counted, and its slot in Spans per trace goes back to the work that took the time. `trace_level: debug` reports every frame unmarked and a span per include with them, which is a page being taken apart rather than a page being served; `off` reports none of the interpreter's own work, and the spans a binding or a `start_span` recorded still arrive.

## The pipeline

`phpfront/app/atkins.yml`:

| Job          | What it does                                                      |
|--------------|-------------------------------------------------------------------|
| `fmt`        | fails if `phpscript fmt -l` names a file of the four source trees |
| `lint`       | `phpscript lint --include stubs/autoload.php`, zero warnings      |
| `test`       | `phpscript test ./tests/...`                                      |
| `models`     | rewrites `code/Task/Model` from `schema/schema.yml`, then formats |
| `precompile` | compiles every template into the `templates_c` beside it          |
| `docs`       | each module's `README.md`, `schema.puml` and `schema.svg`         |
| `rules`      | the semgrep conventions                                           |

`atkins` runs them in that order. The Go side keeps its own: `atkins test` and `splint ./...`.

Run `atkins fmt`, never the bare command. `phpscript fmt -l .` prints nothing on an unformatted tree, because `.` is a directory and not a package pattern; the gate is `phpscript fmt -l ./code/... ./modules/... ./stubs/... ./tools/...`, which is the four trees and not `./...`, because `./...` reaches `ext/`, which is composer's to format, and every `templates_c`, which the next precompile would write back. `-l` exits zero whether or not it named a file, so the job's test is whether it printed anything.

`rules` names its files rather than the tree. Given a directory, semgrep scans what git tracks — it says so in its own summary — so the module directory written an hour ago is held by no rule until it is committed, which is the hour the rules are for. `git ls-files -co --exclude-standard` is the list: tracked, untracked, and nothing git ignores. The Go gate at the repository root keeps the directory, because a named file also escapes semgrep's own ignore list, which holds the Go test files.

---

[Contents](README.md) · [Modules](modules.md) · [Corrections](todo-correction.md) · [Sitemap](sitemap.md)
