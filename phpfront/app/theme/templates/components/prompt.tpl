<details class="promptbox"{if $data_prompt.open} open{/if}{if $data_prompt.hidden} hidden{/if}>
	<summary><h3 class="section-title"><i class="ic ic-chev chev" aria-hidden="true"></i>{data_prompt.title} <small>{data_prompt.note}</small></h3></summary>
{if !$data_prompt.watch}	{eval $data_composer = $data_prompt.composer}{include components/composer.tpl}
{/if}{if $data_prompt.threads}	<ul class="thread-links" data-notes>
{foreach $data_prompt.threads as $thread}		<li><a href="{thread.href}"><i class="ic ic-note" aria-hidden="true"></i> {thread.title}</a>{if $thread.state} <span class="badge note-state {thread.state}">{thread.state}</span>{/if}</li>
{/foreach}	</ul>
{/if}</details>
<p class="faint threads-scope"><a class="threads-link" href="{data_prompt.threads_href}">Threads in this place</a></p>
