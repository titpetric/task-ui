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


<form method="get" action="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs" class="filters">
	<label for="runs-folder">Folder
		<select id="runs-folder" name="folder">
			<option value="">all folders</option>
<?php if(!empty($_v['data']['folders']))foreach($_v['data']['folders'] as $_v['folder']){?><option value="<?php echo htmlspecialchars($_v['folder'], ENT_QUOTES);?>
"<?php if($_v['data']['form']['folder'] === $_v['folder']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['folder'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
	</label>
	<label>Job <input type="text" name="job" value="<?php echo htmlspecialchars($_v['data']['form']['job'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off"></label>
	<label for="runs-state">State
		<select id="runs-state" name="state">
			<option value="">any state</option>
<?php if(!empty($_v['data']['states']))foreach($_v['data']['states'] as $_v['state']){?><option value="<?php echo htmlspecialchars($_v['state'], ENT_QUOTES);?>
"<?php if($_v['data']['form']['state'] === $_v['state']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['state'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
	</label>
	<label for="runs-user">By
		<select id="runs-user" name="user">
			<option value="">anyone</option>
<?php if(!empty($_v['data']['users']))foreach($_v['data']['users'] as $_v['user']){?><option value="<?php echo htmlspecialchars($_v['user'], ENT_QUOTES);?>
"<?php if($_v['data']['form']['user'] === $_v['user']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['user'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
	</label>
	<label class="page-size" for="page-size">Page size
		<select id="page-size" name="n">
<?php if(!empty($_v['data']['sizes']))foreach($_v['data']['sizes'] as $_v['size']){?><option value="<?php echo htmlspecialchars($_v['size'], ENT_QUOTES);?>
"<?php if($_v['data']['form']['n'] === $_v['size']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['size'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
	</label>
	<button class="btn" type="submit">Filter</button>
</form>

<?php if($_v['data']['activity']){?>
<section class="block" data-fold="runs-activity">
	<h3><?php $_v['data_fold'] = $_v['data']['fold'];?>
<button type="button" class="fold-toggle" aria-expanded="true" aria-controls="<?php echo htmlspecialchars($_v['data_fold']['id'], ENT_QUOTES);?>
"><svg class="ic chev"><use href="#i-chev"/></svg><?php echo htmlspecialchars($_v['data_fold']['label'], ENT_QUOTES);?>
</button>
 <small>in the server's time zone</small></h3>
	<div class="fold-body" id="runs-activity">
<?php $_v['data_chart'] = $_v['data']['activity'];if($_v['data_chart']['charts']){?>
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

	</div>
</section>
<?php } if($_v['data']['tooling']){?>
<section class="block" data-fold="runs-tooling">
	<h3><?php $_v['data_fold'] = $_v['data']['toolFold'];?>
<button type="button" class="fold-toggle" aria-expanded="true" aria-controls="<?php echo htmlspecialchars($_v['data_fold']['id'], ENT_QUOTES);?>
"><svg class="ic chev"><use href="#i-chev"/></svg><?php echo htmlspecialchars($_v['data_fold']['label'], ENT_QUOTES);?>
</button>
 <small>what the commands inside the runs spent their time on</small></h3>
	<div class="fold-body" id="runs-tooling">
<?php $_v['data_tooling'] = $_v['data']['tooling'];if($_v['data_tooling']['scopes']){?>
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

	</div>
</section>
<?php }?>
<section class="block" data-sse="runs">
	<h3>Runs <small>newest first</small></h3>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Folder</th><th>Job</th><th>By</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody data-runs-table<?php if($_v['data']['summarize']){?>
 data-summarize<?php }?>>
<?php if(!empty($_v['data']['rows']))foreach($_v['data']['rows'] as $_v['row']){?>			<tr data-run-id="<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
">
				<td class="nowrap" data-label="Started" title="<?php echo htmlspecialchars($_v['row']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['ago'], ENT_QUOTES);?>
</td>
				<td data-label="Folder"><?php if($_v['row']['folderHref']){?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['row']['folderHref'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['place'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['row']['place'], ENT_QUOTES); }?></td>
				<td class="key"><?php if($_v['row']['jobHref']){?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['row']['jobHref'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['job'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['row']['job'], ENT_QUOTES); }?></td>
				<td data-label="By"><span class="who-ran"><?php $_v['data_user'] = $_v['row']['who']['user'];?>
<span class="user-ref"><?php if($_v['data_user']['href']){?>
<a class="user-link" href="<?php echo htmlspecialchars($_v['data_user']['href'], ENT_QUOTES);?>
"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span></a><?php }else{ $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php }?></span>
<?php if($_v['row']['who']['scheduled']){?>
 <span class="badge">scheduled</span><?php }?><?php if($_v['row']['who']['reason']){?>
 <span class="badge bad" title="<?php echo htmlspecialchars($_v['row']['who']['reason'], ENT_QUOTES);?>
">out of hours</span><?php }elseif($_v['row']['who']['forced']){?>
 <span class="badge signal">forced</span><?php }?></span></td>
				<td class="nowrap" data-label="Took" data-took data-started="<?php echo htmlspecialchars($_v['row']['started'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['duration'], ENT_QUOTES);?>
</td>
				<td data-label="Result" data-result><?php $_v['data_state'] = $_v['row']['state']; if($_v['data_state']['id']){?>
<span class="state" data-run-id="<?php echo htmlspecialchars($_v['data_state']['id'], ENT_QUOTES);?>
"<?php if($_v['data_state']['ended']){?>
 data-finished="true"<?php }?>><i class="dot <?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES); if($_v['data_state']['code'] !== ""){?>
 <span class="faint">(<?php echo htmlspecialchars($_v['data_state']['code'], ENT_QUOTES);?>
)</span><?php }?></span><?php }else{?>
<span class="faint">never run</span><?php } if($_v['row']['message']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['row']['message'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['row']['title']){?>
 <span class="term-title" title="<?php echo htmlspecialchars($_v['row']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['title'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['row']['watching']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['row']['watching'], ENT_QUOTES);?>
 watching</span><?php }?></td>
				<td class="actions" data-actions>
					<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['attach'], ENT_QUOTES);?>
</a>
<?php $_v['data_menu'] = $_v['row']['menu'];if(!$_v['data_menu']['empty']){?>
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
<?php }else{?>
			<tr class="empty"><td colspan="7">No runs match. Run a job from its folder and it appears here.</td></tr>
<?php }?>		</tbody>
	</table>
	</div>
<?php $_v['data_pager'] = $_v['data']['pager'];if($_v['data_pager']['prev'] || $_v['data_pager']['next']){?>
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
