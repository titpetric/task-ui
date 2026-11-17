<?php

// @startup

$module = new Task\Module;

$module->setName("taskui/user");
$module->setTitle("User");
$module->setDescription("Who is signed in and who anyone else is: the sign in page, a user's own page and their keys");
$module->setIcon("users");
$module->setTables(array("user", "user_auth_password", "user_auth_token", "user_email", "user_session"));

return $module;
