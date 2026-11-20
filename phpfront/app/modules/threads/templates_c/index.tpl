<?php $_v=&$this->vars;?>
<div class="chat-shell">
	<aside class="chat-list">
		<form method="get" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/threads" class="filters">
			<label>Words <input type="search" name="q" value="<?php echo htmlspecialchars($_v['data']['form']['q'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off"></label>
			<label>Place <input type="text" name="place" value="<?php echo htmlspecialchars($_v['data']['form']['place'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off"></label>
			<label for="threads-state">State
				<select id="threads-state" name="state">
					<option value="">any state</option>
<?php if(!empty($_v['data']['states']))foreach($_v['data']['states'] as $_v['state']){?><option value="<?php echo htmlspecialchars($_v['state'], ENT_QUOTES);?>
"<?php if($_v['data']['form']['state'] === $_v['state']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['state'], ENT_QUOTES);?>
</option>
<?php }?>				</select>
			</label>
			<label>By <input type="text" name="user" value="<?php echo htmlspecialchars($_v['data']['form']['user'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off"></label>
			<button class="btn" type="submit">Filter</button>
		</form>

<?php if($_v['data']['chats']){?>

		<h3 class="section-title">Chats</h3>
		<ol class="timeline">
<?php if(!empty($_v['data']['chats']))foreach($_v['data']['chats'] as $_v['row']){?>			<?php $_v['data_row'] = $_v['row'];?>
<li class="thread-row<?php if($_v['data_row']['state']){?>
 <?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES); }?><?php if($_v['data_row']['unread']){?>
 unread<?php }?><?php if($_v['data_row']['selected']){?>
 selected<?php }?>">
	<a class="thread-open" href="<?php echo htmlspecialchars($_v['data_row']['href'], ENT_QUOTES);?>
" data-thread-id="<?php echo htmlspecialchars($_v['data_row']['id'], ENT_QUOTES);?>
"<?php if($_v['data_row']['selected']){?>
 aria-current="true"<?php }?>>
		<span class="thread-said"><?php echo htmlspecialchars($_v['data_row']['said'], ENT_QUOTES);?>
</span>
		<span class="thread-meta">
			<svg class="ic thread-kind" aria-label="<?php echo htmlspecialchars($_v['data_row']['kind'], ENT_QUOTES);?>
"><use href="#i-<?php if($_v['data_row']['kind'] == 'voice'){?>
mic<?php }elseif($_v['data_row']['kind'] == 'prompt'){?>
prompt<?php }else{?>
note<?php }?>"/></svg>
			<span class="faint" title="<?php echo htmlspecialchars($_v['data_row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_row']['ago'], ENT_QUOTES);?>
</span>
			<?php if($_v['data_row']['place']){?>
<span class="faint mono"><?php echo htmlspecialchars($_v['data_row']['place'], ENT_QUOTES);?>
</span><?php }?>
			<?php if($_v['data_row']['messages']){?>
<span class="faint"><?php echo htmlspecialchars($_v['data_row']['messages'], ENT_QUOTES);?>
 messages</span><?php }?>
			<?php if($_v['data_row']['heard']){?>
<svg class="ic" aria-label="Recorded"><use href="#i-play"/></svg><?php }?>
			<?php if($_v['data_row']['unread']){?>
<span class="badge unread"><?php echo htmlspecialchars($_v['data_row']['unread'], ENT_QUOTES);?>
</span><?php }?>
			<?php if($_v['data_row']['state']){?>
<span class="badge note-state <?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES);?>
</span><?php }?>
		</span>
	</a>
	<span class="thread-who"><?php $_v['data_user'] = $_v['data_row']['person'];?>
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
</li>

<?php }?>		</ol>
<?php }?>

		<h3 class="section-title">Threads</h3>
		<ol class="timeline" data-notes>
<?php if(!empty($_v['data']['rows']))foreach($_v['data']['rows'] as $_v['row']){?>			<?php $_v['data_row'] = $_v['row'];?>
<li class="thread-row<?php if($_v['data_row']['state']){?>
 <?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES); }?><?php if($_v['data_row']['unread']){?>
 unread<?php }?><?php if($_v['data_row']['selected']){?>
 selected<?php }?>">
	<a class="thread-open" href="<?php echo htmlspecialchars($_v['data_row']['href'], ENT_QUOTES);?>
" data-thread-id="<?php echo htmlspecialchars($_v['data_row']['id'], ENT_QUOTES);?>
"<?php if($_v['data_row']['selected']){?>
 aria-current="true"<?php }?>>
		<span class="thread-said"><?php echo htmlspecialchars($_v['data_row']['said'], ENT_QUOTES);?>
</span>
		<span class="thread-meta">
			<svg class="ic thread-kind" aria-label="<?php echo htmlspecialchars($_v['data_row']['kind'], ENT_QUOTES);?>
"><use href="#i-<?php if($_v['data_row']['kind'] == 'voice'){?>
mic<?php }elseif($_v['data_row']['kind'] == 'prompt'){?>
prompt<?php }else{?>
note<?php }?>"/></svg>
			<span class="faint" title="<?php echo htmlspecialchars($_v['data_row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_row']['ago'], ENT_QUOTES);?>
</span>
			<?php if($_v['data_row']['place']){?>
<span class="faint mono"><?php echo htmlspecialchars($_v['data_row']['place'], ENT_QUOTES);?>
</span><?php }?>
			<?php if($_v['data_row']['messages']){?>
<span class="faint"><?php echo htmlspecialchars($_v['data_row']['messages'], ENT_QUOTES);?>
 messages</span><?php }?>
			<?php if($_v['data_row']['heard']){?>
<svg class="ic" aria-label="Recorded"><use href="#i-play"/></svg><?php }?>
			<?php if($_v['data_row']['unread']){?>
<span class="badge unread"><?php echo htmlspecialchars($_v['data_row']['unread'], ENT_QUOTES);?>
</span><?php }?>
			<?php if($_v['data_row']['state']){?>
<span class="badge note-state <?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES);?>
</span><?php }?>
		</span>
	</a>
	<span class="thread-who"><?php $_v['data_user'] = $_v['data_row']['person'];?>
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
</li>

<?php }else{?>
			<?php $_v['data_empty'] = $_v['data']['empty']; if($_v['data_empty']['tag'] == 'li'){?>
<li class="empty-state" role="status"><?php echo htmlspecialchars($_v['data_empty']['text'], ENT_QUOTES);?>
</li><?php }else{?>
<p class="empty-state" role="status"><?php echo htmlspecialchars($_v['data_empty']['text'], ENT_QUOTES); if($_v['data_empty']['href']){?>
 <a href="<?php echo htmlspecialchars($_v['data_empty']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_empty']['label'], ENT_QUOTES);?>
</a><?php }?></p><?php }}?>		</ol>
	</aside>

	<section class="chat-pane" data-thread-pane>
		<div class="conversation">
			<div class="chat-blank">
				<h2><?php echo htmlspecialchars($_v['data']['head']['title'], ENT_QUOTES);?>
</h2>
				<p><?php echo $_v['data']['head']['lead'];?>
</p>
				<?php $_v['data_composer'] = $_v['data']['composer'];?>
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
				<input type="text" name="place" placeholder="folder/dir, empty is the root" spellcheck="false" autocomplete="off">
			</label>
<?php }?>			<span class="composer-space"></span>
<?php if($_v['data_composer']['mic']){?>
			<button class="btn btn-icon" type="button" data-voice-reply data-parent="<?php echo htmlspecialchars($_v['data_composer']['parent'], ENT_QUOTES);?>
" data-channel="<?php echo htmlspecialchars($_v['data_composer']['channel'], ENT_QUOTES);?>
" title="Record a message" aria-label="Record a message"><svg class="ic"><use href="#i-mic"/></svg></button>
<?php }?>			<button class="btn btn-primary" type="submit">Send</button>
		</div>
	</div>
</form>

			</div>
		</div>
	</section>
</div>
