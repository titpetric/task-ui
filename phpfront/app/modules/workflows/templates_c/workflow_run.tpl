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
" title="Record a message" aria-label="Record a message"><svg class="ic"><use href="#i-mic"/></svg></button>
<?php }?>			<button class="btn btn-primary" type="submit" title="Enter sends, Shift+Enter adds a line">Send</button>
		</div>
	</div>
</form>

<?php }?><?php if($_v['data_prompt']['threads']){?>
	<ul class="thread-links">
<?php if(!empty($_v['data_prompt']['threads']))foreach($_v['data_prompt']['threads'] as $_v['thread']){?>		<li><a href="<?php echo htmlspecialchars($_v['thread']['href'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-note"/></svg> <?php echo htmlspecialchars($_v['thread']['title'], ENT_QUOTES);?>
</a><?php if($_v['thread']['state']){?>
 <span class="badge note-state <?php echo htmlspecialchars($_v['thread']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['thread']['state'], ENT_QUOTES);?>
</span><?php }?></li>
<?php }?>	</ul>
<?php }?></details>
<p class="faint threads-scope"><a class="threads-link" href="<?php echo htmlspecialchars($_v['data_prompt']['threads_href'], ENT_QUOTES);?>
">Threads in this place</a></p>

<?php }?></section>
