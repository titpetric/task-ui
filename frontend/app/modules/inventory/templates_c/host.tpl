<?php $_v=&$this->vars; $_v['data_head'] = $_v['data']['head'];?>

<div class="page-head">
	<div>
		<h2><?php if($_v['data_head']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['data_head']['icon'], ENT_QUOTES);?>
"/></svg> <?php }?><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES); if($_v['data_head']['badge']){?>
 <span class="badge<?php if($_v['data_head']['tone']){?>
 <?php echo htmlspecialchars($_v['data_head']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_head']['badge'], ENT_QUOTES);?>
</span><?php }?><?php if(!empty($_v['data_head']['tags']))foreach($_v['data_head']['tags'] as $_v['tag']){?> <span class="badge"><?php echo htmlspecialchars($_v['tag'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_head']['note']){?>
 <span class="term-title" title="<?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
</span><?php }?></h2>
<?php if($_v['data_head']['lead'] || $_v['data_head']['leads']){?>
		<p><?php echo $_v['data_head']['lead'];?>
</p>
<?php }?><?php if($_v['data_head']['faint']){?>
		<p class="faint"><?php echo htmlspecialchars($_v['data_head']['faint'], ENT_QUOTES);?>
</p>
<?php }?>	</div>
<?php if($_v['data_head']['show']){?>
	<div class="actions"<?php if($_v['data_head']['live']){?>
 data-actions<?php }?>>
<?php if($_v['data_head']['run']){?>
		<?php $_v['data_run'] = $_v['data_head']['run'];?>
<div class="run-split">
	<button type="button" class="btn btn-primary btn-sm" data-run data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
" title="Run with arguments after <?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?>><svg class="ic"><use href="#i-play"/></svg>Run<?php if($_v['data_run']['takesArgs']){?>
...<?php }?></button>
<?php if($_v['data_run']['menu']){?>
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
" aria-label="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-down"/></svg></summary>
		<div class="more-list">
			<div class="more-items" role="menu" aria-label="Run options">
<?php if($_v['data_run']['summarize']){?>
				<button type="button" class="more-item" role="menuitem" data-run data-summarize data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?> title="Run and summarize"><svg class="ic"><use href="#i-summary"/></svg><span>Run and summarize</span></button>
<?php }?><?php if($_v['data_run']['force']){?>
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['place'], ENT_QUOTES);?>
" title="Run out of hours, with a reason"><svg class="ic"><use href="#i-play"/></svg><span>Run force</span></button>
<?php }?>			</div>
		</div>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data_head']['buttons']['items']){?>
		<?php $_v['data_buttons'] = $_v['data_head']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }}?><?php if($_v['data_head']['actions']){?>
		<?php $_v['data_menu'] = $_v['data_head']['actions'];?>

		<?php if(!$_v['data_menu']['empty']){?>
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
</details><?php }}?>	</div>
<?php }?></div>

<div class="host-detail" data-host-detail="<?php echo htmlspecialchars($_v['data']['name'], ENT_QUOTES);?>
">
	<p class="host-facts faint">
		<span><code><?php echo htmlspecialchars($_v['data']['host'], ENT_QUOTES);?>
</code></span><?php if($_v['data']['login']){?>
<span>ssh <code><?php echo htmlspecialchars($_v['data']['login'], ENT_QUOTES);?>
</code></span><?php }else{?>
<span>no ssh login</span><?php }?><?php if($_v['data']['cpus']){?>
<span><?php echo htmlspecialchars($_v['data']['cpus'], ENT_QUOTES);?>
 cores</span><?php }?><?php if($_v['data']['up']){?>
<span>up <?php echo htmlspecialchars($_v['data']['uptime'], ENT_QUOTES);?>
</span><?php }else{?>
<span class="bad"><?php echo htmlspecialchars($_v['data']['status'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data']['agent']){?>
<span>Beszel agent <?php echo htmlspecialchars($_v['data']['agent'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data']['read']){?>
<span>read over ssh <?php echo htmlspecialchars($_v['data']['read'], ENT_QUOTES);?>
</span><?php }?>
	</p>

	<h4 class="section-title">Containers <small><?php echo htmlspecialchars($_v['data']['count'], ENT_QUOTES); if($_v['data']['measured']){?>
, measures from Beszel<?php }?><?php if($_v['data']['sized']){?>
, sizes from docker ps<?php }?></small></h4>
<?php if($_v['data']['containers']){?>
	<div class="table-wrap">
	<table class="table host-containers" data-sortable>
		<thead><tr>
			<th data-sort="text">Container</th>
			<th data-sort="text">Status</th>
			<th class="num" data-sort="num" aria-sort="descending">CPU</th>
			<th class="num" data-sort="num">Memory</th>
			<th class="num" data-sort="num">Size</th>
			<th class="num" data-sort="num">Net</th>
		</tr></thead>
		<tbody>
<?php if(!empty($_v['data']['containers']))foreach($_v['data']['containers'] as $_v['row']){?>			<tr>
				<td data-v="<?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
"><span class="key"><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</span><?php if($_v['row']['image']){?>
<span class="faint under"><?php echo htmlspecialchars($_v['row']['image'], ENT_QUOTES);?>
</span><?php }?></td>
				<td data-v="<?php echo htmlspecialchars($_v['row']['status'], ENT_QUOTES);?>
" class="nowrap<?php if($_v['row']['faint']){?>
 faint<?php }?>"><?php if($_v['row']['status']){ echo htmlspecialchars($_v['row']['status'], ENT_QUOTES); }else{?>
<span class="faint">-</span><?php }?></td>
<?php if($_v['row']['measured']){?>
				<td class="num" data-v="<?php echo htmlspecialchars($_v['row']['cpuV'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['cpu'], ENT_QUOTES);?>
%</td>
				<td class="num nowrap" data-v="<?php echo htmlspecialchars($_v['row']['memoryV'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['memory'], ENT_QUOTES); if($_v['row']['shares']){?>
<span class="faint under"><?php echo htmlspecialchars($_v['row']['share'], ENT_QUOTES);?>
% of RAM</span><?php }?></td>
<?php }else{?>
				<td class="num faint" data-v="-1" title="<?php echo htmlspecialchars($_v['data']['measureNote'], ENT_QUOTES);?>
">-</td>
				<td class="num faint" data-v="-1" title="<?php echo htmlspecialchars($_v['data']['measureNote'], ENT_QUOTES);?>
">-</td>
<?php }?><?php if($_v['row']['sized']){?>
				<td class="num nowrap" data-v="<?php echo htmlspecialchars($_v['row']['sizeV'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
<span class="faint under">of <?php echo htmlspecialchars($_v['row']['virtual'], ENT_QUOTES);?>
</span></td>
<?php }else{?>
				<td class="num faint" data-v="-1" title="<?php echo htmlspecialchars($_v['data']['sizeNote'], ENT_QUOTES);?>
">-</td>
<?php }?><?php if($_v['row']['traffic']){?>
				<td class="num nowrap" data-v="<?php echo htmlspecialchars($_v['row']['netV'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['net'], ENT_QUOTES);?>
</td>
<?php }else{?>
				<td class="num faint" data-v="-1" title="<?php if($_v['row']['measured']){?>
Beszel reports no traffic of the container<?php }else{ echo htmlspecialchars($_v['data']['measureNote'], ENT_QUOTES); }?>">-</td>
<?php }?>			</tr>
<?php }?>		</tbody>
	</table>
	</div><?php if(!$_v['data']['sized']){?>
<p class="faint host-note">No sizes: <?php echo htmlspecialchars($_v['data']['unsized'], ENT_QUOTES);?>
.</p><?php } }else{?>
	<p class="empty-state"><?php if($_v['data']['known']){?>
The host runs no containers.<?php }else{?>
Neither Beszel nor docker ps lists containers on the host: <?php echo htmlspecialchars($_v['data']['unsized'], ENT_QUOTES);?>
.<?php }?></p>
<?php }?>
	<h4 class="section-title">Filesystems <small>durable storage<?php if($_v['data']['fromDF']){?>
, from df over ssh<?php }elseif($_v['data']['filesystems']){?>
, the ones Beszel watches<?php }?></small></h4>
<?php if($_v['data']['filesystems']){?>
	<div class="table-wrap">
	<table class="table host-filesystems">
		<thead><tr><th>Mount</th><?php if($_v['data']['fromDF']){?>
<th>Device</th><th>Type</th><?php }?><th class="num">Size</th><th class="num">Used</th><th class="num">Free</th><th>Use</th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['filesystems']))foreach($_v['data']['filesystems'] as $_v['row']){?>			<tr>
				<td class="key"><code><?php echo htmlspecialchars($_v['row']['mount'], ENT_QUOTES);?>
</code></td><?php if($_v['data']['fromDF']){?>
<td class="faint"><code><?php echo htmlspecialchars($_v['row']['device'], ENT_QUOTES);?>
</code></td><td class="faint"><?php echo htmlspecialchars($_v['row']['type'], ENT_QUOTES);?>
</td><?php }?>
				<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['size'], ENT_QUOTES);?>
</td>
				<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['used'], ENT_QUOTES);?>
</td>
				<td class="num nowrap"><?php echo htmlspecialchars($_v['row']['free'], ENT_QUOTES);?>
</td>
				<td><?php $_v['data_meter'] = $_v['row']['use'];?>
<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
			</td>
			</tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php }else{?>
	<p class="empty-state">Neither df over ssh nor Beszel reports the host's filesystems<?php if(!$_v['data']['login']){?>
: task-ui has no ssh login to it, and Beszel no measures<?php }?>.</p>
<?php }?><?php if($_v['data']['charts']){?>
	<h4 class="section-title">The last 30 days <small>Beszel's records</small></h4>
	<div class="charts"><?php if(!empty($_v['data']['charts']))foreach($_v['data']['charts'] as $_v['chart']){?><?php $_v['data_line_chart'] = $_v['chart'];?>
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
<?php }?>	</div>
<?php }else{?>
	<p class="empty-state">Beszel holds no records of the host in the last 30 days.</p>
<?php }?></div>
