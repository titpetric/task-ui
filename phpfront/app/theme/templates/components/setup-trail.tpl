<nav class="setup-trail" aria-label="Setup steps">
{if $data_trail.previous}	<a class="btn btn-sm" href="{data_trail.previous.href}" rel="prev"><i class="ic ic-chev back" aria-hidden="true"></i>Previous: {data_trail.previous.name}</a>
{/if}	<ol class="steps">
{foreach $data_trail.steps as $step}		<li>{if $step.current}<b aria-current="step">{step.name}</b>{else}<a href="{step.href}">{step.name}</a>{/if}</li>
{/foreach}	</ol>
{if $data_trail.next}	<a class="btn btn-sm" href="{data_trail.next.href}" rel="next">Next: {data_trail.next.name}<i class="ic ic-chev" aria-hidden="true"></i></a>
{/if}</nav>
