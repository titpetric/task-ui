<?php $_v=&$this->vars; if($_v['data_chart']['charts']){?>
<div class="activity">
	<p class="activity-sum"><?php echo htmlspecialchars($_v['data_chart']['sum'], ENT_QUOTES); if($_v['data_chart']['failed']){?>
 · <span class="bad"><?php echo htmlspecialchars($_v['data_chart']['failed'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_chart']['median']){?>
 · <?php echo htmlspecialchars($_v['data_chart']['median'], ENT_QUOTES); }?><span class="activity-key"><i class="ok"></i>runs <i class="line avg"></i>average <i class="line max"></i>slowest <i class="line min"></i>fastest</span></p>
	<div class="activity-charts">
<?php if(!empty($_v['data_chart']['charts']))foreach($_v['data_chart']['charts'] as $_v['chart']){?>		<figure class="<?php echo htmlspecialchars($_v['chart']['class'], ENT_QUOTES);?>
">
			<figcaption class="activity-title"><span><?php echo htmlspecialchars($_v['chart']['title'], ENT_QUOTES);?>
</span><?php if($_v['chart']['stats']['text']){?>
<span class="faint" title="<?php echo htmlspecialchars($_v['chart']['stats']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['chart']['stats']['text'], ENT_QUOTES);?>
</span><?php }?></figcaption>
			<div class="activity-bars" style="--n: <?php echo htmlspecialchars($_v['chart']['n'], ENT_QUOTES);?>
" role="img" aria-label="<?php echo htmlspecialchars($_v['chart']['label'], ENT_QUOTES);?>
">
<?php if(!empty($_v['chart']['bars']))foreach($_v['chart']['bars'] as $_v['bar']){?>				<span class="<?php echo htmlspecialchars($_v['bar']['class'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['bar']['tip'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['bar']['tip'], ENT_QUOTES);?>
"><?php if($_v['bar']['other']){?>
<i class="other" style="<?php echo htmlspecialchars($_v['bar']['other'], ENT_QUOTES);?>
"></i><?php }?><?php if($_v['bar']['failed']){?>
<i class="bad" style="<?php echo htmlspecialchars($_v['bar']['failed'], ENT_QUOTES);?>
"></i><?php }?><?php if($_v['bar']['ok']){?>
<i class="ok" style="<?php echo htmlspecialchars($_v['bar']['ok'], ENT_QUOTES);?>
"></i><?php }?></span>
<?php }?>			</div>
<?php if($_v['chart']['lines']){?>
			<div class="activity-line durations" style="--n: <?php echo htmlspecialchars($_v['chart']['n'], ENT_QUOTES);?>
">
				<svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="<?php echo htmlspecialchars($_v['chart']['lines']['label'], ENT_QUOTES);?>
">
					<path class="max" d="<?php echo htmlspecialchars($_v['chart']['lines']['max'], ENT_QUOTES);?>
" vector-effect="non-scaling-stroke"></path>
					<path class="min" d="<?php echo htmlspecialchars($_v['chart']['lines']['min'], ENT_QUOTES);?>
" vector-effect="non-scaling-stroke"></path>
					<path class="avg" d="<?php echo htmlspecialchars($_v['chart']['lines']['avg'], ENT_QUOTES);?>
" vector-effect="non-scaling-stroke"></path>
				</svg>
				<div class="activity-hits">
<?php if(!empty($_v['chart']['bars']))foreach($_v['chart']['bars'] as $_v['bar']){?>					<span class="<?php echo htmlspecialchars($_v['bar']['class'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['bar']['hit'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['bar']['hit'], ENT_QUOTES);?>
"></span>
<?php }?>				</div>
			</div>
<?php }?>			<div class="activity-axis"><?php if(!empty($_v['chart']['ticks']))foreach($_v['chart']['ticks'] as $_v['tick']){?><span style="left: <?php echo htmlspecialchars($_v['tick']['at'], ENT_QUOTES);?>
%"><?php echo htmlspecialchars($_v['tick']['label'], ENT_QUOTES);?>
</span><?php }?></div>
		</figure>
<?php }?>	</div>
<?php if($_v['data_chart']['durations']){?>
	<figure class="activity-chart durations">
		<div class="activity-line">
			<svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="<?php echo htmlspecialchars($_v['data_chart']['durations']['label'], ENT_QUOTES);?>
">
				<path class="wash" d="<?php echo htmlspecialchars($_v['data_chart']['durations']['area'], ENT_QUOTES);?>
"></path>
<?php if(!empty($_v['data_chart']['durations']['points']))foreach($_v['data_chart']['durations']['points'] as $_v['point']){?><?php if($_v['point']['failed']){?>
				<line class="bad" x1="<?php echo htmlspecialchars($_v['point']['at'], ENT_QUOTES);?>
" y1="0" x2="<?php echo htmlspecialchars($_v['point']['at'], ENT_QUOTES);?>
" y2="100" vector-effect="non-scaling-stroke"></line>
<?php }?><?php }?>				<path class="line" d="<?php echo htmlspecialchars($_v['data_chart']['durations']['line'], ENT_QUOTES);?>
" vector-effect="non-scaling-stroke"></path>
			</svg>
<?php if(!empty($_v['data_chart']['durations']['points']))foreach($_v['data_chart']['durations']['points'] as $_v['point']){?>			<span class="<?php echo htmlspecialchars($_v['point']['class'], ENT_QUOTES);?>
" style="<?php echo htmlspecialchars($_v['point']['style'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['point']['title'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['point']['title'], ENT_QUOTES);?>
"></span>
<?php }?>			<span class="activity-longest"><?php echo htmlspecialchars($_v['data_chart']['durations']['longest'], ENT_QUOTES);?>
</span>
		</div>
		<figcaption><span><?php echo htmlspecialchars($_v['data_chart']['durations']['caption'], ENT_QUOTES);?>
</span></figcaption>
	</figure>
<?php }?></div><?php }?>
