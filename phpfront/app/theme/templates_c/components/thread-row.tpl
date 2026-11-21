<?php $_v=&$this->vars;?>
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
			<i class="ic ic-<?php if($_v['data_row']['kind'] == 'voice'){?>
mic<?php }elseif($_v['data_row']['kind'] == 'prompt'){?>
prompt<?php }else{?>
note<?php }?> thread-kind" role="img" aria-label="<?php echo htmlspecialchars($_v['data_row']['kind'], ENT_QUOTES);?>
"></i>
			<span class="faint" title="<?php echo htmlspecialchars($_v['data_row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_row']['ago'], ENT_QUOTES);?>
</span>
			<?php if($_v['data_row']['place']){?>
<span class="faint mono thread-place"><?php echo htmlspecialchars($_v['data_row']['place'], ENT_QUOTES);?>
</span><?php }?>
			<?php if($_v['data_row']['messages'] > 1 && !$_v['data_row']['can_end']){?>
<span class="faint thread-count" title="<?php echo htmlspecialchars($_v['data_row']['messages'], ENT_QUOTES);?>
 messages" aria-label="<?php echo htmlspecialchars($_v['data_row']['messages'], ENT_QUOTES);?>
 messages"><?php echo htmlspecialchars($_v['data_row']['messages'], ENT_QUOTES);?>
</span><?php }?>
			<?php if($_v['data_row']['heard']){?>
<i class="ic ic-play" role="img" aria-label="Recorded"></i><?php }?>
			<?php if($_v['data_row']['unread']){?>
<span class="badge unread"><?php echo htmlspecialchars($_v['data_row']['unread'], ENT_QUOTES);?>
</span><?php }?>
			<?php if($_v['data_row']['state']){?>
<span class="badge note-state <?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_row']['state'], ENT_QUOTES);?>
</span><?php }?>
		</span>
	</a>
<?php if($_v['data_row']['can_end']){?>
	<button class="btn-icon thread-stop" type="button" data-stop="<?php echo htmlspecialchars($_v['data_row']['run'], ENT_QUOTES);?>
" title="Stop the run behind this thread" aria-label="Stop the run behind this thread"><i class="ic ic-stop" aria-hidden="true"></i></button>
<?php }?>	<span class="thread-who"><?php $_v['data_user'] = $_v['data_row']['person'];?>
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
