{if $chrome.can['note/view']}<details class="notes-bell" id="notes-bell" data-notes-bell>
	<summary class="btn-icon bell-icon" aria-label="Chats" aria-haspopup="menu" title="Chats"><svg class="ic"><use href="#i-bell"/></svg><span class="bell-count" id="notes-bell-count" hidden></span></summary>
	<div class="notes-bell-menu" role="menu">
		<p class="notes-bell-empty faint">No chat yet. Pick someone in the orb to start one.</p>
		<div class="notes-bell-items" id="notes-bell-items"></div>
	</div>
</details>
{/if}
