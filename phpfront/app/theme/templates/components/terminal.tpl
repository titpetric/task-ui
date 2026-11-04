{* input scope: data_terminal. See terminal.md. *}
<section class="block" data-live>
	<h3>Terminal <small>as it looks now, {if $data_terminal.watching}{data_terminal.watching} watching{else}nobody attached{/if}</small></h3>
	<a class="tile wide" href="#" data-attach="{data_terminal.id}"><div class="frame" data-frame="{data_terminal.id}" role="img" aria-label="Terminal of {data_terminal.name}">{data_terminal.frame|raw}</div></a>
</section>
