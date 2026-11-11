<?php $_v=&$this->vars;?>
<span class="who-ran"><?php $_v['data_user'] = $_v['data_who']['user'];?>
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
<?php if($_v['data_who']['scheduled']){?>
 <span class="badge">scheduled</span><?php }?><?php if($_v['data_who']['reason']){?>
 <span class="badge bad" title="<?php echo htmlspecialchars($_v['data_who']['reason'], ENT_QUOTES);?>
">out of hours</span><?php }elseif($_v['data_who']['forced']){?>
 <span class="badge signal">forced</span><?php }?></span>
