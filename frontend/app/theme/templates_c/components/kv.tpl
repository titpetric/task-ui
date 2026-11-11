<?php $_v=&$this->vars;?>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
<?php if(!empty($_v['data_kv']['rows']))foreach($_v['data_kv']['rows'] as $_v['row']){?>			<tr><th><?php echo htmlspecialchars($_v['row']['label'], ENT_QUOTES);?>
</th><td<?php if($_v['row']['class']){?>
 class="<?php echo htmlspecialchars($_v['row']['class'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['row']['who']){ $_v['data_who'] = $_v['row']['who'];?>
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
<?php }?><?php if($_v['row']['state']){ $_v['data_state'] = $_v['row']['state']; if($_v['data_state']['id']){?>
<span class="state" data-run-id="<?php echo htmlspecialchars($_v['data_state']['id'], ENT_QUOTES);?>
"<?php if($_v['data_state']['ended']){?>
 data-finished="true"<?php }?>><i class="dot <?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES); if($_v['data_state']['code'] !== ""){?>
 <span class="faint">(<?php echo htmlspecialchars($_v['data_state']['code'], ENT_QUOTES);?>
)</span><?php }?></span><?php }else{?>
<span class="faint">never run</span><?php } }?><?php if($_v['row']['link']['href']){?>
<a href="<?php echo htmlspecialchars($_v['row']['link']['href'], ENT_QUOTES);?>
"><?php if($_v['row']['link']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['row']['link']['icon'], ENT_QUOTES);?>
"/></svg> <?php }?><?php echo htmlspecialchars($_v['row']['link']['text'], ENT_QUOTES);?>
</a><?php }?><?php if(!empty($_v['row']['code']))foreach($_v['row']['code'] as $_v['value']){?><code><?php echo htmlspecialchars($_v['value'], ENT_QUOTES);?>
</code> <?php }?><?php echo htmlspecialchars($_v['row']['text'], ENT_QUOTES); if($_v['row']['faint']){?>
<span class="faint"><?php echo htmlspecialchars($_v['row']['faint'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['row']['after']){?>
 <span class="dim"><?php echo htmlspecialchars($_v['row']['after'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['row']['badge']['text']){?>
 <span class="badge <?php echo htmlspecialchars($_v['row']['badge']['tone'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['badge']['text'], ENT_QUOTES);?>
</span><?php }?></td></tr>
<?php }?>		</tbody>
	</table>
	</div>
