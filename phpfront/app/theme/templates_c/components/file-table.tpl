<?php $_v=&$this->vars;?>
<div class="table-wrap">
<table class="table file-table col-hover">
	<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
	<tbody>
<?php if($_v['data_files']['up']){?>
		<tr class="up"><td class="key"><svg class="ic dir"><use href="#i-folder"/></svg> <a href="<?php echo htmlspecialchars($_v['data_files']['up'], ENT_QUOTES);?>
" title="Up a level">..</a></td><td></td><td class="hide-sm"></td><td></td></tr>
<?php }?><?php if(!empty($_v['data_files']['rows']))foreach($_v['data_files']['rows'] as $_v['row']){?><?php if($_v['row']['sealed']){?>
		<tr class="sealed" aria-disabled="true" title="task-ui lists <?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
 but does not open it">
			<td class="key"><svg class="ic"><use href="#i-secret"/></svg> <span><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</span></td>
			<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
</td>
			<td class="hide-sm nowrap" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</td>
			<td></td>
		</tr>
<?php }else{?>
		<tr class="<?php if($_v['row']['dir']){?>
dir<?php }else{?>
file<?php }?>">
			<td class="key"><?php if($_v['row']['exec']){?>
<i class="shebang" aria-hidden="true">#!</i><?php }else{?>
<svg class="ic<?php if($_v['row']['dir']){?>
 dir<?php }?>"><use href="#i-<?php if($_v['row']['dir']){?>
folder<?php }else{?>
file<?php }?>"/></svg><?php }?><?php if($_v['row']['href']){?>
 <a href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"<?php if(!$_v['row']['dir']){?>
 data-file="<?php echo htmlspecialchars($_v['row']['path'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</a><?php }else{?>
 <span><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['row']['exec']){?>
 <span class="badge exec" title="Runs: executable, or a shebang on its first line">exec</span><?php }?></td>
			<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
</td>
			<td class="hide-sm nowrap" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</td>
			<td class="actions"><?php if($_v['row']['menu']){ $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
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
</details><?php } }?></td>
		</tr>
<?php }?><?php }else{?>
		<tr class="empty"><td colspan="4">This directory is empty.</td></tr>
<?php }?>	</tbody>
</table>
</div>
