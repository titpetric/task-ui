<?php

namespace Task;

// Key draws the keys a run is given: which user gets which, what a key name
// may be, and the rows the keys pages list.
//
// No private material is here. The keystore holds it and it does not enter
// PHP, so nothing on this class answers more than a key's name. The rows are
// the settings module's, in Task\Settings\Keys.
class Key {
	/** SSH is the purpose of a key a run gets, as keystore writes the column. */
	const SSH = Settings\Keys::SSH;

	/** WORKSPACE starts the scope of a key of one workspace, which no run of another gets. */
	const WORKSPACE = Settings\Keys::WORKSPACE;

	/**
	 * NAME is what a key name may be, as the Name field's pattern, which the
	 * keyname define of _layout.tpl writes.
	 *
	 * MiniTPL compiles any {word} into an echo, and the quantifier {0,63} is
	 * one: a template cannot write the literal, so it arrives as a field of
	 * the scope the component reads.
	 */
	const NAME = "[a-z0-9][a-z0-9_\\-]{0,63}";

	/** store is the settings module, which owns the statements over the key tables. */
	var $store;

	function __construct($db = null) {
		$this->store = new Settings\Keys($db);
	}

	/**
	 * ofUsers is the instance SSH keys each user's runs get, by user name; a
	 * user with none has no entry, which is what a picker reads as "no keys".
	 */
	function ofUsers($users) {
		$keys = $this->ssh();
		$roles = $this->store->roles();
		$holders = $this->store->holders();
		$out = array();

		foreach ($users as $user) {
			$name = $user->getName();
			$role = User::role($user);
			$names = array();

			foreach ($keys as $key) {
				if (self::grants($roles, $key, $role) || self::grants($holders, $key, $name)) {
					$names[] = $key;
				}
			}

			if (count($names) > 0) {
				sort($names);

				$out[$name] = $names;
			}
		}

		return $out;
	}

	/** ssh is the name of every instance SSH key, by name; a workspace's own is no run's. */
	function ssh() {
		return $this->store->ssh();
	}

	/** grants reports whether a link table names a value for a key. */
	static function grants($links, $key, $value) {
		return isset($links[$key]) && in_array($value, $links[$key], true);
	}

	/**
	 * folder is the key folder as the keys tables draw it, through
	 * Task\Host::keys(): where it is, the name of the instance's encryption
	 * key and every key in it.
	 *
	 * The key_info row holds no public half. The public half is a file beside
	 * the private one in the folder, so a table that prints an
	 * authorized_keys line reads the folder and not the table.
	 */
	static function folder() {
		return Host::keys();
	}

	/** backupName is the name a rotation keeps the previous material under. */
	static function backupName($name) {
		return "backup/" . $name;
	}

	/**
	 * forUser is the user a key form was opened for, ?for=<user>, "" where the
	 * form names nobody: the Keys block of a profile opens it for its own user.
	 *
	 * The name travels in the query because the page adds a key and the user is
	 * who it is for, which is what userHref() writes. frontend/key.go:129.
	 */
	static function forUser() {
		return Route::param("for");
	}

	/** openName is the name a form opened for a user starts with, "" for none. */
	static function openName($user) {
		return $user === "" ? "" : "user_" . $user;
	}

	/**
	 * nameField is the scope the key name component reads: the name a form
	 * opens with, and the pattern a name must match.
	 */
	static function nameField($value) {
		return array("value" => $value, "pattern" => self::NAME);
	}

	/**
	 * kind is the algorithm and the size a key's name ends with: ed25519,
	 * ecdsa256, rsa4096. keystore/local.go:90.
	 */
	static function kind($key) {
		$algo = Row::text($key, "algo");
		$bits = Row::number($key, "bits");

		if ($algo === "ed25519" || $bits === 0) {
			return $algo;
		}

		return $algo . $bits;
	}

	/**
	 * policies is the who field of a key form: the roles' boxes are named
	 * policies and the users' users, which is what /api/keys takes, and with
	 * nothing checked no run gets the key. _layout.tpl's keypolicies define.
	 */
	static function policies($roles, $users, $all) {
		$field = new WhoField("policies", "set", $roles, $users, $all);

		return $field->named("users");
	}

	/**
	 * mine are the instance SSH keys one user's runs get: the keys of their
	 * role and the keys of their name. A workspace's own key is no run's of
	 * another workspace, and a page of the instance's keys lists none of them.
	 */
	static function mine($keys, $user) {
		if ($user === null) {
			return array();
		}

		$name = $user->getName();
		$role = User::role($user);
		$out = array();

		foreach ($keys as $key) {
			if ($key["purpose"] !== self::SSH || $key["workspace"] !== "") {
				continue;
			}

			if (in_array($role, $key["roles"], true) || in_array($name, $key["users"], true)) {
				$out[] = $key;
			}
		}

		return $out;
	}

	/**
	 * ofWorkspace are the SSH keys of one workspace, by the scope its ID
	 * makes, in the order the key folder answered them.
	 *
	 * The scope is the row's ID and not its folder: a workspace that is
	 * renamed keeps its keys, which is keystore.ScopeWorkspace.
	 */
	static function ofWorkspace($keys, $id) {
		$scope = Workspace::SCOPE . $id;
		$out = array();

		foreach ($keys as $key) {
			if ($key["purpose"] === self::SSH && $key["scope"] === $scope) {
				$out[] = $key;
			}
		}

		return $out;
	}

	/**
	 * wsFoot is the foot of a workspace key's form: the submit button, and a
	 * Cancel that closes the dialog the form is in or goes back to the page
	 * the form was opened from.
	 */
	static function wsFoot($next, $dialog, $label) {
		return array(
			"dialog" => $dialog,
			"cancel" => Route::href($next),
			"label" => $label,
		);
	}

	/**
	 * wsGenerate is the form that generates an SSH key of a workspace: the
	 * endpoint it posts to, the page it returns to, and the name the key
	 * folder gives it.
	 *
	 * Whose runs get it is no field: it is a key of this workspace, so every
	 * run of this workspace gets it and there is nothing left to choose.
	 *
	 * $dialog says the form is drawn inside the dialog the keys page opens,
	 * which carries its own heading and a Cancel that closes it.
	 */
	static function wsGenerate($folder, $dir, $next, $dialog) {
		$foot = self::wsFoot($next, $dialog, "Add key");

		return array(
			"api" => "/api/workspaces/" . $folder . "/keys/generate",
			"next" => $next . "?notice=Key+generated",
			"dialog" => $dialog,
			"folder" => $folder,
			"dir" => $dir,
			"comment" => "task-ui " . $folder,
			"foot" => $foot,
		);
	}

	/**
	 * wsUpload is the form that saves a private key already in use as an SSH
	 * key of a workspace, laid out as wsGenerate.
	 *
	 * The dialog's box is shorter than the page's, because the dialog holds
	 * the heading and the sentence above it.
	 */
	static function wsUpload($folder, $dir, $next, $dialog) {
		$foot = self::wsFoot($next, $dialog, "Add key");

		return array(
			"api" => "/api/workspaces/" . $folder . "/keys/upload",
			"next" => $next . "?notice=Key+uploaded",
			"dialog" => $dialog,
			"folder" => $folder,
			"dir" => $dir,
			"rows" => $dialog ? 6 : 10,
			"comment" => "task-ui " . $folder,
			"foot" => $foot,
		);
	}

	/**
	 * wsDialogs are the two dialogs that add an SSH key to a workspace, which
	 * the pages that list its keys open: the keys page and the workspace's own
	 * page open the same two and return to $next.
	 */
	static function wsDialogs($folder, $dir, $next) {
		$generate = self::wsGenerate($folder, $dir, $next, true);
		$upload = self::wsUpload($folder, $dir, $next, true);

		return array("generate" => $generate, "upload" => $upload);
	}

	/** linked are the instance SSH keys linked to a user by name, which a profile lists. */
	static function linked($keys, $name) {
		$out = array();

		foreach ($keys as $key) {
			if (in_array($name, $key["users"], true)) {
				$out[] = $key;
			}
		}

		return $out;
	}

	/**
	 * rows are the keys a table draws, in the order the folder answered them.
	 *
	 * $next is the page a change returns to, a path of this instance with no
	 * mount in it: the redirect is the API's and the API answers at the root,
	 * so the page that reads it puts the mount back.
	 *
	 * $workspace is the workspace whose page draws the table, "" everywhere
	 * else. There a key of the workspace is removed through the workspace,
	 * which takes workspace/edit, and has neither an Edit links item nor a
	 * link on its name: it was added to one workspace and every run of that
	 * workspace gets it, so there is nothing left to choose.
	 */
	static function rows($chrome, $keys, $next, $workspace = "") {
		$out = array();

		foreach ($keys as $key) {
			$out[] = self::row($chrome, $key, $next, $workspace);
		}

		return $out;
	}

	/** row is one key of a keys table: the cells it draws and its menu. */
	static function row($chrome, $key, $next, $workspace = "") {
		$name = $key["name"];
		$owned = $workspace !== "";
		$edit = self::editHref($name, $next);
		$menu = self::menu($chrome, $key, $next, $workspace);
		$copy = Copy::of($key["public_key"], "the public key of " . $name, null, true);

		return array(
			"name" => $name,
			"href" => $key["encrypts"] || $owned || !$chrome->allowed("key/edit") ? "" : $edit,
			"encrypts" => $key["encrypts"],
			"backup" => $key["backup"],
			"comment" => $key["comment"],
			"algo" => $key["algo"],
			"size" => $key["size"],
			"copy" => $copy,
			"menu" => $menu->data(),
		);
	}

	/**
	 * editHref is the page that sets whose runs get a key. It carries the
	 * page it was opened from, so a change made from the Instance page or a
	 * profile returns there; the Keys page is where the form goes by itself
	 * and names nothing.
	 */
	static function editHref($name, $next) {
		$path = "/settings/keys/" . $name;

		if ($next !== "/settings/keys") {
			$path .= "?next=" . urlencode($next);
		}

		return Route::href($path);
	}

	/**
	 * userHref is the page that makes a key for one user, by generating one or
	 * by taking one that exists: the Keys page of the settings, asked for whom.
	 *
	 * The name travels in the query and not in the path, because the page adds
	 * a key and the user is who it is for; a user with no key yet has no page
	 * of their own under the keys.
	 */
	static function userHref($kind, $user) {
		return Route::href("/settings/keys/" . $kind . "?for=" . Route::queryEscape($user));
	}

	/**
	 * ofUser is the Keys block of one user, which components/user-keys.tpl
	 * draws: the head that names them and the keys linked to them by name.
	 *
	 * A profile and the page that edits a user both draw it, and they differ
	 * only in where a change returns to, so $next is the caller's.
	 *
	 * The keys of their role are not here and the head says so: a role's key
	 * comes with the role and is taken away by changing the role, not by
	 * unlinking a key from one user.
	 */
	static function ofUser($chrome, $name, $keys, $next) {
		$buttons = new Buttons(true);
		$bare = count($keys) === 0;

		if ($chrome->allowed("key/edit")) {
			$buttons->item("key", "Generate key", array("href" => self::userHref("generate", $name)));
			$buttons->item("upload", "Upload key", array("href" => self::userHref("upload", $name)));
		}

		$block = array(
			"title" => "Keys",
			"note" => "linked to " . $name . ", in the ssh agent of their runs",
			"empty" => $bare ? "No key linked to " . $name . "; the role's keys come with the role." : "",
			"buttons" => $buttons->data(),
		);

		return array("block" => $block, "keys" => self::rows($chrome, $keys, $next));
	}

	/**
	 * menu is a key's actions: its details on top, then copying and
	 * downloading the public half, the Edit links page, rotating the key and
	 * removing it.
	 *
	 * Rotate is off while a backup serves beside the key, because the folder
	 * keeps one backup and a second rotation would drop it.
	 */
	static function menu($chrome, $key, $next, $workspace = "") {
		$name = $key["name"];
		$owned = $workspace !== "";
		$canEdit = $chrome->allowed("key/edit");
		$remove = $owned ? "/api/workspaces/" . $workspace . "/keys/" . $name : "/api/keys/" . $name;
		$canRemove = $owned ? $chrome->allowed("workspace/edit") : $chrome->allowed("key/delete");
		$menu = new Menu("Actions on " . $name);

		$menu->detail("Fingerprint", $key["fingerprint"], true);

		if ($key["created_at"] !== "") {
			$menu->detail("Created", Clock::ago($key["created_at"]), false, Clock::read($key["created_at"]));
		}

		if ($key["rotated_at"] !== "") {
			$menu->detail("Rotated", Clock::ago($key["rotated_at"]), false, Clock::read($key["rotated_at"]));
		}

		if ($key["backup"] !== "") {
			$menu->detail("Backup", $key["backup"] . " " . $key["backup_fingerprint"], true);
		}

		$menu->detail("Scope", $key["workspace"] === "" ? "instance" : "workspace " . $key["workspace"]);
		$menu->detail("Purpose", $key["encrypts"] ? "encryption" : "git + ssh", false, $key["git"] ? "git pulls and clones with it" : "");

		if ($chrome->allowed("key/view") || $owned) {
			$menu->detail("Links", self::says($key));
		}

		$menu->item("copy", "Copy public key", array("data-copy" => $key["public_key"]));
		$menu->item("download", "Download public key", array("data-download" => $key["public_key"], "data-filename" => $name . ".pub"));
		$menu->when($canEdit && !$key["encrypts"] && !$owned, "edit", "Edit links", array("href" => self::editHref($name, $next)));
		$menu->sep();

		if ($canEdit && $key["backup"] !== "") {
			$menu->item("rotate", "Rotate key", array("disabled" => "Delete the backup first"));
			$menu->danger("trash", "Delete backup", array(
				"data-delete" => "/api/keys/" . $name . "/backup",
				"data-confirm" => self::dropsBackup($key),
				"data-confirm-ok" => "Delete backup",
				"data-next" => "reload",
			));
		}

		if ($canEdit && $key["backup"] === "") {
			$menu->item("rotate", "Rotate key", array(
				"data-api" => "/api/keys/" . $name . "/rotate",
				"data-method" => "POST",
				"data-next" => $next . ($key["encrypts"] ? "?rekeyed={rekeyed}&what=secret" : "?notice=Key+rotated") . "#key-" . $name,
				"data-confirm" => self::rotates($key),
				"data-confirm-ok" => "Rotate key",
			));
		}

		$menu->whenDanger($canRemove && !$key["encrypts"], "trash", "Remove", array(
			"data-delete" => $remove,
			"data-confirm" => "Remove " . $name . "? " . ($owned ? "The workspace's runs lose it." : "Runs and hosts that use it lose access.") . " Its files stay in the key folder as a backup.",
			"data-confirm-ok" => "Remove",
			"data-next" => "reload",
		));

		return $menu;
	}

	/**
	 * says reads who gets a key: the roles and the users it links, by name.
	 *
	 * An encryption key is handed to nobody, and a workspace's key goes to
	 * every run of its workspace whatever the link rows of an older instance
	 * say, because the scope is the grant.
	 */
	static function says($key) {
		if ($key["encrypts"]) {
			return "never handed out";
		}

		if ($key["workspace"] !== "") {
			return "every run of the workspace";
		}

		$names = array_merge($key["roles"], $key["users"]);

		return count($names) === 0 ? "none" : implode(", ", $names);
	}

	/** dropsBackup is what deleting a key's backup costs, which the confirmation says. */
	static function dropsBackup($key) {
		$lost = $key["encrypts"] ? "Secrets the new key has not re-encrypted stop opening." : "Hosts that know only its public key stop letting runs in.";

		return "Delete the backup " . $key["backup"] . ", " . $key["backup_fingerprint"] . "? " . $lost;
	}

	/** rotates is what rotating a key does, which the confirmation says. */
	static function rotates($key) {
		$kept = $key["encrypts"] ? "still decrypts while task-ui re-encrypts every secret" : "runs and git keep using beside the new key until you delete it: add the new public key to the hosts and remotes first";

		return "Rotate " . $key["name"] . ", " . $key["fingerprint"] . "? task-ui makes a new " . $key["algo"] . " key under its name and keeps this one as " . self::backupName($key["name"]) . ", which " . $kept . ".";
	}
}
