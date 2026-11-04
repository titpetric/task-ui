<?php $_v=&$this->vars;?>

<section class="block mcp-access" id="mcp">
	<h3>MCP access</h3>
<?php if($_v['data_mcp']['on']){?>
	<dl class="copy-list" data-before-reveal>
		<dt>Created</dt>
		<dd title="<?php echo htmlspecialchars($_v['data_mcp']['createdAt'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_mcp']['created'], ENT_QUOTES);?>
</dd>
		<dt>Expires</dt>
		<dd><?php if($_v['data_mcp']['expires']){?>
<span class="badge<?php if($_v['data_mcp']['tone']){?>
 <?php echo htmlspecialchars($_v['data_mcp']['tone'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_mcp']['expiresAt'], ENT_QUOTES);?>
"><?php if($_v['data_mcp']['expired']){?>
expired <?php }?><?php echo htmlspecialchars($_v['data_mcp']['expiresAt'], ENT_QUOTES);?>
</span><?php }else{?>
never: a token from <code>task-ui token</code><?php }?></dd>
		<dt>Last used</dt>
		<dd<?php if($_v['data_mcp']['used']){?>
 title="<?php echo htmlspecialchars($_v['data_mcp']['usedAt'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['data_mcp']['used']){ echo htmlspecialchars($_v['data_mcp']['usedAgo'], ENT_QUOTES); }else{?>
<span class="faint">never</span><?php }?></dd>
	</dl>
<?php if($_v['data_mcp']['manage']){?>
	<div class="actions" data-before-reveal>
		<form class="inline" data-api="<?php echo htmlspecialchars($_v['data_mcp']['api'], ENT_QUOTES);?>
" data-method="POST" data-reveal="#mcp-token" data-confirm="Rotate the MCP token of <?php echo htmlspecialchars($_v['data_mcp']['name'], ENT_QUOTES);?>
? The current token stops working right away. The new one lasts 30 days." data-confirm-ok="Rotate"><button class="btn" type="submit">Rotate</button></form>
		<button type="button" class="btn btn-danger" data-delete="<?php echo htmlspecialchars($_v['data_mcp']['api'], ENT_QUOTES);?>
" data-confirm="Disable MCP for <?php echo htmlspecialchars($_v['data_mcp']['name'], ENT_QUOTES);?>
? The token stops working right away." data-confirm-ok="Disable" data-next="reload">Disable</button>
	</div>
<?php }?><?php }else{?>
	<p data-before-reveal>MCP is off for <?php echo htmlspecialchars($_v['data_mcp']['name'], ENT_QUOTES);?>
. Enable creates a token that lasts 30 days and shows it once.</p>
<?php if(!$_v['data_mcp']['agent']){?>
	<p class="faint">A token needs MCP agent user, under <a href="#roles">Roles</a>, which an admin turns on.</p>
<?php }elseif($_v['data_mcp']['enable']){?>
	<form class="inline" data-api="<?php echo htmlspecialchars($_v['data_mcp']['api'], ENT_QUOTES);?>
" data-method="POST" data-reveal="#mcp-token" data-before-reveal><button class="btn btn-primary" type="submit">Enable MCP</button></form>
<?php }else{?>
	<p class="faint">Enabling needs Allow Edits on Users. Ask someone who has it.</p>
<?php }?><?php }?>	<div id="mcp-token" class="mcp-token" hidden>
		<div class="alert">Copy the token now; you won't see it again. It works until <b data-fill="<?php echo htmlspecialchars($_v['data_mcp']['until'], ENT_QUOTES);?>
"></b>. Rotate it here before then.</div>
		<dl class="copy-list">
			<dt>Token</dt>
			<dd><span class="copyable"><code data-fill="<?php echo htmlspecialchars($_v['data_mcp']['fill'], ENT_QUOTES);?>
"></code><button type="button" class="btn-icon" data-copy="" data-fill="<?php echo htmlspecialchars($_v['data_mcp']['fill'], ENT_QUOTES);?>
" title="Copy the token" aria-label="Copy the token"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>Claude Code</dt>
			<dd><span class="copyable"><code data-fill="<?php echo htmlspecialchars($_v['data_mcp']['add'], ENT_QUOTES);?>
"></code><button type="button" class="btn-icon" data-copy="" data-fill="<?php echo htmlspecialchars($_v['data_mcp']['add'], ENT_QUOTES);?>
" title="Copy the claude mcp add command" aria-label="Copy the claude mcp add command"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
		</dl>
		<a class="btn" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_mcp']['done'], ENT_QUOTES);?>
">Done</a>
	</div>
</section>
