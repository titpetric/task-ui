# Corrections

Two specifications here stand against phpscript and MiniTPL sight unseen: the module system, and the [front end](phpfront.md) that replaces the Go pages. Each one names things neither runtime has. Every requirement is below with the snippet the specification calls for, what the runtime does instead, and the correction the code takes.

Sections 1 to 8 are the module system, 9 to 25 the front end.

A gap of this port's own making is not one of these and is not recorded here: a binding that answered less than the Go page it replaced is a binding to widen, which [the front end](phpfront.md) states as a rule.

Checked against `github.com/titpetric/phpscript` at commit `ee889fe`, which `@main` resolves to as `v0.6.1-0.20261027175813-ee889febddf9`, and `titpetric/minitpl` v1.3.4.

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

Two other parts of a module need a real directory in any case. `stdlib.RegisterFS` reroots the file, image and session bindings at one, and the compile directory has to be writable.

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

## 7. A composer tree cannot be installed into `vendor/`

Specified: `composer.json` requiring `titpetric/minitpl`, with the installed tree committed so a running instance installs nothing and reaches no network.

The requirement holds. The directory name does not. The committed tree is whole in a checkout and gutted in a release: the image installs the binary with `go install github.com/titpetric/task-ui@<ref>`, the Go module proxy builds that zip, and `x/mod/zip` drops what it reads as a vendored package. `isVendoredPackage` finds the `vendor/` element in a path and asks whether a slash follows it, so `vendor/autoload.php` stays and `vendor/composer/autoload_real.php` goes. Of the seventeen files under the module's `vendor/`, one survived, the one at the top, with nothing left for it to load. The module answered its first request on the instance with

```
error including /vendor/composer/autoload_real.php: no such file or directory
```

while passing every test here, because the embed is built from the checkout and the checkout is whole.

Correction: the install directory is `ext`, which composer takes from `config.vendor-dir` and Go treats as any other directory. `composer install` writes it, it ships with the module, and `index.php` requires `ext/autoload.php`. Nothing else changes: the autoloader works, and a dependency is still added by naming it in `composer.json` and running composer.

`TestModulesInstallIntoExt` refuses a `vendor` path in the embedded tree and `TestModulesRequiresResolve` resolves every `require` a module writes against that tree, so a require that would not ship fails here instead of on the instance.

## 8. `phpscript test` roots a fixture at its own directory

Specified: nothing directly. A module's PHP has to be testable without task-ui.

`phpscript test` is the runner and a `.phpt` fixture is the unit test, which is right. The root is the surprise: a fixture runs with the directory it sits in as `/`, so a fixture in `tests/` sees the fixtures beside it and no `lib/`, no `ext/` and no module at all. `require "lib/render.php"` from there is `tests/lib/render.php`, and an absolute `/lib/render.php` is the same path with the slash taken off.

Correction: a `phpscript.yml` in the tree marks a suite root, and the directory holding it is the application root for every fixture below it. The module keeps one at its own root with `test.include: tests/bootstrap.php`, so a fixture reads the tree `index.php` reads and the bootstrap loads the autoloader and the module's declarations ahead of it. [Testing](testing.md) has the shape.

Recorded because the default reads as a bug when it is a default: a fixture tree that holds its own fixtures and nothing else needs no configuration, and a fixture tree inside an application does.

## 9. `fmt` repoints a static call on a root class

Specified: the host calls keep the `Task::` spelling the module system uses, so one contract covers both trees.

`phpscript fmt` rewrites `\Task::base()` to `Task::base()` inside a namespaced file. The two name different classes: the second resolves to `Task\Task`. `\Other\Thing::make()`, `new \Database(...)` and `\strlen(...)` keep their backslash, so it is the single-segment class name in a static call alone. The file still parses, and `phpscript lint` is what reports it:

```
| WARN | Route.php | 26 | static call Task\Task::base(): unknown class |
```

Correction: the host class is registered as `Task\Host`. Every reference is multi-segment, which the formatter leaves alone, and a file inside `namespace Task` writes `Host::base()` with no qualifier at all. `phpfront.HostClass` is the one place the name is spelled.

Filed as [phpscript#187](https://github.com/titpetric/phpscript/issues/187). A formatter that changes name resolution is the one class of bug a formatter must not have, so the fix belongs upstream and `Task\Host` stands on its own merits.

## 10. A property cannot be named by a variable

Specified: one hydrator over every generated model, `fill($object, $fields, $row)`, so the row to model mapping is written once.

`$object->$name` does not parse: `line 15: expected member name`. Neither does `$object->{$name}`: `expected "(" after a braced member name`. The read direction has an answer, because `get_object_vars($object)` and `(array)$object` both work and a generated model declares its columns and nothing else, so its properties are exactly its row.

Correction: `tools/models.php` writes the assignments out, one per column, and `Task\Row::values()` is `get_object_vars`. The repetition lives in `tools/templates/model.tpl`.

Filed as [phpscript#188](https://github.com/titpetric/phpscript/issues/188). The workaround holds for a generated class; hand-written code that maps a row by name has none.

## 11. There is no `gmdate`, and `date` is not UTC

Specified: a row's timestamps written the way `storage.stamp()` writes them, UTC to the millisecond.

`gmdate` is not bound. `date` is, and it reads the process timezone: with `TZ` an hour off UTC it answers `2026-11-02 10:59:04` where the column holds `2026-11-02 09:59:04`. A timestamp column compared as text against another holds one zone, so the hour matters.

Correction: `Task\Clock` reads the Go time binding, which is exact: `DateTime::now()->in(Time\Location::utc())->format(Clock::LAYOUT)`. Reading a column back is `DateTime::parse_in_location`, because `DateTime::parse` also answers local. The layout is a Go layout, `2006-01-02 15:04:05`, and no PHP format string.

## 12. `phpscript list` reports no methods

Specified: a module's `README.md` generated from the API of its storage classes.

`list.Row` holds a route, a filename and the class names of a file. A method, its signature and its docblock are not there, and `phpscript ast` answers a token stream.

Correction: `tools/apidoc.php` reads the classes with `token_get_all()` and writes the API as JSON, which `tools/render.php` feeds to `README.md.tpl`.

Filed as [phpscript#189](https://github.com/titpetric/phpscript/issues/189), proposing `list --api`. The generator re-finds in PHP what the parser already holds.

## 13. A class receives nothing through `extends` and there are no traits

Specified: the row behaviour of `model/user.go` on the generated `Task\Model\User`, as Go puts it in a second file of one package.

`extends` names a parent that a catch clause and `instanceof` filter on, and a class receives no member through it. Traits are unavailable. PHP merges no methods from a second file into a class, so a generated class is the whole class.

Correction: one class owns a table and the behaviour of its rows. `Task\User` holds the statements and `Task\User::admin($row)`, `::sees($row, $folder)`, `::initials($row)` answer what a row means. `Task\Model\User` is data.

Recorded as a design decision of the runtime and not a gap: `docs/design.md` in phpscript argues the absence. It decides the shape of every class under `code/Task`.

## 14. The matched route is not published

Specified: one `mod_` file that holds several `@route` lines and dispatches on the one that matched.

`runner.WithRoutePattern` puts the declared pattern on the request context, and `Context.pathValues` reads it to fill `$_REQUEST` with the path parameters. The pattern itself reaches no superglobal, so a file with four routes cannot ask which of them it is answering.

Correction: `Task\Route::path()` cuts `Host::base()` off `$_SERVER["REQUEST_URI"]` and `::segments()` splits the rest. A handler matches on the method and the segments, as the refinement module's `index.php` does.

## 15. `mig gen` emits Go only

Specified: `Task\Model` generated from the same source the Go types come from.

`mig gen --lang` defaults to `go` and `go` is the only case it handles.

Correction: `tools/models.php` renders `schema/schema.yml`, the file `mig docs --yaml` writes and `mig gen` reads, through `tools/templates/model.tpl`. The Go types and the PHP models come from that one file, written by `atkins schema` and then by `atkins models` in `phpfront/app`.

Not a phpscript gap. Worth folding into mig as `--lang php` once the template has settled.

## 16. A compiled template outlives the component it includes

Specified: components in the theme, included by a module's template, each compiled ahead of time.

MiniTPL resolves `{include}` at compile time and writes the component's body into the compiled template. `load()` recompiles where the source is newer than the compiled file, and the source it stats is the template's own: a changed component leaves every template that includes it compiled against the old body.

Correction: `tools/precompile.php` calls `compile()` for every template, which compiles whatever the timestamps say, and the pipeline runs it. Nothing reaches `load()` for a first compile on a served request.

## 17. An out parameter is lost inside a short circuit

Specified: the names a note mentions, read with the shape a regex that both
tests and captures is written in.

`preg_match_all(..., $names) && isset($names[1])` reads `$names` as unset. The
write lands alone in an `if`, in a preceding assignment, and in the same `&&`
where the right operand does not read it, so it is applied after the
expression and not at the call. `php` 8.5.4 writes it at the call in all five
shapes.

It fails quietly: the branch is skipped, the function answers empty and the
column is written blank, which leaves a `LIKE` over it matching nothing.

Correction: the call gets its own statement and the result is tested
afterwards. `tests/notes.phpt` states the four shapes the column answers.

Filed as [phpscript#190](https://github.com/titpetric/phpscript/issues/190).

## 18. A timestamp column takes two shapes

Specified: a row the front end writes read back by the Go pages while both
answer.

Two things about a timestamp column, found by posting a note in one front end
and listing it in the other.

A nullable timestamp column holds null and not `""`. The Go side writes a nil
`*time.Time` as null, and `""` in the column fails the scan into one.
`storage.NoteStorage.List` discards its error, so the page answers an empty
list and says nothing: a note written with `edited_at = ""` is invisible to
`/notes` and the log is silent.

A write is `time.Time.String()`, `2026-11-02 10:32:12.414 +0000 UTC`, because
that is what the Go side leaves in the column. A read is ISO seconds with no
zone, because the driver parses the column and the bridge renders the time it
got. Both sides read either shape, so ISO writes work; they leave two formats
in one column, and the queries that compare a timestamp compare it as text.

Correction: the generated model names its timestamp columns in `static
$stamps`, and `Task\Row::values()` writes null for an empty one.
`Task\Clock::STORED` is what a write leaves and `::READ` is what a read
answers; `parse()` takes either by looking for the zone. `tests/clock.phpt` states both
shapes, and that a whole second and a fraction of it sort the way they compare
as numbers.

Recorded as a boundary and not a gap: two writers share one table and the
format is the Go driver's. It is the first thing a module's write gets right,
and the only sign of getting it wrong is a page that lists nothing.

## 19. A write inside a fixture fails silently

Specified: a fixture renders a module's page and states the markup, since the
template is the page.

A fixture runs under a sandbox that refuses a write. `file_put_contents`
returns false and leaves no file, while `is_writable` on the same directory
returns true. MiniTPL compiles inside `load()` where the source is newer than
the compiled file, the write fails silently, and `_find_path` then reports
`Template file 'run.tpl' doesn't exist! Is the compile dir writable?`, naming
the condition the fixture reports as true. A page whose template is compiled
renders in a fixture; a new template renders nowhere until `atkins precompile`
writes it.

Correction: a module holds what the template reads in a `scope()` method and
`serve()` renders it, so `tests/run-page.phpt` and `tests/report-page.phpt`
state every key of the scope and no fixture compiles anything.
`is_writable` under the sandbox needs to answer what a write does.

## 20. A modifier tokenizes as a name and the `T_STATIC` family is undefined

Specified: the API table of a module's `README.md`, which carries each method's
docblock, read with `token_get_all()` because [12](#12-phpscript-list-reports-no-methods)
leaves nothing else to read it with.

`token_get_all` answers `T_STRING` for `static`, `public`, `private`,
`protected`, `final` and `abstract`, and the constants that name them are not
declared: `T_STATIC` raises `Undefined constant`. A reader that keeps a
docblock until the token it documents cannot ask whether the token between the
two is a modifier, so a documented static method reads as undocumented and the
table prints `-`.

Correction: `tools/docs.php` holds `MODIFIERS` and matches the token's text. It
is the one place in the tree that reads a modifier, so the list is stated once.

## 21. `fmt` drops the fraction of a whole float literal

Specified: `Task\Chart`, whose share axis pins to 0/25/50/75/100 where the unit
it draws is a percentage and reads against its own hundred, and scales to the
data otherwise.

`phpscript fmt` rewrites `100.0` to `100` and `2.0` to `2`. That is not a
format change: in PHP `100.0 === 100` is false, float against int, so a
comparison the author wrote and tested stops matching the moment the formatter
runs over the file. `1.50` to `1.5` is the same number of the same type and is
the correct half of the rule. The probe is three lines — `$a = 100.0; if ($a
=== 100.0)` comes back as `$a = 100; if ($a === 100)` — and nothing reports it:
`fmt -l` is clean afterwards, the linter sees valid code, and the only symptom
was a chart drawing 0/50/100 instead of 0/25/50/75/100, caught by a text diff
of the SVG.

Correction: `Task\Chart` casts both sides of the comparison,
`(float)$unit["max"] === (float)self::SHARE`, with a comment saying why. It is
the only place in the tree that compares a computed float against a
whole-number literal. The fix belongs in Go: the formatter prints a float
literal with its fraction.

## 22. `self::` in a const array resolves against the reader

Specified: `Task\User` owns the role list once — `const ROLES = array(self::ADMIN, self::RUNNER, self::VIEWER)` beside the three constants it is built from — so the permission matrix's column count and the role select cannot drift apart.

The members are not resolved where the const is declared. They are resolved where it is read: `Task\User::ROLES` from `mod_settings` raises `undefined class constant mod_settings::ADMIN`, and from `mod_user_edit` it raises `mod_user_edit::RUNNER`. The same const read from a method of `Task\User` is correct, which is why a three-line probe passes and the page 500s. Nothing reports it ahead of the request: `fmt` and `lint` are clean, and the linter resolves no constant.

Correction: `Task\User::ROLES` and `::ASSIGNABLE` write the three strings out, with a comment saying why. The fix belongs in Go: a constant expression is evaluated in the scope of the class that declares it.

## 23. `fmt` moves a comment out of the expression it sits in

Specified: a comment sits beside the line it is about. In the thread page's
`scope()` the two halves of a run's report — the images, which are looked at,
and the prose, which is read — are two keys of one array, and which of them the
page draws is the thing to say beside them.

`phpscript fmt` relocates a comment written anywhere inside an expression to
**after the statement that holds it**. Both of these are one pass:

```php
return array(
	"thread" => $folder,
	// The images are the thing to look at and the prose the thing to read.
	"visual" => $report["images"] !== "",
);
```

```php
return array("thread" => $folder, "visual" => $report["images"] !== "");
// The images are the thing to look at and the prose the thing to read.
```

The same happens in an argument list, and in a literal nested one deeper: the
comment leaves every level and lands after the statement. A comment at the head
of a block stays where it is written, so the rule is expressions, not comments.
The result is stable — a second pass changes nothing — which is what makes it
quiet: the comment now reads as being about the statement after it, and the
author who moved it next sees no diff.

`fmt -l` does report the file, so the pipeline does catch it. It cost one
round trip per file and nothing silent.

Correction: nothing in the tree writes a comment inside an expression. Where one
is wanted it goes above the statement, or into the docblock of the method, or
into the template that draws the thing. The fix belongs in Go: a comment inside
an expression keeps its position, the way `gofmt` keeps one inside a composite
literal.

## 24. An empty path parameter lets the query string stand in for it

`runner/request.go:479` writes a path parameter into `$_REQUEST` only when it
matched something:

```go
value := r.PathValue(param.Name)
if value == "" && param.Rest {
	value = r.PathValue("*")
}
if value != "" {
	c.Path[param.Name] = value
}
```

A `{rest...}` wildcard that matched nothing is therefore absent, and `$_REQUEST`
merges the path over the query, so `param("rest")` answers with whatever the
query string called `rest`. The route `/settings/projects/{rest...}` and the URL
`/settings/projects/?rest=widget` drew the edit page for `widget`, where Go
answers 404: the caller supplied a path segment through the query string.

Nothing is escalated by it here — the page the parameter reaches is one the
caller may already ask for by its real URL — but a path parameter that an
untrusted query can set is the wrong shape for any check written over it.

Correction: a wildcard is read from `Task\Route::path()` with the literal prefix
cut, never with `param()`. `/settings/projects/?rest=widget` is 404 on both
front ends, verified. The fix belongs in Go: an empty match of a parameter the
pattern declares should be written as `""` and shadow the query, the way a
matched one does. A declared parameter is either present or the route did not
match, and either way the query has no say in it.

## 25. `html/template` normalises a URL attribute and not a data attribute

`href="{{.}}"` puts Go's escaper in a URL context: it percent-escapes what is
outside the unreserved set and the URL delimiters, and leaves an existing `%XX`
alone. The same value in `data-next="{{.}}"` is plain attribute escaping, which
normalises nothing. MiniTPL has one escaper and applies it to both, so a link
carrying `?workspace=%20labs` matched in the `data-` attribute and diverged in
the `href`.

Correction: `Task\Route::linked()` does the URL normalisation, and a template
asks for it per link rather than having `href()` fold it in — the two attributes
want different answers for one value, so the choice belongs at the call site.
Not a runtime gap in the end, but it is the one place where matching Go's output
needs a second escaper, and it is invisible until a path carries a `%XX` or a
space.

## 26. A host binding cannot declare that it takes an array

The runtime has two spellings of a PHP list. An array a script writes is a
`*model.Array`; what `array_keys`, `array_values` or `explode` answer is a Go
`[]any` (`stdlib/core/arrays.go:69`). `runner/invoke.go:345 coerceArg` converts
neither into the other, and a binding declaring `*model.Array` refuses the
slice. The TypeError it raises names one word twice, because both spellings
print as `array` (`runner/dispatch.go:120`):

```
call(): Argument #1 must be of type array, array given
```

`Task\Folders::warm` asks `Host::folderDirs(array_keys($ask), ...)`. The folder
tree is drawn on every page, and every page answered 500. No fixture saw it:
`stubs/host.php` answers in PHP, where both spellings are one type.

Correction: every binding parameter that takes a list is declared `any` and read
in `phpfront/names` or `phpfront/fields` with `model.LenValues` and
`model.RangeValues`, the runtime's own readers of both spellings. The fix belongs
in Go, in `coerceArg`: a `[]any` bound for a `*model.Array` parameter wraps, the
way an `int` bound for a `float64` converts. Until then a binding signature
states nothing about what it takes, and the error message states nothing about
the fault.

## Not a runtime gap

`~/.claude/skills/dev/skills/phpscript/SKILL.md` states that path parameters arrive in `$_PATH["param"]`. No such superglobal exists. phpscript's own `docs/README.md` says so: "nothing here spells `$_PATH`, a name PHP does not have." Path parameters are merged into `$_REQUEST`, written after the query, the form and the cookies. The skill needs the line corrected.

---

[Contents](README.md) · [Modules](modules.md) · [Development](development.md)
