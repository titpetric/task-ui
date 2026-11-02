# The front end

task-ui's pages are being replaced by a [phpscript](https://github.com/titpetric/phpscript) application: a theme of components, a module per view, and the task-ui data model as PHP classes over the instance's database. The tree is `phpfront/app`, the Go that mounts it is `phpfront/`.

Both front ends answer while the port runs. The Go pages hold the root and the new ones answer under `/development`, which is `app.FrontBase` and the one place the prefix is spelled. A view that is proven equivalent leaves a `TODO: delete, replaced with phpscript, <hash>` on the Go handler, its template and the view helpers only it used; the cutover drops the prefix and every marked file in one commit.

[Corrections](todo-correction.md), sections 9 to 18, records every place the runtime lacks what this needs, with the workaround taken.

## The tree

| Path                          | What it is                                                              |
|-------------------------------|-------------------------------------------------------------------------|
| `composer.json`               | `vendor-dir: ext`, psr-4 over `code/Task`, a classmap over `code/` and `modules/` |
| `ext/`                        | what `composer install` wrote, the autoloader among it                   |
| `code/Module.php`             | the interface every `mod_` class declares                                |
| `code/Task/`                  | the task-ui domain: a DAO per table, and the helpers every page shares   |
| `code/Task/Model/`            | generated from `schema/schema.yml`; do not edit                          |
| `theme/templates/index.tpl`   | the page shell, which places the components                              |
| `theme/templates/components/` | a component and its `.md` per piece of shared markup                     |
| `theme/assets/`               | the stylesheet, the script, the fonts and the icon, served at `/assets/theme/` |
| `modules/<name>/`             | one directory per sidebar entry                                          |
| `stubs/`                      | the host calls, for the linter and the fixtures; nothing requires them at runtime |
| `tools/`                      | the generators and the precompiler                                       |
| `tests/`                      | the `.phpt` fixtures, and `bootstrap.php`                                |

A module directory holds:

| Path                 | What it is                                                     |
|----------------------|-----------------------------------------------------------------|
| `mod_<name>.php`     | one per view: its `@route` lines, its class, its dispatch       |
| `storage/`           | the SQL of a data model the module owns                          |
| `templates/index.tpl`| the module's body; more templates beside it                      |
| `templates_c/`       | the same compiled, written by `tools/precompile.php` and shipped |
| `assets/`            | served at `/assets/<name>/`                                      |
| `api/`               | new JSON endpoints, `@route` annotated                           |
| `docs/`              | the data model and the operations over it                        |
| `README.md`, `schema.puml`, `schema.svg` | generated                                    |

## A module

`mod_<name>.php` carries its own routes, declares its class and runs its dispatch. The pooled runtime drops what an endpoint declared between requests, so declaring the class in the endpoint file is safe, and the dispatch is guarded so a fixture can require the file for the class alone.

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

		if (!$chrome->allowed("note/view")) {
			Task\Route::refused();
			$chrome->heading("Not allowed");

			(new Task\View($this, $chrome))->page(array(), "forbidden.tpl");

			return;
		}

		$chrome->heading("Notes", array(array("text" => "Notes", "href" => "")));

		(new Task\View($this, $chrome))->page(array("threads" => $threads));
	}
}

if (PHP_SAPI === "http") {
	(new mod_notes)->serve();
}
```

`Module` names five methods and an interface here inherits nothing, so each one is written out. `tables()` is what the generated docs read.

A path is written without the prefix. Nothing in the tree spells `/development`; `rules/` refuses the literal.

A file holds no `use` block. `phpscript fmt` inlines an alias and leaves the import behind, so a file writes the qualified name the formatter would leave: `Task\Chrome` from the global namespace of a `mod_` file, `\Task\Chrome` from inside a namespace.

### Routes

`// @route GET /path` and `// @route POST /path`. A line naming no method routes both. A path parameter arrives in `$_REQUEST` under the name the annotation declared, so `/runs/{id}` reads `Task\Route::param("id")`.

The matched pattern reaches no superglobal, so a file carrying several routes matches on the method and the segments: `Task\Route::posted()` and `Task\Route::segments()`.

An `@route` never names a file. The static files are Go's, at `/assets/theme/` and `/assets/<module>/`, and `rules/` refuses a route path with an extension.

### Writes

A handler validates and the storage class writes. One statement where that does, and a transaction inside the storage class where it does not. A handler that changed something answers with a redirect, `Task\Route::go("/notes")`, so a reload does not repeat the work.

A write publishes. `Task\Host::publish($topic, $payload)` sends the event every live label on every open page reads; a write that does not publish leaves the interface stale until a reload.

## The domain

`code/Task` holds a class per table and the helpers every page shares. A page reads rows itself: no binding hands it a workspace, a run or a user.

| Class                 | What it is                                                        |
|-----------------------|---------------------------------------------------------------------|
| `Task\DB`             | `reader()`, which refuses every write, and `writer()`               |
| `Task\Row`            | the readers over a database row, and `values()` over a model        |
| `Task\Model\*`        | a row: its columns, its getters and setters, `TABLE`, `$fields`, `$primary`, `$stamps` |
| `Task\Route`          | the request past the prefix, and the links back                     |
| `Task\Chrome`         | who is reading, what they may do, the sidebar, the crumbs           |
| `Task\View`           | a module's body inside the theme's shell                            |
| `Task\Menu`           | the `[...]` menu of a row or a block                                |
| `Task\Who`            | a user as every page draws them: avatar, name, menu                 |
| `Task\Access`         | what one user may do                                                |
| `Task\Policy`         | a subject's entries, an effect per path                             |
| `Task\Permission`     | `config/permissions.yml`, over `Task\Host::rules()`                 |
| `Task\Clock`          | the two shapes a timestamp column takes                             |
| `Task\Id`             | the 26 character identifier a row carries                           |
| `Task\Text`           | the JSON a column holds where it holds a list or a map              |

A class owns a table and the behaviour of its rows both, because the runtime merges no methods into a generated class: `Task\User` holds the statements and `Task\User::admin($row)` answers what a row means.

**Every SQL statement lives in `code/Task/*.php` or `modules/*/storage/*.php`.** `rules/` refuses one anywhere else. A read page takes `Task\DB::reader()`, which refuses a statement that is not a read, so a view meant to list something cannot write.

A nullable timestamp column holds null and not `""`. The generated model names its timestamp columns and `Task\Row::values()` writes null for an empty one; writing `""` makes the Go scanner fail and the Go page lists nothing, silently.

## The host

`Task\Host::x()` is Go. Every other `Task\` name is PHP. That spelling is the whole of the boundary.

| Call                          | Why it is not PHP                                                |
|-------------------------------|--------------------------------------------------------------------|
| `base()`                      | where the router is mounted                                        |
| `version()`                   | the instance's version                                             |
| `assets()`, `assetVersion()`  | where the static files answer, and their content hash              |
| `whoami()`                    | the name the Guard resolved; the row behind it is `Task\User`      |
| `folders()`                   | `folder.Catalogue` parses Taskfiles to find jobs                   |
| `running()`                   | the runner's memory, where a run lives until it ends               |
| `dirty($dir)`                 | git's answer, not a row                                            |
| `rules()`                     | `config/permissions.yml`, embedded in the binary                   |
| `publish($topic, $payload)`   | the stream every live label reads                                  |
| `markdown($src, $folder)`     | the renderer that owns the content security policy                 |
| `message($src, $prefix, $link)` | the same, with a line break kept and `@name` marked              |
| `artifacts($folder)`, `artifact($path)` | the artifact tree, outside the directory the PHP may open |

A Go slice crosses as a PHP array and a Go struct crosses as a host object whose fields no key reaches, so `phpfront/convert.go` converts: everything that arrives is an array all the way down, keyed by the names `phpfront/host.go` declares as json tags. A binding that answers a shape adds a converter there and a case to `bind_test`.

`stubs/host.php` declares the same calls for the linter and the fixtures, each answering a static a fixture writes. Keep it in step with `phpfront/bind.go`.

## The theme

`theme/templates/index.tpl` is the shell. It places each component and hands it its own slice of the chrome as `data` immediately before including it:

```smarty
{eval $data = $chrome.sidebar}
{include components/sidebar-menu.tpl}
```

Every component has a sibling `.md` stating one thing: its input scope, and the schema of that scope. A component that draws a row inside another names its scope apart — `data_row`, `data_menu`, `data_user` — so neither overwrites the other.

A module's `templates/index.tpl` is the body. `Task\View` renders it with the module's templates and the theme's as the search paths, in that order, so `{include components/table.tpl}` resolves into the theme, then renders the shell around it.

No `set_compile_location` is called. MiniTPL's default writes beside each search path, which is `modules/<name>/templates_c/` and `theme/templates_c/`.

**A module repeats no markup the theme has.** A piece of markup two views share is a component with a `.md`, and the component set starts from the `{{ define }}` list of `frontend/view/templates/_layout.tpl` and the shared parts table of [the sitemap](sitemap.md).

MiniTPL emits a newline after every compiled tag, so an `{if}` directly before punctuation renders a space in front of it. The whole sentence goes inside each branch.

## Running it

```sh
# the instance, with the tree served from the checkout so a template edit
# takes effect on the next request
TASK_UI_WORKSPACE=$PWD/testdata/demo TASK_UI_DATA=.live/data \
  TASK_UI_AUTOLOGIN=admin TASK_UI_PHP_FRONT=$PWD/phpfront/app go run .

# the pipeline
cd phpfront/app && atkins

# one page, both ways
scripts/compare.sh /notes
```

A new `@route` needs a restart: the tree is scanned once, when the front end is built. A template does not.

## The pipeline

`phpfront/app/atkins.yml`:

| Job          | What it does                                                   |
|--------------|------------------------------------------------------------------|
| `fmt`        | fails if `phpscript fmt` would rewrite a file                    |
| `lint`       | `phpscript lint --include stubs/autoload.php`, zero warnings     |
| `test`       | `phpscript test ./tests/...`                                     |
| `models`     | rewrites `code/Task/Model` from `schema/schema.yml`, then formats |
| `precompile` | compiles every template into the `templates_c` beside it         |
| `docs`       | each module's `README.md`, `schema.puml` and `schema.svg`        |
| `rules`      | the semgrep conventions                                          |

`atkins` runs them in that order. The Go side keeps its own: `atkins test` and `splint ./...`.

---

[Contents](README.md) · [Modules](modules.md) · [Corrections](todo-correction.md) · [Sitemap](sitemap.md)
