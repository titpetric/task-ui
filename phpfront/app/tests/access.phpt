name: access
description: Access decides what a user may do, from their own entry then the role's.
---
<?php

Task\Host::$rules = array(
	"paths" => array("note/all", "note/view", "note/edit", "note/delete", "secret/all", "secret/view", "run"),
	"checks" => array("note/view", "note/edit", "note/delete", "secret/view", "run"),
	"extras" => array(),
);

$ops = new Task\Model\User(array("id" => "01", "name" => "ops", "role" => "runner"));
$role = new Task\Policy(array("note/all" => Task\Policy::ALLOW, "run" => Task\Policy::ALLOW));

$set = new Task\Access($ops, $role);

echo var_export($set->allowed("note/edit"), true), " all stands for every rule\n";
echo var_export($set->allowed("run"), true), "\n";
echo var_export($set->allowed("secret/view"), true), " what nobody sets is denied\n";

$own = new Task\Policy(array("note/delete" => Task\Policy::DENY, "secret/view" => Task\Policy::ALLOW));
$set = new Task\Access($ops, $role, $own);

echo var_export($set->allowed("note/delete"), true), " the user's deny beats the role's allow\n";
echo var_export($set->allowed("note/edit"), true), "\n";
echo var_export($set->allowed("secret/view"), true), " the user's allow adds to the role\n";

$secrets = new Task\Policy(array("secret/all" => Task\Policy::DENY));
$denied = new Task\Access($ops, $secrets, $own);

echo var_export($denied->allowed("secret/view"), true), " the user's allow overrides the role's deny\n";

$inherits = new Task\Access($ops, $secrets);

echo var_export($inherits->allowed("secret/view"), true), " without an entry the user inherits the deny\n";

// A deny of all beats an allow of the rule.
$both = new Task\Policy(array("note/all" => Task\Policy::DENY, "note/view" => Task\Policy::ALLOW));

echo $both->effect("note/view"), " a deny of all decides\n";

// An admin may do everything, and the policies leave it as is.
$root = new Task\Model\User(array("id" => "02", "name" => "root", "role" => "admin"));
$nothing = new Task\Policy();
$noRun = new Task\Policy(array("run" => Task\Policy::DENY));
$admin = new Task\Access($root, $nothing, $noRun);

echo var_export($admin->allowed("run"), true), var_export($admin->allowed("secret/delete"), true), "\n";
echo $admin->summary(), "\n";

// Nobody may do nothing.
$none = new Task\Access();

echo var_export($none->allowed("note/view"), true), " ", $none->summary(), "\n";

// effective is every checkable path, and allows the ones that passed.
echo count($set->effective()), " ", implode(",", $set->allows()), "\n";

// A folder the user may not see refuses the run whatever the rule says.
$scoped = new Task\Model\User(array("id" => "03", "name" => "dev", "role" => "runner", "folders" => '["infra"]'));
$run = new Task\Access($scoped, $role);

echo var_export($run->run("infra"), true), var_export($run->run("other"), true), "\n";
echo var_export($run->run("infra", array("dev")), true), var_export($run->run("infra", array("someone")), true), "\n";
echo var_export($run->sees("infra"), true), var_export($run->sees("other"), true), "\n";
---
true all stands for every rule
true
false what nobody sets is denied
false the user's deny beats the role's allow
true
true the user's allow adds to the role
true the user's allow overrides the role's deny
false without an entry the user inherits the deny
deny a deny of all decides
truetrue
everything, as an admin
false nothing
5 note/view,note/edit,secret/view,run
truefalse
truefalse
truefalse
