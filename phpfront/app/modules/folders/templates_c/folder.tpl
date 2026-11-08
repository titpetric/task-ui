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

<?php if($_v['data']['findings']){ $_v['data_findings'] = $_v['data']['findings'];?>
<div class="alert bad pipeline-alert" role="alert">
	<b><?php echo htmlspecialchars($_v['data_findings']['pipelines'], ENT_QUOTES);?>
 <?php echo htmlspecialchars($_v['data_findings']['says'], ENT_QUOTES);?>
 errors.</b> atkins lists no jobs of a pipeline its linter refuses; fix the file and they return.
	<ul class="pipeline-errors">
<?php if(!empty($_v['data_findings']['shown']))foreach($_v['data_findings']['shown'] as $_v['finding']){?><?php $_v['data_finding'] = $_v['finding'];?>
<li><a href="<?php echo htmlspecialchars($_v['data_finding']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_finding']['folder'], ENT_QUOTES);?>
</a> <code><?php echo htmlspecialchars($_v['data_finding']['path'], ENT_QUOTES);?>
</code><?php if($_v['data_finding']['job']){?>
 <span class="dim">job</span> <b><?php echo htmlspecialchars($_v['data_finding']['job'], ENT_QUOTES);?>
</b><?php }?>: <span<?php if($_v['data_finding']['issue']){?>
 title="<?php echo htmlspecialchars($_v['data_finding']['issue'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_finding']['detail'], ENT_QUOTES);?>
</span></li>

<?php }?>	</ul>
<?php if($_v['data_findings']['rest']){?>
	<details class="pipeline-more">
		<summary><svg class="ic chev"><use href="#i-chev"/></svg>Show <?php echo htmlspecialchars($_v['data_findings']['rest'], ENT_QUOTES);?>
 more</summary>
		<ul class="pipeline-errors">
<?php if(!empty($_v['data_findings']['more']))foreach($_v['data_findings']['more'] as $_v['finding']){?><?php $_v['data_finding'] = $_v['finding'];?>
<li><a href="<?php echo htmlspecialchars($_v['data_finding']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_finding']['folder'], ENT_QUOTES);?>
</a> <code><?php echo htmlspecialchars($_v['data_finding']['path'], ENT_QUOTES);?>
</code><?php if($_v['data_finding']['job']){?>
 <span class="dim">job</span> <b><?php echo htmlspecialchars($_v['data_finding']['job'], ENT_QUOTES);?>
</b><?php }?>: <span<?php if($_v['data_finding']['issue']){?>
 title="<?php echo htmlspecialchars($_v['data_finding']['issue'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_finding']['detail'], ENT_QUOTES);?>
</span></li>

<?php }?>		</ul>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data']['git']){ $_v['data_git'] = $_v['data']['git'];?>
<div class="actionsbar">
	<div class="gitline" title="<?php echo htmlspecialchars($_v['data_git']['title'], ENT_QUOTES);?>
">
		<span class="branch"><span class="glyph" aria-hidden="true">⎇</span> <?php if($_v['data_git']['branch']){ echo htmlspecialchars($_v['data_git']['branch'], ENT_QUOTES); }else{?>
<span class="faint">detached</span> <?php echo htmlspecialchars($_v['data_git']['detached'], ENT_QUOTES); }?></span>
<?php if($_v['data_git']['ahead']){?>
		<span class="ahead" title="<?php echo htmlspecialchars($_v['data_git']['ahead'], ENT_QUOTES);?>
 commits to push">↑<?php echo htmlspecialchars($_v['data_git']['ahead'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_git']['behind']){?>
		<span class="behind" title="<?php echo htmlspecialchars($_v['data_git']['behind'], ENT_QUOTES);?>
 commits to pull">↓<?php echo htmlspecialchars($_v['data_git']['behind'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_git']['dirty']){?>
		<button type="button" class="dirty" data-git-status aria-expanded="false" aria-controls="git-status" title="Show the changed files"><span class="glyph" aria-hidden="true">●</span> <?php echo htmlspecialchars($_v['data_git']['files'], ENT_QUOTES);?>
<svg class="ic chev" aria-hidden="true"><use href="#i-chev"/></svg></button>
<?php }?><?php if($_v['data_git']['clean']){?>
		<span class="clean">up to date</span>
<?php }?>		<span class="where">· <?php echo htmlspecialchars($_v['data_git']['where'], ENT_QUOTES);?>
</span>
	</div>
<?php if($_v['data_git']['dirty']){?>
	<div class="gitstatus" id="git-status" hidden>
		<div class="table-wrap">
			<table class="table gitstatus-table">
				<thead><tr><th>Status</th><th>File</th></tr></thead>
				<tbody>
<?php if(!empty($_v['data_git']['changes']))foreach($_v['data_git']['changes'] as $_v['change']){?>					<tr><td class="gitstate"><code><?php echo htmlspecialchars($_v['change']['code'], ENT_QUOTES);?>
</code> <span class="faint"><?php echo htmlspecialchars($_v['change']['state'], ENT_QUOTES);?>
</span></td><td class="gitpath"><?php echo htmlspecialchars($_v['change']['path'], ENT_QUOTES); if($_v['change']['from']){?>
 <span class="faint">from <?php echo htmlspecialchars($_v['change']['from'], ENT_QUOTES);?>
</span><?php }?></td></tr>
<?php }?>				</tbody>
			</table>
		</div>
	</div>
<?php }?></div>

<?php }?><?php if($_v['data']['error']){?>
<div class="alert bad"><b>Could not read a source.</b> <?php echo htmlspecialchars($_v['data']['error'], ENT_QUOTES);?>
</div>
<?php }?><?php if($_v['data']['control']){?>
<section class="block" data-live>
	<h3>Jobs <small>a row per job source, then files, skills and terminals</small></h3>
<?php $_v['data_control'] = $_v['data']['control'];?>
<div class="table-wrap">
<table class="table control col-hover">
	<tbody>
<?php if(!empty($_v['data_control']))foreach($_v['data_control'] as $_v['row']){?>		<tr class="named" data-span>
			<td class="key nowrap"<?php if($_v['row']['title']){?>
 title="<?php echo htmlspecialchars($_v['row']['title'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['row']['href']){?>
<a class="badge" href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['label'], ENT_QUOTES);?>
</a><?php }else{?>
<span class="badge"><?php echo htmlspecialchars($_v['row']['label'], ENT_QUOTES);?>
</span><?php }?></td>
			<td colspan="1"><?php if($_v['row']['empty']){?>
<span class="faint"><?php echo htmlspecialchars($_v['row']['empty'], ENT_QUOTES);?>
</span><?php }else{?>
<div class="cc-row"><?php if(!empty($_v['row']['jobs']))foreach($_v['row']['jobs'] as $_v['job']){?><?php $_v['data_cc'] = $_v['job'];?>
<span class="cc-slot" data-cc="<?php echo htmlspecialchars($_v['data_cc']['key'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_cc']['folder'], ENT_QUOTES);?>
" data-job="<?php echo htmlspecialchars($_v['data_cc']['job'], ENT_QUOTES);?>
" data-name="<?php echo htmlspecialchars($_v['data_cc']['name'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
" data-last="<?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['can']){?>
 data-can<?php }?><?php if($_v['data_cc']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_cc']['dir'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['desc']){?>
 data-desc="<?php echo htmlspecialchars($_v['data_cc']['desc'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_cc']['args'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['confirm']){?>
 data-confirm="<?php echo htmlspecialchars($_v['data_cc']['confirm'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['path']){?>
 data-path="<?php echo htmlspecialchars($_v['data_cc']['path'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['terminal']){?>
 data-interactive<?php }?>><?php if($_v['data_cc']['run']){?>
<a class="cc named<?php if($_v['data_cc']['pinned']){?>
 pinned<?php }?> running" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_cc']['run'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_cc']['run'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></a><?php }elseif($_v['data_cc']['can']){?>
<button type="button" class="cc named<?php if($_v['data_cc']['pinned']){?>
 pinned<?php }?> <?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
" data-run data-folder="<?php echo htmlspecialchars($_v['data_cc']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_cc']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_cc']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cc']['what'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['confirm']){?>
 data-confirm="<?php echo htmlspecialchars($_v['data_cc']['confirm'], ENT_QUOTES);?>
" data-confirm-ok="Run <?php echo htmlspecialchars($_v['data_cc']['name'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_cc']['args'], ENT_QUOTES);?>
"<?php }?> title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-<?php if($_v['data_cc']['terminal']){?>
terminal<?php }else{?>
play<?php }?>"/></svg><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></button><?php }else{?>
<a class="cc named<?php if($_v['data_cc']['pinned']){?>
 pinned<?php }?> idle <?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
" href="<?php echo htmlspecialchars($_v['data_cc']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><i class="shebang" aria-hidden="true">#!</i><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></a><?php }?></span>
<?php }?></div><?php }?></td>
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
</section>
<?php }?><?php if($_v['data']['nothing']){?>
<div class="alert">task-ui found nothing to run in this folder: no pipeline file, Makefile, package scripts or scripts folder.</div>
<?php }?><?php if($_v['data']['prompt']){ $_v['data_prompt'] = $_v['data']['prompt'];?>
<details class="promptbox"<?php if($_v['data_prompt']['open']){?>
 open<?php }?><?php if($_v['data_prompt']['hidden']){?>
 hidden<?php }?>>
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg><?php echo htmlspecialchars($_v['data_prompt']['title'], ENT_QUOTES);?>
 <small><?php echo htmlspecialchars($_v['data_prompt']['note'], ENT_QUOTES);?>
</small></h3></summary>
<?php if(!$_v['data_prompt']['watch']){?>
	<form class="note-form" data-api="<?php echo htmlspecialchars($_v['data_prompt']['api'], ENT_QUOTES);?>
" data-refresh-notes>
<?php if(!empty($_v['data_prompt']['fields']))foreach($_v['data_prompt']['fields'] as $_v['field']){?>		<input type="hidden" name="<?php echo htmlspecialchars($_v['field']['name'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['field']['value'], ENT_QUOTES);?>
">
<?php }?>		<?php $_v['data_markdown'] = $_v['data_prompt']['editor'];?>
<textarea name="<?php echo htmlspecialchars($_v['data_markdown']['name'], ENT_QUOTES);?>
" rows="<?php echo htmlspecialchars($_v['data_markdown']['rows'], ENT_QUOTES);?>
" spellcheck="false" data-markdown data-toolbar="<?php echo htmlspecialchars($_v['data_markdown']['toolbar'], ENT_QUOTES);?>
"<?php if($_v['data_markdown']['mentions']){?>
 data-mentions<?php }?><?php if($_v['data_markdown']['placeholder']){?>
 placeholder="<?php echo htmlspecialchars($_v['data_markdown']['placeholder'], ENT_QUOTES);?>
"<?php }?> aria-label="<?php echo htmlspecialchars($_v['data_markdown']['label'], ENT_QUOTES);?>
"></textarea>
		<div class="form-foot note"><span class="faint run-in"><?php echo htmlspecialchars($_v['data_prompt']['lead'], ENT_QUOTES); if($_v['data_prompt']['options']['users']){?>
<select name="user" aria-label="Runs as"><?php $_v['data_options'] = $_v['data_prompt']['options']; if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select><?php }else{?>
you<?php }?><?php echo htmlspecialchars($_v['data_prompt']['tail'], ENT_QUOTES); if($_v['data_prompt']['picks']){?>
<select name="folder" aria-label="Run in"><?php if(!empty($_v['data_prompt']['picks']))foreach($_v['data_prompt']['picks'] as $_v['pick']){?><option value="<?php echo htmlspecialchars($_v['pick']['value'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['pick']['text'], ENT_QUOTES);?>
</option><?php }?></select><?php }?><?php echo htmlspecialchars($_v['data_prompt']['after'], ENT_QUOTES);?>
</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
	</form>
<?php }?>	<div class="notes prompt-threads" data-notes data-dir="<?php echo htmlspecialchars($_v['data_prompt']['place'], ENT_QUOTES);?>
"<?php if($_v['data_prompt']['area']){?>
 data-area="<?php echo htmlspecialchars($_v['data_prompt']['area'], ENT_QUOTES);?>
"<?php }?>>
<?php if(!empty($_v['data_prompt']['threads']))foreach($_v['data_prompt']['threads'] as $_v['thread']){?><?php $_v['data_note'] = $_v['thread'];?>
	<article class="note<?php if($_v['data_note']['state']){?>
 instruction <?php echo htmlspecialchars($_v['data_note']['state'], ENT_QUOTES); }?>" id="note-<?php echo htmlspecialchars($_v['data_note']['id'], ENT_QUOTES);?>
">
		<header>
			<span class="author"><?php $_v['data_user'] = $_v['data_note']['author'];?>
<span class="user-ref"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php $_v['data_menu'] = $_v['data_user']['menu']; if(!$_v['data_menu']['empty']){?>
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
</span>
			<span class="faint" title="<?php echo htmlspecialchars($_v['data_note']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_note']['ago'], ENT_QUOTES);?>
</span>
<?php if($_v['data_note']['edited']){?>
			<span class="faint edited" title="Edited <?php echo htmlspecialchars($_v['data_note']['editedStamp'], ENT_QUOTES);?>
">edited</span>
<?php }?><?php if($_v['data_note']['state']){?>
			<span class="badge note-state <?php echo htmlspecialchars($_v['data_note']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_note']['state'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_note']['run']){?>
			<a class="faint mono" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_note']['run'], ENT_QUOTES);?>
" title="The run the agent worked in">run <?php echo htmlspecialchars($_v['data_note']['short'], ENT_QUOTES);?>
</a>
<?php }?>			<span class="spacer"></span>
			<?php $_v['data_menu'] = $_v['data_note']['menu'];?>

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
</details><?php }?>

		</header>
		<div class="note-body"<?php if($_v['data_note']['edits']){?>
 data-edit="<?php echo htmlspecialchars($_v['data_note']['id'], ENT_QUOTES);?>
" data-body="<?php echo htmlspecialchars($_v['data_note']['raw'], ENT_QUOTES);?>
" title="Click to edit"<?php }?>><?php echo $_v['data_note']['body'];?>
</div>
<?php if($_v['data_note']['replies']){?>
		<ol class="replies">
<?php if(!empty($_v['data_note']['replies']))foreach($_v['data_note']['replies'] as $_v['reply']){?>			<li<?php if($_v['reply']['working']){?>
 class="rerun"<?php }?>>
				<header><span class="author"><?php $_v['data_user'] = $_v['reply']['author'];?>
<span class="user-ref"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php $_v['data_menu'] = $_v['data_user']['menu']; if(!$_v['data_menu']['empty']){?>
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
</span> <span class="faint" title="<?php echo htmlspecialchars($_v['reply']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['reply']['ago'], ENT_QUOTES);?>
</span><?php if($_v['reply']['edited']){?>
 <span class="faint edited" title="Edited <?php echo htmlspecialchars($_v['reply']['editedStamp'], ENT_QUOTES);?>
">edited</span><?php }?><?php if($_v['reply']['working']){?>
 <span class="badge note-state working">working</span><?php }?><?php if($_v['reply']['run']){?>
 <a class="faint mono" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['reply']['run'], ENT_QUOTES);?>
" title="The run this reply comes from">run <?php echo htmlspecialchars($_v['reply']['short'], ENT_QUOTES);?>
</a><?php }?><span class="spacer"></span><?php $_v['data_menu'] = $_v['reply']['menu']; if(!$_v['data_menu']['empty']){?>
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
</header>
				<div class="note-body"<?php if($_v['reply']['edits']){?>
 data-edit="<?php echo htmlspecialchars($_v['reply']['id'], ENT_QUOTES);?>
" data-body="<?php echo htmlspecialchars($_v['reply']['raw'], ENT_QUOTES);?>
" title="Click to edit"<?php }?>><?php echo $_v['reply']['body'];?>
</div>
			</li>
<?php }?>		</ol>
<?php }?>		<details class="reply">
			<summary>Reply</summary>
			<form class="note-form" data-note-form data-parent="<?php echo htmlspecialchars($_v['data_note']['id'], ENT_QUOTES);?>
">
				<?php $_v['data_markdown'] = $_v['data_note']['editor'];?>
<textarea name="<?php echo htmlspecialchars($_v['data_markdown']['name'], ENT_QUOTES);?>
" rows="<?php echo htmlspecialchars($_v['data_markdown']['rows'], ENT_QUOTES);?>
" spellcheck="false" data-markdown data-toolbar="<?php echo htmlspecialchars($_v['data_markdown']['toolbar'], ENT_QUOTES);?>
"<?php if($_v['data_markdown']['mentions']){?>
 data-mentions<?php }?><?php if($_v['data_markdown']['placeholder']){?>
 placeholder="<?php echo htmlspecialchars($_v['data_markdown']['placeholder'], ENT_QUOTES);?>
"<?php }?> aria-label="<?php echo htmlspecialchars($_v['data_markdown']['label'], ENT_QUOTES);?>
"></textarea>
				<div class="form-foot note"><span></span><button class="btn" type="submit">Reply</button></div>
			</form>
		</details>
	</article>
<?php }?>	</div>
</details>

<?php }?><?php if($_v['data']['files']){?>
<section class="block">
	<h3>Files <small><?php echo htmlspecialchars($_v['data']['files']['note'], ENT_QUOTES);?>
</small></h3>
<?php $_v['data_files'] = $_v['data']['files']['panel'];?>
<div class="table-wrap">
<table class="table file-table col-hover">
	<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
	<tbody>
<?php if($_v['data_files']['up']){?>
		<tr class="up"><td class="key"><svg class="ic dir"><use href="#i-folder"/></svg> <a href="<?php echo htmlspecialchars($_v['data_files']['up'], ENT_QUOTES);?>
" title="Up a level">..</a></td><td></td><td class="hide-sm"></td><td></td></tr>
<?php }?><?php if(!empty($_v['data_files']['rows']))foreach($_v['data_files']['rows'] as $_v['row']){?><?php if($_v['row']['sealed']){?>
		<tr class="sealed" aria-disabled="true" title="task-ui lists <?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
 but does not open it">
			<td class="key"><svg class="ic"><use href="#i-secret"/></svg> <span><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</span></td>
			<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
</td>
			<td class="hide-sm nowrap" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</td>
			<td></td>
		</tr>
<?php }else{?>
		<tr class="<?php if($_v['row']['dir']){?>
dir<?php }else{?>
file<?php }?>">
			<td class="key"><?php if($_v['row']['exec']){?>
<i class="shebang" aria-hidden="true">#!</i><?php }else{?>
<svg class="ic<?php if($_v['row']['dir']){?>
 dir<?php }?>"><use href="#i-<?php if($_v['row']['dir']){?>
folder<?php }else{?>
file<?php }?>"/></svg><?php }?><?php if($_v['row']['href']){?>
 <a href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"<?php if(!$_v['row']['dir']){?>
 data-file="<?php echo htmlspecialchars($_v['row']['path'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</a><?php }else{?>
 <span><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['row']['exec']){?>
 <span class="badge exec" title="Runs: executable, or a shebang on its first line">exec</span><?php }?></td>
			<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
</td>
			<td class="hide-sm nowrap" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</td>
			<td class="actions"><?php if($_v['row']['menu']){ $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
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
</details><?php } }?></td>
		</tr>
<?php }?><?php }else{?>
		<tr class="empty"><td colspan="4">This directory is empty.</td></tr>
<?php }?>	</tbody>
</table>
</div>
</section>
<?php }?><?php if($_v['data']['home']){?>
<section class="block folder-home" id="home">
	<h3>Home <small><?php echo htmlspecialchars($_v['data']['home']['note'], ENT_QUOTES);?>
</small></h3>
<?php $_v['data_files'] = $_v['data']['home']['panel'];?>
<div class="table-wrap">
<table class="table file-table col-hover">
	<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
	<tbody>
<?php if($_v['data_files']['up']){?>
		<tr class="up"><td class="key"><svg class="ic dir"><use href="#i-folder"/></svg> <a href="<?php echo htmlspecialchars($_v['data_files']['up'], ENT_QUOTES);?>
" title="Up a level">..</a></td><td></td><td class="hide-sm"></td><td></td></tr>
<?php }?><?php if(!empty($_v['data_files']['rows']))foreach($_v['data_files']['rows'] as $_v['row']){?><?php if($_v['row']['sealed']){?>
		<tr class="sealed" aria-disabled="true" title="task-ui lists <?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
 but does not open it">
			<td class="key"><svg class="ic"><use href="#i-secret"/></svg> <span><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</span></td>
			<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
</td>
			<td class="hide-sm nowrap" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</td>
			<td></td>
		</tr>
<?php }else{?>
		<tr class="<?php if($_v['row']['dir']){?>
dir<?php }else{?>
file<?php }?>">
			<td class="key"><?php if($_v['row']['exec']){?>
<i class="shebang" aria-hidden="true">#!</i><?php }else{?>
<svg class="ic<?php if($_v['row']['dir']){?>
 dir<?php }?>"><use href="#i-<?php if($_v['row']['dir']){?>
folder<?php }else{?>
file<?php }?>"/></svg><?php }?><?php if($_v['row']['href']){?>
 <a href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"<?php if(!$_v['row']['dir']){?>
 data-file="<?php echo htmlspecialchars($_v['row']['path'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</a><?php }else{?>
 <span><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['row']['exec']){?>
 <span class="badge exec" title="Runs: executable, or a shebang on its first line">exec</span><?php }?></td>
			<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
</td>
			<td class="hide-sm nowrap" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</td>
			<td class="actions"><?php if($_v['row']['menu']){ $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
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
</details><?php } }?></td>
		</tr>
<?php }?><?php }else{?>
		<tr class="empty"><td colspan="4">This directory is empty.</td></tr>
<?php }?>	</tbody>
</table>
</div>
</section>
<?php }?><?php if($_v['data']['detail']){?>
<div class="grid-2">
<section class="block">
	<h3>Folder</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Path</th><td><svg class="ic"><use href="#i-folder"/></svg> <?php echo htmlspecialchars($_v['data']['detail']['path'], ENT_QUOTES);?>
</td></tr>
<?php if(!empty($_v['data']['detail']['sources']))foreach($_v['data']['detail']['sources'] as $_v['source']){?>			<tr><th><?php echo htmlspecialchars($_v['source']['kind'], ENT_QUOTES);?>
</th><td><svg class="ic"><use href="#i-file"/></svg> <?php echo htmlspecialchars($_v['source']['file'], ENT_QUOTES); if($_v['source']['command']){?>
 <span class="dim">runs with</span> <code><?php echo htmlspecialchars($_v['source']['command'], ENT_QUOTES);?>
 &lt;target&gt;</code><?php }?><?php if($_v['source']['shows']){ $_v['data_source'] = $_v['source']['shows'];?>
<button type="button" class="btn btn-sm show-source" data-source="<?php echo htmlspecialchars($_v['data_source']['folder'], ENT_QUOTES);?>
" data-file="<?php echo htmlspecialchars($_v['data_source']['file'], ENT_QUOTES);?>
" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span><?php echo htmlspecialchars($_v['data_source']['label'], ENT_QUOTES);?>
</span></button>
<?php }?></td></tr>
<?php }?><?php if($_v['data']['detail']['checkout']){?>
			<tr><th>Source</th><td><div class="mono-break"><?php echo htmlspecialchars($_v['data']['detail']['checkout']['remote'], ENT_QUOTES);?>
</div><div class="dim"><?php echo htmlspecialchars($_v['data']['detail']['checkout']['branch'], ENT_QUOTES);?>
 <span class="faint"><?php echo htmlspecialchars($_v['data']['detail']['checkout']['commit'], ENT_QUOTES);?>
</span> <?php echo htmlspecialchars($_v['data']['detail']['checkout']['subject'], ENT_QUOTES); if($_v['data']['detail']['checkout']['ago']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['data']['detail']['checkout']['ago'], ENT_QUOTES);?>
</span><?php }?></div><div class="faint">pulled before every run</div></td></tr>
<?php }?><?php if($_v['data']['detail']['space']){?>
			<tr><th>Checkout size</th><td><?php $_v['data_space'] = $_v['data']['detail']['space'];?>
<div class="spacebar" role="img" aria-label="<?php echo htmlspecialchars($_v['data_space']['total'], ENT_QUOTES);?>
 in all: <?php echo htmlspecialchars($_v['data_space']['files'], ENT_QUOTES);?>
 of files, <?php echo htmlspecialchars($_v['data_space']['history'], ENT_QUOTES);?>
 of history">
	<div class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_space']['filesPercent'], ENT_QUOTES);?>
%"></i><i class="history" style="width: <?php echo htmlspecialchars($_v['data_space']['historyPercent'], ENT_QUOTES);?>
%"></i></div>
	<div class="spacebar-legend">
		<span><i class="swatch files"></i>Files <b><?php echo htmlspecialchars($_v['data_space']['files'], ENT_QUOTES);?>
</b></span>
		<span><i class="swatch history"></i>History <b><?php echo htmlspecialchars($_v['data_space']['history'], ENT_QUOTES);?>
</b></span>
		<span class="faint"><?php echo htmlspecialchars($_v['data_space']['total'], ENT_QUOTES);?>
 in all<?php if($_v['data_space']['ignored']){?>
, and <?php echo htmlspecialchars($_v['data_space']['ignored'], ENT_QUOTES);?>
 git ignores<?php }?></span>
	</div>
</div>
			</td></tr>
<?php }?>			<tr><th>Who can run</th><td><?php echo htmlspecialchars($_v['data']['detail']['who'], ENT_QUOTES);?>
</td></tr>
			<tr><th>Hours</th><td><?php $_v['data_hours'] = $_v['data']['detail']['hours']; if($_v['data_hours']['text']){ echo htmlspecialchars($_v['data_hours']['line'], ENT_QUOTES);?>
 <?php if($_v['data_hours']['open']){?>
<span class="badge ok">open now</span><?php }else{?>
<span class="badge bad">out of hours now</span><?php }?><?php }else{?>
<span class="faint">always</span><?php }?>
</td></tr>
<?php if($_v['data']['detail']['includes']){?>
			<tr><th>Includes</th><td><?php if(!empty($_v['data']['detail']['includes']))foreach($_v['data']['detail']['includes'] as $_v['read']){?><div><svg class="ic"><use href="#i-file"/></svg> <?php echo htmlspecialchars($_v['read']['file'], ENT_QUOTES); $_v['data_source'] = $_v['read']['shows'];?>
<button type="button" class="btn btn-sm show-source" data-source="<?php echo htmlspecialchars($_v['data_source']['folder'], ENT_QUOTES);?>
" data-file="<?php echo htmlspecialchars($_v['data_source']['file'], ENT_QUOTES);?>
" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span><?php echo htmlspecialchars($_v['data_source']['label'], ENT_QUOTES);?>
</span></button>
</div><?php }?></td></tr>
<?php }?>			<tr><th>Scheduled</th><td><?php if($_v['data']['detail']['count']){ echo htmlspecialchars($_v['data']['detail']['count'], ENT_QUOTES);?>
 <a href="<?php echo htmlspecialchars($_v['data']['detail']['page'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['detail']['schedules'], ENT_QUOTES);?>
</a><?php }else{?>
<span class="faint">none</span>, <a href="<?php echo htmlspecialchars($_v['data']['detail']['page'], ENT_QUOTES);?>
">add one</a><?php }?></td></tr>
		</tbody>
	</table>
	</div>
</section>
</div>
<?php }?>
