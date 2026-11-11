name: inventory
description: Task\Inventory\Hosts keeps the host table the Inventory page owns, and the page folds the hub's systems into those rows - a system nobody entered goes in on the read, a row the hub says nothing of draws with no measures, and a removal is the only way out.
---
<?php

require "modules/inventory/mod_inventory.php";

$db = Task\DB::writer();
$store = new Task\Inventory\Hosts($db);

// Two hosts of this fixture: one a person entered over an ssh login, one the
// hub reported. The suite shares a connection, so both addresses are its own.
$manual = $store->add("fx-inv-box", "10.77.0.2", Task\Inventory\Hosts::MANUAL, "debian@10.77.0.2", "fx-inv-ann");
$hub = $store->seen("fx-inv-web", "10.77.0.3");

echo var_export($manual !== "", true), " ", var_export($hub !== "", true), "\n";

// The address is the key: the same machine goes in once, whoever asks.
echo var_export($store->add("fx-inv-again", "10.77.0.2", Task\Inventory\Hosts::MANUAL, "", ""), true), "\n";
echo var_export($store->seen("fx-inv-renamed", "10.77.0.2"), true), "\n";

$row = $store->find("10.77.0.2");

echo $row["name"], " ", $row["source"], " ", $row["login"], " ", $row["added_by"], "\n";
echo var_export($store->find("10.77.0.9"), true), "\n";

// A name goes in empty and reads as the address.
$store->add("", "10.77.0.4", Task\Inventory\Hosts::MANUAL, "root@10.77.0.4", "fx-inv-ann");

$named = $store->find("10.77.0.4");

echo $named["name"], "\n";

// names is what the binding folds the hub into: the name of each host by the
// address it answers on.
$held = array();

foreach ($store->all() as $host) {
	if (substr($host->name, 0, 7) === "fx-inv-" || $host->address === "10.77.0.4") {
		$held[] = $host;
	}
}

$names = mod_inventory::names($held);

echo $names["10.77.0.2"], " ", $names["10.77.0.3"], " ", $names["10.77.0.4"], "\n";

// sync writes a host the hub reports and the inventory does not hold, and
// leaves a row that is there as it is: the name a person gave a host outlives
// the hub renaming it.
$mod = new mod_inventory;
$systems = array(
	array("name" => "fx-inv-hub-name", "host" => "10.77.0.2"),
	array("name" => "fx-inv-fresh", "host" => "10.77.0.5"),
	array("name" => "no address", "host" => ""),
);
$by = $mod->sync($store, $held, $systems);

echo $by["10.77.0.2"]->name, " ", $by["10.77.0.5"]->name, " ", $by["10.77.0.5"]->source, "\n";
echo var_export(isset($by[""]), true), "\n";

// A row of the table draws whether the hub says anything of it or not: a
// system with no status reads as no measures, and the menu says where the row
// came from.
$measured = array(
	"id" => "sys-fx-inv", "name" => "fx-inv-box", "host" => "10.77.0.2", "status" => "up", "up" => true,
	"cpus" => 8, "cpu" => 12.5, "memory" => 48, "disk" => 71, "measured" => true,
	"memory_used" => 15.5, "memory_total" => 32, "disk_used" => 355, "disk_total" => 500,
);
$quiet = array(
	"id" => "", "name" => "fx-inv-box", "host" => "10.77.0.2", "status" => "", "up" => false,
	"cpus" => 0, "cpu" => 0, "memory" => 0, "disk" => 0, "measured" => false,
	"memory_used" => 0, "memory_total" => 0, "disk_used" => 0, "disk_total" => 0,
);
$logins = array("10.77.0.2" => "debian@10.77.0.2");
$entry = $by["10.77.0.2"];

$line = $mod->row($measured, $logins, true, $entry, true);

echo $line["status"], " ", var_export($line["drop"] !== "", true), " ", $line["login"], "\n";
echo var_export(strpos($line["drop"], $entry->id) !== false, true), "\n";

$line = $mod->row($quiet, array(), false, $entry, false);

echo $line["status"], " ", var_export($line["drop"], true), " ", var_export($line["measured"], true), "\n";

// A host with no row of its own takes no Remove, and neither does a reader
// without Allow Deletes.
$line = $mod->row($measured, $logins, true, null, true);

echo var_export($line["drop"], true), "\n";

// drop is the only way out, and it takes the row it names.
$store->drop($entry->id);

echo var_export($store->find("10.77.0.2"), true), "\n";

$store->drop($named["id"]);
$store->drop($by["10.77.0.3"]->id);
$store->drop($by["10.77.0.5"]->id);
---
true true
''
''
fx-inv-box manual debian@10.77.0.2 fx-inv-ann
false
10.77.0.4
fx-inv-box fx-inv-web 10.77.0.4
fx-inv-box fx-inv-fresh beszel
false
up true debian@10.77.0.2
true
no measures '' false
''
false
