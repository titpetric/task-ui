<?php $_v=&$this->vars; if($_v['data_field']['kind'] === "who"){?>
							<div class="field wf-wide">
								<span class="label"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES);?>
</span><?php $_v['data_whofield'] = $_v['data_field']['who'];?>
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
</div><?php if($_v['data_field']['help']){?>
<small class="faint"><?php echo htmlspecialchars($_v['data_field']['help'], ENT_QUOTES);?>
</small><?php }?>
							</div>
<?php }elseif($_v['data_field']['kind'] === "env"){?>
							<div class="field wf-wide env-flag">
								<div class="field-head">
									<span class="label"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES);?>
</span><?php $_v['data_switch'] = $_v['data_field']['switch'];?>
<label class="switch env-switch" title="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
 in a run">
	<input type="checkbox" name="<?php echo htmlspecialchars($_v['data_switch']['name'], ENT_QUOTES);?>
" value="1" data-type="toggle"<?php if($_v['data_switch']['on']){?>
 checked<?php }?> aria-label="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
">
	<i aria-hidden="true"></i>
</label>
								</div>
								<small class="faint"><?php if($_v['data_field']['help']){ echo htmlspecialchars($_v['data_field']['help'], ENT_QUOTES); if($_v['data_field']['spaced']){?>
 <?php }?><?php }?><code><?php echo htmlspecialchars($_v['data_field']['env'], ENT_QUOTES);?>
</code></small>
							</div>
<?php }else{?>
							<div class="field<?php if($_v['data_field']['wide']){?>
 wf-wide<?php }?>">
<?php if($_v['data_field']['toggle']){?>
								<div class="field-head">
									<label for="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES); if($_v['data_field']['badge']){?>
 <span class="badge<?php if($_v['data_field']['tone']){?>
 <?php echo htmlspecialchars($_v['data_field']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_field']['badge'], ENT_QUOTES);?>
</span><?php }?></label><?php $_v['data_switch'] = $_v['data_field']['switch'];?>
<label class="switch env-switch" title="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
 in a run">
	<input type="checkbox" name="<?php echo htmlspecialchars($_v['data_switch']['name'], ENT_QUOTES);?>
" value="1" data-type="toggle"<?php if($_v['data_switch']['on']){?>
 checked<?php }?> aria-label="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
">
	<i aria-hidden="true"></i>
</label>
								</div>
<?php }else{?>
								<label for="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES); if($_v['data_field']['badge']){?>
 <span class="badge<?php if($_v['data_field']['tone']){?>
 <?php echo htmlspecialchars($_v['data_field']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_field']['badge'], ENT_QUOTES);?>
</span><?php }?></label>
<?php }?><?php if($_v['data_field']['kind'] === "secret"){?>
								<input type="password" id="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
" name="<?php echo htmlspecialchars($_v['data_field']['name'], ENT_QUOTES);?>
" placeholder="<?php echo htmlspecialchars($_v['data_field']['placeholder'], ENT_QUOTES);?>
" autocomplete="new-password" spellcheck="false">
<?php }elseif($_v['data_field']['kind'] === "select"){?>
								<select id="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
" name="<?php echo htmlspecialchars($_v['data_field']['name'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_field']['options']))foreach($_v['data_field']['options'] as $_v['option']){?>									<option value="<?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
"<?php if($_v['option']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
</option>
<?php }?>								</select>
<?php }else{?>
								<input type="text" id="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
" name="<?php echo htmlspecialchars($_v['data_field']['name'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['data_field']['value'], ENT_QUOTES);?>
" placeholder="<?php echo htmlspecialchars($_v['data_field']['placeholder'], ENT_QUOTES);?>
" autocomplete="off" spellcheck="false">
<?php }?>								<small class="faint"><?php if($_v['data_field']['help']){ echo htmlspecialchars($_v['data_field']['help'], ENT_QUOTES); if($_v['data_field']['spaced']){?>
 <?php }?><?php }?><?php if($_v['data_field']['env']){?>
<code><?php echo htmlspecialchars($_v['data_field']['env'], ENT_QUOTES);?>
</code><?php }?></small>
							</div>
<?php }?>