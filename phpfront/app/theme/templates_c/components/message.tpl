<?php $_v=&$this->vars;?>
<li class="said<?php if($_v['data_message']['mine']){?>
 mine<?php }?><?php if($_v['data_message']['working']){?>
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
		<?php if($_v['data_message']['state']){?>
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
