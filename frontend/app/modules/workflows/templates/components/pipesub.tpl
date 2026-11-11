<li class="node{if $data_sub.state} {data_sub.state}{/if}{if $data_sub.disabled} disabled{/if}" data-step="{data_sub.pos}">
	<span class="rail"><i class="dot{if $data_sub.state} {data_sub.state}{/if}"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="{base}{data_sub.href}"{if $data_sub.runId} data-run-link="{data_sub.runId}"{/if}>{if $data_sub.workflow}<svg class="ic lead"><use href="#i-flow"/></svg>{elseif $data_sub.summary}<svg class="ic lead"><use href="#i-summary"/></svg>{elseif $data_sub.check}<svg class="ic lead"><use href="#i-rotate"/></svg>{elseif $data_sub.prompt}<svg class="ic lead"><use href="#i-prompt"/></svg>{/if}{data_sub.text}</a>
{if $data_sub.expand}			<button type="button" class="btn btn-sm" data-expand="{data_sub.runId}" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand{if $data_sub.lines} {data_sub.lines} lines{/if}</span></button>
{/if}		</div>
		<small>{data_sub.small}</small>
{if $data_sub.prompt}		<p class="step-prompt">{data_sub.prompt}</p>
{/if}{if $data_sub.runId}		<a class="frame" href="{base}/runs/{data_sub.runId}" data-attach="{data_sub.runId}"{if $data_sub.live} data-frame="{data_sub.runId}"{/if} role="img" aria-label="Terminal of {data_sub.text}">{data_sub.frame|raw}</a>
{/if}{if $data_sub.subId}		<ol class="pipeline sub" data-sub="{data_sub.subId}" data-workflow="{data_sub.workflowId}">
		</ol>
{/if}	</div>
</li>
