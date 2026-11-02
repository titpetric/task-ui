{* input scope: data. See sidebar-folders.md. *}
			<nav class="folders">
{foreach $data as $group}
				<div class="group{if $group.workspace} workspace{/if}">
					<a class="name" href="{base}{group.href}"><svg class="ic"><use href="#i-{group.icon}"/></svg><span>{group.title}</span>{if $group.dirty}<i class="dirty" title="{group.dirty} changed files">{group.dirty}</i>{/if}</a>
{foreach $group.jobs as $job}
					<a class="job" href="{base}{job.href}"><i class="shebang" aria-hidden="true">#!</i><span>{job.name}</span></a>
{/foreach}
				</div>
{else}{if $chrome.can['project/edit']}				<div class="empty">No folders yet. <a href="{base}/settings/projects/new">Add a git repository</a>, or put a folder with jobs under the root.</div>
{else}				<div class="empty">No folders yet. Add a git repository in Settings, or put a folder with jobs under the root.</div>
{/if}{/foreach}
			</nav>
