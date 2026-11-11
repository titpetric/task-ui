<nav class="setup-trail" aria-label="Setup steps">
{if $data_trail.previous}	<a class="btn btn-sm" href="{data_trail.previous.href}" rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: {data_trail.previous.name}</a>
{/if}	<ol class="steps">
{foreach $data_trail.steps as $step}		<li>{if $step.current}<b aria-current="step">{step.name}</b>{else}<a href="{step.href}">{step.name}</a>{/if}</li>
{/foreach}	</ol>
{if $data_trail.next}	<a class="btn btn-sm" href="{data_trail.next.href}" rel="next">Next: {data_trail.next.name}<svg class="ic"><use href="#i-chev"/></svg></a>
{/if}</nav>
