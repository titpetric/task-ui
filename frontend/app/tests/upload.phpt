name: upload
description: Task\Upload is where a file dropped into a markdown field lands in the artifact tree - prompts/<workspace>-<project>, one directory per place.
---
<?php

require "ext/autoload.php";

// A place is a folder ID and the directory inside it the page is about. The
// workspace and the project name the directory; what is under them does not,
// so every page of a project writes to one place.
echo Task\Upload::ofPlace("titpetric/mdedit"), "\n";
echo Task\Upload::ofPlace("titpetric/mdedit/docs"), "\n";
echo Task\Upload::ofPlace("malta/apps/incubator/blog"), "\n";

// The root is one step and names itself, and a place of nothing is the root
// too: a page that says where it is reaches this, and one that says nothing
// reaches it through Task\Prompt::root().
echo Task\Upload::ofPlace(Task\Route::INTERNAL), "\n";
echo Task\Upload::ofPlace(""), "\n";
echo Task\Upload::ofPlace("/"), "\n";

// A step is lower case, a run of anything else is one dash, and the dashes at
// either end come off: the store takes a path of plain steps and refuses a
// dot file, so a step of dots slugs to nothing and is dropped.
echo Task\Upload::ofPlace("Malta/Apps Incubator"), "\n";
echo Task\Upload::ofPlace("../etc/passwd"), "\n";
echo Task\Upload::ofPlace(".hidden/x"), "\n";
echo "[", Task\Upload::slug("...."), "]\n";
---
prompts/titpetric-mdedit
prompts/titpetric-mdedit
prompts/malta-apps
prompts/internal
prompts/internal
prompts/internal
prompts/malta-apps-incubator
prompts/etc
prompts/hidden-x
[]
