		<aside class="rail" aria-label="Recents">
			<div class="rail-inner">
				<div class="label side-label"><a href="{base}{data.href}">Recents</a></div>
				<nav class="recents" id="recents">
{foreach $data.groups as $group}
					<div class="recent-group">{group.label}</div>
{foreach $group.entries as $entry}
					<a href="{base}{entry.href}" class="recent{if $entry.flow} flow{/if}" title="{entry.title}">
						<i class="dot {entry.state}"></i>
						<span class="what"><b>{entry.name}</b><small>{if $entry.flow}<svg class="ic lead"><use href="#i-flow"/></svg>{/if}{entry.place}</small></span>
{if !$entry.ended}						<svg class="ic act"><use href="#i-{entry.act}"/></svg>{/if}
						{eval $data_avatar = $entry.avatar}{include components/avatar.tpl}
					</a>
{/foreach}
{/foreach}
{if !$data.groups}
					<div class="empty">Runs started by anyone show up here.</div>
{/if}
					<a href="{base}{data.href}" class="recent more" data-recents-more hidden><svg class="ic"><use href="#i-list"/></svg><span class="what"><b>All runs</b><small></small></span></a>
				</nav>
			</div>
		</aside>
