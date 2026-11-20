<?php

namespace Task\Threads;

// Conversation is one thread as the pane reads it: its title, the messages,
// the agent picker, the reply box and the actions.
//
// Both the thread's own page and the right-hand pane of the Threads page read
// this one scope, so a swap of the pane and a reload of the page draw the same
// conversation and enforce the same permissions.
class Conversation {
	/** MISSING is what a pane says for a thread that is not there. */
	const MISSING = "No thread with that ID. Someone may have deleted it.";

	/** view is one thread as pane.tpl reads it, or an error the pane shows. */
	static function view($note, $chrome, $store, $split = false) {
		$prompt = \Task\Note::isPrompt($note);

		if ($prompt && (!$chrome->allowed("prompt/view") || !\Task\Note::seesPlace($chrome, $note->getDir()))) {
			return ["error" => \Task\View::REFUSED . \Task\View::GRANTED];
		}

		if (!$prompt && !\Task\Note::isChat($note) && !$chrome->allowed("note/view")) {
			return ["error" => \Task\View::REFUSED . \Task\View::GRANTED];
		}

		$chat = \Task\Note::isChat($note);
		$userID = $chrome->user === null ? "" : $chrome->user->getID();
		$member = $chat && $store->participant($note->getId(), $userID);

		if ($chat && !$member && !\Task\User::admin($chrome->user)) {
			return ["error" => self::MISSING];
		}

		$replies = $store->replies([$note->getId()]);
		$under = isset($replies[$note->getId()]) ? $replies[$note->getId()] : [];
		$title = \Task\Note::excerpt(Talk::opening($note, $under));
		$work = Work::byQuestion($store->events($note->getId()), $chrome);
		$users = new \Task\User;
		$agentNames = [];

		foreach ($users->agents() as $user) {
			$agentNames[] = $user->getName();
		}

		$agent = $note->getAgent();

		if ($agent === "" && $prompt) {
			$agent = count($agentNames) > 0 ? $agentNames[0] : "claude";
		}

		$head = new \Task\Head($title);

		$head->menu(self::menu($note, $chrome)->data());
		$seen = new \Task\Seen($chrome);

		return [
			"id" => $note->getId(),
			"title" => $title,
			"head" => $head->data(),
			"split" => $split,
			"channel" => $note->getChannel(),
			"chat_thread" => $chat,
			"messages" => Talk::messages($note, $under, $chrome, $seen, $work),
			"agent" => $agent,
			"agent_names" => $agentNames,
			"pick_agent" => !$chat && $note->getState() !== \Task\Note::WORKING && ($chrome->name() === $note->getUserName() || \Task\User::admin($chrome->user)),
			"editor" => self::field(),
			"replies" => $chat ? $member && $chrome->allowed("note/edit") : ($prompt ? $chrome->allowed("prompt/edit") : $chrome->allowed("note/edit")),
			"mic" => $chrome->allowed("note/edit"),
		];
	}

	/** menu is the thread's own actions: a thread is deleted, never rewritten. */
	static function menu($note, $chrome) {
		$menu = new \Task\Menu("Actions on this thread");

		$menu->whenDanger(\Task\Note::deletes($note, $chrome), "trash", "Delete", [
			"data-delete" => "/api/notes/" . $note->getId(),
			"data-confirm" => "Delete this thread and every message in it? It leaves every listing; nobody reads it again.",
			"data-confirm-ok" => "Delete",
			"data-next" => \Task\Route::href("/threads"),
		]);

		return $menu;
	}

	/** field is the box under the conversation, as components/editor.tpl reads it. */
	static function field() {
		return [
			"name" => "body",
			"rows" => 2,
			"label" => "A reply",
			"placeholder" => "",
			"mentions" => true,
			"toolbar" => "inline",
		];
	}
}
