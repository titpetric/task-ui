{* input scope: data_mcp. See mcp-access.md. *}
<section class="block mcp-access" id="mcp">
	<h3>MCP access</h3>
{if $data_mcp.on}	<dl class="copy-list" data-before-reveal>
		<dt>Created</dt>
		<dd title="{data_mcp.createdAt}">{data_mcp.created}</dd>
		<dt>Expires</dt>
		<dd>{if $data_mcp.expires}<span class="badge{if $data_mcp.tone} {data_mcp.tone}{/if}" title="{data_mcp.expiresAt}">{if $data_mcp.expired}expired {/if}{data_mcp.expiresAt}</span>{else}never: a token from <code>task-ui token</code>{/if}</dd>
		<dt>Last used</dt>
		<dd{if $data_mcp.used} title="{data_mcp.usedAt}"{/if}>{if $data_mcp.used}{data_mcp.usedAgo}{else}<span class="faint">never</span>{/if}</dd>
	</dl>
{if $data_mcp.manage}	<div class="actions" data-before-reveal>
		<form class="inline" data-api="{data_mcp.api}" data-method="POST" data-reveal="#mcp-token" data-confirm="Rotate the MCP token of {data_mcp.name}? The current token stops working right away. The new one lasts 30 days." data-confirm-ok="Rotate"><button class="btn" type="submit">Rotate</button></form>
		<button type="button" class="btn btn-danger" data-delete="{data_mcp.api}" data-confirm="Disable MCP for {data_mcp.name}? The token stops working right away." data-confirm-ok="Disable" data-next="reload">Disable</button>
	</div>
{/if}{else}	<p data-before-reveal>MCP is off for {data_mcp.name}. Enable creates a token that lasts 30 days and shows it once.</p>
{if !$data_mcp.agent}	<p class="faint">A token needs MCP agent user, under <a href="#roles">Roles</a>, which an admin turns on.</p>
{elseif $data_mcp.enable}	<form class="inline" data-api="{data_mcp.api}" data-method="POST" data-reveal="#mcp-token" data-before-reveal><button class="btn btn-primary" type="submit">Enable MCP</button></form>
{else}	<p class="faint">Enabling needs Allow Edits on Users. Ask someone who has it.</p>
{/if}{/if}	<div id="mcp-token" class="mcp-token" hidden>
		<div class="alert">Copy the token now; you won't see it again. It works until <b data-fill="{data_mcp.until}"></b>. Rotate it here before then.</div>
		<dl class="copy-list">
			<dt>Token</dt>
			<dd><span class="copyable"><code data-fill="{data_mcp.fill}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{data_mcp.fill}" title="Copy the token" aria-label="Copy the token"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>Claude Code</dt>
			<dd><span class="copyable"><code data-fill="{data_mcp.add}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{data_mcp.add}" title="Copy the claude mcp add command" aria-label="Copy the claude mcp add command"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
		</dl>
		<a class="btn" href="{base}{data_mcp.done}">Done</a>
	</div>
</section>
