<?php $_v=&$this->vars;?>
<div class="conversation-head">
	<div class="conversation-title">
		<h2><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES);?>
</h2>
<?php if($_v['data_head']['meta']){?>
		<p class="faint"><?php echo htmlspecialchars($_v['data_head']['meta'], ENT_QUOTES);?>
</p>
<?php }?>	</div>
	<div class="actions">
<?php if($_v['data_head']['can_pick']){?>
		<form class="agent-pick" data-agent-form data-thread="<?php echo htmlspecialchars($_v['data_head']['id'], ENT_QUOTES);?>
">
			<label class="faint" for="thread-agent">Agent</label>
			<select id="thread-agent" name="agent">
<?php if(!empty($_v['data_head']['agents']))foreach($_v['data_head']['agents'] as $_v['name']){?>				<option value="<?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
"<?php if($_v['name'] == $_v['data_head']['agent']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
			<button class="btn" type="submit">Save</button>
		</form>
<?php }?>		<?php $_v['data_menu'] = $_v['data_head']['menu']; if(!$_v['data_menu']['empty']){?>
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

	</div>
</div>
