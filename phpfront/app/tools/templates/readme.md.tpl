{*noescape*}# {title}

Generated from the module by `atkins docs`. DO NOT EDIT.

![The tables this module reads](schema.svg)

{description}

## Pages

| Method | Path | Answered by |
|--------|------|-------------|
{foreach $routes as $route}| `{route.method}` | `{route.path}` | [`{route.file}`]({route.file}) |{nl}{/foreach}
## The data model

{foreach $tables as $table}### `{table.name}`

{table.comment}

| Column | Type | Key |
|--------|------|-----|
{foreach $table.columns as $column}| `{column.name}` | `{column.type}` | {column.key} |{nl}{/foreach}
{/foreach}## Operations

{foreach $classes as $class}### `{class.name}`

{class.doc}

| Method | What it does |
|--------|--------------|
{foreach $class.methods as $method}| `{method.signature}` | {method.doc} |{nl}{/foreach}
{/foreach}{if !$classes}This module owns no data model of its own. It reads the instance's tables
through `code/Task`.

{/if}---

The tables are what `{module}::tables()` names. The operations are the classes
of `storage/`, which hold every statement this module runs.
