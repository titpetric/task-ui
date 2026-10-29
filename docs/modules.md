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
| `action/`       | the markdown a `Task::dispatch` call puts into a Claude run's context                |
| `stubs/`        | PHP declarations of the host bindings, for the linter alone                          |
| `composer.json` | the dependencies, `titpetric/minitpl` among them, with `vendor/` committed           |
| `atkins.yml`    | the module's own pipeline: format, lint, precompile                                  |

A module declares no routes of its own. `index.php` reads `$_SERVER['REQUEST_URI']` and answers what it finds there.

## Registration

`modinfo.php` is annotated `@startup`. It builds a module and registers it. A file without the annotation is refused and the reason is logged:

```php
<?php
// @startup
//
// The refinement module: a thread of comments that Claude turns into
// artifacts, and a report that indexes them.

$module = new Task\Module;
$module->setName("titpetric/refinement");
$module->setTitle("Refinement");
$module->setDescription("Refine a design through a thread of comments");
$module->setIcon("flow");

Task::register($module);
```

The name is `<vendor>/<plugin>`. It decides the link, `/m/titpetric/refinement`, and it is the unique key of the `module` table. The title and the icon are what the sidebar row shows; the icon names a symbol of the sprite in `_layout.tpl`, without the `i-` prefix.

task-ui runs `modinfo.php` once per module, in path order, before it serves a request. A module that throws does not register and does not serve; the rest start.

## The bindings

Four host symbols are registered on the runtime of every module.

| Symbol                              | What it does                                                                           |
|-------------------------------------|----------------------------------------------------------------------------------------|
| `new Task\Module`                   | an empty module, with `setName`, `setTitle`, `setDescription`, `setIcon`, `setVersion` |
| `Task::register($module)`           | records the module and its root in the `module` table, and adds its sidebar row        |
| `Task::migrate("schema/")`          | applies a directory of `*.up.sql` into task-ui's database                              |
| `Task::dispatch($comment, $action)` | starts a Claude run, returns the run's identifier                                      |

`Task::dispatch` takes the comment the user wrote and the path of a markdown file under the module's root. The prompt it builds is that file, then the comment, then the run's own environment: the artifact tree, the MCP server and what the user may do. Which module is asking is not an argument; the call reads it from the context task-ui put on the runtime before the request:

```php
$run = Task::dispatch($comment, "action/refinement.md");
```

The run is an ordinary task-ui run. It appears on `/runs`, its output replays, its artifacts land in the instance's artifact tree, and Claude's reply lands as a note.

## Serving a request

`/m/<vendor>/<name>` and everything under it reaches the module. The request is not rewritten:

- a path that names a file under `public/` is served as bytes, with no layout around it, as `public/assets/refinement.js` is;
- anything else runs `index.php`, and its output becomes the body of a task-ui page.

`$_SERVER['REQUEST_URI']` holds the whole path the browser requested, the `/m/<vendor>/<name>` prefix included. Nothing is stripped, so a module builds a link from what it reads.

The page footer prints how long the body took to render.

A module loads JavaScript from its own `public/` tree and calls what the embedding page already defines: `api(method, url, body)`, `toast(text, tone)`, and the `data-api` form contract that posts a form as JSON. The conventions are in [Development](development.md).

## Templates

`index.php` renders through MiniTPL. The sources live in `templates/` and the compiled output in `templates_c/`, both inside the module, and the module resolves the path itself:

```php
require "vendor/autoload.php";

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

A module's tables are task-ui's tables, in task-ui's database, applied by `mig` through `Task::migrate`. The rules are the instance's:

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

`./...` is what makes a path recursive; a bare `.` reads the PHP in the top directory alone. Both skip a directory named `vendor`, so composer's tree is never linted or rewritten. Run `fmt` before `lint`.

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
- `session_start` and `$_SESSION`, PDO, `mysqli_*`, `curl_*`, `proc_*`, `uniqid`, `strftime`, `mktime`, the spaceship operator.

Four behaviours differ from PHP without saying so, and each one has cost an hour somewhere:

- **Arrays are handles.** `$a = $b` on an array shares storage.
- **Declared types are never checked.** A type hint is a comment.
- **There is no exception hierarchy.** `catch` matches on the class name: `Exception` takes any name not ending in `Error`, `Error` takes any name that does, and `$e instanceof Throwable` is false.
- **Dates are Go times with Go layouts.** The format is `2006-01-02`, not `Y-m-d`, and `$t->add(86400)` adds 86.4 microseconds.

A path parameter does not arrive in `$_PATH`; that superglobal does not exist. A module reads `$_SERVER['REQUEST_URI']` instead.

---

[Contents](README.md) · [Development](development.md) · [Corrections](todo-correction.md)
