{foreach $data_workflows as $flow}
{if !$flow.hidden}					<div class="flow{if $flow.active} active{/if}">
						<i class="shebang{if $flow.running} running{/if}" aria-hidden="true">#!</i>
						<a class="name" href="{base}/workflows/{flow.id}" title="{flow.title}">{flow.name}</a>
{if $flow.running}						<a class="play" href="{base}/workflows/{flow.id}" title="{flow.name}, running" aria-label="Watch {flow.name}"><svg class="ic"><use href="#i-eye"/></svg></a>
{elseif $flow.canRun}						<button type="button" class="play" data-workflow-run="{flow.id}"{if $flow.prompted} data-workflow-prompt="{flow.name}"{/if} title="Run {flow.name}" aria-label="Run {flow.name}"><svg class="ic"><use href="#i-play"/></svg></button>
{/if}
					</div>
{/if}
{/foreach}
