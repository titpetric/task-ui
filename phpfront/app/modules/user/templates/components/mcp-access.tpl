<section class="block mcp-access" id="mcp">
	<h3>MCP access</h3>
{if $data_mcp.on}	<div class="table-wrap" data-before-reveal>
	<table class="table cards">
		<thead><tr><th>Token</th><th>Created</th><th>Expires</th><th>Last used</th><th></th></tr></thead>
		<tbody>
{foreach $data_mcp.tokens as $token}			<tr>
				<td class="key nowrap" data-label="Token">{token.name}</td>
				<td class="nowrap" data-label="Created" title="{token.createdAt}">{token.created}</td>
				<td class="nowrap" data-label="Expires">{if $token.expires}<span class="badge{if $token.tone} {token.tone}{/if}" title="{token.expiresAt}">{if $token.expired}expired {/if}{token.expiresAt}</span>{else}<span class="faint">never: a token from <code>task-ui token</code></span>{/if}</td>
				<td class="nowrap" data-label="Last used"{if $token.used} title="{token.usedAt}"{/if}>{if $token.used}{token.usedAgo}{else}<span class="faint">never</span>{/if}</td>
				<td class="actions">{if $data_mcp.manage}<button type="button" class="btn-icon danger" data-delete="{token.delete}" data-confirm="Delete the MCP token {token.name} of {data_mcp.name}? Whatever client holds it stops working right away. The other tokens keep working." data-confirm-ok="Delete" data-next="reload" title="Delete the token {token.name}" aria-label="Delete the token {token.name}"><svg class="ic"><use href="#i-trash"/></svg></button>{/if}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{else}	<p data-before-reveal>{data_mcp.name} holds no MCP token. New token makes one that lasts 30 days and shows it once.</p>
{/if}{if !$data_mcp.agent}	<p class="faint">A token needs MCP agent user, under <a href="#roles">Roles</a>.</p>
{elseif $data_mcp.manage}	<div class="actions" data-before-reveal>
		<form class="inline" data-api="{data_mcp.api}" data-method="POST" data-reveal="#mcp-token"><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-plus"/></svg>New token</button></form>
	</div>
	<p class="faint" data-before-reveal>One token per client, so retiring a client is one Delete and the others keep signing in.</p>
{else}	<p class="faint">Making a token needs Allow Edits on Users.</p>
{/if}	<div id="mcp-token" class="mcp-token" hidden>
		<div class="alert">Copy the token now; you won't see it again. <b data-fill="{data_mcp.named}"></b> works until <b data-fill="{data_mcp.until}"></b>. Make another here, or delete this one, whenever you like.</div>
		<dl class="copy-list">
			<dt>Token</dt>
			<dd><span class="copyable"><code data-fill="{data_mcp.fill}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{data_mcp.fill}" title="Copy the token" aria-label="Copy the token"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>Claude Code</dt>
			<dd><span class="copyable"><code data-fill="{data_mcp.add}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{data_mcp.add}" title="Copy the claude mcp add command" aria-label="Copy the claude mcp add command"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>MCP settings</dt>
			<dd><span class="copyable"><code data-fill="{data_mcp.settings}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{data_mcp.settings}" title="Copy the MCP settings" aria-label="Copy the MCP settings"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
		</dl>
		<a class="btn" href="{base}{data_mcp.done}">Done</a>
	</div>
</section>
