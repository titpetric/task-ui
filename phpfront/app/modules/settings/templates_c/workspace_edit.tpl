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

<?php $_v['set'] = $_v['data']['settings'];?>

<section class="block">
	<h3>Settings <small class="mono-break"><?php echo htmlspecialchars($_v['set']['dir'], ENT_QUOTES);?>
</small></h3>
	<form class="form" data-api="<?php echo htmlspecialchars($_v['set']['action'], ENT_QUOTES);?>
" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['set']['next'], ENT_QUOTES);?>
">
		<div class="field">
			<label for="we-title">Title</label>
			<input type="text" id="we-title" name="title" value="<?php echo htmlspecialchars($_v['set']['title'], ENT_QUOTES);?>
" autocomplete="off">
		</div>
		<div class="field">
			<label for="we-description">Description</label>
			<input type="text" id="we-description" name="description" value="<?php echo htmlspecialchars($_v['set']['description'], ENT_QUOTES);?>
" autocomplete="off">
		</div>
<?php $_v['data_recursive'] = $_v['set']['recursive'];?>
	<div class="field">
		<label class="check"><input type="checkbox" id="<?php echo htmlspecialchars($_v['data_recursive']['id'], ENT_QUOTES);?>
" name="recursive"<?php if($_v['data_recursive']['checked']){?>
 checked<?php }?>> Recursive</label>
		<small class="faint">Lists every folder inside; off lists only the folder and its top-level folders.</small>
	</div>

		<div class="field">
			<label for="we-sidebar">Sidebar</label>
			<input type="number" id="we-sidebar" name="sidebar_limit" data-type="int" min="0" step="1" placeholder="all" value="<?php echo htmlspecialchars($_v['set']['limit'], ENT_QUOTES);?>
" autocomplete="off">
			<small class="faint">How many projects the sidebar lists for this workspace, the most recently committed first. Empty or 0 lists them all, by name.</small>
		</div>
		<div class="field">
			<label for="we-hours">Hours</label>
			<textarea id="we-hours" name="hours" rows="3" spellcheck="false" autocomplete="off" placeholder="mon-fri 9-17&#10;fri 9-13"><?php echo htmlspecialchars($_v['set']['hours'], ENT_QUOTES);?>
</textarea>
		</div>
		<div class="form-foot">
			<a class="btn" href="<?php echo htmlspecialchars($_v['set']['cancel'], ENT_QUOTES);?>
">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
</section>
<?php $_v['home'] = $_v['data']['home'];?>

<section class="block workspace-home" id="home">
<?php $_v['data_block'] = $_v['home']['block'];?>
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
<?php if(!$_v['home']['empty']){?>
	<p><a class="mono-break" href="<?php echo htmlspecialchars($_v['home']['href'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-folder-open"/></svg> <?php echo htmlspecialchars($_v['home']['folder'], ENT_QUOTES);?>
</a> <span class="faint mono-break"><?php echo htmlspecialchars($_v['home']['url'], ENT_QUOTES);?>
</span><?php if($_v['home']['branch']){?>
 <span class="badge"><?php echo htmlspecialchars($_v['home']['branch'], ENT_QUOTES);?>
</span><?php }?></p>
<?php $_v['git'] = $_v['home']['checkout']; if($_v['git']['found']){?>
	<p class="dim"><?php echo htmlspecialchars($_v['git']['branch'], ENT_QUOTES);?>
 <span class="faint"><?php echo htmlspecialchars($_v['git']['commit'], ENT_QUOTES);?>
</span> <?php echo htmlspecialchars($_v['git']['subject'], ENT_QUOTES); if($_v['git']['ago']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['git']['ago'], ENT_QUOTES);?>
</span><?php }?>; pulled before every run.</p>
<?php }else{?>
	<p class="dim">The home holds no checkout: the clone failed, or someone replaced the directory by hand. Replace clones it again.</p>
<?php }?><?php }?>	<p class="dim">Every run in <?php echo htmlspecialchars($_v['home']['workspace'], ENT_QUOTES);?>
 gets <code>HOME</code> here: <code>bash -l</code> reads its <code>.bash_profile</code>, atkins its <code>.atkins/skills</code> and <code>.atkins/atkins.db</code>, Claude its <code>CLAUDE.md</code> and <code>.claude</code> skills. Registry logins, Claude Code's settings and the build caches stay in the server's home.</p>
</section>
<?php if($_v['home']['edit']){?>

<dialog class="confirm" id="home-replace-dialog" aria-labelledby="home-replace-title">
	<form class="form" data-api="<?php echo htmlspecialchars($_v['home']['action'], ENT_QUOTES);?>
" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['home']['next'], ENT_QUOTES);?>
">
		<h3 id="home-replace-title"><?php echo htmlspecialchars($_v['home']['clone'], ENT_QUOTES);?>
 the home repository</h3>
		<p>Removes <span class="mono-break"><?php echo htmlspecialchars($_v['home']['folder'], ENT_QUOTES);?>
/</span> and clones the repository in its place. task-ui refuses while a run of <?php echo htmlspecialchars($_v['home']['workspace'], ENT_QUOTES);?>
 is active. If the clone fails, you see git's output and the home stays an empty directory.</p>
		<div class="field">
			<label for="hr-url">Repository URL</label>
			<input type="text" id="hr-url" name="url" value="<?php echo htmlspecialchars($_v['home']['url'], ENT_QUOTES);?>
" placeholder="git@github.com:org/workspace.git" spellcheck="false" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="hr-branch">Branch <span class="faint">optional</span></label>
			<input type="text" id="hr-branch" name="branch" value="<?php echo htmlspecialchars($_v['home']['branch'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
		</div>
		<div class="form-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Clone</button>
		</div>
	</form>
</dialog>
<?php }?>
<section class="block" id="projects">
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
<?php $_v['data_projects'] = $_v['data']['table']; if($_v['data_projects']['rows']){?>
<div class="table-wrap fit">
<table class="table projects cards">
	<thead><tr><th>Folder</th><th>Sources</th><th>Origin</th><th class="hide-sm hide-narrow">Last commit</th><?php if($_v['data_projects']['sized']){?>
<th class="hide-sm hide-narrow">Size</th><?php }?><th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
<?php if(!empty($_v['data_projects']['rows']))foreach($_v['data_projects']['rows'] as $_v['row']){?>		<tr>
			<td class="key nowrap"><a href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-folder"/></svg> <?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
</a></td>
			<td class="nowrap" data-label="Sources"><?php if($_v['row']['bad']){?>
<span class="badge bad">error</span> <?php }?><?php if($_v['row']['sources']){ echo htmlspecialchars($_v['row']['sources'], ENT_QUOTES);?>
 <span class="faint"><?php echo htmlspecialchars($_v['row']['jobs'], ENT_QUOTES);?>
 jobs</span><?php }else{?>
<span class="faint">nothing to run</span><?php }?></td>
			<td class="origin" data-label="Origin"><?php if($_v['row']['checkout']){?>
<span class="remote mono" title="<?php echo htmlspecialchars($_v['row']['remote'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['remote'], ENT_QUOTES);?>
</span> <span class="badge"><?php echo htmlspecialchars($_v['row']['branch'], ENT_QUOTES);?>
</span><?php }else{?>
<span class="faint">local folder</span><?php }?></td>
			<td class="nowrap hide-sm hide-narrow" data-label="Last commit"><?php if($_v['row']['checkout']){?>
<span class="mono" title="<?php echo htmlspecialchars($_v['row']['subject'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['commit'], ENT_QUOTES);?>
</span><?php if($_v['row']['ago']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</span><?php }?><?php }?></td><?php if($_v['data_projects']['sized']){?>
<td class="hide-sm hide-narrow" data-label="Size"><?php $_v['data_mini'] = $_v['row']['space'];?>
<span class="spacemini" title="<?php echo htmlspecialchars($_v['data_mini']['title'], ENT_QUOTES);?>
"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_mini']['filesPercent'], ENT_QUOTES);?>
%"></i><i class="history" style="width: <?php echo htmlspecialchars($_v['data_mini']['historyPercent'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_mini']['total'], ENT_QUOTES);?>
</span></span></td><?php }?>
			<td class="actions"><?php $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
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
<?php }?>	</tbody>
</table>
</div>
<?php }?></section>
<?php if($_v['data']['keys']){ $_v['keys'] = $_v['data']['keys'];?>

<section class="block workspace-keys" id="keys">
<?php $_v['data_block'] = $_v['keys']['block'];?>
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
<?php if($_v['keys']['rows']){?>
	<p class="dim">A key goes to every run in the workspace, through the run's ssh agent.</p>
<?php $_v['data_keys'] = $_v['keys']['rows']; if($_v['data_keys']){?>
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
<?php } }?></section>
<?php if($_v['data']['dialogs']){ $_v['data_wsdialogs'] = $_v['data']['dialogs'];?>
<dialog class="confirm" id="key-generate-dialog" aria-labelledby="key-generate-title">
	<?php $_v['data_wsgen'] = $_v['data_wsdialogs']['generate'];?>
<form class="form" data-api="<?php echo htmlspecialchars($_v['data_wsgen']['api'], ENT_QUOTES);?>
" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data_wsgen']['next'], ENT_QUOTES);?>
">
<?php if($_v['data_wsgen']['dialog']){?>
	<h3 id="key-generate-title">Generate key</h3>
	<p>Generates an SSH key of <?php echo htmlspecialchars($_v['data_wsgen']['folder'], ENT_QUOTES);?>
 in <?php echo htmlspecialchars($_v['data_wsgen']['dir'], ENT_QUOTES);?>
, named <span class="mono">workspace_<?php echo htmlspecialchars($_v['data_wsgen']['folder'], ENT_QUOTES);?>
_&lt;name&gt;_&lt;algorithm&gt;</span>.</p>
<?php }?>	<div class="field">
		<label for="wkg-name">Name</label>
		<input type="text" id="wkg-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wkg-algo">Algorithm</label>
		<select id="wkg-algo" name="algo" data-key-algo>
			<option value="ed25519" selected>ed25519</option>
			<option value="ecdsa" data-bits="256">ECDSA P-256</option>
			<option value="ecdsa" data-bits="384">ECDSA P-384</option>
			<option value="ecdsa" data-bits="521">ECDSA P-521</option>
			<option value="rsa" data-bits="2048">RSA 2048</option>
			<option value="rsa" data-bits="3072">RSA 3072</option>
			<option value="rsa" data-bits="4096">RSA 4096</option>
		</select>
		<input type="hidden" name="bits" value="0" data-type="int">
	</div>
	<div class="field">
		<label for="wkg-comment">Comment</label>
		<input type="text" id="wkg-comment" name="comment" value="<?php echo htmlspecialchars($_v['data_wsgen']['comment'], ENT_QUOTES);?>
">
	</div>
<?php $_v['data_wsfoot'] = $_v['data_wsgen']['foot'];?>
	<div class="form-foot">
<?php if($_v['data_wsfoot']['dialog']){?>
		<button type="button" class="btn" data-dialog-close>Cancel</button>
<?php }else{?>
		<a class="btn" href="<?php echo htmlspecialchars($_v['data_wsfoot']['cancel'], ENT_QUOTES);?>
">Cancel</a>
<?php }?>		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_wsfoot']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>
</dialog>
<dialog class="confirm" id="key-upload-dialog" aria-labelledby="key-upload-title">
	<?php $_v['data_wsup'] = $_v['data_wsdialogs']['upload'];?>
<form class="form" data-api="<?php echo htmlspecialchars($_v['data_wsup']['api'], ENT_QUOTES);?>
" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data_wsup']['next'], ENT_QUOTES);?>
">
<?php if($_v['data_wsup']['dialog']){?>
	<h3 id="key-upload-title">Upload key</h3>
	<p>Saves a private key you already use as an SSH key of <?php echo htmlspecialchars($_v['data_wsup']['folder'], ENT_QUOTES);?>
. The folder refuses a key it holds, under any name.</p>
<?php }?>	<div class="field">
		<label for="wku-file">Private key file</label>
		<input type="file" id="wku-file" data-read-into="wku-content">
	</div>
	<div class="field">
		<label for="wku-content">Or paste the private key</label>
		<textarea id="wku-content" name="content" rows="<?php echo htmlspecialchars($_v['data_wsup']['rows'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	<div class="field">
		<label for="wku-name">Name</label>
		<input type="text" id="wku-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wku-comment">Comment</label>
		<input type="text" id="wku-comment" name="comment" value="<?php echo htmlspecialchars($_v['data_wsup']['comment'], ENT_QUOTES);?>
">
	</div>
<?php $_v['data_wsfoot'] = $_v['data_wsup']['foot'];?>
	<div class="form-foot">
<?php if($_v['data_wsfoot']['dialog']){?>
		<button type="button" class="btn" data-dialog-close>Cancel</button>
<?php }else{?>
		<a class="btn" href="<?php echo htmlspecialchars($_v['data_wsfoot']['cancel'], ENT_QUOTES);?>
">Cancel</a>
<?php }?>		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_wsfoot']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>
</dialog>
<?php }?>
<dialog class="confirm workspace-export" id="export-dialog" aria-labelledby="export-title">
	<form class="form" data-export="/api/workspaces/<?php echo htmlspecialchars($_v['data']['export'], ENT_QUOTES);?>
/export">
		<h3 id="export-title">Export <?php echo htmlspecialchars($_v['data']['export'], ENT_QUOTES);?>
</h3>
		<p>Downloads <span class="mono-break"><?php echo htmlspecialchars($_v['data']['export'], ENT_QUOTES);?>
.tui.age</span> for Import on another task-ui's Workspaces page: the workspace, its projects, env files and grants, its secrets and the SSH keys its runs get, encrypted to the passphrase. Anyone with the file and the passphrase has the secrets and the keys. No instance key and no encryption key goes; schedules, workflows, runs and artifacts stay here.</p>
		<div class="field">
			<label for="wx-passphrase">Passphrase <span class="faint">at least 12 characters</span></label>
			<div class="passphrase">
				<input type="password" id="wx-passphrase" name="passphrase" minlength="12" autocomplete="new-password" spellcheck="false" required>
				<button type="button" class="btn-icon" data-copy="" data-passphrase-copy title="Copy passphrase" aria-label="Copy passphrase" hidden><svg class="ic"><use href="#i-copy"/></svg></button>
				<button type="button" class="btn" data-passphrase-generate>Generate</button>
			</div>
		</div>
		<div class="field" data-passphrase-confirm>
			<label for="wx-confirm">Confirm passphrase</label>
			<input type="password" id="wx-confirm" name="confirm" autocomplete="new-password" spellcheck="false">
		</div>
		<p class="passphrase-note" data-passphrase-note hidden>Store this passphrase before you export: task-ui keeps no copy, and nothing recovers it.</p>
		<div class="form-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		</div>
	</form>
</dialog>
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
