{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="orb-room" data-orb-room
	data-meeting="{data.room.id}" data-state="{data.room.state}"
	data-self="{data.room.self}" data-self-name="{data.room.self_name}"
	data-host="{if $data.room.host}1{else}0{/if}"
	data-can-edit="{if $data.room.can_edit}1{else}0{/if}"
	data-can-delete="{if $data.room.can_delete}1{else}0{/if}"
	data-started="{data.room.started}" data-limit="{data.room.limit}"
	data-api="{base}/api" data-base="{base}">
	<div class="orb-people" data-people>
{foreach $data.people as $person}		<span class="orb-person{if $person.self} self{/if}{if $person.here} here{else} gone{/if}" data-person="{person.id}" data-here="{if $person.here}1{else}0{/if}">
			<span class="orb-dot"></span><span class="orb-name">{person.name}</span>{if $person.joins > 1} <span class="orb-joins" title="Times joined">{person.joins}</span>{/if}
		</span>
{/foreach}	</div>
{if $data.room.open}	<div class="orb-controls">
		<div class="orb-action">
{if $data.room.can_edit}			<button class="btn btn-primary btn-lg" type="button" data-orb-join><i class="ic ic-users" aria-hidden="true"></i>Join meeting</button>
			<div class="orb-speak" data-orb-speak hidden>
				<button class="orb-mic" type="button" data-orb-mic aria-label="Start speaking" aria-pressed="false" title="Start speaking"><i class="ic ic-mic" aria-hidden="true"></i></button>
				<span class="orb-miclabel faint" data-orb-miclabel>Microphone off</span>
			</div>
{else}			<span class="faint">You can read this room but not speak in it.</span>
{/if}		</div>
		<span class="orb-timer" data-orb-timer title="Time left"></span>
	</div>
{if $data.room.pick_agent}	<form class="orb-agent" data-orb-agent>
		<label class="faint" for="orb-agent">Agent</label>
		<select id="orb-agent" name="agent">
			<option value="">No agent</option>
{foreach $data.room.agent_names as $name}			<option value="{name}"{if $name == $data.room.agent} selected{/if}>{name}</option>
{/foreach}		</select>
		<button class="btn" type="submit">Save</button>
	</form>{/if}
	<p class="orb-status faint" data-orb-status aria-live="polite">{if $data.room.can_edit}Join to hear the room and speak.{else}You need <code>meeting/edit</code> to join.{/if}</p>
	<p class="faint orb-hint">Headphones keep the room's own voices out of your microphone.</p>
{else}	<p class="orb-status faint" data-orb-status>This meeting ended {data.room.ended}.</p>
{/if}

{if !$data.room.open}<section class="block orb-summary" data-orb-summary data-state="{data.room.summary_state}">
	<h3>Summary{if $data.room.summary_state == "working"} <small class="faint">Claude is reading the transcript...</small>{/if}</h3>
	<div class="prose" data-summary-body>{data.room.summary|raw}</div>
	<p class="faint" data-summary-message>{data.room.summary_message}</p>
{if $data.room.can_edit}	<button class="btn" type="button" data-summary-retry{if $data.room.summary_state != "failed"} hidden{/if}>Ask Claude again</button>
{/if}</section>
{/if}

<details class="block orb-log-block"{if $data.room.open} open{/if}>
	<summary>Transcript{if !$data.room.open}, with joins and parts{/if}</summary>
	<div class="orb-log" data-orb-log aria-live="polite">
{foreach $data.lines as $line}{if $line.said}		<div class="orb-line" data-line-id="{line.id}">
			<span class="orb-who">{line.name}</span>
			<span class="orb-said">{line.body}</span>
			<span class="orb-at" data-at="{line.created}" title="{line.at}">{line.at}</span>
			<button class="btn-icon" type="button" data-play data-user="{line.user}" data-start="{line.start}" data-end="{line.end}" title="Play this line"><i class="ic ic-play" aria-hidden="true"></i></button>
		</div>
{else}		<div class="orb-line system" data-line-id="{line.id}"><span>{line.body}</span><span class="orb-at" data-at="{line.created}">{line.at}</span></div>
{/if}{/foreach}		<p class="faint orb-empty"{if $data.lines} hidden{/if} data-orb-empty>Nothing said yet.</p>
	</div>
</details>
</section>
