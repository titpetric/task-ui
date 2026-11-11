<?php $_v=&$this->vars;?>
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
