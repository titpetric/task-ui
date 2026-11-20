<?php $_v=&$this->vars;?>
<div class="chat-shell thread-page">
	<section class="chat-pane" data-thread-pane>
<?php $_v['data_pane'] = $_v['data'];?>
<div class="conversation">
<a class="conversation-back" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads"><svg class="ic"><use href="#i-back"/></svg> All threads</a>
<?php $_v['data_head'] = $_v['data_pane']['head_view'];?>
<div class="conversation-head">
	<div class="conversation-title">
		<h2><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES);?>
</h2>
<?php if($_v['data_head']['meta']){?>
		<p class="faint"><?php echo htmlspecialchars($_v['data_head']['meta'], ENT_QUOTES);?>
</p>
<?php }?>	</div>
	<div class="actions">
<?php if($_v['data_head']['can_pick']){?>
		<form class="agent-pick" data-agent-form data-thread="<?php echo htmlspecialchars($_v['data_head']['id'], ENT_QUOTES);?>
">
			<label class="faint" for="thread-agent">Agent</label>
			<select id="thread-agent" name="agent">
<?php if(!empty($_v['data_head']['agents']))foreach($_v['data_head']['agents'] as $_v['name']){?>				<option value="<?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
"<?php if($_v['name'] == $_v['data_head']['agent']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
			<button class="btn" type="submit">Save</button>
		</form>
<?php }?>		<?php $_v['data_menu'] = $_v['data_head']['menu']; if(!$_v['data_menu']['empty']){?>
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

	</div>
</div>

<?php $_v['data_status'] = $_v['data_pane']['status']; if($_v['data_status']['working']){?>
<div class="conversation-status" data-status data-state="<?php echo htmlspecialchars($_v['data_status']['state'], ENT_QUOTES);?>
" data-started="<?php echo htmlspecialchars($_v['data_status']['started'], ENT_QUOTES);?>
">
	<p aria-live="polite"><i class="dot working"></i> <span data-status-text><?php echo htmlspecialchars($_v['data_status']['label'], ENT_QUOTES);?>
</span> <span class="faint" data-status-elapsed></span></p>
<?php if($_v['data_status']['can_end']){?>
	<div class="actions">
		<button class="btn btn-sm" type="button" data-stop="<?php echo htmlspecialchars($_v['data_status']['run'], ENT_QUOTES);?>
" title="Stop this run and let it settle">Stop</button>
		<button class="btn btn-sm btn-danger" type="button" data-kill="<?php echo htmlspecialchars($_v['data_status']['run'], ENT_QUOTES);?>
" title="Kill this run and everything it started">Kill</button>
	</div>
<?php }?></div>
<?php }?>


	<ol class="talk" data-talk data-thread="<?php echo htmlspecialchars($_v['data_pane']['id'], ENT_QUOTES);?>
" data-last-event="<?php echo htmlspecialchars($_v['data_pane']['last_event'], ENT_QUOTES);?>
"<?php if($_v['data_pane']['chat_thread']){?>
 data-chat-thread="<?php echo htmlspecialchars($_v['data_pane']['id'], ENT_QUOTES);?>
"<?php }?>>
<?php if(!empty($_v['data_pane']['messages']))foreach($_v['data_pane']['messages'] as $_v['message']){?>		<?php $_v['data_message'] = $_v['message'];?>
<li class="said<?php if($_v['data_message']['mine']){?>
 mine<?php }?><?php if($_v['data_message']['agent']){?>
 agent<?php }?><?php if($_v['data_message']['working']){?>
 working<?php }?>" id="note-<?php echo htmlspecialchars($_v['data_message']['id'], ENT_QUOTES);?>
">
	<header>
		<span class="author"><?php $_v['data_user'] = $_v['data_message']['author'];?>
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
		<span class="faint" title="<?php echo htmlspecialchars($_v['data_message']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_message']['ago'], ENT_QUOTES);?>
</span>
		<?php if($_v['data_message']['agent']){?>
			<span class="badge agent" title="The agent's answer">agent</span>
<?php }?><?php if($_v['data_message']['state']){?>
			<span class="badge note-state <?php echo htmlspecialchars($_v['data_message']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_message']['state'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_message']['run']){?>
			<a class="faint mono" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_message']['run'], ENT_QUOTES);?>
" title="The run this message comes from">run <?php echo htmlspecialchars($_v['data_message']['short'], ENT_QUOTES);?>
</a>
<?php }?>	</header>
	<div class="said-body prose"><?php echo $_v['data_message']['body'];?>
</div>
<?php if(!empty($_v['data_message']['work']))foreach($_v['data_message']['work'] as $_v['run']){?><?php $_v['data_work'] = $_v['run'];?>
<div class="work-run" data-run="<?php echo htmlspecialchars($_v['data_work']['id'], ENT_QUOTES);?>
">
<?php if($_v['data_work']['stats']){?>
	<div class="work-stats"><?php echo htmlspecialchars($_v['data_work']['stats'], ENT_QUOTES);?>
</div>
<?php }?><?php if(!empty($_v['data_work']['events']))foreach($_v['data_work']['events'] as $_v['event']){?><?php if($_v['event']['kind'] == 'text' || $_v['event']['kind'] == 'result'){?>
	<div class="work-text prose" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><?php echo $_v['event']['prose'];?>
</div>
<?php }elseif($_v['event']['kind'] == 'thinking'){?>
	<details class="work-thinking" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary>Thinking</summary><pre><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'tool_use'){?>
	<details class="work-tool" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary><code><?php echo htmlspecialchars($_v['event']['tool'], ENT_QUOTES);?>
</code><?php if($_v['event']['body']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</span><?php }?></summary><pre><?php echo htmlspecialchars($_v['event']['input'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'tool_result'){?>
	<details class="work-result<?php if($_v['event']['error']){?>
 error<?php }?>" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary><?php if($_v['event']['error']){?>
Error<?php }else{?>
Result<?php }?></summary><pre><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'system'){?>
	<details class="work-system" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary>Session</summary><pre><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'error'){?>
	<div class="work-error" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</div>
<?php }?><?php }?></div>

<?php }?><?php if($_v['data_message']['audio']){?>
	<audio class="said-audio" controls preload="none" src="<?php echo htmlspecialchars($_v['data_message']['audio'], ENT_QUOTES);?>
"></audio>
<?php }?></li>

<?php }?>	</ol>

<?php if($_v['data_pane']['replies']){ $_v['data_composer'] = $_v['data_pane']['composer'];?>
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

<?php }?></div>

	</section>
</div>
