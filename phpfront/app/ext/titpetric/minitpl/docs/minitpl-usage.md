MiniTPL - Mini Template
=======================

Last update - Tue 18 Dec 2012 11:24:08 AM CET

by Tit Petrič ( tit.petric@gmail.com )


Introduction
------------

This is a production stable version of our internal templating
system, which supports template compilation, and is PHP4 and
PHP5 compatible. It uses some very advanced PHP functionality,
to keep down the size of the code and at the same time still
provide exceptional functionality.

The total code is 13.5KB and is propperly indented, contains
minimal commenting and perfectly readable with all white space
left in tact.

The code falls under the MIT license, see LICENSE.

Installation and Requirement
----------------------------

MiniTPL requires PHP version 4.3.0 or later.

It can work with lower php versions also, if you provide
your own [file_get_contents()](http://php.net/file_get_contents) function.

Depending on the folder location where you use the template object,
it will need the following folders:

> templates/<br/>
> templates_c/ <span class="red">*</span>

Since this templating system is a compiling template system,
you will need a writable compile directory. By default the compiled
templates are stored in a `templates_c/` directory next to the
`templates/` directory, and it is created when missing.


Language reference
------------------

Since syntax for the template system is loosely based on Smarty
and PHP language syntax, knowing some PHP basics while using this
template system will help you a long way.

1. <a href="#variables">Language constructs</a>
2. <a href="#loops">Loops</a>
3. <a href="#conditions">Conditions</a>
4. <a href="#blocks">Block and Inline definitions</a>
5. <a href="#advanced">Advanced, embedding php code</a>
6. <a href="#php">PHP usage reference</a>
7. <a href="#hooks">Hooks</a>

----------------------------------------------------------

<h3 id="variables">1. Language constructs <a href="#top">Δ Jump to top</a></h3>

Theese basic language constructs provide you with some insight
into the workings of the template system, so you can start to
create your own templates.

Let me just start by saying, all language constructs are
defined between curly brace tokens (`{` and `}`). In case
the template engine doesn't recognise the token, it leaves
it as-is, with curly braces in tact. This should make the
template engine javascript/json safe, no escaping is required.

#### 1.1. Variables

Using variables from templates is easy. Variables are enclosed
in curly braces like so: `{variable}`. Since variables are parsed
on the last stage, prefixing variables with the dollar sign is
optional. So, `{$variable}` is the same as `{variable}`.

```smarty
{variable} is the same as {$variable}
```

#### 1.2. Arrays

There is a shorthand syntax for using arrays inside a template.
The array index separator is a dot (`.`). Consecutive dots can be
used for traversing into array depth like `{$array.items.0}`, or
even variables starting with `$`, like `{$sections.$news_section.title}`.
PHP syntax for arrays is also supported.

```smarty
{$array.items} is the same as {$array['items']}
{$array.items.0} is the same as {$array['items'][0]}
{$array.$items.0} is the same as {$array[$items][0]}
```

#### 1.3. Modifiers

Modifiers are PHP functions, which take one arbitrary value
(usually, string), and return a string, which gets shown.
Some functions in PHP can be used as modifiers
(strtoupper, ucfirst, str_rot13, strrev, count, ...).

```php
/** Custom modifier example */

function add_it_up($array)
{
	$size = 0;
	foreach ($array as $value) {
		$size += $value['size'];
	}
	return $size;
}
```

To use the above function in a template, just do `{$variable|add_it_up}`,
which will loop trough all `$variable` items and add up the value of
the size element and return the total sum of all sizes.

You can't pass additional parameters to modifiers. If you need to
do that, then take a look at <a href="#advanced">advanced templating</a>,
which will show you a way to embed php code inside a template.

Modifiers can be chained. They are applied left to right, so
`{variable|tolower|add_it_up}` lowercases the value first and passes
the result to `add_it_up`.

Aditional template modifiers are `toupper` for `strtoupper` and
`tolower` for `strtolower`. No additional functions are created
for theese special modifiers.

```smarty
{variable|toupper}
{variable|add_id_up}
```

The above code gets compiled to:

```php
<?php
	$_v = &$this->vars;
	echo htmlspecialchars(strtoupper($_v['variable']), ENT_QUOTES);
	echo htmlspecialchars(add_id_up($_v['variable']), ENT_QUOTES);
?>
```

The `htmlspecialchars` calls come from auto-escaping, described next.

#### 1.3.1. Escaping

Printed variables are escaped with `htmlspecialchars($value, ENT_QUOTES)`
by default, so a value cannot inject markup or break out of an attribute.
The compiler scans the template to decide this, and it distinguishes:

| Context | Escaped |
|---|---|
| text node | yes |
| open tag, quoted or unquoted attribute value | yes |
| html comment | yes |
| `<script>` and `<style>` body | no |

Script and style bodies are left alone because they are not markup;
escaping there would corrupt the javascript or css rather than protect it.

None of the following needs a modifier:

```smarty
<input type="text" value="{title}"/>
<textarea name="content">{content}</textarea>
<a href="{news.link}" title="{news.title}">Read more ...</a>
<h3>{site.title}</h3>
```

A value that is already markup opts out with the `raw` modifier, or its
`unescape` alias. The opt-out is per tag, so the same variable can be
escaped in one place and raw in another.

```smarty
<div class="body">{article.html|raw}</div>
<div class="body">{article.html|unescape}</div>
```

The `escape` modifier still works and still compiles to a
`htmlspecialchars($left, ENT_QUOTES)` call. It is now redundant in markup
contexts, and useful to force escaping inside a `<script>` body.

Escaping is the outermost step whatever order the modifiers came in, so
`{variable|escape|toupper}` and `{variable|toupper|escape}` both compile to
`htmlspecialchars(strtoupper(...), ENT_QUOTES)`.

A template that is not markup - json, a plain text mail, a CSV - opts its
whole body out with the `{*noescape*}` directive:

```smarty
{*noescape*}
{ldelim}"id": {id}, "title": {title|json_encode}{rdelim}
```

The same switch is available to the caller, for a whole `Template`
instance:

```php
$tpl = new MiniTPL\Template();
$tpl->set_escape(false);
```

Auto-escaping changes compiled output, so compiled templates from before
this feature have to be cleared once on upgrade.

A common use for modifiers is outputting data for javascript,
using the php function `json_encode`.

#### 1.4. Objects

Objects used by templates must be assigned like any other template
variable. You can call their methods or access their properties.

```smarty
{$memcache->get()}
{$memcache->variable}
```

Both objects are resolved from the template's assigned variables:

```php
<?php
	$_v = &$this->vars;
	echo $_v['memcache']->get();
	echo $_v['memcache']->variable;
?>
```

#### 1.5. Constants

Value starting with `_` is assumed to be a constant. The template
system will output the value of the constant if defined,
or just the name of the constant. Item `{_MY_CONSTANT}` will be
used as a constant, because of those rules, however `{$_my_variable}`
wouldnt be, since it starts with the variable identifier `$`.

These are meant for interface language: `{_SAVE}`, `{_SUBMIT}`,
`{_CANCEL}` and the rest of the short strings a template repeats on every
page. The application defines them once, per locale:

```php
define("_SAVE", "Save");
define("_SUBMIT", "Submit");
```

The point of the rule is the compiled output. A constant becomes a plain
`echo _SAVE;` with no `$_v` lookup, so a label costs nothing to render and
nothing to assign - it never travels through the template variables at
all. That is the whole reason the prefix exists.

It is not a general way to reach php constants, and is not meant to become
one. A name without the leading underscore is a variable, so `{APP_NAME}`
reads `$_v['APP_NAME']` and `{PHP_EOL}` reads `$_v['PHP_EOL']`. Inside an
expression the rule does not apply at all, because only variables are
rewritten there: `{if APP_DEBUG}` compiles to `if(APP_DEBUG){`.

```smarty
{_MY_CONSTANT}
{_this_is_also_a_constant}
{$_my_variable}
```

The above code gets compiled to:

```php
<?php
	$_v = &$this->vars;
	echo htmlspecialchars(_MY_CONSTANT, ENT_QUOTES);
	echo htmlspecialchars(_this_is_also_a_constant, ENT_QUOTES);
	echo htmlspecialchars($_v['_my_variable'], ENT_QUOTES);
?>
```

Interface language is exactly the sort of value that carries a quote. A
label reused as an attribute would break the tag it sits in, so constants
escape by context like any other value, and take the same modifiers:

```smarty
<button title="{_SAVE}">{_SAVE}</button>
<div>{_FOOTER_HTML|raw}</div>
```

With `_SAVE` defined as `L'enregistrer`, and again as `Save "as"`, that
button renders as:

```html
<button title="L&#039;enregistrer">L&#039;enregistrer</button>
<button title="Save &quot;as&quot;">Save &quot;as&quot;</button>
```

An apostrophe in a translation is common enough that this is not a
theoretical case, and it breaks a single-quoted attribute just as a double
quote breaks a double-quoted one. `ENT_QUOTES` covers both.

Constants are replaced earlier in the compile than variables, before the
body of a `<script type="text/template">` or `<script type="text/x-jquery">`
is set aside untouched. That is why a constant is still substituted in
there while a variable is left as written:

```smarty
<script type="text/template">
	You can go {nuts} in here, no variables will be parsed.
	Except for {_CONSTANTS}, those will work.
</script>
```

#### 1.6. Includes

The statement `{include filename.tpl}` gets replaced with the content
of the specified template, before it does any processing. In case
you change the filename.tpl template, you have to regenerate the
cache of the main template manually (by deleting the cache it or
updating the source template, so the timestamp is modified).

There is no limit to the number of files you can include.

```smarty
<html>
<head>
	<title>{$title}</title>
</head>
<body>
{include site_header.tpl}
{$contents}
{include site_footer.tpl}
</body>
</html>
```

If you need to include template files dynamically, you can use `{load}` to
do this. This will solve some caching problems and enable you to use templates
based on what you pass to the main template as variables via `assign`.

Keep in mind that with this you can't use the `inline` and `block` definitions
outside the loaded template, but with `include` you can.

```smarty
<html>
<head>
	<title>{$title}</title>
</head>
<body>
{load $dynamic_header}
{load $dynamic_contents_template}
{load $dynamic_footer}
</body>
</html>
```


#### 1.7. Comments

Comments are enclosed between `{*` and `*}`. Comments are stripped
at compile time, and don't show up in the compiled code or in
the template engine output. This is useful for extensive documentation
if needed, since it doesn't give any overhead at run-time.

```smarty
{* This is a comment that won't be shown anywhere,
   except in the source template, only to developers. *}

Hello world!
```

The above compiles to:

`Hello world!`

<h3 id="loops">2. Loops <a href="#top">Δ Jump to top</a></h3>

#### for, foreach, foreach / else, while

Since the main goal when doing pseudo code here is to make it
friendly for the developer, syntax is PHP inspired. While
this isn't ideal for web designers that make templates, it
is ideal for the majority of the PHP developers, since they
don't need to learn a new language for templating.

```smarty
<html><body>
{for $i=0; $i<count($array); $i++}
	Hop number {i}<br/>
{/for}
{foreach $array as $value}
	I like values like I like: {value}<br/>
{/foreach}
{foreach $array as $key=>$value}
	I like my value {value} to have a key {key}.<br/>
{/foreach}
{foreach $array as $key=>$value}
	I like my value {value} to have a key {key}.<br/>
{else}
	I'm sorry, I have nothing in the array.<br/>
{/foreach}
{while ($k++ < 10)}
	Well, k is {k}<br/>
{/while}
</body></html>
```

You see, this is very friendly for a PHP developer. Lets compare
it to the original PHP code (two ways):

```php
echo '<html><body>';
for ($i=0; $i<count($array); $i++) {
	echo 'Hop number '.$i.'<br/>';
}
foreach ($array as $value) {
	echo 'I like values like I like: '.$value.'<br/>';
}
foreach ($array as $key=>$value) {
	echo 'I like my value '.$value.' to have a key '.$key.'.<br/>';
}
if (!empty($array)) {
	foreach ($array as $key=>$value) {
		echo 'I like my value '.$value.' to have a key '.$key.'.<br/>';
	}
} else {
	echo "I'm sorry, I have nothing in the array.<br/>";
}
while ($k++ < 10) {
	echo "Well, k is ".$k."<br/>";
}
echo '</body></html>';
```

While this example illustrates the similarity in syntax,
it also shows how writing templates is actually less time
consuming and more readable in the long run.

If you notice, the statements all support full PHP syntax.

While not very obvious with `foreach`, you can see that
the code is practically identical for other statements,
where you can mix the complete PHP syntax along with
template features like the table addressing shorthand.

```smarty
<p>{foreach $items.news as $newsitem}<b>-</b>{/foreach}</p>
<p>{foreach $items['news'] as $newsitem}<b>+</b>{/foreach}</p>
<p>{for $i=0; $i<count($items.news); $i++}<b>!</b>{/for}</p>
<p>{for $i=0; $i<count($items.news)+count($items['news']); $i+=2}<b>?</b>{/for}</p>
```


<h3 id="conditions">3. Conditions <a href="#top">Δ Jump to top</a></h3>

The syntax for the `if` and `elseif` statements are
PHP compatible. You can call functions, methods, use
constants and arithmetic operations. Variables are parsed
with the templating extensions for using arrays, so both
syntaxes can be used at the same time.

#### 3.1. if statement

```smarty
{if $is.admin && $user['name']=="black"}
        {* Hello black! Only you can edit things,
           but only as long as you stay an admin. *}
  ...
{/if}
```

#### 3.2. if/elseif/else statement

```smarty
{if $is.admin && $user['name']=="black"}
        {* Hello black! Only you can edit things,
           but only as long as you stay an admin. *}
  ...
{elseif $is_moderator}
        {* Hello moderator! You can do some things. *}
{else}
	{* You are a nobody and you earn nothing! *}
{/if}
```

#### 3.3. foreach/else statement

```smarty
{foreach $newsitems.$section.items as $item}
	<div class="newsitem">
		<h3 class="title">{item.title}</h3>
		<div class="content">{item.content}</div>
	</div>
{else}
	<div class="notice">
	No newsitems exist in the chosen section.
	</div>
{/if}
```

#### 3.4. nocache

Minitpl template files are translated into PHP code to give you
the best possible execution speed. Using `include` statements
sometimes makes it hard to invalidate this cache, so for
development purposes we have included a `nocache` directive.
This erases your template cache after the file has been used once.

```smarty
{*nocache*}
```

This behaviour needs to be activated by setting `$tpl->_nocache` to `true`.

<h3 id="blocks">4. Block and Inline definitions <a href="#top">Δ Jump to top</a></h3>

Depending on the usage, you might want to reuse pieces of
the template multiple times in the same or multiple templates.
You can achieve this by using Block and Inline definitions.
The difference between a block and inline definition is,
that the `block` definition defines a PHP function, and
can be used recursively in the template, for example, to
traverse a tree structure. The `inline` definition only
allows the same piece of template code to be reused multiple
times.

#### 4.1 block definition and usage

```smarty
{block recurse}
{if $i++ < 10}
        call {i}
        {block:recurse}
{/if}
{/block}

{block:recurse}
```

The compiled template will look something like this:

```php
<?php

/* This code has been cleaned up some, for
   documentation purposes. It illustrates
   the compile aspect of blocks, but is not
   a carbon copy of the compiled template. */

	$_v = &$this->vars;
	function recurse_1213891842_673($_v) {
		if ($_v['i']++ < 10) {
	        	echo "call ".$_v['i'];
        		recurse_1213891842_673(&$_v);
		}
	}
	recurse_1213891842_673(&$_v);
?>
```

As you can see, the block definition gets compiled into
a function definition, which lives in the same variable space.

While traversing a tree this way isn't very practical,
it is however possible. You probablly won't ever need
this functionality. In our experience recursion itself
is very rare, and even if used, it is handled on the PHP
level, and not the templating level.

You can use `block` definitions instead of `inline` definitions
if you are worried about code overhead. I would consider this
 when an `inline` definition is multiple kilobytes in size and
is beeing used extensively in the same template.

#### 4.2 inline definition and usage

The inline keyword comes from `C++`, where the compiler
would replace the calls to an inline function definition
with the function itself. This is a speed gain for
`C++`, since calling a function many times is more
expensive, than copying the code around. The practical
reason for this inside a template goes along the same
train of thought.

```smarty
{inline newsitem}
<div class="news">
	<h3 class="title">{news.title}</h3>
	<div class="content">{news.content}</div>
</div>
{/inline}

<div class="top_news">
	{foreach $newsitems.top.items as $news}{inline:newsitem}{/foreach}
</div>

<div class="other_news">
{foreach $newsitems.$section.items as $news}{inline:newsitem}{/foreach}
</div>
```

Usage of `inline` definition decreases template size before
compilation, and helps with design standardisation. Same page
components can be literally re-used inside the template and
changing one aspect of the design is done in one place instead
of every place the same code snippet is used.

<h3 id="advanced">5. Advanced, embedding php code <a href="#top">Δ Jump to top</a></h3>

Since the above language constructs only take care of the
read only aspect of template programming, there is sometimes
also a need to modify variables inside templates for various
uses. The `eval` and `eval_literal` constructs take care
of that. The `php` construct is ment for more advanced
operations, and like `eval` lives in the template variable
space.

#### 5.1 eval

The eval construct allows quick operations on local
template variables. For example, if you want to build
a table with alternating css styles on rows, you would
do something like this:

```smarty
{eval $style="even";}
<table>
{foreach $rows as $row}
{eval $style = ($style=="even") ? "odd" : "even"}
<tr class="{style}">
<td>{row.message}</td>
</tr>
{/foreach}
</table>
```

All variables when using the `eval` construct, are
mapped to local template variables.

It is not ok to use variables named or starting with
`$_v`, since that will result in naming errors on
compile time, and you won't be able to use your data
in the template, since it's in the wrong location.

#### 5.2 eval_literal

When you need to execute more complex code, you can use the
`eval_literal` construct. The code inside does not get evaluated,
meaning it is kept as-is. Assigned variables remain available through
the internal `$_v` array.

```smarty
{eval_literal
	$_v['menu_data'] =
		$_v['cms_module']->get_menu("branch", array("item","menu")); }
```

#### 5.3 php

```smarty
{php}
function mygettime()
{
        return array("time"=>time(),"date"=>date("r"),"microtime"=>microtime());
}
$mygettime = mygettime();
{/php}

{mygettime|var_dump}
```

The above code gets compiled into:

```php
<?php
	$_v = &$this->vars;

	function mygettime()
	{
	        return array("time"=>time(),"date"=>date("r"),"microtime"=>microtime());
	}

	$_v['mygettime'] = mygettime();

	echo var_dump($_v['mygettime']);
?>
```

This is by far the least used construct and also least tested.
If you want to build objects or functions or use a lot of PHP
code inside the template, you are definetly doing something wrong,
even if supported by the templating system.

It is not ok to use variables named or starting with
`$_v`, since that will result in naming errors on
compile time, and you won't be able to use your data
in the template, since it's in the wrong location.

<h3 id="php">6. PHP usage reference <a href="#top">Δ Jump to top</a></h3>

This section is about functionality of the template system in PHP.
You already know how to make templates like the ones above, this
teaches you how to create the PHP code, that uses the templates.

#### 6.1 Basic usage

For basic usage, you need to include the file `class.template.php`.
The functions used include `class.template_compiler.php` if needed.
The following methods are exported:

`load`, `assign`, `render`, `get`, `set_paths`, `set_compile_location`, `compile`,
`add_hook`, `clear_hooks`

You will probablly only use the first four methods unless you
want to change the default paths, where the system is searching for
templates, or if you want to do your own compiling of templates
for whatever reason.

##### 6.1. load ( string $filename )

With this method you specify which template file to load. It will
search trough the configured paths (Default: `templates/`) and
use the template. Calling this method resets the template to it's
defaults, and you have to fill up the content using the `assign`
method.

##### 6.2. assign ( mixed $key, [ mixed $value = '' ] )

This is the most complex method available, and it's behaviour is
dependant on the parameters supplied.

If the first parameter is an array, and the second parameter stays
at the default value, an entry will be created for each key and value
pair in the first parameter.

If the first parameter is an array, and the second parameter is
a string value, an entry will be created for each key and value
pair in the first parameter, using the second parameter as a prefix
for each key.

If the first parameter is a string, the second parameter will be
assigned as it's value. The second parameter can be any PHP type.

```php
$data = array(); // some example data
$data['title'] = "Leno promises smooth transition to O'Brien";
$data['content'] = "For months, Fallon has been widely considered
                    the top choice to succeed O’Brien when he steps
                    down next year. On Thursday, published reports ...";

/* This will define an entry {timestamp} */

$tpl->assign("timestamp", time());

/* This will define entries {title} and {content} */

$tpl->assign($data);

/* This example will define {news_title} and {news_content} */

$tpl->assign($data, "news");

/* This will define the item {news}, which contains an array.
   You can output the fields with {news.title} and {news.content} */

$tpl->assign("news", $data);
```

##### 6.3. render and get

Theese methods dont have any arguments. Template compilation is
done in the background, if needed. The method `render` outputs
the data to standard output, while the method `get` simply returns
it, in order to use it in PHP.

##### 6.4. set_paths ( [ $paths = false ] )

The method `set_paths`, takes an array, with possible locations
for template files. The compiled location is set with the
`set_compile_location` function documented later. Paths passed to
this function, must end with a trailing slash.

The template system will traverse the `$paths` array searching
for template files, until one is found. If none are found, an
error is printed, and the script execution is terminated.

This is useful if you have multiple template locations, which
are overridable. For example, you can have the following CMS
structure.

```php
$paths = array();

/* This is the most important location, everything
   can be overriden from inside the theme. */

$paths[] = "theme/templates/";

/* This is the second most important location,
   it usually defines the look of the cms modules */

$paths[] = "modules/".$module_name."/templates/";

/* This is the least important template location,
   it usually provides system wide templates, like
   a paginator template, an xml / rss template, or
   other very general templates. */

$paths[] = "include/templates/";

$tpl->set_paths($paths);
```

Templating constructs like `{include}` are affected by the path
settings. The included templates use the same paths in the same
order, until a needed template is found.

The following syntax is also allowed since 2012/05/24:

```php
$tpl->set_paths("theme/templates/","modules/".$module_name."/templates/","include/templates/");
```

##### 6.5. $this->set_compile_location( [ $compile_location [ , $is_absolute = false ] ] );

By default the compile location of the templates is relative to the
source template folder, in a directory beside it with `_c` appended:
a search path of `templates/` compiles into `templates_c/`, and
`theme/templates/` into `theme/templates_c/`. The directory is created
when it is missing and only has to be writable, so the default paths
need no configuration.

You can modify this behavior with the `set_compile_location` function.
You can set the cache location to a common folder, and mark it as
an absolute location (as opposed to the default, relative path).

```php
$tpl->set_compile_location("/tmp/minitpl", true);
```

The example will store all your templates under `/tmp/minitpl`. This
location needs to be writable. The compiled files under the location
can be safely deleted any time, if you want to regenerate them.

When caching, a full path structure will be created, to avoid any 
conflicts between templates, that use same names. For example:

If you have a template named `site.tpl` under two distinct locations like
`theme/templates/` and `theme/issue2012/templates/` and use both sources,
the conflict is resolved by creating the following structure in the
cache location:

```text
/tmp/minitpl/theme/templates/site.tpl
/tmp/minitpl/theme/issue2012/templates/site.tpl
```

This can let you change your source folders with `set_paths~ and and the
same time only have one cache location for all your templates. When
you want to expunge any stale cache files or use minitpl in a more complex
system, this is a much simpler approach for cache management.

Earlier versions compiled into a `cache/` directory inside every template
folder. The default location moved to the adjacent `templates_c/`, so a
cache written by an older version has to be removed once on upgrade.

<h3 id="hooks">7. Hooks <a href="#top">Δ Jump to top</a></h3>

Hooks rewrite template source around compilation. A hook implements
the `MiniTPL\Hook` interface:

```php
use MiniTPL\Hook;

class TrimHook implements Hook
{
	function execute($filename, $contents)
	{
		return trim($contents);
	}
}
```

`execute` receives the template filename and the template contents,
and returns the new contents.

##### 7.1. add_hook ( Hook $hook, string $position )

Registers a hook on the template. The position is one of the two
`Hook` interface constants:

| Position | Runs |
|---|---|
| `Hook::POSITION_PRE` | before compilation, on the template source |
| `Hook::POSITION_POST` | after compilation, on the generated php |

```php
$tpl = new MiniTPL\Template();
$tpl->add_hook(new TrimHook(), Hook::POSITION_PRE);
```

Hooks at the same position run in the order they were added. They
apply at compile time, so a template already in the compile cache is
not affected until it recompiles.

##### 7.2. clear_hooks ( )

Removes every registered hook from the template.

<h3 id="cli">8. Command line <a href="#top">Δ Jump to top</a></h3>

The `cmd/minitpl` script compiles and renders templates from the shell,
for builds and for trying a template out. It is a phpscript script with a
`#!/usr/bin/env phpscript` shebang: run it with `phpscript cmd/minitpl`,
or make it executable and run `./cmd/minitpl`, from a checkout of the
repository after `composer install`. It loads `vendor/autoload.php` from
the root of the clone and uses `MiniTPL\Compiler` and `MiniTPL\Template`
for the work.

phpscript roots its source filesystem at the directory the command is
started in and reads every path below it, so run the command from the
checkout and name the templates and data relative to it.

##### 8.1. compile

`minitpl compile [templates_dir] [output_dir]` transpiles every `.tpl`
file under `templates_dir` (default: `templates`). Subdirectories keep
their names, `{include}` is resolved against `templates_dir`, and the
output goes to `output_dir`, which defaults to the `templates_c/`
directory beside the source. `-o DIR` names it explicitly, `-q` prints
nothing but errors.

```sh
minitpl compile templates
minitpl compile -o build templates
```

##### 8.2. render

`minitpl render <template.tpl> [data.yml]` renders a template and writes
the result to standard output. The template variables come from a yaml
mapping, read from the data file or, when it is omitted, from standard
input; pass `-` to read standard input explicitly. Rendering compiles
the template when needed, into the `templates_c/` directory beside it.

```sh
minitpl render templates/page.tpl data.yml
cat data.yml | minitpl render templates/page.tpl
```

The data is decoded by the runtime, not by the command line: phpscript
registers `yaml_decode()`, which `phpscript list --bindings` lists, and
the command uses it. A runtime without the yaml binding decodes json
instead, so a `minitpl` run under plain php reads json data.
