<?php $_v=&$this->vars;?>
<form class="matrix-form<?php if($_v['data_matrix']['fixed']){?>
 fixed<?php }?>"<?php if($_v['data_matrix']['editable']){?>
 data-api="<?php echo htmlspecialchars($_v['data_matrix']['api'], ENT_QUOTES);?>
" data-method="PUT" data-next="reload" data-notice="Permissions saved" data-matrix<?php }?>>
	<div class="table-wrap">
	<table class="table matrix<?php if($_v['data_matrix']['effective']){?>
 effective<?php }?>">
		<thead><tr><th>Section</th><?php if(!empty($_v['data_matrix']['rules']))foreach($_v['data_matrix']['rules'] as $_v['rule']){?><th><?php echo htmlspecialchars($_v['rule']['title'], ENT_QUOTES);?>
</th><?php }?></tr></thead>
		<tbody>
<?php if(!empty($_v['data_matrix']['sections']))foreach($_v['data_matrix']['sections'] as $_v['section']){?>			<tr><th scope="row"><?php echo htmlspecialchars($_v['section']['title'], ENT_QUOTES);?>
</th><?php if(!empty($_v['section']['cells']))foreach($_v['section']['cells'] as $_v['cell']){?><td><?php $_v['data_cell'] = $_v['cell'];?>
<span class="tri-cell"><button type="button" class="tri" data-tri="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cell']['title'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
"<?php if($_v['data_cell']['locked']){?>
 disabled<?php }?>><?php echo htmlspecialchars($_v['data_cell']['mark'], ENT_QUOTES);?>
</button><?php if($_v['data_cell']['field']){?>
<input type="hidden" name="<?php echo htmlspecialchars($_v['data_cell']['path'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-was="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" disabled><?php }?><?php if($_v['data_cell']['shown']){?>
<span class="eff <?php echo htmlspecialchars($_v['data_cell']['tone'], ENT_QUOTES);?>
" title="Result: <?php echo htmlspecialchars($_v['data_cell']['says'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_cell']['result'], ENT_QUOTES);?>
</span><?php }?></span>
</td><?php }?></tr>
<?php }?>			<tr class="group"><th colspan="<?php echo htmlspecialchars($_v['data_matrix']['columns'], ENT_QUOTES);?>
">Application</th></tr>
<?php if(!empty($_v['data_matrix']['permissions']))foreach($_v['data_matrix']['permissions'] as $_v['cell']){?>			<tr><th scope="row"><?php echo htmlspecialchars($_v['cell']['title'], ENT_QUOTES);?>
</th><td colspan="<?php echo htmlspecialchars($_v['data_matrix']['span'], ENT_QUOTES);?>
"><?php $_v['data_cell'] = $_v['cell'];?>
<span class="tri-cell"><button type="button" class="tri" data-tri="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cell']['title'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
"<?php if($_v['data_cell']['locked']){?>
 disabled<?php }?>><?php echo htmlspecialchars($_v['data_cell']['mark'], ENT_QUOTES);?>
</button><?php if($_v['data_cell']['field']){?>
<input type="hidden" name="<?php echo htmlspecialchars($_v['data_cell']['path'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-was="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" disabled><?php }?><?php if($_v['data_cell']['shown']){?>
<span class="eff <?php echo htmlspecialchars($_v['data_cell']['tone'], ENT_QUOTES);?>
" title="Result: <?php echo htmlspecialchars($_v['data_cell']['says'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_cell']['result'], ENT_QUOTES);?>
</span><?php }?></span>
</td></tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php if($_v['data_matrix']['editable']){?>
	<div class="form-foot">
		<span class="faint" data-matrix-changed>No changes yet.</span>
		<button class="btn btn-primary" type="submit">Save permissions</button>
	</div>
<?php }?></form>
