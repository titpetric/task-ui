<?php $_v=&$this->vars;?>
<figure class="chart">
	<figcaption class="chart-title"><?php echo htmlspecialchars($_v['data_line_chart']['title'], ENT_QUOTES);?>
</figcaption>
<?php if($_v['data_line_chart']['empty']){?>
	<p class="faint chart-empty">No records in the last 30 days.</p>
<?php }else{?>
	<?php echo $_v['data_line_chart']['svg'];?>

	<table class="chart-legend">
		<thead><tr><th></th><th>min</th><th>avg</th><th>max</th></tr></thead>
		<tbody>
<?php if(!empty($_v['data_line_chart']['legend']))foreach($_v['data_line_chart']['legend'] as $_v['row']){?>			<tr><td><i class="swatch s<?php echo htmlspecialchars($_v['row']['color'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['row']['min'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['row']['avg'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['row']['max'], ENT_QUOTES);?>
</td></tr>
<?php }?><?php if($_v['data_line_chart']['total']){?>
			<tr class="total"><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['name'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['min'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['avg'], ENT_QUOTES);?>
</td><td><?php echo htmlspecialchars($_v['data_line_chart']['total']['max'], ENT_QUOTES);?>
</td></tr>
<?php }?>		</tbody>
	</table>
<?php }?></figure>
