{if $data_pager.prev || $data_pager.next}	<nav class="pager" aria-label="Pages">
{if $data_pager.prev}		<a class="btn btn-sm" href="{base}{data_pager.prev}" rel="prev">Previous</a>
{/if}{if $data_pager.next}		<a class="btn btn-sm" href="{base}{data_pager.next}" rel="next">Next</a>
{/if}	</nav>
{/if}
