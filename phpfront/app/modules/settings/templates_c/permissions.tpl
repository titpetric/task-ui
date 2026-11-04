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


<div class="alert">
	Admins can do everything; you can't change that. <b>Allow All</b> sets every rule in its row, and a rule you set on its own overrides it. Folder access and each job's allow list still limit what a user can see and run.
</div>
<?php if(!empty($_v['data']['matrices']))foreach($_v['data']['matrices'] as $_v['matrix']){?><section class="block matrix-role" id="<?php echo htmlspecialchars($_v['matrix']['id'], ENT_QUOTES);?>
">
	<h3><span class="avatar sm role-<?php echo htmlspecialchars($_v['matrix']['title'], ENT_QUOTES);?>
" aria-hidden="true"><?php echo htmlspecialchars($_v['matrix']['initials'], ENT_QUOTES);?>
</span> <?php echo htmlspecialchars($_v['matrix']['title'], ENT_QUOTES);?>
 <?php if($_v['matrix']['fixed']){?>
<small>can do everything</small><?php }elseif(!$_v['matrix']['editable']){?>
<small>read only</small><?php }?></h3>
	<p class="role-members">
		<span class="label">Users</span><?php if(!empty($_v['matrix']['users']))foreach($_v['matrix']['users'] as $_v['user']){?> <a class="key-chip" href="<?php echo htmlspecialchars($_v['user']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['user']['name'], ENT_QUOTES);?>
</a><?php }?><?php if(!$_v['matrix']['users']){?>
 <span class="faint">none</span><?php }?><?php if($_v['data']['addUser']){?>
 <a href="<?php echo htmlspecialchars($_v['data']['addUser'], ENT_QUOTES);?>
">Add user</a><?php }?>
	</p><?php if($_v['data']['viewsKeys']){?>

	<p class="role-members">
		<span class="label">Keys</span><?php if(!empty($_v['matrix']['keys']))foreach($_v['matrix']['keys'] as $_v['key']){?> <a class="key-chip" href="<?php echo htmlspecialchars($_v['key']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['key']['name'], ENT_QUOTES);?>
</a><?php }?><?php if(!$_v['matrix']['keys']){?>
 <span class="faint">none</span><?php }?><?php if($_v['data']['generate']){?>
 <a href="<?php echo htmlspecialchars($_v['data']['generate'], ENT_QUOTES);?>
">Generate key</a><?php }?>
	</p><?php }?><?php $_v['data_matrix'] = $_v['matrix']['matrix'];?>

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
