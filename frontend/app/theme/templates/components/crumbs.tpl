{if $data.crumbs}
			<div class="crumbs"><a href="{data.base}/">Dashboard</a>{foreach $data.crumbs as $crumb}<span>{if $crumb.href}<a href="{data.base}{crumb.href}">{crumb.text}</a>{else}{crumb.text}{/if}</span>{/foreach}</div>
{/if}
