<?php $_v=&$this->vars;?>
<div class="conversation">
<a class="conversation-back" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads"><i class="ic ic-back" aria-hidden="true"></i> All threads</a>
<?php $_v['data_head'] = $_v['data_pane']['head_view'];?>
<div class="conversation-head">
	<div class="conversation-title">
		<h2><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES);?>
</h2>
<?php if($_v['data_head']['meta'] || $_v['data_head']['usage']){?>
		<p class="faint"><?php echo htmlspecialchars($_v['data_head']['meta'], ENT_QUOTES); if($_v['data_head']['usage']){?>
 · <span title="What this thread read and wrote"><?php echo htmlspecialchars($_v['data_head']['usage'], ENT_QUOTES);?>
</span><?php }?></p>
<?php }?>	</div>
	<div class="actions">
<?php if($_v['data_head']['can_pick']){?>
		<form class="agent-pick" data-agent-form data-thread="<?php echo htmlspecialchars($_v['data_head']['id'], ENT_QUOTES);?>
">
			<label class="faint" for="thread-agent">Agent</label>
			<span class="agent-select"><select id="thread-agent" name="agent">
<?php if(!empty($_v['data_head']['agents']))foreach($_v['data_head']['agents'] as $_v['name']){?>				<option value="<?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
"<?php if($_v['name'] == $_v['data_head']['agent']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
</option>
<?php }?>			</select></span>
			<button class="btn" type="submit">Save</button>
		</form>
<?php }?>		<?php $_v['data_menu'] = $_v['data_head']['menu']; if(!$_v['data_menu']['empty']){?>
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

	</div>
</div>


	<ol class="talk" data-talk data-thread="<?php echo htmlspecialchars($_v['data_pane']['id'], ENT_QUOTES);?>
" data-last-event="<?php echo htmlspecialchars($_v['data_pane']['last_event'], ENT_QUOTES);?>
"<?php if($_v['data_pane']['chat_thread']){?>
 data-chat-thread="<?php echo htmlspecialchars($_v['data_pane']['id'], ENT_QUOTES);?>
"<?php }?>>
<?php if(!empty($_v['data_pane']['messages']))foreach($_v['data_pane']['messages'] as $_v['message']){?>		<?php $_v['data_message'] = $_v['message'];?>
<li class="<?php echo htmlspecialchars($_v['data_message']['className'], ENT_QUOTES);?>
" id="note-<?php echo htmlspecialchars($_v['data_message']['id'], ENT_QUOTES);?>
" data-run-id="<?php echo htmlspecialchars($_v['data_message']['run'], ENT_QUOTES);?>
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
<?php if($_v['data_message']['body']){?>
	<div class="said-body prose"><?php echo $_v['data_message']['body'];?>
</div>
<?php }?><?php if($_v['data_message']['working']){ $_v['data_working'] = $_v['data_message']['working_view'];?>
<div class="said-live" data-status data-state="<?php echo htmlspecialchars($_v['data_working']['state'], ENT_QUOTES);?>
" data-started="<?php echo htmlspecialchars($_v['data_working']['started'], ENT_QUOTES);?>
">
	<div class="<?php echo htmlspecialchars($_v['data_working']['progressClass'], ENT_QUOTES);?>
" data-work-progress><?php if($_v['data_working']['progressHtml']){ echo $_v['data_working']['progressHtml']; }?></div>
	<p class="said-state" aria-live="polite"><i class="dot working"></i> <span data-status-text><?php echo htmlspecialchars($_v['data_working']['label'], ENT_QUOTES);?>
</span> <span class="faint" data-status-elapsed></span><?php if($_v['data_working']['canEnd']){?>
<span class="actions"><button class="btn-icon" type="button" data-stop="<?php echo htmlspecialchars($_v['data_working']['run'], ENT_QUOTES);?>
" title="Stop this run and let it settle" aria-label="Stop this run"><i class="ic ic-stop" aria-hidden="true"></i></button><button class="btn-icon danger" type="button" data-kill="<?php echo htmlspecialchars($_v['data_working']['run'], ENT_QUOTES);?>
" title="Kill this run and everything it started" aria-label="Kill this run and everything it started"><i class="ic ic-trash" aria-hidden="true"></i></button></span><?php }?></p>
<?php if($_v['data_working']['agents']){?>
	<ul class="status-agents" data-status-agents>
<?php if(!empty($_v['data_working']['agents']))foreach($_v['data_working']['agents'] as $_v['agent']){?>		<li><details class="<?php echo htmlspecialchars($_v['agent']['className'], ENT_QUOTES);?>
" data-agent="<?php echo htmlspecialchars($_v['agent']['id'], ENT_QUOTES);?>
">
			<summary><i class="ic ic-users" aria-hidden="true"></i> <span><?php echo htmlspecialchars($_v['agent']['task'], ENT_QUOTES);?>
</span><?php if($_v['agent']['kind']){?>
 <small class="faint"><?php echo htmlspecialchars($_v['agent']['kind'], ENT_QUOTES);?>
</small><?php }?> <span class="<?php echo htmlspecialchars($_v['agent']['badge'], ENT_QUOTES);?>
" data-status-agent-state><?php echo htmlspecialchars($_v['agent']['state'], ENT_QUOTES);?>
</span></summary>
<?php if($_v['agent']['prompt']){?>
			<pre class="status-agent-prompt"><?php echo htmlspecialchars($_v['agent']['prompt'], ENT_QUOTES);?>
</pre>
<?php }?><?php if($_v['agent']['result']){?>
			<pre class="status-agent-result"><?php echo htmlspecialchars($_v['agent']['result'], ENT_QUOTES);?>
</pre>
<?php }?>		</details></li>
<?php }?>	</ul>
<?php }?>	<p class="said-live-foot"><span class="work-stats" data-work-stats data-tools="<?php echo htmlspecialchars($_v['data_working']['tools'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_working']['stats'], ENT_QUOTES);?>
</span></p>
</div>

<?php }else{ if(!empty($_v['data_message']['work']))foreach($_v['data_message']['work'] as $_v['run']){?><?php $_v['data_work'] = $_v['run']; if($_v['data_work']['stats'] || $_v['data_work']['trace']['states']){?>
<div class="work-run" data-run-id="<?php echo htmlspecialchars($_v['data_work']['id'], ENT_QUOTES);?>
">
<?php if($_v['data_work']['trace']['states']){?>
	<button type="button" class="work-trace" data-trace data-mode="duration" data-ms="<?php echo htmlspecialchars($_v['data_work']['trace']['ms'], ENT_QUOTES);?>
" data-count="<?php echo htmlspecialchars($_v['data_work']['trace']['count'], ENT_QUOTES);?>
" title="Switch the bar between duration, count and time">
		<span class="work-trace-mode" data-trace-mode>Duration</span>
		<span class="work-trace-bar"><?php if(!empty($_v['data_work']['trace']['states']))foreach($_v['data_work']['trace']['states'] as $_v['state']){?><span class="work-trace-state <?php echo htmlspecialchars($_v['state']['key'], ENT_QUOTES);?>
" style="width:<?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
%" data-ms="<?php echo htmlspecialchars($_v['state']['ms'], ENT_QUOTES);?>
" data-count="<?php echo htmlspecialchars($_v['state']['count'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['state']['label'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['state']['label'], ENT_QUOTES);?>
: <?php echo htmlspecialchars($_v['state']['span'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
%</span><?php }?></span>
	</button>
<?php }?><?php if($_v['data_work']['stats']){?>
	<div class="work-stats"><?php echo htmlspecialchars($_v['data_work']['stats'], ENT_QUOTES);?>
</div>
<?php }?></div><?php }}?><?php }?><?php if($_v['data_message']['audio']){?>
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
" title="Record a message" aria-label="Record a message"><i class="ic ic-mic" aria-hidden="true"></i></button>
<?php }?>			<button class="btn btn-primary" type="submit" title="Enter sends, Shift+Enter adds a line">Send</button>
		</div>
	</div>
</form>

<?php }?><button type="button" class="scroll-latest" data-scroll-latest title="Scroll to the latest" aria-label="Scroll to the latest" hidden><i class="ic ic-down" aria-hidden="true"></i></button></div>
