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

<?php if($_v['data']['off']){?>
<p class="empty-state">The Beszel integration isn't set up yet.<?php if($_v['data']['integrations']){?>
 Add its URL and token under <a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/integrations#beszel">Settings, Integrations</a>.<?php }?></p>
<?php }else{ if($_v['data']['form']){?>
<section class="block inventory-run">
	<h3>Run on the hosts</h3>
	<form data-inventory-run>
		<div class="run-commands">
<?php if(!empty($_v['data']['form']['commands']))foreach($_v['data']['form']['commands'] as $_v['command']){?>			<button type="button" class="btn btn-sm" data-command="<?php echo htmlspecialchars($_v['command'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['command'], ENT_QUOTES);?>
</button>
<?php }?>		</div>
		<textarea name="command" rows="2" required spellcheck="false" placeholder="a shell command to run on every picked host" aria-label="The command"></textarea>
		<div class="run-hosts" data-host-picker>
			<div class="host-tools">
				<span class="faint" data-host-count></span>
				<button type="button" class="btn btn-sm" data-hosts="all">All</button>
				<button type="button" class="btn btn-sm" data-hosts="none">None</button>
				<button type="button" class="btn btn-sm" data-hosts="invert">Invert</button>
			</div>
			<div class="host-picks">
<?php if(!empty($_v['data']['form']['picks']))foreach($_v['data']['form']['picks'] as $_v['pick']){?>				<label class="host-pick"><input type="checkbox" name="hosts" value="<?php echo htmlspecialchars($_v['pick']['host'], ENT_QUOTES);?>
" checked><span><?php echo htmlspecialchars($_v['pick']['login'], ENT_QUOTES);?>
</span></label>
<?php }?>			</div>
		</div>
		<div class="form-foot note"><span></span><span class="actions"><?php if($_v['data']['form']['summarize']){?>
<button class="btn" type="submit" name="summarize" value="1"><svg class="ic"><use href="#i-note"/></svg>Run and summarize</button><?php }?><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run</button></span></div>
	</form>
		<div class="inventory-results" data-inventory-results>
<?php if($_v['data']['batch']){?>
			<div class="block-head"<?php if($_v['data']['batch']['live']){?>
 data-batch-live<?php }?>>
				<h4>Your last run <small><code><?php echo htmlspecialchars($_v['data']['batch']['args'], ENT_QUOTES);?>
</code> on <?php echo htmlspecialchars($_v['data']['batch']['count'], ENT_QUOTES);?>
 host<?php if($_v['data']['batch']['count'] !== 1){?>
s<?php }?>, <?php echo htmlspecialchars($_v['data']['batch']['ago'], ENT_QUOTES);?>
</small></h4><?php if($_v['data']['batch']['id']){?>
<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs?batch=<?php echo htmlspecialchars($_v['data']['batch']['id'], ENT_QUOTES);?>
">In Runs</a><?php }?>
			</div>
			<div class="table-wrap">
			<table class="table run-results">
				<thead><tr><th>Host</th><th>Result</th><th>Output</th></tr></thead>
				<tbody>
<?php if(!empty($_v['data']['batch']['rows']))foreach($_v['data']['batch']['rows'] as $_v['run']){?>					<tr>
						<td class="nowrap"><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['run']['id'], ENT_QUOTES);?>
"><code><?php echo htmlspecialchars($_v['run']['target'], ENT_QUOTES);?>
</code></a></td>
						<td class="nowrap"><?php $_v['data_state'] = $_v['run']['state']; if($_v['data_state']['id']){?>
<span class="state" data-run-id="<?php echo htmlspecialchars($_v['data_state']['id'], ENT_QUOTES);?>
"<?php if($_v['data_state']['ended']){?>
 data-finished="true"<?php }?>><i class="dot <?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES); if($_v['data_state']['code'] !== ""){?>
 <span class="faint">(<?php echo htmlspecialchars($_v['data_state']['code'], ENT_QUOTES);?>
)</span><?php }?></span><?php }else{?>
<span class="faint">never run</span><?php }?>
 <span class="faint"><?php echo htmlspecialchars($_v['run']['duration'], ENT_QUOTES);?>
</span> <button type="button" class="btn btn-sm" data-attach="<?php echo htmlspecialchars($_v['run']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['run']['attach'], ENT_QUOTES);?>
</button></td>
						<td class="run-output"><?php if($_v['run']['queued']){?>
<span class="faint">Queued: waits for the run on <?php echo htmlspecialchars($_v['run']['target'], ENT_QUOTES);?>
 before it.</span><?php }elseif($_v['run']['running']){?>
<div class="frame" role="img" aria-label="Terminal of <?php echo htmlspecialchars($_v['run']['target'], ENT_QUOTES);?>
"><?php echo $_v['run']['frame'];?>
</div><?php }elseif($_v['run']['cast']){?>
<pre class="output" data-output="<?php echo htmlspecialchars($_v['run']['id'], ENT_QUOTES);?>
" data-cols="<?php echo htmlspecialchars($_v['run']['cols'], ENT_QUOTES);?>
" data-rows="<?php echo htmlspecialchars($_v['run']['rows'], ENT_QUOTES);?>
"><span class="faint">Reading the recording…</span></pre><?php }else{?>
<span class="faint">No recording.</span><?php }?></td>
					</tr>
<?php }?>				</tbody>
			</table>
			</div>
<?php }?>		</div>
<?php if($_v['data']['prompt']){ $_v['data_prompt'] = $_v['data']['prompt'];?>

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
<?php }?>		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="<?php echo htmlspecialchars($_v['data_prompt']['placeholder'], ENT_QUOTES);?>
" aria-label="The prompt"></textarea>
		<div class="form-foot note"><span class="faint run-in"><?php echo htmlspecialchars($_v['data_prompt']['lead'], ENT_QUOTES); if($_v['data_prompt']['options']['users']){?>
<select name="user" aria-label="Runs as"><?php $_v['data_options'] = $_v['data_prompt']['options'];if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
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
</span> <span class="faint" title="<?php echo htmlspecialchars($_v['reply']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['reply']['ago'], ENT_QUOTES);?>
</span><?php if($_v['reply']['edited']){?>
 <span class="faint edited" title="Edited <?php echo htmlspecialchars($_v['reply']['editedStamp'], ENT_QUOTES);?>
">edited</span><?php }?><?php if($_v['reply']['working']){?>
 <span class="badge note-state working">working</span><?php }?><?php if($_v['reply']['run']){?>
 <a class="faint mono" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['reply']['run'], ENT_QUOTES);?>
" title="The run this reply comes from">run <?php echo htmlspecialchars($_v['reply']['short'], ENT_QUOTES);?>
</a><?php }?><span class="spacer"></span><?php $_v['data_menu'] = $_v['reply']['menu'];if(!$_v['data_menu']['empty']){?>
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
				<textarea name="body" rows="2" required data-mentions aria-label="A reply"></textarea>
				<div class="form-foot note"><span></span><button class="btn" type="submit">Reply</button></div>
			</form>
		</details>
	</article>
<?php }?>	</div>
</details>
<?php }?></section>
<?php }?><section class="block" data-hosts-live>
<?php if($_v['data']['cluster']){?>
<div class="cluster">
	<h3 class="section-title">Cluster <small><?php echo htmlspecialchars($_v['data']['cluster']['count'], ENT_QUOTES);?>
 host<?php if($_v['data']['cluster']['count'] !== 1){?>
s<?php }?>, <?php echo htmlspecialchars($_v['data']['cluster']['up'], ENT_QUOTES);?>
 up<?php if($_v['data']['cluster']['down']){?>
, <span class="bad"><?php echo htmlspecialchars($_v['data']['cluster']['down'], ENT_QUOTES);?>
 down</span><?php }?></small></h3>
	<div class="pools">
<?php if(!empty($_v['data']['cluster']['pools']))foreach($_v['data']['cluster']['pools'] as $_v['pool']){?>		<div class="pool">
			<div class="pool-head">
				<span class="pool-name"><?php echo htmlspecialchars($_v['pool']['name'], ENT_QUOTES);?>
</span>
				<span class="pool-figure"><?php echo htmlspecialchars($_v['pool']['figure'], ENT_QUOTES);?>
</span><?php if($_v['pool']['share']){?>
<span class="pool-pct"><?php echo htmlspecialchars($_v['pool']['pct'], ENT_QUOTES);?>
%</span><?php }?>
			</div>
			<div class="pool-bar" role="img" aria-label="<?php echo htmlspecialchars($_v['pool']['name'], ENT_QUOTES);?>
: <?php echo htmlspecialchars($_v['pool']['figure'], ENT_QUOTES);?>
"><?php if(!empty($_v['pool']['segments']))foreach($_v['pool']['segments'] as $_v['segment']){?><span class="seg c<?php echo htmlspecialchars($_v['segment']['color'], ENT_QUOTES);?>
" style="flex-grow: <?php echo htmlspecialchars($_v['segment']['width'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['segment']['title'], ENT_QUOTES);?>
"><i style="width: <?php echo htmlspecialchars($_v['segment']['fill'], ENT_QUOTES);?>
%"></i></span><?php }?>
			</div>
		</div>
<?php }?>	</div>
	<ul class="pool-legend">
<?php if(!empty($_v['data']['cluster']['hosts']))foreach($_v['data']['cluster']['hosts'] as $_v['host']){?>		<li<?php if($_v['host']['left']){?>
 class="left"<?php }?>><i class="swatch<?php if(!$_v['host']['left']){?>
 c<?php echo htmlspecialchars($_v['host']['color'], ENT_QUOTES); }?>"></i><?php echo htmlspecialchars($_v['host']['name'], ENT_QUOTES);?>
</li>
<?php }?>	</ul>
<?php if($_v['data']['cluster']['left']){?>
	<p class="faint pool-note"><?php echo htmlspecialchars($_v['data']['cluster']['left'], ENT_QUOTES);?>
 host<?php if($_v['data']['cluster']['left'] !== 1){?>
s are down or send<?php }else{?>
 is down or sends<?php }?> no measures, greyed in the legend and left out of the pools.</p>
<?php }?><?php if($_v['data']['cluster']['charts']){?>
	<div class="charts"><?php if(!empty($_v['data']['cluster']['charts']))foreach($_v['data']['cluster']['charts'] as $_v['chart']){?><?php $_v['data_line_chart'] = $_v['chart'];?>

<figure class="chart">
	<figcaption class="chart-title"><?php echo htmlspecialchars($_v['data_line_chart']['title'], ENT_QUOTES);?>
</figcaption>
<?php if($_v['data_line_chart']['empty']){?>
	<p class="faint chart-empty">No records in the last 30 days.</p>
<?php }else{?>
	<?php echo $_v['data_line_chart']['svg'];?>

	<table class="chart-legend">
		<thead><tr><th></th><th>min</th><th>avg</th><th>max</th></tr></thead>
		<tbody>
<?php if(!empty($_v['data_line_chart']['legend']))foreach($_v['data_line_chart']['legend'] as $_v['row']){?>			<tr><td><i class="swatch s<?php echo htmlspecialchars($_v['row']['color'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['row']['min'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['row']['avg'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['row']['max'], ENT_QUOTES);?>
</td></tr>
<?php }?><?php if($_v['data_line_chart']['total']){?>
			<tr class="total"><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['name'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['min'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['avg'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['max'], ENT_QUOTES);?>
</td></tr>
<?php }?>		</tbody>
	</table>
<?php }?></figure>
<?php }?>	</div>
<?php }?></div>
<?php }?><h3>Hosts</h3>
<div class="table-wrap">
<table class="table hosts">
	<thead><tr><th>Host</th><?php if($_v['data']['logins']){?>
<th>SSH</th><?php }?><th class="center">Cores</th><th>CPU</th><th colspan="2">Memory <small>GB</small></th><th colspan="2">Disk <small>GB</small></th><th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
<?php if(!empty($_v['data']['rows']))foreach($_v['data']['rows'] as $_v['row']){?>		<tr<?php if($_v['row']['open']){?>
 class="host-row" data-host-row="<?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
" data-host-url="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
" tabindex="0" aria-expanded="false"<?php }?>>
			<td class="key nowrap" title="<?php echo htmlspecialchars($_v['row']['host'], ENT_QUOTES);?>
, <?php echo htmlspecialchars($_v['row']['status'], ENT_QUOTES);?>
"><?php if($_v['row']['open']){?>
<svg class="ic chev" aria-hidden="true"><use href="#i-chev"/></svg><?php }?><svg class="ic host-state <?php if($_v['row']['up']){?>
up<?php }else{?>
down<?php }?>" role="img" aria-label="<?php echo htmlspecialchars($_v['row']['status'], ENT_QUOTES);?>
"><use href="#i-server"/></svg> <?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</td>
<?php if($_v['data']['logins']){?>
			<td class="nowrap"><?php if($_v['row']['login']){ $_v['data_copy'] = $_v['row']['copy'];?>
<span class="copyable<?php if($_v['data_copy']['fill']){?>
 fill<?php }?>"><code<?php if($_v['data_copy']['fill'] || $_v['data_copy']['shown'] !== $_v['data_copy']['value']){?>
 title="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_copy']['shown'], ENT_QUOTES);?>
</code><button type="button" class="btn-icon" data-copy="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
" title="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
" aria-label="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-copy"/></svg></button></span>
<?php }else{?>
<span class="faint">-</span><?php }?></td>
<?php }?>			<td class="center"><?php if($_v['row']['cpus']){ echo htmlspecialchars($_v['row']['cpus'], ENT_QUOTES); }else{?>
<span class="faint">-</span><?php }?></td>
			<td><?php $_v['data_meter'] = $_v['row']['cpu'];?>

<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
			</td>
			<td><?php $_v['data_meter'] = $_v['row']['memory'];?>

<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
			</td>
			<td class="size faint nowrap"><?php if($_v['row']['measured']){ echo htmlspecialchars($_v['row']['memoryUsed'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['row']['memoryTotal'], ENT_QUOTES); }?></td>
			<td><?php $_v['data_meter'] = $_v['row']['disk'];?>

<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
			</td>
			<td class="size faint nowrap"><?php if($_v['row']['measured']){ echo htmlspecialchars($_v['row']['diskUsed'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['row']['diskTotal'], ENT_QUOTES); }?></td>
			<td class="actions"><?php $_v['data_menu'] = $_v['row']['menu'];if(!$_v['data_menu']['empty']){?>
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
		<tr class="empty"><td colspan="9">Beszel isn't monitoring any hosts.</td></tr>
<?php }?>	</tbody>
</table>
</div>
<p class="faint live-stamp">Updated <?php echo htmlspecialchars($_v['data']['now'], ENT_QUOTES);?>
</p>
</section>
<?php }?>
