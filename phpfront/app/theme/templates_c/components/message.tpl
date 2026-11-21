<?php $_v=&$this->vars;?>
<li class="said<?php if($_v['data_message']['mine']){?>
 mine<?php }?><?php if($_v['data_message']['agent']){?>
 agent<?php }?><?php if($_v['data_message']['working']){?>
 working<?php }?>" id="note-<?php echo htmlspecialchars($_v['data_message']['id'], ENT_QUOTES);?>
" data-run="<?php echo htmlspecialchars($_v['data_message']['run'], ENT_QUOTES);?>
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
<?php }?><?php if(!empty($_v['data_message']['work']))foreach($_v['data_message']['work'] as $_v['run']){?><?php $_v['data_work'] = $_v['run']; if($_v['data_work']['stats']){?>
<div class="work-run" data-run="<?php echo htmlspecialchars($_v['data_work']['id'], ENT_QUOTES);?>
">
	<div class="work-stats"><?php echo htmlspecialchars($_v['data_work']['stats'], ENT_QUOTES);?>
</div>
</div><?php } }?><?php if($_v['data_message']['audio']){?>
	<audio class="said-audio" controls preload="none" src="<?php echo htmlspecialchars($_v['data_message']['audio'], ENT_QUOTES);?>
"></audio>
<?php }?></li>
