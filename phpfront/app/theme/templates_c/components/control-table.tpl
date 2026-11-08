<?php $_v=&$this->vars;?>
<div class="table-wrap">
<table class="table control col-hover">
	<tbody>
<?php if(!empty($_v['data_control']))foreach($_v['data_control'] as $_v['row']){?>		<tr class="named" data-span>
			<td class="key nowrap"<?php if($_v['row']['title']){?>
 title="<?php echo htmlspecialchars($_v['row']['title'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['row']['href']){?>
<a class="badge" href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['label'], ENT_QUOTES);?>
</a><?php }else{?>
<span class="badge"><?php echo htmlspecialchars($_v['row']['label'], ENT_QUOTES);?>
</span><?php }?></td>
			<td colspan="1"><?php if($_v['row']['empty']){?>
<span class="faint"><?php echo htmlspecialchars($_v['row']['empty'], ENT_QUOTES);?>
</span><?php }else{?>
<div class="cc-row"><?php if(!empty($_v['row']['jobs']))foreach($_v['row']['jobs'] as $_v['job']){?><?php $_v['data_cc'] = $_v['job'];?>
<span class="cc-slot" data-cc="<?php echo htmlspecialchars($_v['data_cc']['key'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_cc']['folder'], ENT_QUOTES);?>
" data-job="<?php echo htmlspecialchars($_v['data_cc']['job'], ENT_QUOTES);?>
" data-name="<?php echo htmlspecialchars($_v['data_cc']['name'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
" data-last="<?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['can']){?>
 data-can<?php }?><?php if($_v['data_cc']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_cc']['dir'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['desc']){?>
 data-desc="<?php echo htmlspecialchars($_v['data_cc']['desc'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_cc']['args'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['confirm']){?>
 data-confirm="<?php echo htmlspecialchars($_v['data_cc']['confirm'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['path']){?>
 data-path="<?php echo htmlspecialchars($_v['data_cc']['path'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['terminal']){?>
 data-interactive<?php }?>><?php if($_v['data_cc']['run']){?>
<a class="cc named<?php if($_v['data_cc']['pinned']){?>
 pinned<?php }?> running" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_cc']['run'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_cc']['run'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></a><?php }elseif($_v['data_cc']['can']){?>
<button type="button" class="cc named<?php if($_v['data_cc']['pinned']){?>
 pinned<?php }?> <?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
" data-run data-folder="<?php echo htmlspecialchars($_v['data_cc']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_cc']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_cc']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cc']['what'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['confirm']){?>
 data-confirm="<?php echo htmlspecialchars($_v['data_cc']['confirm'], ENT_QUOTES);?>
" data-confirm-ok="Run <?php echo htmlspecialchars($_v['data_cc']['name'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_cc']['args'], ENT_QUOTES);?>
"<?php }?> title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-<?php if($_v['data_cc']['terminal']){?>
terminal<?php }else{?>
play<?php }?>"/></svg><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></button><?php }else{?>
<a class="cc named<?php if($_v['data_cc']['pinned']){?>
 pinned<?php }?> idle <?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
" href="<?php echo htmlspecialchars($_v['data_cc']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><i class="shebang" aria-hidden="true">#!</i><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></a><?php }?></span>
<?php }?></div><?php }?></td>
			<td class="actions"><?php $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
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
</td>
		</tr>
<?php }?>	</tbody>
</table>
</div>
