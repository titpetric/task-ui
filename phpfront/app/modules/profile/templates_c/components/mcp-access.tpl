<?php $_v=&$this->vars;?>
<section class="block mcp-access" id="mcp">
	<h3>MCP access</h3>
<?php if($_v['data_mcp']['on']){?>
	<div class="table-wrap" data-before-reveal>
	<table class="table cards">
		<thead><tr><th>Token</th><th>Created</th><th>Expires</th><th>Last used</th><th></th></tr></thead>
		<tbody>
<?php if(!empty($_v['data_mcp']['tokens']))foreach($_v['data_mcp']['tokens'] as $_v['token']){?>			<tr>
				<td class="key nowrap" data-label="Token"><?php echo htmlspecialchars($_v['token']['name'], ENT_QUOTES);?>
</td>
				<td class="nowrap" data-label="Created" title="<?php echo htmlspecialchars($_v['token']['createdAt'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['token']['created'], ENT_QUOTES);?>
</td>
				<td class="nowrap" data-label="Expires"><?php if($_v['token']['expires']){?>
<span class="badge<?php if($_v['token']['tone']){?>
 <?php echo htmlspecialchars($_v['token']['tone'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['token']['expiresAt'], ENT_QUOTES);?>
"><?php if($_v['token']['expired']){?>
expired <?php }?><?php echo htmlspecialchars($_v['token']['expiresAt'], ENT_QUOTES);?>
</span><?php }else{?>
<span class="faint">never: a token from <code>task-ui token</code></span><?php }?></td>
				<td class="nowrap" data-label="Last used"<?php if($_v['token']['used']){?>
 title="<?php echo htmlspecialchars($_v['token']['usedAt'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['token']['used']){ echo htmlspecialchars($_v['token']['usedAgo'], ENT_QUOTES); }else{?>
<span class="faint">never</span><?php }?></td>
				<td class="actions"><?php if($_v['data_mcp']['manage']){?>
<button type="button" class="btn-icon danger" data-delete="<?php echo htmlspecialchars($_v['token']['delete'], ENT_QUOTES);?>
" data-confirm="Delete the MCP token <?php echo htmlspecialchars($_v['token']['name'], ENT_QUOTES);?>
 of <?php echo htmlspecialchars($_v['data_mcp']['name'], ENT_QUOTES);?>
? Whatever client holds it stops working right away. The other tokens keep working." data-confirm-ok="Delete" data-next="reload" title="Delete the token <?php echo htmlspecialchars($_v['token']['name'], ENT_QUOTES);?>
" aria-label="Delete the token <?php echo htmlspecialchars($_v['token']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-trash"/></svg></button><?php }?></td>
			</tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php }else{?>
	<p data-before-reveal><?php echo htmlspecialchars($_v['data_mcp']['name'], ENT_QUOTES);?>
 holds no MCP token. New token makes one that lasts 30 days and shows it once.</p>
<?php }?><?php if(!$_v['data_mcp']['agent']){?>
	<p class="faint">A token needs MCP agent user, under <a href="#roles">Roles</a>.</p>
<?php }elseif($_v['data_mcp']['manage']){?>
	<div class="actions" data-before-reveal>
		<form class="inline" data-api="<?php echo htmlspecialchars($_v['data_mcp']['api'], ENT_QUOTES);?>
" data-method="POST" data-reveal="#mcp-token"><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-plus"/></svg>New token</button></form>
	</div>
	<p class="faint" data-before-reveal>One token per client, so retiring a client is one Delete and the others keep signing in.</p>
<?php }else{?>
	<p class="faint">Making a token needs Allow Edits on Users.</p>
<?php }?>	<div id="mcp-token" class="mcp-token" hidden>
		<div class="alert">Copy the token now; you won't see it again. <b data-fill="<?php echo htmlspecialchars($_v['data_mcp']['named'], ENT_QUOTES);?>
"></b> works until <b data-fill="<?php echo htmlspecialchars($_v['data_mcp']['until'], ENT_QUOTES);?>
"></b>. Make another here, or delete this one, whenever you like.</div>
		<dl class="copy-list">
			<dt>Token</dt>
			<dd><span class="copyable"><code data-fill="<?php echo htmlspecialchars($_v['data_mcp']['fill'], ENT_QUOTES);?>
"></code><button type="button" class="btn-icon" data-copy="" data-fill="<?php echo htmlspecialchars($_v['data_mcp']['fill'], ENT_QUOTES);?>
" title="Copy the token" aria-label="Copy the token"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>Claude Code</dt>
			<dd><span class="copyable"><code data-fill="<?php echo htmlspecialchars($_v['data_mcp']['add'], ENT_QUOTES);?>
"></code><button type="button" class="btn-icon" data-copy="" data-fill="<?php echo htmlspecialchars($_v['data_mcp']['add'], ENT_QUOTES);?>
" title="Copy the claude mcp add command" aria-label="Copy the claude mcp add command"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>MCP settings</dt>
			<dd><span class="copyable"><code data-fill="<?php echo htmlspecialchars($_v['data_mcp']['settings'], ENT_QUOTES);?>
"></code><button type="button" class="btn-icon" data-copy="" data-fill="<?php echo htmlspecialchars($_v['data_mcp']['settings'], ENT_QUOTES);?>
" title="Copy the MCP settings" aria-label="Copy the MCP settings"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
		</dl>
		<a class="btn" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_mcp']['done'], ENT_QUOTES);?>
">Done</a>
	</div>
</section>
