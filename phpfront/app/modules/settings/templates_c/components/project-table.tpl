<?php $_v=&$this->vars;?>
<div class="table-wrap fit">
<table class="table projects cards">
	<thead><tr><th>Folder</th><th>Sources</th><th>Origin</th><th class="hide-sm hide-narrow">Last commit</th><?php if($_v['data_projects']['sized']){?>
<th class="hide-sm hide-narrow">Size</th><?php }?><th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
<?php if(!empty($_v['data_projects']['rows']))foreach($_v['data_projects']['rows'] as $_v['row']){?>		<tr>
			<td class="key nowrap"><a href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-folder"/></svg> <?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
</a></td>
			<td class="nowrap" data-label="Sources"><?php if($_v['row']['bad']){?>
<span class="badge bad">error</span> <?php }?><?php if($_v['row']['sources']){ echo htmlspecialchars($_v['row']['sources'], ENT_QUOTES);?>
 <span class="faint"><?php echo htmlspecialchars($_v['row']['jobs'], ENT_QUOTES);?>
 jobs</span><?php }else{?>
<span class="faint">nothing to run</span><?php }?></td>
			<td class="origin" data-label="Origin"><?php if($_v['row']['checkout']){?>
<span class="remote mono" title="<?php echo htmlspecialchars($_v['row']['remote'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['remote'], ENT_QUOTES);?>
</span> <span class="badge"><?php echo htmlspecialchars($_v['row']['branch'], ENT_QUOTES);?>
</span><?php }else{?>
<span class="faint">local folder</span><?php }?></td>
			<td class="nowrap hide-sm hide-narrow" data-label="Last commit"><?php if($_v['row']['checkout']){?>
<span class="mono" title="<?php echo htmlspecialchars($_v['row']['subject'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['commit'], ENT_QUOTES);?>
</span><?php if($_v['row']['ago']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</span><?php }?><?php }?></td><?php if($_v['data_projects']['sized']){?>
<td class="hide-sm hide-narrow" data-label="Size"><?php $_v['data_mini'] = $_v['row']['space'];?>
<span class="spacemini" title="<?php echo htmlspecialchars($_v['data_mini']['title'], ENT_QUOTES);?>
"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_mini']['filesPercent'], ENT_QUOTES);?>
%"></i><i class="history" style="width: <?php echo htmlspecialchars($_v['data_mini']['historyPercent'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_mini']['total'], ENT_QUOTES);?>
</span></span></td><?php }?>
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