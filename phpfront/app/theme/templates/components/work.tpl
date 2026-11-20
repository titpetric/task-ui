<div class="work-run" data-run="{data_work.id}">
{if $data_work.stats}	<div class="work-stats">{data_work.stats}</div>
{/if}{foreach $data_work.events as $event}{if $event.kind == 'text' || $event.kind == 'result'}	<div class="work-text prose" data-seq="{event.seq}">{event.prose|raw}</div>
{elseif $event.kind == 'thinking'}	<details class="work-thinking" data-seq="{event.seq}"><summary>Thinking</summary><pre>{event.body}</pre></details>
{elseif $event.kind == 'tool_use'}	<details class="work-tool" data-seq="{event.seq}"><summary><code>{event.tool}</code>{if $event.body} <span class="faint">{event.body}</span>{/if}</summary><pre>{event.input}</pre></details>
{elseif $event.kind == 'tool_result'}	<details class="work-result{if $event.error} error{/if}" data-seq="{event.seq}"><summary>{if $event.error}Error{else}Result{/if}</summary><pre>{event.body}</pre></details>
{elseif $event.kind == 'system'}	<details class="work-system" data-seq="{event.seq}"><summary>Session</summary><pre>{event.body}</pre></details>
{elseif $event.kind == 'error'}	<div class="work-error" data-seq="{event.seq}">{event.body}</div>
{/if}{/foreach}</div>
