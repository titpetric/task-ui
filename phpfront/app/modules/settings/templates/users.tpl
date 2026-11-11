{eval $data_head = $data.head}
{include components/page-head.tpl}

<div class="alert">
	Users come from the password file; <b>Add user</b> or <code>task-ui passwd &lt;user&gt;</code> adds one or resets a password. A new user starts as a viewer. An admin sets the admin flag, the role and MCP on the user's edit page, or with <code>task-ui user</code>.
</div>

<section class="block">
	<h3>Users</h3>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>User</th><th>Role</th><th>Flags</th><th>Folders</th><th>Since</th><th></th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr class="row-link" data-href="{row.href}">
				<td class="key nowrap"><span class="user-ref">{eval $data_avatar = $row.avatar}{include components/avatar.tpl} <a class="user-name" href="{row.href}">{row.name}</a></span></td>
				<td data-label="Role"><a class="badge{if $row.admin} signal{/if}" href="{row.roleHref}" title="What {row.role} may do">{row.role}</a></td>
				<td class="nowrap" data-label="Flags">{if $row.mcp}<span class="badge" title="MCP agent user">mcp</span>{/if}{if $row.reply} <span class="badge" title="Runs instance Claude when mentioned">agent</span>{/if}{if !$row.mcp}<span class="faint">none</span>{/if}</td>
				<td class="desc" data-label="Folders">{if $row.folders}{row.folders}{else}<span class="faint">all</span>{/if}</td>
				<td class="nowrap" data-label="Since" title="{row.at}">{row.ago}</td>
				<td class="actions">{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
			</tr>
{/foreach}{if !$data.rows}			<tr class="empty"><td colspan="6">No users yet. Add one with <code>task-ui passwd &lt;user&gt;</code> and it appears here.</td></tr>
{/if}		</tbody>
	</table>
	</div>
</section>

<section class="block mcp-connect" id="mcp">
	<h3>Connect an MCP client</h3>
	<p>task-ui is an MCP server. An agent such as Claude Code can list jobs and workflows, run them and read the results, with the permissions of the user whose token it uses.</p>
	<dl class="copy-list">
		<dt>Base link</dt>
		<dd>{eval $data_copy = $data.mcp.base}{include components/copyable.tpl}</dd>
		<dt>MCP endpoint</dt>
		<dd>{eval $data_copy = $data.mcp.endpoint}{include components/copyable.tpl}</dd>
	</dl>
	<ol class="mcp-steps">
		<li>{if $data.mcp.name}Make a token under <a href="{data.mcp.edit}">your MCP access</a>{else}Make a token under your MCP access{/if} and copy it. It shows once and lasts 30 days, or until you delete it; make one per client, so retiring a client takes nothing else with it. A token needs MCP agent user on, which an admin sets, and making one needs Allow Edits on Users; if you don't have them, ask someone who does.</li>
		<li>Add the server to Claude Code:
			{eval $data_copy = $data.mcp.add}{include components/copyable.tpl}</li>
		<li>Other clients connect to the endpoint over HTTP with the header <code>Authorization: Bearer &lt;token&gt;</code>. Check with <code>claude mcp list</code>: it shows <code>task</code> as connected.</li>
	</ol>
	<p class="faint"><a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/mcp.md">The MCP chapter of the guide</a></p>
</section>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
