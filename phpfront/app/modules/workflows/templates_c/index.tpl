<?php $_v=&$this->vars; $_v['data_head'] = $_v['data']['head'];?>

<div class="page-head">
	<div>
		<h2><?php if($_v['data_head']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['data_head']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i> <?php }?><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES); if($_v['data_head']['badge']){?>
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
"<?php }?>><i class="ic ic-play" aria-hidden="true"></i>Run<?php if($_v['data_run']['takesArgs']){?>
...<?php }?></button>
<?php if($_v['data_run']['menu']){?>
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
" aria-label="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
"><i class="ic ic-down" aria-hidden="true"></i></summary>
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
"<?php }?> title="Run and summarize"><i class="ic ic-summary" aria-hidden="true"></i><span>Run and summarize</span></button>
<?php }?><?php if($_v['data_run']['force']){?>
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['place'], ENT_QUOTES);?>
" title="Run out of hours, with a reason"><i class="ic ic-play" aria-hidden="true"></i><span>Run force</span></button>
<?php }?>			</div>
		</div>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data_head']['buttons']['items']){?>
		<?php $_v['data_buttons'] = $_v['data_head']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }}?><?php if($_v['data_head']['actions']){?>
		<?php $_v['data_menu'] = $_v['data_head']['actions'];?>

		<?php if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><i class="ic ic-more" aria-hidden="true"></i></summary>
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
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }}?>	</div>
<?php }?></div>


<?php if($_v['data']['prompt']){ $_v['data_prompt'] = $_v['data']['prompt'];?>

<details class="promptbox"<?php if($_v['data_prompt']['open']){?>
 open<?php }?><?php if($_v['data_prompt']['hidden']){?>
 hidden<?php }?>>
	<summary><h3 class="section-title"><i class="ic ic-chev chev" aria-hidden="true"></i><?php echo htmlspecialchars($_v['data_prompt']['title'], ENT_QUOTES);?>
 <small><?php echo htmlspecialchars($_v['data_prompt']['note'], ENT_QUOTES);?>
</small></h3></summary>
<?php if(!$_v['data_prompt']['watch']){?>
	<?php $_v['data_composer'] = $_v['data_prompt']['composer'];?>
<form class="composer" data-send="<?php echo htmlspecialchars($_v['data_composer']['mode'], ENT_QUOTES);?>
"<?php if($_v['data_composer']['url']){?>
 data-send-url="<?php echo htmlspecialchars($_v['data_composer']['url'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_composer']['next']){?>
 data-next="<?php echo htmlspecialchars($_v['data_composer']['next'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_composer']['pane']){?>
 data-pane<?php }?>>
<?php if($_v['data_composer']['parent']){?>
	<input type="hidden" name="parent" value="<?php echo htmlspecialchars($_v['data_composer']['parent'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['channel']){?>
	<input type="hidden" name="channel" value="<?php echo htmlspecialchars($_v['data_composer']['channel'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['folder']){?>
	<input type="hidden" name="folder" value="<?php echo htmlspecialchars($_v['data_composer']['folder'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['dir']){?>
	<input type="hidden" name="dir" value="<?php echo htmlspecialchars($_v['data_composer']['dir'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['run']){?>
	<input type="hidden" name="run" value="<?php echo htmlspecialchars($_v['data_composer']['run'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['area']){?>
	<input type="hidden" name="area" value="<?php echo htmlspecialchars($_v['data_composer']['area'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['origin']){?>
	<input type="hidden" name="origin" value="<?php echo htmlspecialchars($_v['data_composer']['origin'], ENT_QUOTES);?>
">
<?php }?>	<div class="composer-box">
		<?php $_v['data_markdown'] = $_v['data_composer']['editor'];?>
<textarea name="<?php echo htmlspecialchars($_v['data_markdown']['name'], ENT_QUOTES);?>
" rows="<?php echo htmlspecialchars($_v['data_markdown']['rows'], ENT_QUOTES);?>
" spellcheck="false" data-markdown data-toolbar="<?php echo htmlspecialchars($_v['data_markdown']['toolbar'], ENT_QUOTES);?>
"<?php if($_v['data_markdown']['mentions']){?>
 data-mentions<?php }?><?php if($_v['data_markdown']['placeholder']){?>
 placeholder="<?php echo htmlspecialchars($_v['data_markdown']['placeholder'], ENT_QUOTES);?>
"<?php }?> aria-label="<?php echo htmlspecialchars($_v['data_markdown']['label'], ENT_QUOTES);?>
"></textarea>
		<div class="composer-foot">
<?php if($_v['data_composer']['modes']){?>
			<div class="composer-modes" role="group" aria-label="What to send">
<?php if(!empty($_v['data_composer']['modes']))foreach($_v['data_composer']['modes'] as $_v['mode']){?>				<button type="button" class="btn btn-sm" data-composer-mode="<?php echo htmlspecialchars($_v['mode']['key'], ENT_QUOTES);?>
" aria-pressed="<?php if($_v['mode']['key'] == $_v['data_composer']['mode']){?>
true<?php }else{?>
false<?php }?>"><?php echo htmlspecialchars($_v['mode']['label'], ENT_QUOTES);?>
</button>
<?php }?>			</div>
<?php }?><?php if($_v['data_composer']['agents']){?>
			<label class="composer-field" data-composer-agent>Agent
				<select name="agent" aria-label="Answers as"><?php if(!empty($_v['data_composer']['agents']))foreach($_v['data_composer']['agents'] as $_v['name']){?><option value="<?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
"<?php if($_v['name'] == $_v['data_composer']['agent']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
</option><?php }?></select>
			</label>
<?php }?><?php if($_v['data_composer']['show_place']){?>
			<label class="composer-field" data-composer-place>In
				<input type="text" name="place" placeholder="folder/dir, empty is the root" spellcheck="false" autocomplete="off"<?php if($_v['data_composer']['places']){?>
 list="composer-places"<?php }?>>
<?php if($_v['data_composer']['places']){?>
				<datalist id="composer-places"><?php if(!empty($_v['data_composer']['places']))foreach($_v['data_composer']['places'] as $_v['place']){?><option value="<?php echo htmlspecialchars($_v['place'], ENT_QUOTES);?>
"><?php }?></datalist>
<?php }?>			</label>
<?php }?>			<span class="composer-space"></span>
<?php if($_v['data_composer']['mic']){?>
			<button class="btn btn-icon" type="button" data-voice-reply data-parent="<?php echo htmlspecialchars($_v['data_composer']['parent'], ENT_QUOTES);?>
" data-channel="<?php echo htmlspecialchars($_v['data_composer']['channel'], ENT_QUOTES);?>
" title="Record a message" aria-label="Record a message"><i class="ic ic-mic" aria-hidden="true"></i></button>
<?php }?>			<button class="btn btn-primary" type="submit" title="Enter sends, Shift+Enter adds a line">Send</button>
		</div>
	</div>
</form>

<?php }?><?php if($_v['data_prompt']['threads']){?>
	<ul class="thread-links" data-notes>
<?php if(!empty($_v['data_prompt']['threads']))foreach($_v['data_prompt']['threads'] as $_v['thread']){?>		<li><a href="<?php echo htmlspecialchars($_v['thread']['href'], ENT_QUOTES);?>
"><i class="ic ic-note" aria-hidden="true"></i> <?php echo htmlspecialchars($_v['thread']['title'], ENT_QUOTES);?>
</a><?php if($_v['thread']['state']){?>
 <span class="badge note-state <?php echo htmlspecialchars($_v['thread']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['thread']['state'], ENT_QUOTES);?>
</span><?php }?></li>
<?php }?>	</ul>
<?php }?></details>
<p class="faint threads-scope"><a class="threads-link" href="<?php echo htmlspecialchars($_v['data_prompt']['threads_href'], ENT_QUOTES);?>
">Threads in this place</a></p>

<?php }?>
<section class="block">
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
 <?php echo $_v['button']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }?>
	</div>
<?php }?></div>

<?php if($_v['data']['rows']){?>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Workflow</th><th>Last run</th><th>By</th><th></th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['rows']))foreach($_v['data']['rows'] as $_v['row']){?>			<tr>
				<td class="key"><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/workflows/<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</a><?php if($_v['row']['description']){?>
<small class="faint wf-desc"><?php echo $_v['row']['description'];?>
</small><?php }?><?php if(!empty($_v['row']['schedules']))foreach($_v['row']['schedules'] as $_v['hint']){?><small class="faint wf-desc" title="<?php echo htmlspecialchars($_v['hint']['title'], ENT_QUOTES);?>
"><i class="ic ic-clock lead" aria-hidden="true"></i><code><?php echo htmlspecialchars($_v['hint']['expression'], ENT_QUOTES);?>
</code><?php if(!$_v['hint']['enabled']){?>
 disabled<?php }?></small><?php }?></td>
				<td class="nowrap" data-label="Last run"><?php if($_v['row']['running']){?>
<span class="state"><i class="dot running"></i>running</span><?php }elseif($_v['row']['last']){?>
<span class="state"><i class="dot <?php echo htmlspecialchars($_v['row']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['row']['state'], ENT_QUOTES);?>
</span> <span class="faint" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</span><?php }else{?>
<span class="faint">never run</span><?php }?></td>
				<td data-label="By"><?php $_v['data_user'] = $_v['row']['user'];?>
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
<?php if($_v['row']['runAs']){?>
 <span class="faint nowrap">· runs as <?php echo htmlspecialchars($_v['row']['runAs'], ENT_QUOTES);?>
</span><?php }?></td>
				<td class="actions"><?php if($_v['row']['canRun']){ if($_v['row']['running']){?>

					<button type="button" class="btn btn-sm" data-workflow-stop="<?php echo htmlspecialchars($_v['row']['running'], ENT_QUOTES);?>
"><i class="ic ic-stop" aria-hidden="true"></i>Stop</button><?php }else{?>

					<button type="button" class="btn btn-primary btn-sm" data-workflow-run="<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
"<?php if($_v['row']['prompted']){?>
 data-workflow-prompt="<?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
"<?php }?>><i class="ic ic-play" aria-hidden="true"></i>Run</button><?php }?><?php }elseif($_v['row']['runDenied']){?>

					<span class="faint run-denied"><?php echo htmlspecialchars($_v['row']['runDenied'], ENT_QUOTES);?>
</span><?php }?><?php $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><i class="ic ic-more" aria-hidden="true"></i></summary>
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
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
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
<?php }?></section>
<section class="block">
<?php $_v['data_block'] = $_v['data']['runsBlock'];?>

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
 <?php echo $_v['button']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }?>
	</div>
<?php }?></div>

<?php if($_v['data']['runs']){?>
	<div class="table-wrap">
	<table class="table cards wf-runs">
		<thead><tr><th>Run</th><th>By</th><th>Started</th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['runs']))foreach($_v['data']['runs'] as $_v['run']){?>			<tr data-run-id="<?php echo htmlspecialchars($_v['run']['id'], ENT_QUOTES);?>
">
				<td class="key"><?php $_v['data_wffold'] = $_v['run']['fold']; if($_v['data_wffold']['answer']){?>
<button type="button" class="btn-icon wf-fold" data-answer-toggle aria-expanded="<?php if($_v['data_wffold']['open']){?>
true<?php }else{?>
false<?php }?>" aria-controls="wf-answer-<?php echo htmlspecialchars($_v['data_wffold']['id'], ENT_QUOTES);?>
" title="<?php if($_v['data_wffold']['open']){?>
Hide<?php }else{?>
Show<?php }?>" aria-label="<?php if($_v['data_wffold']['open']){?>
Hide<?php }else{?>
Show<?php }?>"><i class="ic ic-chev chev" aria-hidden="true"></i></button><?php }else{?>
<span class="wf-fold" aria-hidden="true"></span><?php }?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/workflows/<?php echo htmlspecialchars($_v['run']['workflowId'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['run']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['run']['title'], ENT_QUOTES);?>
"><i class="dot <?php echo htmlspecialchars($_v['run']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['run']['name'], ENT_QUOTES);?>
</a></td>
				<td data-label="By"><?php $_v['data_user'] = $_v['run']['user'];?>
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
</td>
				<td class="nowrap" data-label="Started" title="<?php echo htmlspecialchars($_v['run']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['run']['ago'], ENT_QUOTES);?>
</td>
			</tr>
<?php $_v['data_wfanswer'] = $_v['run']['answer'];if($_v['data_wfanswer']['body']){?>
			<tr class="wf-answer-row" id="wf-answer-<?php echo htmlspecialchars($_v['data_wfanswer']['id'], ENT_QUOTES);?>
" data-answer-of="<?php echo htmlspecialchars($_v['data_wfanswer']['id'], ENT_QUOTES);?>
"<?php if(!$_v['data_wfanswer']['open']){?>
 hidden<?php }?>><td colspan="<?php echo htmlspecialchars($_v['data_wfanswer']['cols'], ENT_QUOTES);?>
"><div class="note-body"><?php echo $_v['data_wfanswer']['body'];?>
</div></td></tr>
<?php }}?>		</tbody>
	</table>
	</div>
<?php }?></section>
