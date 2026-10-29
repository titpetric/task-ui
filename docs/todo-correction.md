# Corrections

The module system has a specification written against phpscript and MiniTPL sight unseen, and six of its requirements name something neither one has. Each requirement is below with the snippet the specification calls for, what the runtime does instead, and the correction the code takes.

Checked against `github.com/titpetric/phpscript` at commit `ee889fe`, which `@main` resolves to as `v0.6.1-0.20261027175813-ee889febddf9`, and `titpetric/minitpl` v1.3.3.

## 1. `new Task::Module()` does not parse

Specified:

```php
$module = new Task::Module();
$module->setName("titpetric/mysql");
Task::register($module);
```

`::` after a name resolves a class constant, a static property or a static method. In a `new` expression the class name ends at the `::`, so `new Task::Module()` is a parse error. `Task::Module` can only name a constant of a class called `Task`, and `helperClassConst` reads those from `model.Class.Consts` alone.

A host class registers through `rt.RegisterClass(*model.Class)`, which takes PHP declarations, so a Go method cannot be attached to it. Go methods reach PHP one way: `rt.RegisterConstructor(name, ctor)` publishes every exported method of the value the constructor returns. Constructor names may hold a namespace.

Correction: the constructor is registered as `Task\Module`, and a module writes a namespaced `new`:

```php
$module = new Task\Module;
$module->setName("titpetric/mysql");
Task::register($module);
```

`Task::register`, `Task::migrate` and `Task::dispatch` keep the `::` spelling: a host function whose name holds `::` is the first thing `helperStaticCall` resolves, so a static call on a class that does not exist reaches it. The two spellings coexist. Only the constructor moved.

## 2. MiniTPL has no precompiler and no `templates_c`

Specified: a `templates/` and a `templates_c/` in every module, the second precompiled, so every template that runs is compiled ahead of time.

MiniTPL compiles lazily, inside `load()`, when the source is newer than the compiled file. Its compile directory defaults to `cache/` under each template path, and `set_compile_location($path, $absolute)` moves it. The package is pure PHP with three classes and a composer autoload, it ships no command line tool, and `templates_c/` is a Smarty name.

Correction: `precompile.php` in the module walks `templates/` and calls the compiler per file, which does exist:

```php
$tpl = new MiniTPL\Template("templates/");
$tpl->set_compile_location("templates_c/", true);
foreach (glob("templates/*.tpl") as $source) {
	$tpl->compile($source, "templates_c/" . basename($source) . ".php");
}
```

`atkins precompile` in the module runs it, so `templates_c/` is a shipped artefact and the pipeline keeps it in step. `index.php` compiles into the same directory, so a request finds every template already compiled.

## 3. An embedded tree cannot be served with options

Specified: `phpscripts/modules/` embedded in the binary with `go:embed` and served from there.

`server.NewHandler(fs.FS)` serves an `fs.FS` but takes no options: the document root is fixed at `public`, the flat stack is off and `runner.Options` is the zero value. `server.NewModule` takes options but its root is a string and it calls `os.DirFS` itself, so an `embed.FS` cannot reach it. `annotations.NewRoute(fs.FS, ...)` takes both, and it registers `@route` endpoints, which a module does not declare.

Three other parts of a module need a real directory in any case. `stdlib.RegisterFS` reroots the file, image and session bindings at one; `composer`'s autoloader reads `vendor/`; and the compile directory has to be writable.

Correction: `phpscripts.Extract` writes the embedded tree to `<data>/modules` on first start, and discovery walks that directory with the working directory. The binary still holds the modules; the filesystem the runtime reads is a real one.

## 4. The handler strips no path prefix

Specified: the PHP receives the full request including the `/m` prefix in `$_SERVER['REQUEST_URI']`.

There is nothing to correct. No `path_prefix`, no `base_path` and no `http.StripPrefix` exists in the module, and `REQUEST_URI` derives from `r.URL.RequestURI()`. The requirement holds by doing nothing: the handler resolves the module from the path itself and passes the request through unrewritten. `SCRIPT_NAME`, `PHP_SELF` and `DOCUMENT_ROOT` are set from the module's root so a script still knows where it lives.

Recorded because wrapping the handler in `http.StripPrefix` would strip `/m/<vendor>/<name>` out of `REQUEST_URI` and break every link a module builds.

## 5. A module cannot write to the host's /tmp

Specified: the compile directory is something in `/tmp`.

`stdlib.RegisterFS(rt, dir)` roots the file, image, auth and session bindings at `dir`, and `root.resolve` joins every path a script writes onto it. A path written from `/` names the root itself, anything else is relative to the working directory, and no spelling reaches a host path above the root: phpscript's own comment says so, because a runtime serving an embedded tree could not honour one. `WritablePaths` narrows that set and does not widen it. So `/tmp/templates_c/` in a module means `<root>/tmp/templates_c/`.

Correction: the requirement is withdrawn, and the compile directory is the module's own `templates_c`:

```php
$tpl = new MiniTPL\Template("templates/");
$tpl->set_compile_location("templates_c/", true);
```

The module resolves the path, so Go holds no setting for it and the Manager prepares nothing. MiniTPL's compiler creates the directory with a recursive mkdir of its own when a template is newer than what is there, and the module ships the directory compiled, so an ordinary request reads a file that is already there.

A compiled template lands at `templates_c/templates/<name>.tpl`: the absolute form of `set_compile_location` prefixes the compile location to the template path. `precompile.php` writes that path, and `_find_path` reads it.

## 6. Two option fields are parsed and ignored

Specified: nothing directly. A module that loops forever would hang a page.

`runner.Options.TimeLimit` and `runner.Options.ConcurrencyLimit` are read from the configuration and held on the struct. Neither is enforced.

Correction: the handler bounds each render with `context.WithTimeout` and the pool bounds how many runtimes exist, so Go keeps the two limits. The two fields stay at their defaults.

## Not a runtime gap

`~/.claude/skills/dev/skills/phpscript/SKILL.md` states that path parameters arrive in `$_PATH["param"]`. No such superglobal exists. phpscript's own `docs/README.md` says so: "nothing here spells `$_PATH`, a name PHP does not have." Path parameters are merged into `$_REQUEST`, written after the query, the form and the cookies. The skill needs the line corrected.

---

[Contents](README.md) · [Modules](modules.md) · [Development](development.md)
