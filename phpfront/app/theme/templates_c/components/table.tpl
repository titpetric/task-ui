<?php $_v=&$this->vars;?>

<div class="table-wrap<?php if($_v['data_table']['wrap']){?>
 <?php echo htmlspecialchars($_v['data_table']['wrap'], ENT_QUOTES); }?>">
<table class="table<?php if($_v['data_table']['class']){?>
 <?php echo htmlspecialchars($_v['data_table']['class'], ENT_QUOTES); }?>">
<?php if($_v['data_table']['header']){?>
	<thead>
<?php if(!empty($_v['data_table']['header']))foreach($_v['data_table']['header'] as $_v['row']){?>		<tr><?php if(!empty($_v['row']['cells']))foreach($_v['row']['cells'] as $_v['cell']){?><th<?php if($_v['cell']['class']){?>
 class="<?php echo htmlspecialchars($_v['cell']['class'], ENT_QUOTES);?>
"<?php }?><?php if($_v['cell']['span']){?>
 colspan="<?php echo htmlspecialchars($_v['cell']['span'], ENT_QUOTES);?>
"<?php }?>><?php echo $_v['cell']['html'];?>
</th><?php }?></tr>
<?php }?>	</thead>
<?php }?>	<tbody>
<?php if(!empty($_v['data_table']['body']))foreach($_v['data_table']['body'] as $_v['row']){?>		<?php $_v['data_row'] = $_v['row'];?>
<tr<?php if($_v['data_row']['id']){?>
 id="<?php echo htmlspecialchars($_v['data_row']['id'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['run']){?>
 data-run-id="<?php echo htmlspecialchars($_v['data_row']['run'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['class']){?>
 class="<?php echo htmlspecialchars($_v['data_row']['class'], ENT_QUOTES);?>
"<?php }?>><?php if(!empty($_v['data_row']['cells']))foreach($_v['data_row']['cells'] as $_v['cell']){?><td<?php if($_v['cell']['class']){?>
 class="<?php echo htmlspecialchars($_v['cell']['class'], ENT_QUOTES);?>
"<?php }?><?php if($_v['cell']['span']){?>
 colspan="<?php echo htmlspecialchars($_v['cell']['span'], ENT_QUOTES);?>
"<?php }?>><?php echo $_v['cell']['html'];?>
</td><?php }?></tr>

<?php }?>	</tbody>
<?php if($_v['data_table']['footer']){?>
	<tfoot>
<?php if(!empty($_v['data_table']['footer']))foreach($_v['data_table']['footer'] as $_v['row']){?>		<?php $_v['data_row'] = $_v['row'];?>
<tr<?php if($_v['data_row']['id']){?>
 id="<?php echo htmlspecialchars($_v['data_row']['id'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['run']){?>
 data-run-id="<?php echo htmlspecialchars($_v['data_row']['run'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['class']){?>
 class="<?php echo htmlspecialchars($_v['data_row']['class'], ENT_QUOTES);?>
"<?php }?>><?php if(!empty($_v['data_row']['cells']))foreach($_v['data_row']['cells'] as $_v['cell']){?><td<?php if($_v['cell']['class']){?>
 class="<?php echo htmlspecialchars($_v['cell']['class'], ENT_QUOTES);?>
"<?php }?><?php if($_v['cell']['span']){?>
 colspan="<?php echo htmlspecialchars($_v['cell']['span'], ENT_QUOTES);?>
"<?php }?>><?php echo $_v['cell']['html'];?>
</td><?php }?></tr>

<?php }?>	</tfoot>
<?php }?></table>
</div>
