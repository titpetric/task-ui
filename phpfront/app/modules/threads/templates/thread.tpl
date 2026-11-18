{eval $data_head = $data.head}
{include components/page-head.tpl}

{if $data.pick_agent}<form class="agent-pick" data-agent-form data-thread="{data.id}">
	<label class="faint" for="thread-agent">Agent</label>
	<select id="thread-agent" name="agent">
{foreach $data.agent_names as $name}		<option value="{name}"{if $name == $data.agent} selected{/if}>{name}</option>
{/foreach}	</select>
	<button class="btn" type="submit">Save</button>
</form>{/if}

<ol class="talk" data-notes data-thread="{data.id}"{if $data.chat_thread} data-chat-thread="{data.id}"{/if}>
{foreach $data.messages as $message}	<li class="said{if $message.mine} mine{/if}{if $message.working} working{/if}" id="note-{message.id}">
		<header>
			<span class="author">{eval $data_user = $message.author}{include components/user.tpl}</span>
			<span class="faint" title="{message.stamp}">{message.ago}</span>
{if $message.working}			<span class="badge note-state working">working</span>
{/if}{if $message.run}			<a class="faint mono" href="{base}/runs/{message.run}" title="The run this message comes from">run {message.short}</a>
{/if}		</header>
		<div class="said-body prose">{message.body|raw}</div>
{if $message.work}		<div class="work">
{foreach $message.work as $run}			<div class="work-run" data-run="{run.id}">
{foreach $run.events as $event}{if $event.kind == 'status'}				<div class="work-status">{event.body}</div>
{elseif $event.kind == 'text' || $event.kind == 'result'}				<div class="work-text prose">{event.prose|raw}</div>
{elseif $event.kind == 'thinking'}				<details class="work-thinking"><summary>Thinking</summary><pre>{event.body}</pre></details>
{elseif $event.kind == 'tool_use'}				<details class="work-tool"><summary><code>{event.tool}</code></summary><pre>{event.input}</pre></details>
{elseif $event.kind == 'tool_result'}				<details class="work-result{if $event.error} error{/if}"><summary>{if $event.error}Error{else}Result{/if}</summary><pre>{event.body}</pre></details>
{elseif $event.kind == 'system'}				<details class="work-system"><summary>Session</summary><pre>{event.body}</pre></details>
{elseif $event.kind == 'error'}				<div class="work-error">{event.body}</div>
{/if}
{/foreach}			</div>
{/foreach}		</div>
{/if}{if $message.audio}		<audio class="said-audio" controls preload="none" src="{message.audio}"></audio>
{/if}
	</li>
{/foreach}</ol>

{if $data.replies}<form class="note-form" data-note-form data-parent="{data.id}" data-channel="{data.channel}">
	{eval $data_markdown = $data.editor}{include components/editor.tpl}
	<div class="form-foot note"><span class="faint">Ctrl+Enter posts.</span>{if $data.mic}<button class="btn btn-ghost" type="button" data-voice-reply data-parent="{data.id}" data-channel="{data.channel}" title="Record a reply" aria-label="Record a reply"><svg class="ic"><use href="#i-mic"/></svg></button>{/if}<button class="btn btn-primary" type="submit">Reply</button></div>
</form>{/if}
