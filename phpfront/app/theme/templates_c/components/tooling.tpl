<?php $_v=&$this->vars;if($_v['data_tooling']['scopes']){?>
<div class="tooling">
<?php if(!empty($_v['data_tooling']['scopes']))foreach($_v['data_tooling']['scopes'] as $_v['scope']){?>	<section class="tooling-scope">
		<p class="tooling-sum"><span class="tooling-db" title="<?php echo htmlspecialchars($_v['scope']['database'], ENT_QUOTES);?>
">&lrm;<?php echo htmlspecialchars($_v['scope']['database'], ENT_QUOTES);?>
</span><span><span class="key"><?php echo htmlspecialchars($_v['scope']['name'], ENT_QUOTES);?>
</span><?php if($_v['scope']['repository']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['scope']['repository'], ENT_QUOTES);?>
</span><?php }?> · <?php echo htmlspecialchars($_v['scope']['sum'], ENT_QUOTES);?>
</span></p>
		<div class="tooling-charts">
<?php if(!empty($_v['scope']['charts']))foreach($_v['scope']['charts'] as $_v['chart']){?>			<figure class="<?php echo htmlspecialchars($_v['chart']['class'], ENT_QUOTES);?>
">
				<figcaption class="tooling-title"><span><?php echo htmlspecialchars($_v['chart']['title'], ENT_QUOTES);?>
</span><small><?php echo htmlspecialchars($_v['chart']['stats'], ENT_QUOTES);?>
</small></figcaption>
				<div class="<?php echo htmlspecialchars($_v['chart']['bars']['class'], ENT_QUOTES);?>
" style="<?php echo htmlspecialchars($_v['chart']['bars']['style'], ENT_QUOTES);?>
" role="img" aria-label="<?php echo htmlspecialchars($_v['chart']['bars']['label'], ENT_QUOTES);?>
">
<?php if(!empty($_v['chart']['bars']['slots']))foreach($_v['chart']['bars']['slots'] as $_v['bar']){?><?php if($_v['bar']['segments']){?>
					<span aria-label="<?php echo htmlspecialchars($_v['bar']['title'], ENT_QUOTES);?>
"><?php if(!empty($_v['bar']['segments']))foreach($_v['bar']['segments'] as $_v['part']){?><?php if($_v['part']['title']){?>
<b class="<?php echo htmlspecialchars($_v['part']['class'], ENT_QUOTES);?>
" style="<?php echo htmlspecialchars($_v['part']['style'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['part']['title'], ENT_QUOTES);?>
"></b><?php }else{?>
<b></b><?php }?><?php }?></span>
<?php }else{?>
					<span class="<?php echo htmlspecialchars($_v['bar']['class'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['bar']['title'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['bar']['title'], ENT_QUOTES);?>
"></span>
<?php }?><?php }?>				</div>
				<div class="activity-axis"><?php if(!empty($_v['chart']['ticks']))foreach($_v['chart']['ticks'] as $_v['tick']){?><span style="left: <?php echo htmlspecialchars($_v['tick']['at'], ENT_QUOTES);?>
%"><?php echo htmlspecialchars($_v['tick']['label'], ENT_QUOTES);?>
</span><?php }?></div>
				<ul class="tooling-legend">
<?php if($_v['chart']['legend']){ if(!empty($_v['chart']['legend']))foreach($_v['chart']['legend'] as $_v['item']){?>					<li><span class="swatch <?php echo htmlspecialchars($_v['item']['color'], ENT_QUOTES);?>
"></span><span class="key" title="<?php echo htmlspecialchars($_v['item']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['item']['name'], ENT_QUOTES);?>
</span><span class="nowrap"><?php echo htmlspecialchars($_v['item']['total'], ENT_QUOTES);?>
</span><small class="faint nowrap"><?php echo htmlspecialchars($_v['item']['runs'], ENT_QUOTES);?>
</small></li>
<?php }?><?php }else{?>
					<li class="empty">Nothing ran in this window.</li>
<?php }?>				</ul>
			</figure>
<?php }?>		</div>
<?php if($_v['scope']['latest']){?>
		<div class="tooling-latest">
			<h4>Latest commands</h4>
			<ol>
<?php if(!empty($_v['scope']['latest']))foreach($_v['scope']['latest'] as $_v['ran']){?>				<li><time datetime="<?php echo htmlspecialchars($_v['ran']['stamp'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['ran']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['ran']['when'], ENT_QUOTES);?>
</time><span class="swatch <?php echo htmlspecialchars($_v['ran']['color'], ENT_QUOTES);?>
"></span><span class="key" title="<?php echo htmlspecialchars($_v['ran']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['ran']['command'], ENT_QUOTES); if($_v['ran']['job']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['ran']['job'], ENT_QUOTES);?>
</span><?php }?></span><span class="nowrap"><?php echo htmlspecialchars($_v['ran']['took'], ENT_QUOTES);?>
</span><?php if($_v['ran']['failed']){?>
<span class="bad" title="exit <?php echo htmlspecialchars($_v['ran']['exit'], ENT_QUOTES);?>
">failed</span><?php }else{?>
<span class="ok">ok</span><?php }?></li>
<?php }?>			</ol>
		</div>
<?php }?><?php if($_v['scope']['commands']){?>
		<details class="tooling-details">
			<summary><svg class="ic chev"><use href="#i-chev"/></svg><?php echo htmlspecialchars($_v['scope']['toggle'], ENT_QUOTES);?>
</summary>
			<div class="table-wrap">
			<table class="table cards">
				<thead><tr><th>Command</th><th>Tool</th><th>Runs</th><th>Total</th><th>Average</th><th>Slowest</th><th>Failed</th><th>Last</th></tr></thead>
				<tbody>
<?php if(!empty($_v['scope']['commands']))foreach($_v['scope']['commands'] as $_v['row']){?>					<tr>
						<td class="key"><?php echo htmlspecialchars($_v['row']['command'], ENT_QUOTES);?>
</td>
						<td data-label="Tool"><?php echo htmlspecialchars($_v['row']['tool'], ENT_QUOTES);?>
</td>
						<td data-label="Runs"><?php echo htmlspecialchars($_v['row']['runs'], ENT_QUOTES);?>
</td>
						<td class="nowrap" data-label="Total"><?php echo htmlspecialchars($_v['row']['total'], ENT_QUOTES);?>
</td>
						<td class="nowrap" data-label="Average"><?php echo htmlspecialchars($_v['row']['average'], ENT_QUOTES);?>
</td>
						<td class="nowrap" data-label="Slowest"><?php echo htmlspecialchars($_v['row']['max'], ENT_QUOTES);?>
</td>
						<td data-label="Failed"><?php if($_v['row']['failures']){?>
<span class="bad"><?php echo htmlspecialchars($_v['row']['failures'], ENT_QUOTES);?>
</span><?php }else{?>
0<?php }?></td>
						<td class="nowrap" data-label="Last"><?php echo htmlspecialchars($_v['row']['last'], ENT_QUOTES);?>
</td>
					</tr>
<?php }?>				</tbody>
			</table>
			</div>
		</details>
<?php }?>	</section>
<?php }?></div><?php }?>
