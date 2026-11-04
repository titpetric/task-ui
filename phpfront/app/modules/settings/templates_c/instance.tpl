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

<section class="block instance-section" id="general">
	<h3>General</h3>
	<p class="section-intro">Where the instance answers, who signs in, and the clock it reads.</p>
	<div class="table-wrap">
	<table class="table instance-table cards">
		<colgroup><col class="c-key"><col class="c-value"><col class="c-var"><col class="c-desc"></colgroup>
		<thead><tr><th>Setting</th><th>Value</th><th>Variable</th><th>Description</th></tr></thead>
		<tbody>

		<tr><td class="key">Base link</td><td class="value mono" data-label="Value"><?php echo htmlspecialchars($_v['data']['site'], ENT_QUOTES);?>
</td><td class="var" data-label="Variable"><code>TASK_UI_URL</code></td><td class="desc" data-label="About"><span><?php if($_v['data']['instance']['url']){?>
The link pages and MCP clients use.<?php }else{?>
Empty: task-ui reads the link from each request.<?php }?></span></td></tr>
		<tr><td class="key">Version</td><td class="value mono ro" data-label="Value"><?php echo htmlspecialchars($_v['chrome']['version'], ENT_QUOTES);?>
</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>The build that serves this page.</span></td></tr>
		<tr><td class="key">Listen</td><td class="value mono" data-label="Value"><?php echo htmlspecialchars($_v['data']['instance']['listen'], ENT_QUOTES);?>
</td><td class="var" data-label="Variable"><code>TASK_UI_LISTEN</code></td><td class="desc" data-label="About"><span>The address the server listens on.</span></td></tr>
		<tr><td class="key">Time zone</td><td class="value mono ro" data-label="Value"><?php if($_v['data']['instance']['time_zone']){ echo htmlspecialchars($_v['data']['instance']['time_zone'], ENT_QUOTES); }else{?>
<span class="faint">local</span><?php }?></td><td class="var" data-label="Variable"><code>TZ</code></td><td class="desc" data-label="About"><span>The clock of the server, which the hours of a folder and the schedules read.</span></td></tr>
		<tr><td class="key">Sign in</td><td class="value" data-label="Value"><?php if($_v['data']['instance']['auto_login']){?>
<span class="badge bad">auto login</span> <?php $_v['data_user'] = $_v['data']['signin'];?>
<span class="user-ref"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php $_v['data_menu'] = $_v['data_user']['menu'];if(!$_v['data_menu']['empty']){?>
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
</span>
<?php }elseif($_v['chrome']['loginRequired']){?>
<span class="badge ok">required</span><?php }else{?>
<span class="badge">open</span><?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_AUTOLOGIN</code></td><td class="desc" data-label="About"><span>Auto login signs everyone in as one user; for demos only.</span></td></tr>
		<tr><td class="key">Session lifetime</td><td class="value mono" data-label="Value"><?php if($_v['data']['instance']['session_ttl']){ echo htmlspecialchars($_v['data']['instance']['session_ttl'], ENT_QUOTES); }else{?>
<span class="faint">default</span><?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_SESSION_TTL</code></td><td class="desc" data-label="About"><span>How long a browser stays signed in.</span></td></tr>
		<tr><td class="key">Password file</td><td class="value mono" data-label="Value"><?php if($_v['data']['instance']['htpasswd']){ echo htmlspecialchars($_v['data']['instance']['htpasswd'], ENT_QUOTES); }else{?>
<span class="faint">none</span><?php }?></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Who can sign in: <code>.htpasswd</code> in the data directory.</span></td></tr>
		<tr><td class="key">Roles</td><td class="value" data-label="Value"><a href="<?php echo htmlspecialchars($_v['data']['users'], ENT_QUOTES);?>
">Users</a></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>The admin flag, the role and MCP of each user live on their profile, and <code>task-ui user</code> sets them.</span></td></tr>
		<tr><td class="key">MCP</td><td class="value mono" data-label="Value"><?php echo htmlspecialchars($_v['data']['site'], ENT_QUOTES);?>
/api/mcp</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Where MCP clients connect, with a user's token.</span></td></tr>
	</tbody>
	</table>
	</div>

</section>

<section class="block instance-section" id="runs">
	<h3>Runs</h3>
	<p class="section-intro">How many runs do work at once, and how many may wait for a free slot. Terminals take no slot.</p>
	<div class="table-wrap">
	<table class="table instance-table cards">
		<colgroup><col class="c-key"><col class="c-value"><col class="c-var"><col class="c-desc"></colgroup>
		<thead><tr><th>Setting</th><th>Value</th><th>Variable</th><th>Description</th></tr></thead>
		<tbody>

<?php if($_v['data']['instance']['running']){?>
		<tr><td class="key">Parallel runs</td><td class="value mono ro" data-label="Value"><?php echo htmlspecialchars($_v['data']['instance']['backlog']['parallel'], ENT_QUOTES);?>
<small class="faint"><?php echo htmlspecialchars($_v['data']['instance']['backlog']['going'], ENT_QUOTES);?>
 going, <?php echo htmlspecialchars($_v['data']['instance']['backlog']['waiting'], ENT_QUOTES);?>
 waiting</small></td><td class="var" data-label="Variable"><span class="faint">computed</span></td><td class="desc" data-label="About"><span>One run per CPU of the machine; a run over that waits, queued, in order.</span></td></tr>
		<tr><td class="key">Automatic backlog</td><td class="value mono ro" data-label="Value"><?php echo htmlspecialchars($_v['data']['instance']['backlog']['auto'], ENT_QUOTES);?>
</td><td class="var" data-label="Variable"><code>backlog_length_auto</code></td><td class="desc" data-label="About"><span>20 runs per CPU: the length of the line while an admin sets no backlog length.</span></td></tr>
		<tr><td class="key">Backlog length</td><td class="value" data-label="Value">
<?php if($_v['data']['admin']){?>
			<form class="inline-setting" data-api="/api/instance" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['data']['backlog'], ENT_QUOTES);?>
">
				<input type="text" name="backlog_length" value="<?php if($_v['data']['instance']['backlog']['configured']){ echo htmlspecialchars($_v['data']['instance']['backlog']['configured'], ENT_QUOTES); }?>" placeholder="<?php echo htmlspecialchars($_v['data']['instance']['backlog']['auto'], ENT_QUOTES);?>
" inputmode="numeric" autocomplete="off" spellcheck="false" aria-label="Backlog length, empty for the automatic <?php echo htmlspecialchars($_v['data']['instance']['backlog']['auto'], ENT_QUOTES);?>
">
				<button class="btn btn-sm" type="submit">Save</button>
			</form>
<?php }else{?>
			<span class="mono"><?php if($_v['data']['instance']['backlog']['configured']){ echo htmlspecialchars($_v['data']['instance']['backlog']['configured'], ENT_QUOTES); }else{?>
<span class="faint">automatic</span><?php }?></span>
<?php }?>			<small class="faint">in effect: <?php echo htmlspecialchars($_v['data']['instance']['backlog']['length'], ENT_QUOTES);?>
</small></td>
			<td class="var" data-label="Variable"><code>backlog_length</code></td><td class="desc" data-label="About"><span>How many runs may wait for a slot, up to 100000; empty is the automatic length. A start past it answers <code>429</code>.</span></td></tr>
<?php }?>		<tr><td class="key">Keep attachable</td><td class="value mono" data-label="Value"><?php if($_v['data']['instance']['keep']){ echo htmlspecialchars($_v['data']['instance']['keep'], ENT_QUOTES); }else{?>
<span class="faint">default</span><?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_KEEP</code></td><td class="desc" data-label="About"><span>How long a finished run stays in memory to attach to; its row and recording stay.</span></td></tr>
		<tr><td class="key">Stop grace</td><td class="value mono" data-label="Value"><?php if($_v['data']['instance']['grace']){ echo htmlspecialchars($_v['data']['instance']['grace'], ENT_QUOTES); }else{?>
<span class="faint">default</span><?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_GRACE</code></td><td class="desc" data-label="About"><span>How long a stopped run gets before the kill.</span></td></tr>
		<tr><td class="key">Artifact history</td><td class="value mono" data-label="Value"><?php if($_v['data']['instance']['artifact_history']){ echo htmlspecialchars($_v['data']['instance']['artifact_history'], ENT_QUOTES);?>
 versions<?php }else{?>
every version<?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_ARTIFACT_HISTORY</code></td><td class="desc" data-label="About"><span>How many versions the artifact history keeps before it squashes to one.</span></td></tr>
		<tr><td class="key">Retention</td><td class="value" data-label="Value"><a href="<?php echo htmlspecialchars($_v['data']['retention'], ENT_QUOTES);?>
">Data retention</a></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>How long runs, recordings, prompt threads, workflow runs, artifact history and the secret read log stay.</span></td></tr>
	</tbody>
	</table>
	</div>

</section>

<?php if($_v['data']['instance']['parent_env']){?>
<section class="block instance-section" id="environment">
	<h3>Parent environment</h3>
	<p class="section-intro">How much of the server's own environment a run starts with. The variables of the folder and of the integrations reach a run under every choice; a script reads any other variable with <code>task-ui get-secret NAME</code>, as <a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/settings.md#secrets-in-scripts">Secrets in scripts</a> shows.</p>
<?php if($_v['data']['admin']){?>
	<form class="form parent-env" data-api="/api/instance" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['data']['forward'], ENT_QUOTES);?>
" data-parent-env>
		<div class="field" role="radiogroup" aria-labelledby="parent-env-label">
			<span class="label" id="parent-env-label">Forward <span class="faint">applies to the next run</span></span>
<?php if(!empty($_v['data']['modes']))foreach($_v['data']['modes'] as $_v['mode']){?>			<label class="check"><input type="radio" name="parent_env" value="<?php echo htmlspecialchars($_v['mode']['value'], ENT_QUOTES);?>
"<?php if($_v['mode']['checked']){?>
 checked<?php }?>> <strong><?php echo htmlspecialchars($_v['mode']['label'], ENT_QUOTES);?>
</strong></label>
			<small class="faint"><?php echo htmlspecialchars($_v['mode']['note'], ENT_QUOTES);?>
</small>
<?php }?>		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
	<p class="faint">Base variables: <?php if(!empty($_v['data']['base']))foreach($_v['data']['base'] as $_v['item']){?><?php echo htmlspecialchars($_v['item']['sep'], ENT_QUOTES);?>
<code><?php echo htmlspecialchars($_v['item']['name'], ENT_QUOTES);?>
</code><?php }?>. A name holding <?php if(!empty($_v['data']['words']))foreach($_v['data']['words'] as $_v['item']){?><?php echo htmlspecialchars($_v['item']['sep'], ENT_QUOTES);?>
<code><?php echo htmlspecialchars($_v['item']['name'], ENT_QUOTES);?>
</code><?php }?> is sensitive.</p>
<?php if($_v['data']['env']){?>
	<div class="table-wrap">
	<table class="table env-table cards">
		<thead><tr><th>Variable</th><th>Forwarded</th><th>Flag</th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['env']))foreach($_v['data']['env'] as $_v['row']){?>		<tr data-env-none="<?php echo htmlspecialchars($_v['row']['none'], ENT_QUOTES);?>
" data-env-auto="<?php echo htmlspecialchars($_v['row']['auto'], ENT_QUOTES);?>
"><td class="mono"><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</td>
			<td data-label="Forwarded" data-env-forwarded><?php if($_v['row']['forwarded']){?>
<span class="badge ok">yes</span><?php }else{?>
<span class="badge">no</span><?php }?></td>
			<td data-label="Flag"><?php if($_v['row']['sensitive']){?>
<span class="badge bad" title="The name looks like a credential's">sensitive</span><?php }?></td></tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php }?><?php }else{?>
	<p>In effect: <strong><?php echo htmlspecialchars($_v['data']['instance']['parent_env'], ENT_QUOTES);?>
</strong>; an admin changes it here.</p>
<?php }?></section>
<?php }?>
<section class="block instance-section" id="paths">
	<h3>Paths</h3>
	<p class="section-intro">Where task-ui keeps its data, the folders it runs and its keys.</p>
	<div class="table-wrap">
	<table class="table instance-table cards">
		<colgroup><col class="c-key"><col class="c-value"><col class="c-var"><col class="c-desc"></colgroup>
		<thead><tr><th>Setting</th><th>Value</th><th>Variable</th><th>Description</th></tr></thead>
		<tbody>

		<tr><td class="key">Workspace folder</td><td class="value mono" data-label="Value"><?php echo htmlspecialchars($_v['data']['instance']['root'], ENT_QUOTES);?>
</td><td class="var" data-label="Variable"><code>TASK_UI_WORKSPACE</code></td><td class="desc" data-label="About"><span>The folders with jobs to run, by default <code>workspaces</code> in the data directory.</span></td></tr>
		<tr><td class="key">Discovery</td><td class="value" data-label="Value"><?php if($_v['data']['instance']['discover']){?>
<span class="badge ok">on</span><?php }else{?>
<span class="badge">off</span><?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_DISCOVER</code></td><td class="desc" data-label="About"><span>List every folder under the root with something to run.</span></td></tr>
		<tr><td class="key">Data</td><td class="value mono" data-label="Value"><?php echo htmlspecialchars($_v['data']['instance']['data_dir'], ENT_QUOTES);?>
</td><td class="var" data-label="Variable"><code>TASK_UI_DATA</code></td><td class="desc" data-label="About"><span>The database, recordings and artifacts.</span></td></tr>
		<tr><td class="key">Data on the host</td><td class="value mono" data-label="Value"><?php if($_v['data']['instance']['host_data']){ echo htmlspecialchars($_v['data']['instance']['host_data'], ENT_QUOTES); }else{?>
<span class="faint">the same</span><?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_HOST_DATA</code></td><td class="desc" data-label="About"><span>The data directory as the host's docker daemon sees it; runs get it, and their own directory in <code>TASK_UI_HOST_PWD</code>.</span></td></tr>
		<tr><td class="key">Key folder</td><td class="value mono" data-label="Value"><?php echo htmlspecialchars($_v['data']['instance']['ssh_dir'], ENT_QUOTES);?>
</td><td class="var" data-label="Variable"><code>TASK_UI_SSH_DIR</code></td><td class="desc" data-label="About"><span>Every key of the instance, and the backups of rotated keys in <code>backup/</code>.</span></td></tr>
		<tr><td class="key">Compose project</td><td class="value mono" data-label="Value"><?php if($_v['data']['instance']['compose_project']){ echo htmlspecialchars($_v['data']['instance']['compose_project'], ENT_QUOTES); }else{?>
<span class="faint">none</span><?php }?></td><td class="var" data-label="Variable"><code>TASK_UI_COMPOSE_PROJECT</code></td><td class="desc" data-label="About"><span>The docker compose project task-ui runs in, whose up and down ask first.</span></td></tr>
	</tbody>
	</table>
	</div>

</section>
<?php if($_v['data']['instance']['measured']){?>
<section class="block instance-section" id="storage">
	<h3>Storage</h3>
	<p class="section-intro">What the instance keeps on disk: its database, the recordings and the artifacts of the runs, the tool history of each workspace's home and the instance's own copy, and the checkouts under each workspace's <code>src</code>. Measured now and then; an ellipsis stands for a measure still under way.</p>
	<div class="table-wrap">
	<table class="table storage-table cards">
		<colgroup><col class="c-item"><col class="c-where"><col class="c-size"><col class="c-files"></colgroup>
		<thead><tr><th>Item</th><th>Where</th><th class="num">Size</th><th class="num">Files</th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['storage']['rows']))foreach($_v['data']['storage']['rows'] as $_v['row']){?>		<tr><td class="key"><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</td><td class="where mono" data-label="Where"><?php echo htmlspecialchars($_v['row']['where'], ENT_QUOTES);?>
</td><td class="num mono" data-label="Size"><?php if($_v['row']['pending']){?>
<span class="faint pending" aria-label="not measured yet"></span><?php }else{ echo htmlspecialchars($_v['row']['size'], ENT_QUOTES); }?></td><td class="num mono" data-label="Files"><?php if($_v['row']['pending']){?>
<span class="faint pending" aria-label="not measured yet"></span><?php }else{ echo htmlspecialchars($_v['row']['files'], ENT_QUOTES); }?></td></tr>
<?php }?>		</tbody>
		<tfoot><tr><th>Total</th><td></td><td class="num mono" data-label="Size"><?php if($_v['data']['storage']['total']['pending']){?>
<span class="faint pending" aria-label="not measured yet"></span><?php }else{ echo htmlspecialchars($_v['data']['storage']['total']['size'], ENT_QUOTES); }?></td><td class="num mono" data-label="Files"><?php if($_v['data']['storage']['total']['pending']){?>
<span class="faint pending" aria-label="not measured yet"></span><?php }else{ echo htmlspecialchars($_v['data']['storage']['total']['files'], ENT_QUOTES); }?></td></tr></tfoot>
	</table>
	</div>
</section>
<?php }?>
<section class="block instance-section" id="tooling">
	<h3>Tooling</h3>
	<p class="section-intro">The tools task-ui found in its PATH at the last check, and what each is used for. Check probes one tool; Check all probes every tool and replaces the list. A tool that is not there is left out.</p>
<?php if($_v['data']['admin']){?>
	<form class="inline" data-api="/api/instance/tools/check" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data']['checked'], ENT_QUOTES);?>
"><button class="btn btn-sm" type="submit">Check all</button></form>
<?php }?><?php if(!empty($_v['data']['tooling']))foreach($_v['data']['tooling'] as $_v['group']){?>	<h4><?php echo htmlspecialchars($_v['group']['title'], ENT_QUOTES);?>
</h4>
	<div class="table-wrap">
	<table class="table toolset-table cards">
		<colgroup><col class="c-tool"><col class="c-version"><col class="c-path"><col class="c-use"><col class="c-actions"></colgroup>
		<thead><tr><th>Tool</th><th>Version</th><th>Path</th><th>Used for</th><th></th></tr></thead>
		<tbody>
<?php if(!empty($_v['group']['tools']))foreach($_v['group']['tools'] as $_v['tool']){?>		<tr><td class="key mono" data-label="Tool"><?php echo htmlspecialchars($_v['tool']['name'], ENT_QUOTES);?>
</td><td class="mono" data-label="Version"><?php if($_v['tool']['version']){ echo htmlspecialchars($_v['tool']['version'], ENT_QUOTES); }else{?>
<span class="faint">unknown</span><?php }?></td><td class="path mono" data-label="Path"><?php echo htmlspecialchars($_v['tool']['path'], ENT_QUOTES);?>
</td><td class="desc" data-label="Used for"><?php echo htmlspecialchars($_v['tool']['description'], ENT_QUOTES);?>
</td><td class="actions"><?php if($_v['data']['admin']){?>
<form class="inline" data-api="/api/instance/tools/<?php echo htmlspecialchars($_v['tool']['id'], ENT_QUOTES);?>
/check" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data']['probed'], ENT_QUOTES);?>
"><button class="btn btn-sm" type="submit">Check</button></form><?php }?></td></tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php }?><?php if(!$_v['data']['tooling']){?>
	<p>No tooling was detected: Check all probes the PATH.</p>
<?php }?></section>
<?php if($_v['data']['keys']['shown']){?>
<section class="block instance-section" id="keys">
	<h3>Keys <small>in the key folder</small></h3>
	<p class="section-intro">The encryption key of the instance's secrets and the SSH key git pulls and clones with. <a href="<?php echo htmlspecialchars($_v['data']['keys']['every'], ENT_QUOTES);?>
">Every key</a></p>
	<div class="table-wrap fit">
	<table class="table instance-keys key-table cards">
		<colgroup><col class="c-name"><col class="c-algo"><col class="c-size"><col class="c-pub"><col class="c-actions"></colgroup>
		<thead><tr><th>Name</th><th class="hide-narrow">Algorithm</th><th class="hide-narrow">Size</th><th>Public key</th><th class="actions" aria-label="Actions"></th></tr></thead>

		<tbody>
<?php if($_v['data']['keys']['encrypt']){ $_v['data_key'] = $_v['data']['keys']['encrypt'];?>
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
<?php }else{?>
			<tr><td class="key">Encryption key</td><td colspan="4" data-label="Key"><span class="badge bad">missing</span> <span class="faint">The integrations' settings and the secrets need one.</span></td>
				<td class="actions"><?php if($_v['data']['keys']['edits']){?>
<form class="inline" data-api="/api/keys/encrypt/rotate" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data']['keys']['next'], ENT_QUOTES);?>
" data-confirm="Generate the instance's encryption key?" data-confirm-ok="Generate"><button class="btn btn-sm" type="submit">Generate</button></form> <a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data']['keys']['upload'], ENT_QUOTES);?>
">Upload</a><?php }?></td></tr>
<?php }?><?php if($_v['data']['keys']['git']){ $_v['data_key'] = $_v['data']['keys']['git'];?>
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
<?php }else{?>
			<tr><td class="key">Git key</td><td colspan="4" data-label="Key"><span class="faint">none: git uses ssh's own</span></td>
				<td class="actions"><?php if($_v['data']['keys']['edits']){?>
<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data']['keys']['every'], ENT_QUOTES);?>
">Choose</a><?php }?></td></tr>
<?php }?>		</tbody>
	</table>
	</div>
</section>
<?php }?>
<section class="block instance-section" id="integrations">
	<h3>Integrations</h3>
	<p class="section-intro">The services task-ui reads from.</p>
	<div class="table-wrap">
	<table class="table instance-table cards">
		<colgroup><col class="c-key"><col class="c-value"><col class="c-var"><col class="c-desc"></colgroup>
		<thead><tr><th>Setting</th><th>Value</th><th>Variable</th><th>Description</th></tr></thead>
		<tbody>

		<tr><td class="key">Inventory</td><td class="value" data-label="Value"><?php if($_v['data']['instance']['beszel']){?>
<span class="badge ok">on</span><?php }else{?>
<span class="badge">off</span><?php }?></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Hosts from the Beszel <a href="<?php echo htmlspecialchars($_v['data']['integrations'], ENT_QUOTES);?>
">integration</a>.</span></td></tr>
	</tbody>
	</table>
	</div>

</section>
<?php if($_v['data']['trail']){ $_v['data_trail'] = $_v['data']['trail'];?>

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
