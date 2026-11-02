<?php $_v=&$this->vars;?>

<div class="table-wrap">
<table>
<?php if($_v['data_table']['columns']){?>
	<colgroup><?php if(!empty($_v['data_table']['columns']))foreach($_v['data_table']['columns'] as $_v['column']){?><col<?php if($_v['column']['class']){?>
 class="<?php echo htmlspecialchars($_v['column']['class'], ENT_QUOTES);?>
"<?php }?>><?php }?></colgroup>
	<thead><tr><?php if(!empty($_v['data_table']['columns']))foreach($_v['data_table']['columns'] as $_v['column']){?><th<?php if($_v['column']['numeric']){?>
 class="num"<?php }?>><?php echo htmlspecialchars($_v['column']['title'], ENT_QUOTES);?>
</th><?php }?></tr></thead>
<?php }?>	<tbody>
<?php if(!empty($_v['data_table']['rows']))foreach($_v['data_table']['rows'] as $_v['row']){?>		<?php $_v['data_row'] = $_v['row'];?>
<tr<?php if($_v['data_row']['id']){?>
 id="<?php echo htmlspecialchars($_v['data_row']['id'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['run']){?>
 data-run-id="<?php echo htmlspecialchars($_v['data_row']['run'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['class']){?>
 class="<?php echo htmlspecialchars($_v['data_row']['class'], ENT_QUOTES);?>
"<?php }?>><?php if(!empty($_v['data_row']['cells']))foreach($_v['data_row']['cells'] as $_v['cell']){?><td<?php if($_v['cell']['class']){?>
 class="<?php echo htmlspecialchars($_v['cell']['class'], ENT_QUOTES);?>
"<?php }?>><?php echo $_v['cell']['html'];?>
</td><?php }?></tr>

<?php }else{?>
		<tr class="empty-row"><td colspan="<?php echo htmlspecialchars($_v['data_table']['width'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_table']['empty'], ENT_QUOTES);?>
</td></tr>
<?php }?>	</tbody>
</table>
</div>
