<div class="actionsbar">
	<div class="gitline" title="{data_git.title}">
		<span class="branch"><span class="glyph" aria-hidden="true">⎇</span> {if $data_git.branch}{data_git.branch}{else}<span class="faint">detached</span> {data_git.detached}{/if}</span>
{if $data_git.ahead}		<span class="ahead" title="{data_git.ahead} commits to push">↑{data_git.ahead}</span>
{/if}{if $data_git.behind}		<span class="behind" title="{data_git.behind} commits to pull">↓{data_git.behind}</span>
{/if}{if $data_git.dirty}		<button type="button" class="dirty" data-git-status aria-expanded="false" aria-controls="git-status" title="Show the changed files"><span class="glyph" aria-hidden="true">●</span> {data_git.files}<svg class="ic chev" aria-hidden="true"><use href="#i-chev"/></svg></button>
{/if}{if $data_git.clean}		<span class="clean">up to date</span>
{/if}		<span class="where">· {data_git.where}</span>
	</div>
{if $data_git.dirty}	<div class="gitstatus" id="git-status" hidden>
		<div class="table-wrap">
			<table class="table gitstatus-table">
				<thead><tr><th>Status</th><th>File</th></tr></thead>
				<tbody>
{foreach $data_git.changes as $change}					<tr><td class="gitstate"><code>{change.code}</code> <span class="faint">{change.state}</span></td><td class="gitpath">{change.path}{if $change.from} <span class="faint">from {change.from}</span>{/if}</td></tr>
{/foreach}				</tbody>
			</table>
		</div>
	</div>
{/if}</div>
