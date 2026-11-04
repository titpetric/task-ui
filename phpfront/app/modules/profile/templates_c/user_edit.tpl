<?php $_v=&$this->vars;$_v['data_head'] = $_v['data']['head'];?>


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
		<?php $_v['data_buttons'] = $_v['data_head']['buttons'];if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
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


<section class="block" id="roles">
	<h3>Roles <small>what <?php echo htmlspecialchars($_v['data']['name'], ENT_QUOTES);?>
 is</small></h3>
<?php if($_v['data']['roles']['edits']){?>
	<form class="form" data-api="<?php echo htmlspecialchars($_v['data']['api'], ENT_QUOTES);?>
" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['data']['roles']['next'], ENT_QUOTES);?>
">
		<div class="field">
			<label class="check"><input type="checkbox" name="admin"<?php if($_v['data']['roles']['admin']){?>
 checked<?php }?>> Administrative privileges (all)</label>
			<small class="faint">Every permission, for any role; the last admin keeps it.</small>
		</div>
		<div class="field">
			<label for="u-role">Role <span class="faint">without the admin flag</span></label>
			<select id="u-role" name="role">
<?php if(!empty($_v['data']['roles']['options']))foreach($_v['data']['roles']['options'] as $_v['option']){?>				<option value="<?php echo htmlspecialchars($_v['option'], ENT_QUOTES);?>
"<?php if($_v['data']['roles']['role'] === $_v['option']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['option'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="mcp"<?php if($_v['data']['roles']['mcp']){?>
 checked<?php }?>> MCP agent user</label>
			<small class="faint">A token reaches the MCP endpoint; an agent learns of work from the notes that name the user.</small>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="agent_reply"<?php if($_v['data']['roles']['reply']){?>
 checked<?php }?>> Run instance Claude when mentioned</label>
			<small class="faint">A note naming <?php echo htmlspecialchars($_v['data']['name'], ENT_QUOTES);?>
 starts the Notes agent as them; needs MCP agent user.</small>
		</div>
		<div class="form-foot">
			<a class="btn" href="<?php echo htmlspecialchars($_v['data']['view'], ENT_QUOTES);?>
">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
<?php }else{?>
	<p><?php if($_v['data']['roles']['admin']){?>
Administrative privileges (all). <?php }?>Role: <?php echo htmlspecialchars($_v['data']['roles']['role'], ENT_QUOTES);?>
.<?php if($_v['data']['roles']['mcp']){?>
 MCP agent user.<?php }?><?php if($_v['data']['roles']['reply']){?>
 Runs instance Claude when mentioned.<?php }?></p>
<?php }?></section>

<section class="block" id="folders">
	<h3>Folders <small>what <?php echo htmlspecialchars($_v['data']['name'], ENT_QUOTES);?>
 can see</small></h3>
<?php if($_v['data']['folders']['edits']){?>
	<form class="form" data-api="<?php echo htmlspecialchars($_v['data']['api'], ENT_QUOTES);?>
" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['data']['folders']['next'], ENT_QUOTES);?>
">
<?php $_v['data_pick'] = $_v['data']['folders']['pick'];?>
	<div class="field" data-folder-pick<?php if($_v['data_pick']['label']){?>
 role="group" aria-labelledby="folders-label"<?php }?>>
<?php if($_v['data_pick']['label']){?>
		<span class="label" id="folders-label">Folders</span>
<?php }?>		<label class="check"><input type="radio" name="scope" value="all"<?php if($_v['data_pick']['all']){?>
 checked<?php }?>> All folders</label>
		<label class="check"><input type="radio" name="scope" value="custom"<?php if(!$_v['data_pick']['all']){?>
 checked<?php }?>> Custom</label>
		<div class="folder-choices">
<?php if(!empty($_v['data_pick']['choices']))foreach($_v['data_pick']['choices'] as $_v['choice']){?>			<label class="check"><input type="checkbox" name="folders" value="<?php echo htmlspecialchars($_v['choice']['id'], ENT_QUOTES);?>
" data-type="set"<?php if($_v['choice']['checked']){?>
 checked<?php }?><?php if($_v['data_pick']['all']){?>
 disabled<?php }?>> <span class="mono"><?php echo htmlspecialchars($_v['choice']['id'], ENT_QUOTES);?>
</span></label>
<?php }?><?php if(!$_v['data_pick']['choices']){?>
			<span class="faint">No folders yet.</span>
<?php }?>		</div>
	</div>
		<div class="form-foot">
			<a class="btn" href="<?php echo htmlspecialchars($_v['data']['view'], ENT_QUOTES);?>
">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
<?php }elseif($_v['data']['folders']['list']){?>
	<p><?php if(!empty($_v['data']['folders']['list']))foreach($_v['data']['folders']['list'] as $_v['index'] => $_v['folder']){?><?php if($_v['index']){?>
, <?php }?><span class="mono"><?php echo htmlspecialchars($_v['folder'], ENT_QUOTES);?>
</span><?php }?></p>
<?php }else{?>
	<p>All folders.</p>
<?php }?></section>

<section class="block" id="permissions">
	<h3>Permissions <small><?php echo htmlspecialchars($_v['data']['name'], ENT_QUOTES);?>
's own rules, before the role's</small></h3>
	<p>Role: <?php if($_v['data']['permissions']['links']){?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/permissions#role-<?php echo htmlspecialchars($_v['data']['permissions']['role'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['permissions']['role'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['data']['permissions']['role'], ENT_QUOTES); }?>.<?php if($_v['data']['permissions']['matrix']['fixed']){?>
 Admins can do everything; this can't change.<?php }?></p>
<?php if(!$_v['data']['permissions']['matrix']['fixed']){?>
	<p>An allow or a deny here is <?php echo htmlspecialchars($_v['data']['name'], ENT_QUOTES);?>
's own and decides the rule for them. Inherit, the empty box, leaves it to the role's rules, where a deny denies the action even where another rule allows it. The mark after each box shows the result.</p>
<?php }?><?php $_v['data_matrix'] = $_v['data']['permissions']['matrix'];?>

<form class="matrix-form<?php if($_v['data_matrix']['fixed']){?>
 fixed<?php }?>"<?php if($_v['data_matrix']['editable']){?>
 data-api="<?php echo htmlspecialchars($_v['data_matrix']['api'], ENT_QUOTES);?>
" data-method="PUT" data-next="reload" data-notice="Permissions saved" data-matrix<?php }?>>
	<div class="table-wrap">
	<table class="table matrix<?php if($_v['data_matrix']['effective']){?>
 effective<?php }?>">
		<thead><tr><th>Section</th><?php if(!empty($_v['data_matrix']['rules']))foreach($_v['data_matrix']['rules'] as $_v['rule']){?><th><?php echo htmlspecialchars($_v['rule']['title'], ENT_QUOTES);?>
</th><?php }?></tr></thead>
		<tbody>
<?php if(!empty($_v['data_matrix']['sections']))foreach($_v['data_matrix']['sections'] as $_v['section']){?>			<tr><th scope="row"><?php echo htmlspecialchars($_v['section']['title'], ENT_QUOTES);?>
</th><?php if(!empty($_v['section']['cells']))foreach($_v['section']['cells'] as $_v['cell']){?><td><?php $_v['data_cell'] = $_v['cell'];?>
<span class="tri-cell"><button type="button" class="tri" data-tri="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cell']['title'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
"<?php if($_v['data_cell']['locked']){?>
 disabled<?php }?>><?php echo htmlspecialchars($_v['data_cell']['mark'], ENT_QUOTES);?>
</button><?php if($_v['data_cell']['field']){?>
<input type="hidden" name="<?php echo htmlspecialchars($_v['data_cell']['path'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-was="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" disabled><?php }?><?php if($_v['data_cell']['shown']){?>
<span class="eff <?php echo htmlspecialchars($_v['data_cell']['tone'], ENT_QUOTES);?>
" title="Result: <?php echo htmlspecialchars($_v['data_cell']['says'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_cell']['result'], ENT_QUOTES);?>
</span><?php }?></span>
</td><?php }?></tr>
<?php }?>			<tr class="group"><th colspan="<?php echo htmlspecialchars($_v['data_matrix']['columns'], ENT_QUOTES);?>
">Application</th></tr>
<?php if(!empty($_v['data_matrix']['permissions']))foreach($_v['data_matrix']['permissions'] as $_v['cell']){?>			<tr><th scope="row"><?php echo htmlspecialchars($_v['cell']['title'], ENT_QUOTES);?>
</th><td colspan="<?php echo htmlspecialchars($_v['data_matrix']['span'], ENT_QUOTES);?>
"><?php $_v['data_cell'] = $_v['cell'];?>
<span class="tri-cell"><button type="button" class="tri" data-tri="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cell']['title'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
"<?php if($_v['data_cell']['locked']){?>
 disabled<?php }?>><?php echo htmlspecialchars($_v['data_cell']['mark'], ENT_QUOTES);?>
</button><?php if($_v['data_cell']['field']){?>
<input type="hidden" name="<?php echo htmlspecialchars($_v['data_cell']['path'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-was="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" disabled><?php }?><?php if($_v['data_cell']['shown']){?>
<span class="eff <?php echo htmlspecialchars($_v['data_cell']['tone'], ENT_QUOTES);?>
" title="Result: <?php echo htmlspecialchars($_v['data_cell']['says'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_cell']['result'], ENT_QUOTES);?>
</span><?php }?></span>
</td></tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php if($_v['data_matrix']['editable']){?>
	<div class="form-foot">
		<span class="faint" data-matrix-changed>No changes yet.</span>
		<button class="btn btn-primary" type="submit">Save permissions</button>
	</div>
<?php }?></form>

</section>
<?php if($_v['data']['keys']){ $_v['data_user_keys'] = $_v['data']['keys'];?>

<section class="block" id="keys">
<?php $_v['data_block'] = $_v['data_user_keys']['block'];?>

<div class="block-head<?php if($_v['data_block']['empty']){?>
 is-empty<?php }?>">
	<h3><?php echo htmlspecialchars($_v['data_block']['title'], ENT_QUOTES); if($_v['data_block']['note']){?>
 <small><?php echo htmlspecialchars($_v['data_block']['note'], ENT_QUOTES);?>
</small><?php }?></h3>
<?php if($_v['data_block']['empty']){?>
	<p class="empty-line"><?php echo htmlspecialchars($_v['data_block']['empty'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_block']['buttons']['items']){?>
	<div class="actions"><?php $_v['data_buttons'] = $_v['data_block']['buttons'];if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
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

<?php $_v['data_keys'] = $_v['data_user_keys']['keys']; if($_v['data_keys']){?>
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
				<td class="actions"><?php $_v['data_menu'] = $_v['data_key']['menu'];if(!$_v['data_menu']['empty']){?>
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
<?php } $_v['data_mcp'] = $_v['data']['mcp'];?>

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

