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
		<div class="note-body prose"><?php echo $_v['data_note']['body'];?>
</div>
<?php if($_v['data_note']['replies']){?>
		<ol class="replies">
<?php if(!empty($_v['data_note']['replies']))foreach($_v['data_note']['replies'] as $_v['reply']){?>			<li<?php if($_v['reply']['working']){?>
 class="rerun"<?php }?>>
				<header><span class="author"><?php $_v['data_user'] = $_v['reply']['author'];?>
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
				<div class="note-body prose"><?php echo $_v['reply']['body'];?>
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
"><svg class="ic lead"><use href="#i-clock"/></svg><code><?php echo htmlspecialchars($_v['hint']['expression'], ENT_QUOTES);?>
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
"><svg class="ic"><use href="#i-stop"/></svg>Stop</button><?php }else{?>

					<button type="button" class="btn btn-primary btn-sm" data-workflow-run="<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
"<?php if($_v['row']['prompted']){?>
 data-workflow-prompt="<?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
"<?php }?>><svg class="ic"><use href="#i-play"/></svg>Run</button><?php }?><?php }elseif($_v['row']['runDenied']){?>

					<span class="faint run-denied"><?php echo htmlspecialchars($_v['row']['runDenied'], ENT_QUOTES);?>
</span><?php }?><?php $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
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
Show<?php }?>"><svg class="ic chev"><use href="#i-chev"/></svg></button><?php }else{?>
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
