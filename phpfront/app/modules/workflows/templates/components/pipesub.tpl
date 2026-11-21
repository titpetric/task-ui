<li class="node{if $data_sub.state} {data_sub.state}{/if}{if $data_sub.disabled} disabled{/if}" data-step="{data_sub.pos}">
	<span class="rail"><i class="dot{if $data_sub.state} {data_sub.state}{/if}"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="{base}{data_sub.href}"{if $data_sub.runId} data-run-link="{data_sub.runId}"{/if}>{if $data_sub.workflow}<i class="ic ic-flow lead" aria-hidden="true"></i>{elseif $data_sub.summary}<i class="ic ic-summary lead" aria-hidden="true"></i>{elseif $data_sub.check}<i class="ic ic-rotate lead" aria-hidden="true"></i>{elseif $data_sub.prompt}<i class="ic ic-prompt lead" aria-hidden="true"></i>{/if}{data_sub.text}</a>
{if $data_sub.expand}			<button type="button" class="btn btn-sm" data-expand="{data_sub.runId}" aria-expanded="false"><i class="ic ic-expand" aria-hidden="true"></i><span>Expand{if $data_sub.lines} {data_sub.lines} lines{/if}</span></button>
{/if}		</div>
		<small>{data_sub.small}</small>
{if $data_sub.prompt}		<p class="step-prompt">{data_sub.prompt}</p>
{/if}{if $data_sub.runId}		<a class="frame" href="{base}/runs/{data_sub.runId}" data-attach="{data_sub.runId}"{if $data_sub.live} data-frame="{data_sub.runId}"{/if} role="img" aria-label="Terminal of {data_sub.text}">{data_sub.frame|raw}</a>
{/if}{if $data_sub.subId}		<ol class="pipeline sub" data-sub="{data_sub.subId}" data-workflow="{data_sub.workflowId}">
		</ol>
{/if}	</div>
</li>
