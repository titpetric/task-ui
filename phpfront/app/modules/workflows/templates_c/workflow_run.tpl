<?php $_v=&$this->vars;?>
<section class="block" data-pipeline-view>
	<div class="page-head">
		<div>
			<h2><?php echo htmlspecialchars($_v['data']['flow']['name'], ENT_QUOTES);?>
</h2>
			<p class="faint"><span class="state"><i class="dot <?php echo htmlspecialchars($_v['data']['run']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data']['run']['state'], ENT_QUOTES);?>
</span><?php if($_v['data']['run']['message']){?>
: <?php echo htmlspecialchars($_v['data']['run']['message'], ENT_QUOTES);?>
;<?php }?> started <span title="<?php echo htmlspecialchars($_v['data']['run']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['run']['ago'], ENT_QUOTES);?>
</span> by <?php echo htmlspecialchars($_v['data']['run']['userName'], ENT_QUOTES); if($_v['data']['run']['finished']){?>
, took <?php echo htmlspecialchars($_v['data']['run']['took'], ENT_QUOTES); }?></p>
		</div>
	</div>
<?php $_v['data_thread'] = $_v['data']['thread'];?>

<div class="notes wf-thread" data-workflow-thread="<?php echo htmlspecialchars($_v['data_thread']['id'], ENT_QUOTES);?>
"<?php if($_v['data_thread']['running']){?>
 data-follow="<?php echo htmlspecialchars($_v['data_thread']['workflowId'], ENT_QUOTES);?>
" data-page="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_thread']['page'], ENT_QUOTES);?>
"<?php }?>>
	<article class="note instruction <?php echo htmlspecialchars($_v['data_thread']['noteState'], ENT_QUOTES);?>
">
		<header>
			<span class="author"><?php $_v['data_user'] = $_v['data_thread']['user'];?>
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
			<span class="faint" title="<?php echo htmlspecialchars($_v['data_thread']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_thread']['ago'], ENT_QUOTES);?>
</span>
			<span class="spacer"></span>
		</header>
<?php if($_v['data_thread']['question']){?>
		<div class="note-body"><?php echo $_v['data_thread']['question'];?>
</div>
<?php }else{?>
		<div class="note-body"><p class="faint">Ran <?php echo htmlspecialchars($_v['data_thread']['name'], ENT_QUOTES);?>
, with no input.</p></div>
<?php }?>		<ol class="replies">
			<li>
				<header>
<?php if($_v['data_thread']['agent']){?>
					<span class="author"><?php $_v['data_user'] = $_v['data_thread']['agent'];?>
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
<?php }else{?>
					<span class="author"><?php echo htmlspecialchars($_v['data_thread']['name'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_thread']['running']){?>
 <span class="badge note-state working">working</span>
<?php }else{?>
 <span class="faint" title="<?php echo htmlspecialchars($_v['data_thread']['doneStamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_thread']['doneAgo'], ENT_QUOTES);?>
</span> <span class="state"><i class="dot <?php echo htmlspecialchars($_v['data_thread']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data_thread']['state'], ENT_QUOTES);?>
</span>
<?php }?>				</header>
<?php if($_v['data_thread']['answer']){?>
				<div class="note-body"><?php echo $_v['data_thread']['answer'];?>
</div>
<?php }elseif($_v['data_thread']['running']){?>
				<div class="note-body"><p class="faint"><?php echo htmlspecialchars($_v['data_thread']['waiting'], ENT_QUOTES);?>
</p></div>
<?php }else{?>
				<div class="note-body"><p><?php echo htmlspecialchars($_v['data_thread']['ended'], ENT_QUOTES);?>
</p></div>
<?php }?>			</li>
		</ol>
	</article>
</div>

	<h3 class="wf-steps-title">Steps</h3>
	<ol class="pipeline">
<?php if(!empty($_v['data']['pipeline']))foreach($_v['data']['pipeline'] as $_v['data_node']){?><li class="node<?php if($_v['data_node']['state']){?>
 <?php echo htmlspecialchars($_v['data_node']['state'], ENT_QUOTES); }?><?php if($_v['data_node']['disabled']){?>
 disabled<?php }?>" data-step="<?php echo htmlspecialchars($_v['data_node']['pos'], ENT_QUOTES);?>
">
	<span class="rail"><i class="dot<?php if($_v['data_node']['state']){?>
 <?php echo htmlspecialchars($_v['data_node']['state'], ENT_QUOTES); }?>"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_node']['href'], ENT_QUOTES);?>
"<?php if($_v['data_node']['runId']){?>
 data-run-link="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['data_node']['workflow']){?>
<svg class="ic lead"><use href="#i-flow"/></svg><?php }elseif($_v['data_node']['summary']){?>
<svg class="ic lead"><use href="#i-summary"/></svg><?php }elseif($_v['data_node']['check']){?>
<svg class="ic lead"><use href="#i-rotate"/></svg><?php }elseif($_v['data_node']['prompt']){?>
<svg class="ic lead"><use href="#i-prompt"/></svg><?php }?><?php echo htmlspecialchars($_v['data_node']['text'], ENT_QUOTES);?>
</a>
<?php if($_v['data_node']['expand']){?>
			<button type="button" class="btn btn-sm" data-expand="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand<?php if($_v['data_node']['lines']){?>
 <?php echo htmlspecialchars($_v['data_node']['lines'], ENT_QUOTES);?>
 lines<?php }?></span></button>
<?php }?>		</div>
		<small><?php echo htmlspecialchars($_v['data_node']['small'], ENT_QUOTES);?>
</small>
<?php if($_v['data_node']['prompt']){?>
		<p class="step-prompt"><?php echo htmlspecialchars($_v['data_node']['prompt'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_node']['runId']){?>
		<a class="frame" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
"<?php if($_v['data_node']['live']){?>
 data-frame="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
"<?php }?> role="img" aria-label="Terminal of <?php echo htmlspecialchars($_v['data_node']['text'], ENT_QUOTES);?>
"><?php echo $_v['data_node']['frame'];?>
</a>
<?php }?><?php if($_v['data_node']['subId']){?>
		<ol class="pipeline sub" data-sub="<?php echo htmlspecialchars($_v['data_node']['subId'], ENT_QUOTES);?>
" data-workflow="<?php echo htmlspecialchars($_v['data_node']['workflowId'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_node']['sub']))foreach($_v['data_node']['sub'] as $_v['data_sub']){?><li class="node<?php if($_v['data_sub']['state']){?>
 <?php echo htmlspecialchars($_v['data_sub']['state'], ENT_QUOTES); }?><?php if($_v['data_sub']['disabled']){?>
 disabled<?php }?>" data-step="<?php echo htmlspecialchars($_v['data_sub']['pos'], ENT_QUOTES);?>
">
	<span class="rail"><i class="dot<?php if($_v['data_sub']['state']){?>
 <?php echo htmlspecialchars($_v['data_sub']['state'], ENT_QUOTES); }?>"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_sub']['href'], ENT_QUOTES);?>
"<?php if($_v['data_sub']['runId']){?>
 data-run-link="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['data_sub']['workflow']){?>
<svg class="ic lead"><use href="#i-flow"/></svg><?php }elseif($_v['data_sub']['summary']){?>
<svg class="ic lead"><use href="#i-summary"/></svg><?php }elseif($_v['data_sub']['check']){?>
<svg class="ic lead"><use href="#i-rotate"/></svg><?php }elseif($_v['data_sub']['prompt']){?>
<svg class="ic lead"><use href="#i-prompt"/></svg><?php }?><?php echo htmlspecialchars($_v['data_sub']['text'], ENT_QUOTES);?>
</a>
<?php if($_v['data_sub']['expand']){?>
			<button type="button" class="btn btn-sm" data-expand="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand<?php if($_v['data_sub']['lines']){?>
 <?php echo htmlspecialchars($_v['data_sub']['lines'], ENT_QUOTES);?>
 lines<?php }?></span></button>
<?php }?>		</div>
		<small><?php echo htmlspecialchars($_v['data_sub']['small'], ENT_QUOTES);?>
</small>
<?php if($_v['data_sub']['prompt']){?>
		<p class="step-prompt"><?php echo htmlspecialchars($_v['data_sub']['prompt'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_sub']['runId']){?>
		<a class="frame" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php if($_v['data_sub']['live']){?>
 data-frame="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php }?> role="img" aria-label="Terminal of <?php echo htmlspecialchars($_v['data_sub']['text'], ENT_QUOTES);?>
"><?php echo $_v['data_sub']['frame'];?>
</a>
<?php }?><?php if($_v['data_sub']['subId']){?>
		<ol class="pipeline sub" data-sub="<?php echo htmlspecialchars($_v['data_sub']['subId'], ENT_QUOTES);?>
" data-workflow="<?php echo htmlspecialchars($_v['data_sub']['workflowId'], ENT_QUOTES);?>
">
		</ol>
<?php }?>	</div>
</li>
<?php }?>		</ol>
<?php }?>	</div>
</li>
<?php }?>	</ol>
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
				<textarea name="body" rows="2" required data-mentions aria-label="A reply"></textarea>
				<div class="form-foot note"><span></span><button class="btn" type="submit">Reply</button></div>
			</form>
		</details>
	</article>
<?php }?>	</div>
</details>

<?php }?></section>
