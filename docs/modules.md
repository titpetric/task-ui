# Modules

A module is a plugin written in PHP. task-ui embeds [phpscript](https://github.com/titpetric/phpscript), a PHP runtime written in Go, finds the modules under a working directory at startup, records each one in the database, and serves it at `/m/<vendor>/<name>`. A module renders HTML into the task-ui layout: the sidebar, the header and the footer are the instance's, the body is the module's. There is no iframe.

phpscript is not PHP. It runs a subset and it diverges, so read [what the runtime refuses](#what-the-runtime-refuses) before writing a line. [docs/todo-correction.md](todo-correction.md) records every place the runtime lacks what this design needs, with the correction taken.

## The directory

A module is a directory with a `modinfo.php` in it. The directory is the module's root, and the root is the only path a module may read.

| Path            | What it is                                                                           |
|-----------------|--------------------------------------------------------------------------------------|
| `modinfo.php`   | the registration, run once at startup, annotated `@startup`                          |
| `index.php`     | the catch-all: every request the `public/` tree does not answer arrives here         |
| `migrate.php`   | applies the module's `schema/` through `Task::migrate`                               |
| `schema/`       | the module's own `*.up.sql` migrations, applied into task-ui's database              |
| `templates/`    | the MiniTPL sources                                                                  |
| `templates_c/`  | the same templates compiled, written by `precompile.php` and shipped with the module |
| `public/`       | files served as bytes under `/m/<vendor>/<name>/`, the `assets/` tree among them     |
| `code/`         | the module's classes, autoloaded: its store, its manager, its pages                  |
| `ext/`          | what `composer install` wrote, the autoloader among it                               |
| `tests/`        | the `.phpt` fixtures `phpscript test` runs, and their bootstrap                      |
| `action/`       | the markdown a `Task::dispatch` call puts into a Claude run's context                |
| `stubs/`        | PHP declarations of the host bindings, for the linter alone                          |
| `composer.json` | the dependencies, and the classmap over `code/`                                      |
| `atkins.yml`    | the module's own pipeline: format, lint, test, precompile                            |

A module declares no routes of its own. `index.php` reads `$_SERVER['REQUEST_URI']` and answers what it finds there.

**A module writes classes, not functions, and autoloads them.** `composer.json` carries a classmap over `code/`, `composer install` writes `ext/`, and `index.php` requires `ext/autoload.php` and nothing else: naming a class is what loads it. Free functions in a shared include are what this stops — a module is one page's worth of PHP and a handful of types, and the types are easier to read and to test than a namespace full of `module_thing_verb()`.

The refinement module is four: `RefinementStore` holds every statement, `RefinementManager` decides what to ask the store for, `RefinementPrompt` composes what a run is sent, and `RefinementPage` routes and draws. The shape is phpscript's own and Go's: a manager over a store, made once per request.

A path in a `require` is read from the module's root, the directory `modinfo.php` sits in, whatever file the `require` is written in: the runtime roots the module there and nothing above it resolves.

**A module installs its PHP dependencies with composer, into `ext/` and never `vendor/`.** The installed tree ships with the module, so a running instance installs nothing and reaches no network. The directory is `ext` because the binary is installed with `go install <module>@<ref>` and the zip the Go module proxy builds for that drops every file more than one level under a directory named `vendor`: a composer tree there arrives as `vendor/autoload.php` alone, with nothing left for it to load, and the instance answers the first request with `no such file or directory` while every test here passes. `composer.json` names it:

```json
"config": { "vendor-dir": "ext" }
```

`index.php` then requires `ext/autoload.php` and the autoloader finds the rest. `TestModulesInstallIntoExt` and `TestModulesRequiresResolve` in `phpscripts/` are the gate: the first refuses a `vendor` path in the embedded tree, the second resolves every `require` a module writes against that tree.

A module's forms post to the module, which is the one place in the front end that takes a `POST`, so those routes carry the same `access.SameOrigin` check `/api` has. A form that changed something answers with a redirect, `header("Location: ...")` and `http_response_code(303)`, which task-ui sends on rather than wrapping in a layout: without it a reload repeats the work.

## Registration

`modinfo.php` is annotated `@startup`. It builds a module and registers it. A file without the annotation is refused and the reason is logged:

```php
<?php
// @startup
//
// The refinement module: a thread of comments that Claude turns into
// artifacts, and a report on what they are.

$module = new Task\Module;
$module->setName("titpetric/refinement");
$module->setTitle("Refinement");
$module->setDescription("Refine a design through a thread of comments");
$module->setIcon("flow");

Task::register($module);
```

The name is `<vendor>/<plugin>`. It decides the link, `/m/titpetric/refinement`, and it is the unique key of the `module` table. The title and the icon are what the sidebar row shows; the icon names a symbol of the sprite in `_layout.tpl`, without the `i-` prefix.

task-ui runs `modinfo.php` once per module, in path order, before it serves a request. A module that throws does not register and does not serve; the rest start.

Then it runs every other `@startup` file of the module, in path order, on the same runtime. A module's migrations go in one of those rather than in a `require` from `modinfo.php`, so a startup step is a file with an annotation and the list of them is the directory:

```php
<?php
// @startup

Task::migrate("schema/");
```

A startup file that throws keeps the module from registering, which is what should happen when its tables are the thing that failed.

## The bindings

These reach the PHP of every module.

| Symbol                              | What it does                                                                           |
|-------------------------------------|----------------------------------------------------------------------------------------|
| `new Task\Module`                   | an empty module, with `setName`, `setTitle`, `setDescription`, `setIcon`, `setVersion` |
| `Task::register($module)`           | records the module and its root in the `module` table, and adds its sidebar row        |
| `Task::migrate("schema/")`          | applies a directory of `*.up.sql` into task-ui's database                              |
| `Task::dispatch($comment, $action, $folder)` | starts a Claude run in that folder of the tree, returns the identifier of the note that holds it |
| `Task::artifacts($folder)`          | the files under one folder of the artifact tree, by path relative to it                |
| `Task::artifact($path)`             | the text of one artifact, from the top of the tree, or `""`                            |
| `Task::markdown($src, $folder)`     | markdown as HTML, through the renderer that owns the content policy                    |
| `new Database("task-ui")`           | the instance's database, with `query`, `get`, `get_all` and `insert`                   |

`Task::VERSION` is the instance's version and `Task::PREFIX` is `/m`.

A trailing argument is optional: a Go binding spells an optional parameter as an extra one, and a call that leaves it out is zero-padded. `Task::artifacts()` is the whole tree and `Task::dispatch($comment, $action)` works at the top of it.

A module reads the artifact tree through the two bindings and not with `fopen`: the filesystem boundary roots every path a script opens at the module's own directory. Both are read only, because a module shows what a run wrote and the runs write.

### A folder per thread

One artifact tree serves every run of the instance, which makes a module's work and everyone else's one pile. A module that keeps several pieces of work apart gives each one a folder of the tree and names it in the three calls that take one: the run writes there, the module lists there, and the links the page draws carry it back.

The refinement module names each folder after the thread, so a new thread starts on an empty directory and a correction reads what that thread's own runs left. The paths never carry the folder: a run writes `schema/user.up.sql` inside it, `Task::artifacts($thread)` answers `schema/user.up.sql`, and `Task::markdown($report, $thread)` is what resolves the report's own `[schema/user.up.sql](schema/user.up.sql)` to the thread's copy.

`Task::artifact` is the exception and takes a whole path, because reading one file is not a listing: a module inside a folder spells it `Task::artifact($thread . "/README.md")`.

`Task::markdown` resolves a relative link into the artifact tree, under the folder named, which is where a run put the files its report names. It is the instance's own renderer: raw HTML shows as text and a link keeps to a scheme it knows, so markdown a reader typed is as safe to draw as markdown a run wrote. What the folder argument decides is only what a relative link means.

`Task::dispatch` takes the comment the user wrote, the path of a markdown file under the module's root, and the folder the run works in. The prompt it builds is that file, then where to work, then the comment, then what the folder holds now. Which module is asking is not an argument; the call reads it from the context task-ui put on the runtime before the request:

```php
$run = Task::dispatch($comment, "action/refinement.md", $thread);
```

The run is an ordinary task-ui run. It appears on `/runs`, its output replays, its artifacts land in the instance's artifact tree, and Claude's reply lands as a note.

## Serving a request

`/m/<vendor>/<name>` and everything under it reaches the module. The request is not rewritten:

- a path that names a file under `public/` is served as bytes, with no layout around it, as `public/assets/refinement.js` is;
- anything else runs `index.php`, and its output becomes the body of a task-ui page.

`$_SERVER['REQUEST_URI']` holds the whole path the browser requested, the `/m/<vendor>/<name>` prefix included. Nothing is stripped, so a module builds a link from what it reads.

The page footer prints how long the body took to render.

A module loads JavaScript from its own `public/` tree. What the instance's own script offers it is markup rather than functions: [the front end's shared markup](framework.md) is the contract, and a module writes the attribute and ships no handler.

The one a module that dispatches work wants is the run status label. `Task::run($note)` answers the run a dispatch turned into, and

```html
<span class="state" data-run-id="01K8Z3..."></span>
```

is filled and kept current by the instance, over the event stream every other page of it already shares.

## Templates

`index.php` renders through MiniTPL. The sources live in `templates/` and the compiled output in `templates_c/`, both inside the module, and the module resolves the path itself:

```php
require "ext/autoload.php";

$tpl = new MiniTPL\Template("templates/");
$tpl->set_compile_location("templates_c/", true);
$tpl->load("thread.tpl");
$tpl->assign("thread", $thread);
echo $tpl->get();
```

A module ships `templates_c/` compiled, so no request compiles a template. MiniTPL has no command to do that; `precompile.php` in the module calls `$tpl->compile($source, $destination)` per file and `atkins precompile` runs it. A template newer than its compiled file is compiled on the next load, and MiniTPL's compiler creates the directory it writes into.

The absolute form of `set_compile_location` prefixes the compile location to the template path, so `templates/thread.tpl` compiles to `templates_c/templates/thread.tpl`. `precompile.php` writes that path, and `_find_path` reads it.

A module cannot reach the host's `/tmp`: the filesystem boundary roots every path a script writes at the module's own directory, and `/` to a script is that directory. [Corrections](todo-correction.md) records it.

## The database

A module's tables are task-ui's tables, in task-ui's database, applied by `mig` through `Task::migrate` and read through `new Database("task-ui")`. There is one connection and every module shares it: the pool is sqlite on one connection, so a module's work belongs in a statement and not in a long transaction.

Each module migrates under a project of its own, so two cannot claim one applied row. The name is the plugin cut to seven characters and eight hex of the digest of the full name, because the column mig records it in holds sixteen: `titpetric/refinement` is twenty.

### The store, and the manager over it

Every statement a module runs lives in one class, `<Module>Store`, the
way a Go package keeps its queries in `storage/` and nothing outside it
issues one. A page that writes its own SQL is the thing this stops.

```php
class RefinementStore {
	var $db;

	function __construct($db) {
		$this->db = $db;
	}

	function getThread($id) {
		return $this->db->get("SELECT * FROM refinement_thread WHERE id = ?", $id);
	}

	function updateRefinementThread($data) {
		return $this->db->update("refinement_thread", $data, "id");
	}
}
```

- A read is `get<Thing>()` and answers rows or a value. One that needs
  more than one statement composes them in the class, not in the page.
- A write is a CRUD verb and the table it writes:
  `createRefinementThread`, `updateRefinementThread`. It takes one array
  whose keys are the table's columns, so the call site reads as the row
  it writes and a new column is a new key rather than a new argument.
- A statement that returns no rows goes through `query()`. The binding
  has no `delete`.
- `get()` answers a missing row with `false`. A caller tests
  `is_array($row)`.
- Nothing in the store renders, routes or dispatches.

`<Module>Manager` is what decides. It holds the store, it is made once
per request, and the work of the module is its methods: what a thread
does, what a correction is counted against, what a page load reconciles.
Its static side is the part that needs no host binding — parsing a
title, grouping rows for a page — and that is the part a `.phpt` fixture
calls, which is why it is static rather than free.

The rules are the instance's:

- one file owns one table, named for it, and the files are append-only once applied anywhere;
- a primary key is a ULID in `CHAR(26)`;
- no foreign keys;
- `created_at` and `updated_at` are native timestamps;
- an index is `idx_<table>_<columns>`, a unique one `uidx_<table>_<columns>`;
- the table name is prefixed with the module, so `refinement_thread` cannot collide with a core table.

task-ui records the modules themselves in `module`, and the sidebar in `sidebar`. A module gets a sidebar row when it registers, hidden or shown from [Settings, Sidebar](guide/settings.md#sidebar).

## Telemetry

The runtime reports to the same `oida` tracer as task-ui's own middleware, so one trace in `/debug/oida` holds the spans of the request and the spans of the PHP that ran under it. A module adds its own with `start_span`:

```php
$span = start_span("load the thread", "database");
$span->set_attribute("thread", $id);
$span->end();
```

With telemetry off the call records nothing and costs a function call. [Settings, Telemetry](guide/settings.md#telemetry) holds the switch and the rest of the tracer's options.

## The pipeline

A module ships an `atkins.yml` whose `default` runs three jobs in order: `fmt`, `lint`, `precompile`.

`./...` is what makes a path recursive; a bare `.` reads the PHP in the top directory alone. Run `fmt` before `lint`.

Two things about the jobs are not obvious, and each one has cost an hour:

- **`phpscript fmt -l` exits zero whether or not it listed a file.** A step that calls it alone reports success over an unformatted tree, so the job tests what it printed:

  ```yaml
  fmt:
    desc: "Fail if phpscript fmt would rewrite a file"
    steps:
      - |
        out="$(phpscript fmt -l . ./stubs/...)"
        if [ -n "${out}" ]; then
            echo "${out}" >&2
            exit 1
        fi
  ```

- **The linter reads PHP declarations and a host binding is Go.** `Task\Module` and the three `Task::` calls are registered on the runtime, and `MiniTPL\Template` arrives through composer's autoloader, so a plain `phpscript lint .` reports four undefined classes and a real one reads as more of the same. `stubs/` declares them in PHP for the linter, and nothing requires that directory at runtime:

  ```bash
  phpscript lint --include stubs/autoload.php .
  ```

## What the runtime refuses

phpscript decided against these, and they will not arrive. The full list is in the runtime's own `docs/design.md`; these are the ones that catch a plugin author.

- `extends`, `parent::`, traits, `abstract`, `final`, `readonly`, `new self()`, `new static()`, enums, cloning. An `interface` with `implements` is the only contract there is, and it is checked.
- Generators, `yield`, fibers, `eval`, `goto`, `global`, references outside `foreach`, `$$name`.
- `??` and `??=`, the null coalescing operators: the parser refuses the second `?`. Write `isset($a["k"]) ? $a["k"] : $fallback`. The short ternary, `$a ?: $b`, does parse.
- `strtok`. Use `explode` and take the first entry.
- `date()`, which answers with the layout it was handed rather than a
  time. `DateTime::now()->format("2006-01-02 15:04:05.000000")` is the
  one that reads the clock, and a row written through `date()` carries
  the layout as its timestamp with nothing to say it went wrong.
- A monotonic `ulid()`. Two in one millisecond share the timestamp and
  carry unordered random tails, so an identifier is no tiebreak for a
  second-resolution stamp: order rows by a timestamp that holds
  microseconds.
- `session_start` and `$_SESSION`, PDO, `mysqli_*`, `curl_*`, `proc_*`, `uniqid`, `strftime`, `mktime`, the spaceship operator. `ulid()` is there, which is what a primary key wants.

`$db->get()` answers a missing row with `false`, the way PHP does, and not with `null`. Test it with `is_array($row)`: a `=== null` reads every missing row as a hit, and the page draws an empty one.

Four behaviours differ from PHP without saying so, and each one has cost an hour somewhere:

- **Arrays are handles.** `$a = $b` on an array shares storage.
- **Declared types are never checked.** A type hint is a comment.
- **There is no exception hierarchy.** `catch` matches on the class name: `Exception` takes any name not ending in `Error`, `Error` takes any name that does, and `$e instanceof Throwable` is false.
- **Dates are Go times with Go layouts.** The format is `2006-01-02`, not `Y-m-d`, and `$t->add(86400)` adds 86.4 microseconds.

A path parameter does not arrive in `$_PATH`; that superglobal does not exist. A module reads `$_SERVER['REQUEST_URI']` instead.

---

[Contents](README.md) · [Development](development.md) · [Corrections](todo-correction.md)
