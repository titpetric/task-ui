<?php

namespace Task\Mcp;

// Connect is how a client is told to connect with a token it has just been
// given: the command for Claude Code, the same connection as a settings file,
// and the placeholders task-ui.js replaces in either of them.
//
// A placeholder is a constant here and not markup of a template, because
// MiniTPL compiles any {word} into an echo: a template cannot write the
// literal, so it arrives as a field of the scope. The same strings are read by
// the user edit page, which reveals them, and by the one-click buttons of the
// profile and the users list, which copy them.
class Connect {
	/** TOKEN, UNTIL and NAMED are the fields of the answer that made the token. */
	const TOKEN = "{token}";

	const UNTIL = "{expires_at}";

	const NAMED = "{name}";

	/** ADD is the line that adds this instance to Claude Code, with TOKEN in it. */
	const ADD = 'claude mcp add --transport http --scope user task %s/api/mcp --header "Authorization: Bearer {token}"';

	/** SETTINGS is the same connection as a settings file, for a client that takes one. */
	const SETTINGS = '{"mcpServers":{"task":{"type":"http","url":"%s/api/mcp","headers":{"Authorization":"Bearer {token}"}}}}';

	/** add is ADD against a site, which is the instance's own base link. */
	static function add($site) {
		return sprintf(self::ADD, $site);
	}

	/** settings is SETTINGS against a site. */
	static function settings($site) {
		return sprintf(self::SETTINGS, $site);
	}
}
