<?php

// @startup
//
// User. Who is signed in and who anyone else is: the sign in page, a user's
// own page, their keys and the form that edits them.
//
// task-ui runs this file once, before it serves a request, and the module
// draws its pages only if Task::register takes what it built. The tables are
// the ones this module's schema/ creates and no other module may.

$module = new Task\Module;

$module->setName("taskui/user");
$module->setTitle("User");
$module->setDescription("Who is signed in and who anyone else is: the sign in page, a user's own page and their keys");
$module->setIcon("users");

// None yet: the user table's migration is still under the settings module and
// the password, token, session and email tables under internal/access, which
// is the Guard that writes them. The statements over the user table are this
// module's, in Task\User\Users.
$module->setTables(array());

Task::register($module);
