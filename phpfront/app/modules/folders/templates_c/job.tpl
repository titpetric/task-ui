<?php $_v=&$this->vars;$_v['data_head'] = $_v['data']['head'];?>


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
		<?php $_v['data_buttons'] = $_v['data_head']['buttons'];if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
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

<?php if($_v['data']['running']){?>
<div class="alert signal">
	<span class="state"><i class="dot running"></i><b>Running</b></span>, started by <b><?php echo htmlspecialchars($_v['data']['running']['who'], ENT_QUOTES);?>
</b> <?php echo htmlspecialchars($_v['data']['running']['ago'], ENT_QUOTES); if($_v['data']['running']['watching']){?>
, <?php echo htmlspecialchars($_v['data']['running']['watching'], ENT_QUOTES);?>
 watching<?php }?>.<?php if($_v['data']['running']['kills']){?>
 Attach to follow it, or kill it.
<?php }else{?>
 Attach to follow it.
<?php }?></div>
<?php }?><?php if($_v['data']['denied']){?>
<div class="alert">You can see this job and its history, but you don't have permission to run it.</div>
<?php }?><?php if($_v['data']['summary']){?>
<pre class="summary"><?php echo htmlspecialchars($_v['data']['summary'], ENT_QUOTES);?>
</pre>
<?php }?><section class="block">
	<h3>Job</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Folder</th><td><a href="<?php echo htmlspecialchars($_v['data']['job']['folderHref'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-folder"/></svg> <?php echo htmlspecialchars($_v['data']['job']['folder'], ENT_QUOTES);?>
</a> <span class="dim"><?php echo htmlspecialchars($_v['data']['job']['title'], ENT_QUOTES);?>
</span></td></tr>
<?php if($_v['data']['job']['dir']){?>
			<tr><th>Directory</th><td><a href="<?php echo htmlspecialchars($_v['data']['job']['dirHref'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-folder-open"/></svg> <?php echo htmlspecialchars($_v['data']['job']['dir'], ENT_QUOTES);?>
/</a> <span class="dim">the job runs here</span></td></tr>
<?php }?>			<tr><th>Source</th><td><svg class="ic"><use href="#i-file"/></svg> <?php echo htmlspecialchars($_v['data']['job']['source'], ENT_QUOTES); if($_v['data']['job']['shows']){ $_v['data_source'] = $_v['data']['job']['shows'];?>
<button type="button" class="btn btn-sm show-source" data-source="<?php echo htmlspecialchars($_v['data_source']['folder'], ENT_QUOTES);?>
" data-file="<?php echo htmlspecialchars($_v['data_source']['file'], ENT_QUOTES);?>
" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span><?php echo htmlspecialchars($_v['data_source']['label'], ENT_QUOTES);?>
</span></button>
<?php }?></td></tr>
<?php if($_v['data']['job']['includes']){?>
			<tr><th>Includes</th><td><?php if(!empty($_v['data']['job']['includes']))foreach($_v['data']['job']['includes'] as $_v['read']){?><div><svg class="ic"><use href="#i-file"/></svg> <?php echo htmlspecialchars($_v['read']['file'], ENT_QUOTES); $_v['data_source'] = $_v['read']['shows'];?>
<button type="button" class="btn btn-sm show-source" data-source="<?php echo htmlspecialchars($_v['data_source']['folder'], ENT_QUOTES);?>
" data-file="<?php echo htmlspecialchars($_v['data_source']['file'], ENT_QUOTES);?>
" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span><?php echo htmlspecialchars($_v['data_source']['label'], ENT_QUOTES);?>
</span></button>
</div><?php }?></td></tr>
<?php }?>			<tr><th>Command</th><td><code><?php echo htmlspecialchars($_v['data']['job']['command'], ENT_QUOTES);?>
</code></td></tr>
<?php if($_v['data']['job']['depends']){?>
			<tr><th>Depends on</th><td><?php if(!empty($_v['data']['job']['depends']))foreach($_v['data']['job']['depends'] as $_v['dep']){?><?php if($_v['dep']['href']){?>
<a href="<?php echo htmlspecialchars($_v['dep']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['dep']['name'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['dep']['text'], ENT_QUOTES); }?> <?php }?></td></tr>
<?php }?>			<tr><th>Steps</th><td class="steps"><?php if(!empty($_v['data']['job']['steps']))foreach($_v['data']['job']['steps'] as $_v['step']){?><code><?php echo htmlspecialchars($_v['step'], ENT_QUOTES);?>
</code><?php }else{?>
<span class="faint">none listed</span><?php }?></td></tr>
<?php if($_v['data']['job']['aliases']){?>
			<tr><th>Aliases</th><td><?php if(!empty($_v['data']['job']['aliases']))foreach($_v['data']['job']['aliases'] as $_v['alias']){?><span class="badge"><?php echo htmlspecialchars($_v['alias'], ENT_QUOTES);?>
</span> <?php }?></td></tr>
<?php }?>			<tr><th>Who can run</th><td><?php echo htmlspecialchars($_v['data']['job']['who'], ENT_QUOTES);?>
</td></tr>
			<tr><th>Hours</th><td><?php $_v['data_hours'] = $_v['data']['job']['hours']; if($_v['data_hours']['text']){ echo htmlspecialchars($_v['data_hours']['line'], ENT_QUOTES);?>
 <?php if($_v['data_hours']['open']){?>
<span class="badge ok">open now</span><?php }else{?>
<span class="badge bad">out of hours now</span><?php }?><?php }else{?>
<span class="faint">always</span><?php }?>
</td></tr>
			<tr><th>Input</th><td><?php if($_v['data']['job']['interactive']){?>
the job takes keyboard input<?php }else{?>
output only<?php }?></td></tr>
		</tbody>
	</table>
	</div>
</section>
<section class="block" data-live>
	<h3>History<?php if($_v['data']['count']){?>
 <small><?php echo htmlspecialchars($_v['data']['count'], ENT_QUOTES);?>
 runs</small><?php }?></h3>
<?php if($_v['data']['activity']){ $_v['data_chart'] = $_v['data']['activity'];if($_v['data_chart']['charts']){?>
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
<?php }?></div><?php }}?>	<h4>Runs <small>newest first</small></h4>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Runs as</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['runs']))foreach($_v['data']['runs'] as $_v['run']){?>			<tr>
				<td class="nowrap key" title="<?php echo htmlspecialchars($_v['run']['ago'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['run']['stamp'], ENT_QUOTES);?>
</td>
				<td data-label="By"><?php $_v['data_who'] = $_v['run']['who'];?>
<span class="who-ran"><?php $_v['data_user'] = $_v['data_who']['user'];?>
<span class="user-ref"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php $_v['data_menu'] = $_v['data_user']['menu'];if(!$_v['data_menu']['empty']){?>
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
</span>
<?php if($_v['data_who']['scheduled']){?>
 <span class="badge">scheduled</span><?php }?><?php if($_v['data_who']['reason']){?>
 <span class="badge bad" title="<?php echo htmlspecialchars($_v['data_who']['reason'], ENT_QUOTES);?>
">out of hours</span><?php }elseif($_v['data_who']['forced']){?>
 <span class="badge signal">forced</span><?php }?></span>
</td>
				<td class="nowrap" data-label="Took"><?php echo htmlspecialchars($_v['run']['duration'], ENT_QUOTES);?>
</td>
				<td data-label="Result"><?php $_v['data_state'] = $_v['run']['state']; if($_v['data_state']['id']){?>
<span class="state" data-run-id="<?php echo htmlspecialchars($_v['data_state']['id'], ENT_QUOTES);?>
"<?php if($_v['data_state']['ended']){?>
 data-finished="true"<?php }?>><i class="dot <?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES); if($_v['data_state']['code'] !== ""){?>
 <span class="faint">(<?php echo htmlspecialchars($_v['data_state']['code'], ENT_QUOTES);?>
)</span><?php }?></span><?php }else{?>
<span class="faint">never run</span><?php }?>
</td>
				<td class="actions">
					<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['run']['href'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['run']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['run']['attach'], ENT_QUOTES);?>
</a><?php if($_v['run']['pending']){?>
 <span class="faint">summarizing</span><?php } $_v['data_menu'] = $_v['run']['menu'];?>


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
</details><?php }?>

				</td>
			</tr>
<?php if($_v['run']['summary']){?>
			<tr class="summary-row">
				<td colspan="5"><details class="summary-details">
					<summary>Summary<?php if($_v['run']['summary']['by']){?>
 <span class="faint">asked by <?php echo htmlspecialchars($_v['run']['summary']['by'], ENT_QUOTES);?>
</span><?php }?></summary>
					<div class="markdown-doc"><?php echo $_v['run']['summary']['body'];?>
</div>
					<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['run']['summary']['href'], ENT_QUOTES);?>
">Report</a>
				</details></td>
			</tr>
<?php }?><?php }else{?>
			<tr class="empty"><td colspan="5">No runs yet.<?php if($_v['data']['can']){?>
 Press Run to start the first one.<?php }?></td></tr>
<?php }?>		</tbody>
	</table>
	</div>
</section>
<?php if($_v['data']['schedules']){?>
<section class="block">
<?php $_v['data_block'] = $_v['data']['schedules']['block'];?>

<div class="block-head<?php if($_v['data_block']['empty']){?>
 is-empty<?php }?>">
	<h3><?php echo htmlspecialchars($_v['data_block']['title'], ENT_QUOTES); if($_v['data_block']['note']){?>
 <small><?php echo htmlspecialchars($_v['data_block']['note'], ENT_QUOTES);?>
</small><?php }?></h3>
<?php if($_v['data_block']['empty']){?>
	<p class="empty-line"><?php echo htmlspecialchars($_v['data_block']['empty'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_block']['buttons']['items']){?>
	<div class="actions"><?php $_v['data_buttons'] = $_v['data_block']['buttons'];if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
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
<?php }?><?php }?>
	</div>
<?php }?></div>

<?php $_v['data_sched'] = $_v['data']['schedules']['table']; if($_v['data_sched']['rows']){?>
	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
<?php if(!empty($_v['data_sched']['rows']))foreach($_v['data_sched']['rows'] as $_v['row']){?>			<tr>
				<td class="key" data-label="Expression"><code><?php echo htmlspecialchars($_v['row']['expression'], ENT_QUOTES);?>
</code></td>
				<td data-label="Next run"><?php if($_v['row']['on']){ if(!empty($_v['row']['next']))foreach($_v['row']['next'] as $_v['part']){?><span class="nowrap"><?php echo htmlspecialchars($_v['part'], ENT_QUOTES);?>
</span> <?php }?><?php }else{?>
<span class="faint">-</span><?php }?></td>
				<td class="nowrap hide-narrow" data-label="Last run"><?php if($_v['row']['last'] !== ""){ echo htmlspecialchars($_v['row']['last'], ENT_QUOTES); }else{?>
<span class="faint">never</span><?php }?></td>
				<td data-label="Runs as"><?php if($_v['row']['pick']){?>
<form class="run-as" data-api="<?php echo htmlspecialchars($_v['row']['api'], ENT_QUOTES);?>
" data-method="PUT" data-next="reload" data-notice="Schedule saved"><?php $_v['data_avatar'] = $_v['row']['user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <select name="user" data-submit aria-label="Runs as">
<?php $_v['data_options'] = $_v['row']['options'];if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select><?php $_v['data_menu'] = $_v['row']['user']['menu'];if(!$_v['data_menu']['empty']){?>
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
</form><?php }else{ $_v['data_user'] = $_v['row']['user'];?>
<span class="user-ref"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php $_v['data_menu'] = $_v['data_user']['menu'];if(!$_v['data_menu']['empty']){?>
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
</span>
<?php }?></td>
				<td data-label="State"><?php if($_v['row']['on']){?>
<span class="badge ok">enabled</span><?php }else{?>
<span class="badge">disabled</span><?php }?><?php if($_v['row']['triage']){?>
 <span class="badge signal" title="Claude triages a failed run">triage</span><?php }?></td>
				<td class="actions"><?php if($_v['data_sched']['can']){ $_v['data_menu'] = $_v['row']['menu'];if(!$_v['data_menu']['empty']){?>
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
<?php }?>		</tbody>
	</table>
	</div>
<?php } if($_v['data']['can']){?>
	<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload" data-notice="Schedule added" data-sched>
		<input type="hidden" name="folder" value="<?php echo htmlspecialchars($_v['data']['schedules']['folder'], ENT_QUOTES);?>
">
		<input type="hidden" name="job" value="<?php echo htmlspecialchars($_v['data']['schedules']['job'], ENT_QUOTES);?>
">
		<div class="field whole">
			<label class="check"><input type="checkbox" id="sch-triage" name="triage"> Triage failures with Claude</label>
		<small class="faint">When the run fails, Claude reads the run in its folder, fixes the cause when it can, replies on the run's page and posts a note.</small>
		</div>
<?php $_v['data_runas'] = $_v['data']['schedules']['runAs']; if($_v['data_runas']['options']['users']){?>
	<div class="field">
		<label for="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
">Runs as</label>
		<select id="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
" name="user">
<?php $_v['data_options'] = $_v['data_runas']['options'];if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
		<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>
	</div>
<?php }?>	<div class="field">
		<label for="sch-every">Repeat</label>
		<select id="sch-every" data-sched-mode>
			<option value="hourly">Hourly</option>
			<option value="daily" selected>Daily</option>
			<option value="weekly">Weekly</option>
			<option value="monthly">Monthly</option>
			<option value="expression">Expression</option>
		</select>
		<small class="faint">Expression leaves the five fields to you.</small>
	</div>
	<div class="field" data-sched-when="hourly">
		<label for="sch-minute">At minute</label>
		<input type="number" id="sch-minute" min="0" max="59" step="1" value="0" data-sched-minute>
		<small class="faint">Past every hour, 0 to 59.</small>
	</div>
	<div class="field" data-sched-when="daily weekly monthly">
		<label for="sch-time">At</label>
		<input type="time" id="sch-time" value="03:00" data-sched-time>
		<small class="faint">The server's clock, not yours.</small>
	</div>
	<div class="field whole" data-sched-when="weekly">
		<span class="label">On these days</span>
		<div class="checks">
			<label class="check"><input type="checkbox" value="1" data-sched-day> Mon</label>
			<label class="check"><input type="checkbox" value="2" data-sched-day> Tue</label>
			<label class="check"><input type="checkbox" value="3" data-sched-day> Wed</label>
			<label class="check"><input type="checkbox" value="4" data-sched-day> Thu</label>
			<label class="check"><input type="checkbox" value="5" data-sched-day> Fri</label>
			<label class="check"><input type="checkbox" value="6" data-sched-day> Sat</label>
			<label class="check"><input type="checkbox" value="0" data-sched-day> Sun</label>
			<button class="btn btn-sm" type="button" data-sched-days="1,2,3,4,5">Weekdays</button>
			<button class="btn btn-sm" type="button" data-sched-days="0,6">Weekend</button>
			<button class="btn btn-sm" type="button" data-sched-days="">Every day</button>
		</div>
	</div>
	<div class="field whole" data-sched-when="monthly">
		<span class="label">On this day of the month</span>
		<div class="checks">
			<input type="number" id="sch-day" min="1" max="28" step="1" value="1" aria-label="Day of the month" data-sched-dom>
			<button class="btn btn-sm" type="button" data-sched-pick="1">First</button>
			<button class="btn btn-sm" type="button" data-sched-pick="15">Middle</button>
			<button class="btn btn-sm" type="button" data-sched-pick="28">Last</button>
			<span class="faint">1 to 28, the days every month has.</span>
		</div>
	</div>
	<div class="field whole">
		<label for="sch-expr">Expression</label>
		<input type="text" id="sch-expr" name="expression" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required data-sched-expr>
		<small class="faint" data-sched-says></small>
	</div>
	<div class="field whole">
		<div class="table-wrap fit">
		<table class="table fit-cards sched-legend">
			<thead><tr><th>Position</th><th>Field</th><th>Range</th><th>This one</th><th>What it says</th></tr></thead>
			<tbody>
				<tr><td class="key" data-label="Position">1</td><td data-label="Field">minute</td><td data-label="Range"><code>0-59</code></td><td data-label="This one"><code data-sched-pos="0"></code></td><td data-label="What it says">The minute of the hour it starts on.</td></tr>
				<tr><td class="key" data-label="Position">2</td><td data-label="Field">hour</td><td data-label="Range"><code>0-23</code></td><td data-label="This one"><code data-sched-pos="1"></code></td><td data-label="What it says">The hour of the day, on the server's clock.</td></tr>
				<tr><td class="key" data-label="Position">3</td><td data-label="Field">day of the month</td><td data-label="Range"><code>1-31</code></td><td data-label="This one"><code data-sched-pos="2"></code></td><td data-label="What it says">The date. A day past the end of a short month is skipped that month.</td></tr>
				<tr><td class="key" data-label="Position">4</td><td data-label="Field">month</td><td data-label="Range"><code>1-12</code></td><td data-label="This one"><code data-sched-pos="3"></code></td><td data-label="What it says">The month, by number or by name, <code>jan</code> to <code>dec</code>.</td></tr>
				<tr><td class="key" data-label="Position">5</td><td data-label="Field">day of the week</td><td data-label="Range"><code>0-6</code></td><td data-label="This one"><code data-sched-pos="4"></code></td><td data-label="What it says">Sunday is 0, and 7 is Sunday too, or by name, <code>sun</code> to <code>sat</code>.</td></tr>
			</tbody>
		</table>
		</div>
		<small class="faint"><code>*</code> is every, <code>*/5</code> every fifth, <code>1-5</code> a range and <code>1,15</code> a list. A day of the month and a day of the week both named must both hold, so <code>0 3 1 * mon</code> is the 1st when the 1st is a Monday. The whole-period names are <code>@hourly</code>, <code>@daily</code>, <code>@weekly</code>, <code>@monthly</code> and <code>@yearly</code>; <code>@midnight</code> is <code>@daily</code>.</small>
	</div>
	<div class="form-foot">
		<button class="btn" type="button" data-sched-cancel hidden>Cancel</button>
		<button class="btn btn-primary" type="submit"><svg class="ic" data-sched-icon><use href="#i-plus"/></svg><span data-sched-label>Add schedule</span></button>
	</div>
	</form>
<?php }?></section>
<?php }?>
