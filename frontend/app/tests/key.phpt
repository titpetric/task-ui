name: key
description: Task\Key reads the keys a run is given - the instance SSH keys, and the two tables that say who gets one, by the role a user acts with and by their name.
---
<?php

require "ext/autoload.php";

// Every statement here reads a whole table, so each answer is narrowed to the
// rows this fixture owns: one connection serves the suite and another
// fixture's key is as real as these.

$db = Task\DB::writer();

// The rows are written with replace, because a name is unique and another
// fixture may have written the one a row is about.
//
// Three keys, of which one is a run's: the encryption key is no run's, and a
// key scoped to one workspace is no run of another's.
$db->replace("key_info", array("name" => "fx-key-deploy", "purpose" => "ssh", "algo" => "ed25519", "fingerprint" => "fx:key:1", "scope" => "instance"));
$db->replace("key_info", array("name" => "fx-key-scoped", "purpose" => "ssh", "algo" => "ed25519", "fingerprint" => "fx:key:2", "scope" => "workspace:fx-key-alpha"));
$db->replace("key_info", array("name" => "fx-key-age", "purpose" => "encrypt", "algo" => "age", "fingerprint" => "fx:key:3", "scope" => "instance"));

// The one key reaches a role and a name, which are the two ways a run gets one.
$db->replace("key_role", array("id" => "fx-key-kr1", "key_name" => "fx-key-deploy", "role" => "runner"));
$db->replace("key_user", array("id" => "fx-key-ku1", "key_name" => "fx-key-deploy", "user_name" => "fx-key-bob"));

$keys = new Task\Key;

// mine is the entries of a list this fixture wrote, in the order the statement
// answered them.
function mine($values) {
	$out = array();

	foreach ($values as $value) {
		if (strpos($value, "fx-key-") === 0) {
			$out[] = $value;
		}
	}

	return $out;
}

// The instance SSH keys: an encryption key is not one, and a workspace's own is
// not one either.
echo implode(",", mine($keys->ssh())), "\n";

// A link table is the values of each key name, and a key no row names has no
// entry.
$links = $keys->links("SELECT key_name, role AS value FROM key_role");

echo implode(",", $links["fx-key-deploy"]), "\n";
echo var_export(isset($links["fx-key-scoped"]), true), "\n";

// grants reads one: the key names the value, or it does not, and a key with no
// entry names nothing.
echo var_export(Task\Key::grants($links, "fx-key-deploy", "runner"), true), "\n";
echo var_export(Task\Key::grants($links, "fx-key-deploy", "viewer"), true), "\n";
echo var_export(Task\Key::grants($links, "fx-key-gone", "runner"), true), "\n";

// The keys each user's runs get, which is what reads both link tables: ann by
// her role, bob by his name, and carol by neither, so the picker says she has
// none.
$ann = new Task\Model\User(array("id" => "fx-key-u1", "name" => "fx-key-ann", "role" => "runner"));
$bob = new Task\Model\User(array("id" => "fx-key-u2", "name" => "fx-key-bob", "role" => "viewer"));
$carol = new Task\Model\User(array("id" => "fx-key-u3", "name" => "fx-key-carol", "role" => "viewer"));
$held = $keys->ofUsers(array($ann, $bob, $carol));

foreach (array("fx-key-ann", "fx-key-bob", "fx-key-carol") as $name) {
	$got = isset($held[$name]) ? mine($held[$name]) : array();

	echo $name, ": [", implode(",", $got), "]\n";
}
---
fx-key-deploy
runner
false
true
false
false
fx-key-ann: [fx-key-deploy]
fx-key-bob: [fx-key-deploy]
fx-key-carol: []
