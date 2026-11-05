<?php $_v=&$this->vars;?>
<div class="whofield<?php if($_v['data_whofield']['compact']){?>
 compact<?php }?>" data-who-empty="<?php echo htmlspecialchars($_v['data_whofield']['empty'], ENT_QUOTES);?>
">
	<div class="checks who-roles">
		<span class="faint">Roles</span><?php if(!empty($_v['data_whofield']['roles']))foreach($_v['data_whofield']['roles'] as $_v['role']){?>
		<label class="check"><input type="checkbox" name="<?php echo htmlspecialchars($_v['data_whofield']['name'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['role']['name'], ENT_QUOTES);?>
" data-type="<?php echo htmlspecialchars($_v['data_whofield']['type'], ENT_QUOTES);?>
"<?php if($_v['role']['checked']){?>
 checked<?php }?>> <?php echo htmlspecialchars($_v['role']['name'], ENT_QUOTES);?>
 <small class="count">· <?php echo htmlspecialchars($_v['role']['count'], ENT_QUOTES);?>
</small></label><?php }?>
	</div>
	<div class="checks who-users">
		<span class="faint">Users</span>
		<ul class="who-list"><?php if(!empty($_v['data_whofield']['users']))foreach($_v['data_whofield']['users'] as $_v['user']){?>
		<li><label class="check who-user" data-role="<?php echo htmlspecialchars($_v['user']['role'], ENT_QUOTES);?>
"><input type="checkbox" name="<?php echo htmlspecialchars($_v['data_whofield']['userName'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['user']['name'], ENT_QUOTES);?>
" data-type="<?php echo htmlspecialchars($_v['data_whofield']['type'], ENT_QUOTES);?>
"<?php if($_v['user']['checked']){?>
 checked<?php }?>> <?php echo htmlspecialchars($_v['user']['name'], ENT_QUOTES);?>
 <small class="role"><?php echo htmlspecialchars($_v['user']['role'], ENT_QUOTES);?>
</small> <small class="who-state"></small></label></li><?php }?><?php if(!empty($_v['data_whofield']['extra']))foreach($_v['data_whofield']['extra'] as $_v['other']){?>
		<li><label class="check who-user" data-role=""><input type="checkbox" name="<?php echo htmlspecialchars($_v['data_whofield']['userName'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['other'], ENT_QUOTES);?>
" data-type="<?php echo htmlspecialchars($_v['data_whofield']['type'], ENT_QUOTES);?>
" checked> <?php echo htmlspecialchars($_v['other'], ENT_QUOTES);?>
 <small class="role">no such user</small> <small class="who-state"></small></label></li><?php }?>
		</ul>
	</div><?php if(!$_v['data_whofield']['compact']){?>

	<p class="dim who-summary"></p><?php }?>
</div>