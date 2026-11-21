		<aside class="rail" aria-label="Recents">
			<div class="rail-inner">
				<div class="label side-label"><a href="{base}{data.href}">Recents</a></div>
				<nav class="recents" id="recents">
{foreach $data.groups as $group}
					<div class="recent-group">{group.label}</div>
{foreach $group.entries as $entry}
					<a href="{base}{entry.href}" class="recent{if $entry.flow} flow{/if}" title="{entry.title}">
						<i class="dot {entry.state}"></i>
						<span class="what"><b>{entry.name}</b><small>{if $entry.icon}<i class="ic ic-{entry.icon} lead" aria-hidden="true"></i>{/if}{entry.place}</small></span>
{if $entry.children}						<span class="badge" title="{if $entry.thread}Replies{else}Runs it started{/if}">{entry.children}</span>{/if}
{if !$entry.ended}						<i class="ic ic-{entry.act} act" aria-hidden="true"></i>{/if}
						{eval $data_avatar = $entry.avatar}{include components/avatar.tpl}
					</a>
{/foreach}
{/foreach}
{if !$data.groups}
					<div class="empty">Runs started by anyone show up here.</div>
{/if}
					<a href="{base}{data.href}" class="recent more" data-recents-more hidden><i class="ic ic-list" aria-hidden="true"></i><span class="what"><b>All runs</b><small></small></span></a>
				</nav>
			</div>
		</aside>
