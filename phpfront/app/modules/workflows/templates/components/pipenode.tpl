{* input scope: data_node. See pipenode.md. *}<li class="node{if $data_node.state} {data_node.state}{/if}{if $data_node.disabled} disabled{/if}" data-step="{data_node.pos}">
	<span class="rail"><i class="dot{if $data_node.state} {data_node.state}{/if}"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="{base}{data_node.href}"{if $data_node.runId} data-run-link="{data_node.runId}"{/if}>{if $data_node.workflow}<svg class="ic lead"><use href="#i-flow"/></svg>{elseif $data_node.summary}<svg class="ic lead"><use href="#i-summary"/></svg>{elseif $data_node.check}<svg class="ic lead"><use href="#i-rotate"/></svg>{elseif $data_node.prompt}<svg class="ic lead"><use href="#i-prompt"/></svg>{/if}{data_node.text}</a>
{if $data_node.expand}			<button type="button" class="btn btn-sm" data-expand="{data_node.runId}" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand{if $data_node.lines} {data_node.lines} lines{/if}</span></button>
{/if}		</div>
		<small>{data_node.small}</small>
{if $data_node.prompt}		<p class="step-prompt">{data_node.prompt}</p>
{/if}{if $data_node.runId}		<a class="frame" href="{base}/runs/{data_node.runId}" data-attach="{data_node.runId}"{if $data_node.live} data-frame="{data_node.runId}"{/if} role="img" aria-label="Terminal of {data_node.text}">{data_node.frame|raw}</a>
{/if}{if $data_node.subId}		<ol class="pipeline sub" data-sub="{data_node.subId}" data-workflow="{data_node.workflowId}">
{foreach $data_node.sub as $data_sub}{include components/pipesub.tpl}{/foreach}		</ol>
{/if}	</div>
</li>
