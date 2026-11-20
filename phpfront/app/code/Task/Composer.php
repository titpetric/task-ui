<?php

namespace Task;

// Composer is the one box that sends a message: a question to Claude, a note
// on the board or a chat message. The page says where it posts and which
// fields it has, and components/composer.tpl draws it.
class Composer {
	/** ASK and NOTE are the two ways the Threads box starts a thread. */
	const ASK = "ask";

	const NOTE = "note";

	/** reply is the box under a conversation, in the mode its thread takes. */
	static function reply($pane) {
		$mode = self::NOTE;

		if ($pane["chat_thread"]) {
			$mode = "chat";
		} elseif (!empty($pane["prompt"])) {
			$mode = self::ASK;
		}

		return self::scope($mode, $pane["editor"], [
			"id" => $pane["id"],
			"parent" => $pane["id"],
			"channel" => $pane["channel"],
			"mic" => $pane["mic"],
			"pane" => $pane["split"],
			"next" => Route::href("/threads/" . Route::pathEscape($pane["id"])),
		]);
	}

	/** blank is the Threads box that starts a thread, asking by default. */
	static function blank($chrome, $agent) {
		$agents = Prompt::agents($chrome);

		return self::scope(self::ASK, Prompt::field("what should we work on?"), [
			"modes" => [
				["key" => self::ASK, "label" => "Ask Claude"],
				["key" => self::NOTE, "label" => "Note"],
			],
			"agents" => $agents["names"],
			"agent" => $agent === "" ? $agents["current"] : $agent,
			"folder" => Route::INTERNAL,
			"show_place" => true,
			"origin" => "threads",
		]);
	}

	/** prompt is a page's prompt box, from the scope Task\Prompt builds. */
	static function prompt($panel) {
		$place = [
			"folder" => "",
			"dir" => "",
			"run" => "",
			"area" => "",
		];

		foreach ($panel["fields"] as $field) {
			if (isset($place[$field["name"]])) {
				$place[$field["name"]] = $field["value"];
			}
		}

		return self::scope(self::ASK, $panel["editor"], [
			"id" => "{id}",
			"agents" => $panel["agents"],
			"agent" => $panel["agent"],
			"folder" => $place["folder"],
			"dir" => $place["dir"],
			"run" => $place["run"],
			"area" => $place["area"],
			"url" => $panel["api"],
			"next" => $panel["next"],
		]);
	}

	/** scope is the composer's fields, with the caller's own laid over them. */
	static function scope($mode, $editor, $over = []) {
		return array_merge([
			"mode" => $mode,
			"modes" => [],
			"editor" => $editor,
			"agents" => [],
			"agent" => "",
			"id" => "",
			"parent" => "",
			"channel" => "note",
			"folder" => "",
			"dir" => "",
			"run" => "",
			"origin" => "",
			"area" => "",
			"place" => "",
			"show_place" => false,
			"mic" => false,
			"pane" => false,
			"url" => Prompt::API,
			"next" => "",
		], $over);
	}
}
