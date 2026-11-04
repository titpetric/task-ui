<?php $_v=&$this->vars;?>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
<?php if(!empty($_v['data_kv']['rows']))foreach($_v['data_kv']['rows'] as $_v['row']){?>			<tr><th><?php echo htmlspecialchars($_v['row']['label'], ENT_QUOTES);?>
</th><td<?php if($_v['row']['class']){?>
 class="<?php echo htmlspecialchars($_v['row']['class'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['row']['who']){ $_v['data_who'] = $_v['row']['who'];?>
<span class="who-ran"><?php $_v['data_user'] = $_v['data_who']['user'];?>
<span class="user-ref"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php $_v['data_menu'] = $_v['data_user']['menu'];if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-more"/></svg></summary>
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
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }?>
</span>
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
