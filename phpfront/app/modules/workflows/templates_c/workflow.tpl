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


<section class="block" data-sse="workflow:<?php echo htmlspecialchars($_v['data']['id'], ENT_QUOTES);?>
" data-user="<?php echo htmlspecialchars($_v['data']['user'], ENT_QUOTES);?>
">
	<h3>Pipeline</h3>
<?php if($_v['data']['latest']){?>
	<p class="faint" data-latest>Latest run <span class="state"><i class="dot <?php echo htmlspecialchars($_v['data']['latest']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data']['latest']['state'], ENT_QUOTES);?>
</span><?php if($_v['data']['latest']['message']){?>
 <?php echo htmlspecialchars($_v['data']['latest']['message'], ENT_QUOTES);?>
,<?php }?> <span title="<?php echo htmlspecialchars($_v['data']['latest']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['latest']['ago'], ENT_QUOTES);?>
</span> by <?php echo htmlspecialchars($_v['data']['latest']['userName'], ENT_QUOTES);?>
</p>
<?php }else{?>
	<p class="faint" data-latest>Not run yet.</p>
<?php }?>	<ol class="pipeline" data-pipeline>
<?php if(!empty($_v['data']['pipeline']))foreach($_v['data']['pipeline'] as $_v['data_node']){?><li class="node<?php if($_v['data_node']['state']){?>
 <?php echo htmlspecialchars($_v['data_node']['state'], ENT_QUOTES); }?><?php if($_v['data_node']['disabled']){?>
 disabled<?php }?>" data-step="<?php echo htmlspecialchars($_v['data_node']['pos'], ENT_QUOTES);?>
">
	<span class="rail"><i class="dot<?php if($_v['data_node']['state']){?>
 <?php echo htmlspecialchars($_v['data_node']['state'], ENT_QUOTES); }?>"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_node']['href'], ENT_QUOTES);?>
"<?php if($_v['data_node']['runId']){?>
 data-run-link="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['data_node']['workflow']){?>
<svg class="ic lead"><use href="#i-flow"/></svg><?php }elseif($_v['data_node']['summary']){?>
<svg class="ic lead"><use href="#i-summary"/></svg><?php }elseif($_v['data_node']['check']){?>
<svg class="ic lead"><use href="#i-rotate"/></svg><?php }elseif($_v['data_node']['prompt']){?>
<svg class="ic lead"><use href="#i-prompt"/></svg><?php }?><?php echo htmlspecialchars($_v['data_node']['text'], ENT_QUOTES);?>
</a>
<?php if($_v['data_node']['expand']){?>
			<button type="button" class="btn btn-sm" data-expand="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand<?php if($_v['data_node']['lines']){?>
 <?php echo htmlspecialchars($_v['data_node']['lines'], ENT_QUOTES);?>
 lines<?php }?></span></button>
<?php }?>		</div>
		<small><?php echo htmlspecialchars($_v['data_node']['small'], ENT_QUOTES);?>
</small>
<?php if($_v['data_node']['prompt']){?>
		<p class="step-prompt"><?php echo htmlspecialchars($_v['data_node']['prompt'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_node']['runId']){?>
		<a class="frame" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
"<?php if($_v['data_node']['live']){?>
 data-frame="<?php echo htmlspecialchars($_v['data_node']['runId'], ENT_QUOTES);?>
"<?php }?> role="img" aria-label="Terminal of <?php echo htmlspecialchars($_v['data_node']['text'], ENT_QUOTES);?>
"><?php echo $_v['data_node']['frame'];?>
</a>
<?php }?><?php if($_v['data_node']['subId']){?>
		<ol class="pipeline sub" data-sub="<?php echo htmlspecialchars($_v['data_node']['subId'], ENT_QUOTES);?>
" data-workflow="<?php echo htmlspecialchars($_v['data_node']['workflowId'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_node']['sub']))foreach($_v['data_node']['sub'] as $_v['data_sub']){?><li class="node<?php if($_v['data_sub']['state']){?>
 <?php echo htmlspecialchars($_v['data_sub']['state'], ENT_QUOTES); }?><?php if($_v['data_sub']['disabled']){?>
 disabled<?php }?>" data-step="<?php echo htmlspecialchars($_v['data_sub']['pos'], ENT_QUOTES);?>
">
	<span class="rail"><i class="dot<?php if($_v['data_sub']['state']){?>
 <?php echo htmlspecialchars($_v['data_sub']['state'], ENT_QUOTES); }?>"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_sub']['href'], ENT_QUOTES);?>
"<?php if($_v['data_sub']['runId']){?>
 data-run-link="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['data_sub']['workflow']){?>
<svg class="ic lead"><use href="#i-flow"/></svg><?php }elseif($_v['data_sub']['summary']){?>
<svg class="ic lead"><use href="#i-summary"/></svg><?php }elseif($_v['data_sub']['check']){?>
<svg class="ic lead"><use href="#i-rotate"/></svg><?php }elseif($_v['data_sub']['prompt']){?>
<svg class="ic lead"><use href="#i-prompt"/></svg><?php }?><?php echo htmlspecialchars($_v['data_sub']['text'], ENT_QUOTES);?>
</a>
<?php if($_v['data_sub']['expand']){?>
			<button type="button" class="btn btn-sm" data-expand="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand<?php if($_v['data_sub']['lines']){?>
 <?php echo htmlspecialchars($_v['data_sub']['lines'], ENT_QUOTES);?>
 lines<?php }?></span></button>
<?php }?>		</div>
		<small><?php echo htmlspecialchars($_v['data_sub']['small'], ENT_QUOTES);?>
</small>
<?php if($_v['data_sub']['prompt']){?>
		<p class="step-prompt"><?php echo htmlspecialchars($_v['data_sub']['prompt'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_sub']['runId']){?>
		<a class="frame" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php if($_v['data_sub']['live']){?>
 data-frame="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php }?> role="img" aria-label="Terminal of <?php echo htmlspecialchars($_v['data_sub']['text'], ENT_QUOTES);?>
"><?php echo $_v['data_sub']['frame'];?>
</a>
<?php }?><?php if($_v['data_sub']['subId']){?>
		<ol class="pipeline sub" data-sub="<?php echo htmlspecialchars($_v['data_sub']['subId'], ENT_QUOTES);?>
" data-workflow="<?php echo htmlspecialchars($_v['data_sub']['workflowId'], ENT_QUOTES);?>
">
		</ol>
<?php }?>	</div>
</li>
<?php }?>		</ol>
<?php }?>	</div>
</li>
<?php }?>	</ol>
</section>

<section class="block">
	<h3>Runs</h3>
<?php if($_v['data']['activity']){ $_v['data_chart'] = $_v['data']['activity'];?>


<?php if($_v['data_chart']['charts']){?>
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
<?php }?></div><?php }}?>	<form method="get" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/workflows/<?php echo htmlspecialchars($_v['data']['id'], ENT_QUOTES);?>
" class="filters">
		<label class="page-size" for="page-size">Page size
			<select id="page-size" name="n">
<?php if(!empty($_v['data']['sizes']))foreach($_v['data']['sizes'] as $_v['size']){?><option value="<?php echo htmlspecialchars($_v['size'], ENT_QUOTES);?>
"<?php if($_v['data']['form']['n'] === $_v['size']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['size'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
		</label>
		<button class="btn" type="submit">Filter</button>
	</form>
	<div class="table-wrap">
	<table class="table cards wf-runs">
		<thead><tr><th>Run</th><th>By</th><th>Steps</th><th></th></tr></thead>
		<tbody data-runs<?php if($_v['data']['summarize']){?>
 data-summarize<?php }?>>
<?php if(!empty($_v['data']['rows']))foreach($_v['data']['rows'] as $_v['row']){?>			<tr data-run-id="<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
">
				<td class="key" data-result><?php $_v['data_wffold'] = $_v['row']['fold']; if($_v['data_wffold']['answer']){?>
<button type="button" class="btn-icon wf-fold" data-answer-toggle aria-expanded="<?php if($_v['data_wffold']['open']){?>
true<?php }else{?>
false<?php }?>" aria-controls="wf-answer-<?php echo htmlspecialchars($_v['data_wffold']['id'], ENT_QUOTES);?>
" title="<?php if($_v['data_wffold']['open']){?>
Hide<?php }else{?>
Show<?php }?> the answer" aria-label="<?php if($_v['data_wffold']['open']){?>
Hide<?php }else{?>
Show<?php }?> the answer"><svg class="ic chev"><use href="#i-chev"/></svg></button><?php }else{?>
<span class="wf-fold" aria-hidden="true"></span><?php }?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><i class="dot <?php echo htmlspecialchars($_v['row']['state'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['row']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</a><?php if($_v['row']['message']){?>
<small class="faint wf-desc"><?php echo htmlspecialchars($_v['row']['message'], ENT_QUOTES);?>
</small><?php }?></td>
				<td data-label="By"><?php $_v['data_user'] = $_v['row']['user'];?>
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
</td>
				<td data-label="Steps"><span class="chain"><?php if(!empty($_v['row']['chain']))foreach($_v['row']['chain'] as $_v['link']){?><?php if($_v['link']['then']){?>
<span class="faint then" aria-hidden="true"></span><?php }?><a class="nowrap" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['link']['href'], ENT_QUOTES);?>
" data-pos="<?php echo htmlspecialchars($_v['link']['pos'], ENT_QUOTES);?>
"><i class="dot <?php echo htmlspecialchars($_v['link']['state'], ENT_QUOTES);?>
"></i><?php if($_v['link']['workflow']){?>
<svg class="ic lead"><use href="#i-flow"/></svg><?php }elseif($_v['link']['summary']){?>
<svg class="ic lead"><use href="#i-summary"/></svg><?php }elseif($_v['link']['check']){?>
<svg class="ic lead"><use href="#i-rotate"/></svg><?php }elseif($_v['link']['prompt']){?>
<svg class="ic lead"><use href="#i-prompt"/></svg><?php }?><?php echo htmlspecialchars($_v['link']['text'], ENT_QUOTES);?>
</a><?php }?></span></td>
				<td class="actions"><a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
" data-workflow-view>View</a><?php $_v['data_menu'] = $_v['row']['menu'];if(!$_v['data_menu']['empty']){?>
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
<?php $_v['data_wfanswer'] = $_v['row']['answer'];if($_v['data_wfanswer']['body']){?>
			<tr class="wf-answer-row" id="wf-answer-<?php echo htmlspecialchars($_v['data_wfanswer']['id'], ENT_QUOTES);?>
" data-answer-of="<?php echo htmlspecialchars($_v['data_wfanswer']['id'], ENT_QUOTES);?>
"<?php if(!$_v['data_wfanswer']['open']){?>
 hidden<?php }?>><td colspan="<?php echo htmlspecialchars($_v['data_wfanswer']['cols'], ENT_QUOTES);?>
"><div class="note-body"><?php echo $_v['data_wfanswer']['body'];?>
</div></td></tr>
<?php }}else{?>
			<tr class="empty"><td colspan="4">Not run yet.</td></tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php $_v['data_pager'] = $_v['data']['pager'];?>


<?php if($_v['data_pager']['prev'] || $_v['data_pager']['next']){?>
	<nav class="pager" aria-label="Pages">
<?php if($_v['data_pager']['prev']){?>
		<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_pager']['prev'], ENT_QUOTES);?>
" rel="prev">Previous</a>
<?php }?><?php if($_v['data_pager']['next']){?>
		<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_pager']['next'], ENT_QUOTES);?>
" rel="next">Next</a>
<?php }?>	</nav>
<?php }?>

</section>
