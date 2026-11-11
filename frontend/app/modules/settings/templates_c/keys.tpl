<?php $_v=&$this->vars; $_v['data_head'] = $_v['data']['head'];?>

<div class="page-head">
	<div>
		<h2><?php if($_v['data_head']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['data_head']['icon'], ENT_QUOTES);?>
"/></svg> <?php }?><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES); if($_v['data_head']['badge']){?>
 <span class="badge<?php if($_v['data_head']['tone']){?>
 <?php echo htmlspecialchars($_v['data_head']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_head']['badge'], ENT_QUOTES);?>
</span><?php }?><?php if(!empty($_v['data_head']['tags']))foreach($_v['data_head']['tags'] as $_v['tag']){?> <span class="badge"><?php echo htmlspecialchars($_v['tag'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_head']['note']){?>
 <span class="term-title" title="<?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
</span><?php }?></h2>
<?php if($_v['data_head']['lead'] || $_v['data_head']['leads']){?>
		<p><?php echo $_v['data_head']['lead'];?>
</p>
<?php }?><?php if($_v['data_head']['faint']){?>
		<p class="faint"><?php echo htmlspecialchars($_v['data_head']['faint'], ENT_QUOTES);?>
</p>
<?php }?>	</div>
<?php if($_v['data_head']['show']){?>
	<div class="actions"<?php if($_v['data_head']['live']){?>
 data-actions<?php }?>>
<?php if($_v['data_head']['run']){?>
		<?php $_v['data_run'] = $_v['data_head']['run'];?>
<div class="run-split">
	<button type="button" class="btn btn-primary btn-sm" data-run data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
" title="Run with arguments after <?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?>><svg class="ic"><use href="#i-play"/></svg>Run<?php if($_v['data_run']['takesArgs']){?>
...<?php }?></button>
<?php if($_v['data_run']['menu']){?>
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
" aria-label="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-down"/></svg></summary>
		<div class="more-list">
			<div class="more-items" role="menu" aria-label="Run options">
<?php if($_v['data_run']['summarize']){?>
				<button type="button" class="more-item" role="menuitem" data-run data-summarize data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?> title="Run and summarize"><svg class="ic"><use href="#i-summary"/></svg><span>Run and summarize</span></button>
<?php }?><?php if($_v['data_run']['force']){?>
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['place'], ENT_QUOTES);?>
" title="Run out of hours, with a reason"><svg class="ic"><use href="#i-play"/></svg><span>Run force</span></button>
<?php }?>			</div>
		</div>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data_head']['buttons']['items']){?>
		<?php $_v['data_buttons'] = $_v['data_head']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }}?><?php if($_v['data_head']['actions']){?>
		<?php $_v['data_menu'] = $_v['data_head']['actions'];?>

		<?php if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-more"/></svg></summary>
	<div class="more-list">
<?php if($_v['data_menu']['details']){?>
		<dl class="more-details">
<?php if(!empty($_v['data_menu']['details']))foreach($_v['data_menu']['details'] as $_v['detail']){?>			<dt><?php echo htmlspecialchars($_v['detail']['label'], ENT_QUOTES);?>
</dt><dd<?php if($_v['detail']['mono']){?>
 class="mono"<?php }?><?php if($_v['detail']['title']){?>
 title="<?php echo htmlspecialchars($_v['detail']['title'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['detail']['value'], ENT_QUOTES);?>
</dd>
<?php }?>		</dl>
<?php }?>		<div class="more-items" role="menu" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_menu']['groups']))foreach($_v['data_menu']['groups'] as $_v['index'] => $_v['group']){?><?php if($_v['index']){?>
			<hr class="more-sep">
<?php }?><?php if(!empty($_v['group']))foreach($_v['group'] as $_v['item']){?><?php if($_v['item']['href']){?>
			<a class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem" href="<?php echo htmlspecialchars($_v['item']['href'], ENT_QUOTES);?>
"<?php if($_v['item']['download']){?>
 download<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }}?>	</div>
<?php }?></div>


<section class="block key-overview">
	<h3>Scopes <small>the keys of the instance and each workspace</small></h3>
	<div class="table-wrap">
	<table class="table key-purposes">
		<thead><tr><th>Scope</th><th>Variables</th><th>Encryption key</th><th>SSH keys</th></tr></thead>
		<tbody>
			<tr class="key-instance" id="instance">
				<td><b>Instance</b><div class="faint">every run, by policy</div></td>
				<td><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/integrations">integrations</a><div class="faint">their settings</div></td>
				<td>
<?php if($_v['data']['scopes']['instance']['set']){?>
					<a class="badge ok" href="<?php echo htmlspecialchars($_v['data']['scopes']['instance']['anchor'], ENT_QUOTES);?>
" title="Its [...] in Every key rotates it">set</a><?php if($_v['data']['scopes']['instance']['backup']){?>
 <span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: <?php echo htmlspecialchars($_v['data']['scopes']['instance']['backup'], ENT_QUOTES);?>
</span><?php }?>
					<div class="key-fp faint" title="<?php echo htmlspecialchars($_v['data']['scopes']['instance']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['scopes']['instance']['fingerprint'], ENT_QUOTES);?>
</div>
<?php }else{?>
					<span class="badge bad">missing</span> <span class="faint">task-ui creates it on start</span>
<?php }?>				</td>
				<td>
<?php if($_v['data']['scopes']['instance']['count']){?>
					<a href="#instance-ssh"><?php echo htmlspecialchars($_v['data']['scopes']['instance']['count'], ENT_QUOTES);?>
 key<?php if($_v['data']['scopes']['instance']['count'] !== 1){?>
s<?php }?></a>
					<div class="key-names"><?php if(!empty($_v['data']['scopes']['instance']['chips']))foreach($_v['data']['scopes']['instance']['chips'] as $_v['chip']){?><a class="key-chip" href="<?php echo htmlspecialchars($_v['chip']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['chip']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['chip']['name'], ENT_QUOTES);?>
</a><?php }?></div>
<?php }else{?>
					<span class="faint">none: runs get no SSH key from the instance</span>
<?php }?>				</td>
			</tr>
<?php if(!empty($_v['data']['scopes']['workspaces']))foreach($_v['data']['scopes']['workspaces'] as $_v['ws']){?>			<tr id="ws-<?php echo htmlspecialchars($_v['ws']['folder'], ENT_QUOTES);?>
">
				<td>
					<a href="<?php echo htmlspecialchars($_v['ws']['href'], ENT_QUOTES);?>
"><b><?php echo htmlspecialchars($_v['ws']['title'], ENT_QUOTES);?>
</b></a><?php if($_v['ws']['badge']){?>
 <span class="badge"><?php echo htmlspecialchars($_v['ws']['badge'], ENT_QUOTES);?>
</span><?php }?>
					<div class="faint key-projects"><?php if($_v['ws']['projects']){ echo htmlspecialchars($_v['ws']['projects'], ENT_QUOTES); }else{?>
no projects<?php }?></div>
				</td>
				<td><?php if($_v['ws']['secrets']){?>
<a href="<?php echo htmlspecialchars($_v['ws']['secretsHref'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['ws']['secrets'], ENT_QUOTES);?>
 variable<?php if($_v['ws']['secrets'] !== 1){?>
s<?php }?></a><?php }else{?>
<a class="faint" href="<?php echo htmlspecialchars($_v['ws']['secretsHref'], ENT_QUOTES);?>
">none</a><?php }?></td>
				<td>
					<span class="faint">the instance's</span>
				</td>
				<td>
<?php if($_v['ws']['ssh']){?>
					<a href="<?php echo htmlspecialchars($_v['ws']['keysHref'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['ws']['count'], ENT_QUOTES);?>
 key<?php if($_v['ws']['count'] !== 1){?>
s<?php }?></a>
					<div class="key-names"><?php if(!empty($_v['ws']['ssh']))foreach($_v['ws']['ssh'] as $_v['chip']){?><a class="key-chip" href="<?php echo htmlspecialchars($_v['chip']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['chip']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['chip']['name'], ENT_QUOTES);?>
</a><?php }?></div>
<?php }else{?>
					<a class="faint" href="<?php echo htmlspecialchars($_v['ws']['keysHref'], ENT_QUOTES);?>
">none</a>
					<div class="faint">Its runs get only the instance SSH keys their policies grant.</div>
<?php }?>				</td>
			</tr>
<?php }?>		</tbody>
	</table>
	</div>
</section>

<section class="block" id="instance-ssh">
<?php $_v['data_block'] = $_v['data']['block'];?>
<div class="block-head<?php if($_v['data_block']['empty']){?>
 is-empty<?php }?>">
	<h3><?php echo htmlspecialchars($_v['data_block']['title'], ENT_QUOTES); if($_v['data_block']['note']){?>
 <small><?php echo htmlspecialchars($_v['data_block']['note'], ENT_QUOTES);?>
</small><?php }?></h3>
<?php if($_v['data_block']['empty']){?>
	<p class="empty-line"><?php echo htmlspecialchars($_v['data_block']['empty'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_block']['buttons']['items']){?>
	<div class="actions"><?php $_v['data_buttons'] = $_v['data_block']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }?>
	</div>
<?php }?></div>

<?php if($_v['data']['keys']){?>
	<p class="dim">Every key lives in the key folder as <code>&lt;name&gt;.key</code> and <code>&lt;name&gt;.pub</code>, the name ending in its algorithm and size. An instance SSH key goes to the runs of the roles and users it links, through their ssh agent, and the inventory signs in to hosts with it; a workspace's SSH key goes to the runs of that workspace. The git key, which a key's page sets, pulls and clones the checkouts. An encryption key never leaves task-ui.</p>
<?php }?><?php $_v['data_keys'] = $_v['data']['keys']; if($_v['data_keys']){?>
	<div class="table-wrap fit">
	<table class="table keys key-table cards">
		<colgroup><col class="c-name"><col class="c-algo"><col class="c-size"><col class="c-pub"><col class="c-actions"></colgroup>
		<thead><tr><th>Name</th><th class="hide-narrow">Algorithm</th><th class="hide-narrow">Size</th><th>Public key</th><th class="actions" aria-label="Actions"></th></tr></thead>

		<tbody>
<?php if(!empty($_v['data_keys']))foreach($_v['data_keys'] as $_v['row']){?><?php $_v['data_key'] = $_v['row'];?>
			<tr id="key-<?php echo htmlspecialchars($_v['data_key']['name'], ENT_QUOTES);?>
">
				<td class="key"><?php if($_v['data_key']['href']){?>
<a class="mono" href="<?php echo htmlspecialchars($_v['data_key']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_key']['name'], ENT_QUOTES);?>
</a><?php }else{?>
<span class="mono"><?php echo htmlspecialchars($_v['data_key']['name'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_key']['encrypts']){?>
 <span class="badge signal">encryption</span><?php }?><?php if($_v['data_key']['backup']){?>
<span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: <?php echo htmlspecialchars($_v['data_key']['backup'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_key']['comment']){?>
<div class="faint"><?php echo htmlspecialchars($_v['data_key']['comment'], ENT_QUOTES);?>
</div><?php }?></td>
				<td class="mono hide-narrow" data-label="Algorithm"><?php echo htmlspecialchars($_v['data_key']['algo'], ENT_QUOTES);?>
</td>
				<td class="mono nowrap hide-narrow" data-label="Size"><?php echo htmlspecialchars($_v['data_key']['size'], ENT_QUOTES);?>
 bits</td>
				<td class="pubkey" data-label="Public key"><?php $_v['data_copy'] = $_v['data_key']['copy'];?>
<span class="copyable<?php if($_v['data_copy']['fill']){?>
 fill<?php }?>"><code<?php if($_v['data_copy']['fill'] || $_v['data_copy']['shown'] !== $_v['data_copy']['value']){?>
 title="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_copy']['shown'], ENT_QUOTES);?>
</code><button type="button" class="btn-icon" data-copy="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
" title="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
" aria-label="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-copy"/></svg></button></span>
</td>
				<td class="actions"><?php $_v['data_menu'] = $_v['data_key']['menu']; if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-more"/></svg></summary>
	<div class="more-list">
<?php if($_v['data_menu']['details']){?>
		<dl class="more-details">
<?php if(!empty($_v['data_menu']['details']))foreach($_v['data_menu']['details'] as $_v['detail']){?>			<dt><?php echo htmlspecialchars($_v['detail']['label'], ENT_QUOTES);?>
</dt><dd<?php if($_v['detail']['mono']){?>
 class="mono"<?php }?><?php if($_v['detail']['title']){?>
 title="<?php echo htmlspecialchars($_v['detail']['title'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['detail']['value'], ENT_QUOTES);?>
</dd>
<?php }?>		</dl>
<?php }?>		<div class="more-items" role="menu" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_menu']['groups']))foreach($_v['data_menu']['groups'] as $_v['index'] => $_v['group']){?><?php if($_v['index']){?>
			<hr class="more-sep">
<?php }?><?php if(!empty($_v['group']))foreach($_v['group'] as $_v['item']){?><?php if($_v['item']['href']){?>
			<a class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem" href="<?php echo htmlspecialchars($_v['item']['href'], ENT_QUOTES);?>
"<?php if($_v['item']['download']){?>
 download<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }?>
</td>
			</tr>

<?php }?>		</tbody>
	</table>
	</div>
<?php }?>
</section>

<section class="block">
	<h3>Who gets which key <small>through the ssh agent of their runs</small></h3>
	<div class="table-wrap">
	<table class="table key-holders">
		<thead><tr><th>Policy or user</th><th>Keys</th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['holders']))foreach($_v['data']['holders'] as $_v['holder']){?>			<tr>
<?php if($_v['holder']['user']){?>
				<td class="nowrap"><?php $_v['data_user'] = $_v['holder']['user'];?>
<span class="user-ref"><?php if($_v['data_user']['href']){?>
<a class="user-link" href="<?php echo htmlspecialchars($_v['data_user']['href'], ENT_QUOTES);?>
"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span></a><?php }else{ $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php }?></span>
 <span class="faint"><?php echo htmlspecialchars($_v['holder']['role'], ENT_QUOTES);?>
</span></td>
<?php }else{?>
				<td class="nowrap"><span class="badge"><?php echo htmlspecialchars($_v['holder']['role'], ENT_QUOTES);?>
</span> <span class="faint">every <?php echo htmlspecialchars($_v['holder']['role'], ENT_QUOTES);?>
</span></td>
<?php }?>				<td><?php if(!empty($_v['holder']['direct']))foreach($_v['holder']['direct'] as $_v['chip']){?><a class="key-chip" href="<?php echo htmlspecialchars($_v['chip']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['chip']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['chip']['name'], ENT_QUOTES);?>
</a><?php }?><?php if(!empty($_v['holder']['viaRole']))foreach($_v['holder']['viaRole'] as $_v['chip']){?><a class="key-chip faint" href="<?php echo htmlspecialchars($_v['chip']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['chip']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['chip']['name'], ENT_QUOTES);?>
 <small>via role</small></a><?php }?><?php if(!$_v['holder']['direct'] && !$_v['holder']['viaRole']){?>
<span class="faint">none</span><?php }?>
				</td>
			</tr>
<?php }?>		</tbody>
	</table>
	</div>
</section>
<?php if(!empty($_v['data']['folders']))foreach($_v['data']['folders'] as $_v['folder']){?><div class="alert signal"><b><?php echo htmlspecialchars($_v['folder']['project'], ENT_QUOTES);?>
</b> has private keys in <b><?php echo htmlspecialchars($_v['folder']['dir'], ENT_QUOTES);?>
</b>: <?php echo htmlspecialchars($_v['folder']['keys'], ENT_QUOTES);?>
. Its runs don't get them. Upload a key the project needs as an instance key, then link it to roles or users.</div>
<?php }?><?php if($_v['data']['trail']){ $_v['data_trail'] = $_v['data']['trail'];?>
<nav class="setup-trail" aria-label="Setup steps">
<?php if($_v['data_trail']['previous']){?>
	<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data_trail']['previous']['href'], ENT_QUOTES);?>
" rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: <?php echo htmlspecialchars($_v['data_trail']['previous']['name'], ENT_QUOTES);?>
</a>
<?php }?>	<ol class="steps">
<?php if(!empty($_v['data_trail']['steps']))foreach($_v['data_trail']['steps'] as $_v['step']){?>		<li><?php if($_v['step']['current']){?>
<b aria-current="step"><?php echo htmlspecialchars($_v['step']['name'], ENT_QUOTES);?>
</b><?php }else{?>
<a href="<?php echo htmlspecialchars($_v['step']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['step']['name'], ENT_QUOTES);?>
</a><?php }?></li>
<?php }?>	</ol>
<?php if($_v['data_trail']['next']){?>
	<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data_trail']['next']['href'], ENT_QUOTES);?>
" rel="next">Next: <?php echo htmlspecialchars($_v['data_trail']['next']['name'], ENT_QUOTES);?>
<svg class="ic"><use href="#i-chev"/></svg></a>
<?php }?></nav>
<?php }?>
