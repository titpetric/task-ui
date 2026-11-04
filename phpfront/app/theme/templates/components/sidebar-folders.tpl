{* input scope: data. See sidebar-folders.md. *}
			<nav class="folders">
{foreach $data as $group}
{if $group.workspace}{eval $data_workspace = $group}{include components/sidebar-workspace.tpl}
{else}{eval $data_project = $group}{include components/sidebar-project.tpl}
{/if}
{else}
				<div class="empty">No folders yet. {if $chrome.can['project/edit']}<a href="{base}/settings/projects/new">Add a git repository</a>{else}Add a git repository in Settings{/if}, or put a folder with jobs under the root.</div>
{/foreach}
			</nav>
