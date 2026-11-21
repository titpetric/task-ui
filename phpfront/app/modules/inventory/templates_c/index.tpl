<?php $_v=&$this->vars; $_v['data_head'] = $_v['data']['head'];?>

<div class="page-head">
	<div>
		<h2><?php if($_v['data_head']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['data_head']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i> <?php }?><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES); if($_v['data_head']['badge']){?>
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
"<?php }?>><i class="ic ic-play" aria-hidden="true"></i>Run<?php if($_v['data_run']['takesArgs']){?>
...<?php }?></button>
<?php if($_v['data_run']['menu']){?>
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
" aria-label="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
"><i class="ic ic-down" aria-hidden="true"></i></summary>
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
"<?php }?> title="Run and summarize"><i class="ic ic-summary" aria-hidden="true"></i><span>Run and summarize</span></button>
<?php }?><?php if($_v['data_run']['force']){?>
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['place'], ENT_QUOTES);?>
" title="Run out of hours, with a reason"><i class="ic ic-play" aria-hidden="true"></i><span>Run force</span></button>
<?php }?>			</div>
		</div>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data_head']['buttons']['items']){?>
		<?php $_v['data_buttons'] = $_v['data_head']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="<?php echo htmlspecialchars($_v['button']['className'], ENT_QUOTES);?>
"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }}?><?php if($_v['data_head']['actions']){?>
		<?php $_v['data_menu'] = $_v['data_head']['actions'];?>

		<?php if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><i class="ic ic-more" aria-hidden="true"></i></summary>
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
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }}?>	</div>
<?php }?></div>

<?php if($_v['data']['off']){?>
<p class="empty-state">The hosts below are the inventory's own. Beszel isn't set up, so no host has measures.<?php if($_v['data']['integrations']){?>
 Add its URL and token under <a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/integrations#beszel">Settings, Integrations</a>.<?php }?></p>
<?php }?><?php if($_v['data']['form']){?>
<section class="block inventory-run">
	<h3>Run on the hosts</h3>
	<form data-inventory-run>
		<div class="run-commands">
<?php if(!empty($_v['data']['form']['commands']))foreach($_v['data']['form']['commands'] as $_v['command']){?>			<button type="button" class="btn btn-sm" data-command="<?php echo htmlspecialchars($_v['command'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['command'], ENT_QUOTES);?>
</button>
<?php }?>		</div>
		<textarea name="command" rows="2" required spellcheck="false" placeholder="a shell command to run on every picked host" aria-label="The command"></textarea>
		<div class="run-hosts" data-host-picker>
			<div class="host-tools">
				<span class="faint" data-host-count></span>
				<button type="button" class="btn btn-sm" data-hosts="all">All</button>
				<button type="button" class="btn btn-sm" data-hosts="none">None</button>
				<button type="button" class="btn btn-sm" data-hosts="invert">Invert</button>
			</div>
			<div class="host-picks">
<?php if(!empty($_v['data']['form']['picks']))foreach($_v['data']['form']['picks'] as $_v['pick']){?>				<label class="host-pick"><input type="checkbox" name="hosts" value="<?php echo htmlspecialchars($_v['pick']['host'], ENT_QUOTES);?>
" checked><span><?php echo htmlspecialchars($_v['pick']['login'], ENT_QUOTES);?>
</span></label>
<?php }?>			</div>
		</div>
		<div class="form-foot note"><span></span><span class="actions"><?php if($_v['data']['form']['summarize']){?>
<button class="btn" type="submit" name="summarize" value="1"><i class="ic ic-note" aria-hidden="true"></i>Run and summarize</button><?php }?><button class="btn btn-primary" type="submit"><i class="ic ic-play" aria-hidden="true"></i>Run</button></span></div>
	</form>
		<div class="inventory-results" data-inventory-results></div>
<?php if($_v['data']['prompt']){ $_v['data_prompt'] = $_v['data']['prompt'];?>
<details class="promptbox"<?php if($_v['data_prompt']['open']){?>
 open<?php }?><?php if($_v['data_prompt']['hidden']){?>
 hidden<?php }?>>
	<summary><h3 class="section-title"><i class="ic ic-chev chev" aria-hidden="true"></i><?php echo htmlspecialchars($_v['data_prompt']['title'], ENT_QUOTES);?>
 <small><?php echo htmlspecialchars($_v['data_prompt']['note'], ENT_QUOTES);?>
</small></h3></summary>
<?php if(!$_v['data_prompt']['watch']){?>
	<?php $_v['data_composer'] = $_v['data_prompt']['composer'];?>
<form class="composer" data-send="<?php echo htmlspecialchars($_v['data_composer']['mode'], ENT_QUOTES);?>
"<?php if($_v['data_composer']['url']){?>
 data-send-url="<?php echo htmlspecialchars($_v['data_composer']['url'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_composer']['next']){?>
 data-next="<?php echo htmlspecialchars($_v['data_composer']['next'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_composer']['pane']){?>
 data-pane<?php }?>>
<?php if($_v['data_composer']['parent']){?>
	<input type="hidden" name="parent" value="<?php echo htmlspecialchars($_v['data_composer']['parent'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['channel']){?>
	<input type="hidden" name="channel" value="<?php echo htmlspecialchars($_v['data_composer']['channel'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['folder']){?>
	<input type="hidden" name="folder" value="<?php echo htmlspecialchars($_v['data_composer']['folder'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['dir']){?>
	<input type="hidden" name="dir" value="<?php echo htmlspecialchars($_v['data_composer']['dir'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['run']){?>
	<input type="hidden" name="run" value="<?php echo htmlspecialchars($_v['data_composer']['run'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['area']){?>
	<input type="hidden" name="area" value="<?php echo htmlspecialchars($_v['data_composer']['area'], ENT_QUOTES);?>
">
<?php }?><?php if($_v['data_composer']['origin']){?>
	<input type="hidden" name="origin" value="<?php echo htmlspecialchars($_v['data_composer']['origin'], ENT_QUOTES);?>
">
<?php }?>	<div class="composer-box">
		<?php $_v['data_markdown'] = $_v['data_composer']['editor'];?>
<textarea name="<?php echo htmlspecialchars($_v['data_markdown']['name'], ENT_QUOTES);?>
" rows="<?php echo htmlspecialchars($_v['data_markdown']['rows'], ENT_QUOTES);?>
" spellcheck="false" data-markdown data-toolbar="<?php echo htmlspecialchars($_v['data_markdown']['toolbar'], ENT_QUOTES);?>
"<?php if($_v['data_markdown']['mentions']){?>
 data-mentions<?php }?><?php if($_v['data_markdown']['placeholder']){?>
 placeholder="<?php echo htmlspecialchars($_v['data_markdown']['placeholder'], ENT_QUOTES);?>
"<?php }?> aria-label="<?php echo htmlspecialchars($_v['data_markdown']['label'], ENT_QUOTES);?>
"></textarea>
		<div class="composer-foot">
<?php if($_v['data_composer']['agents']){?>
			<label class="composer-field" data-composer-agent>Agent
				<select name="agent" aria-label="Answers as"><?php if(!empty($_v['data_composer']['agents']))foreach($_v['data_composer']['agents'] as $_v['name']){?><option value="<?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
"<?php if($_v['name'] == $_v['data_composer']['agent']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
</option><?php }?></select>
			</label>
<?php }?><?php if($_v['data_composer']['show_place']){?>
			<label class="composer-field" data-composer-place>In
				<input type="text" name="place" placeholder="folder/dir, empty is the root" spellcheck="false" autocomplete="off"<?php if($_v['data_composer']['places']){?>
 list="composer-places"<?php }?>>
<?php if($_v['data_composer']['places']){?>
				<datalist id="composer-places"><?php if(!empty($_v['data_composer']['places']))foreach($_v['data_composer']['places'] as $_v['place']){?><option value="<?php echo htmlspecialchars($_v['place'], ENT_QUOTES);?>
"><?php }?></datalist>
<?php }?>			</label>
<?php }?>			<span class="composer-space"></span>
<?php if($_v['data_composer']['mic']){?>
			<button class="btn btn-icon" type="button"<?php if($_v['data_composer']['parent']){?>
 data-voice-reply data-parent="<?php echo htmlspecialchars($_v['data_composer']['parent'], ENT_QUOTES);?>
" data-channel="<?php echo htmlspecialchars($_v['data_composer']['channel'], ENT_QUOTES);?>
" title="Record a message" aria-label="Record a message"<?php }else{?>
 data-voice-draft title="Record a prompt and transcribe it" aria-label="Record a prompt and transcribe it"<?php }?>><i class="ic ic-mic" aria-hidden="true"></i></button>
<?php }?>			<button class="btn btn-primary" type="submit" title="Enter sends, Shift+Enter adds a line">Send</button>
		</div>
	</div>
</form>

<?php }?><?php if($_v['data_prompt']['threads']){?>
	<ul class="thread-links" data-notes>
<?php if(!empty($_v['data_prompt']['threads']))foreach($_v['data_prompt']['threads'] as $_v['thread']){?>		<li><a href="<?php echo htmlspecialchars($_v['thread']['href'], ENT_QUOTES);?>
"><i class="ic ic-note" aria-hidden="true"></i> <?php echo htmlspecialchars($_v['thread']['title'], ENT_QUOTES);?>
</a><?php if($_v['thread']['state']){?>
 <span class="badge note-state <?php echo htmlspecialchars($_v['thread']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['thread']['state'], ENT_QUOTES);?>
</span><?php }?></li>
<?php }?>	</ul>
<?php }?></details>
<p class="faint threads-scope"><a class="threads-link" href="<?php echo htmlspecialchars($_v['data_prompt']['threads_href'], ENT_QUOTES);?>
">Threads in this place</a></p>
<?php }?></section>
<?php }?><section class="block" data-hosts-live>
<?php if($_v['data']['cluster']){?>
<div class="cluster">
	<h3 class="section-title">Cluster <small><?php echo htmlspecialchars($_v['data']['cluster']['count'], ENT_QUOTES);?>
 host<?php if($_v['data']['cluster']['count'] !== 1){?>
s<?php }?>, <?php echo htmlspecialchars($_v['data']['cluster']['up'], ENT_QUOTES);?>
 up<?php if($_v['data']['cluster']['down']){?>
, <span class="bad"><?php echo htmlspecialchars($_v['data']['cluster']['down'], ENT_QUOTES);?>
 down</span><?php }?></small></h3>
	<div class="pools">
<?php if(!empty($_v['data']['cluster']['pools']))foreach($_v['data']['cluster']['pools'] as $_v['pool']){?>		<div class="pool">
			<div class="pool-head">
				<span class="pool-name"><?php echo htmlspecialchars($_v['pool']['name'], ENT_QUOTES);?>
</span>
				<span class="pool-figure"><?php echo htmlspecialchars($_v['pool']['figure'], ENT_QUOTES);?>
</span><?php if($_v['pool']['share']){?>
<span class="pool-pct"><?php echo htmlspecialchars($_v['pool']['pct'], ENT_QUOTES);?>
%</span><?php }?>
			</div>
			<div class="pool-bar" role="img" aria-label="<?php echo htmlspecialchars($_v['pool']['name'], ENT_QUOTES);?>
: <?php echo htmlspecialchars($_v['pool']['figure'], ENT_QUOTES);?>
"><?php if(!empty($_v['pool']['segments']))foreach($_v['pool']['segments'] as $_v['segment']){?><span class="seg c<?php echo htmlspecialchars($_v['segment']['color'], ENT_QUOTES);?>
" style="flex-grow: <?php echo htmlspecialchars($_v['segment']['width'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['segment']['title'], ENT_QUOTES);?>
"><i style="width: <?php echo htmlspecialchars($_v['segment']['fill'], ENT_QUOTES);?>
%"></i></span><?php }?>
			</div>
		</div>
<?php }?>	</div>
	<ul class="pool-legend">
<?php if(!empty($_v['data']['cluster']['hosts']))foreach($_v['data']['cluster']['hosts'] as $_v['host']){?>		<li<?php if($_v['host']['left']){?>
 class="left"<?php }?>><i class="swatch<?php if(!$_v['host']['left']){?>
 c<?php echo htmlspecialchars($_v['host']['color'], ENT_QUOTES); }?>"></i><?php echo htmlspecialchars($_v['host']['name'], ENT_QUOTES);?>
</li>
<?php }?>	</ul>
<?php if($_v['data']['cluster']['left']){?>
	<p class="faint pool-note"><?php echo htmlspecialchars($_v['data']['cluster']['left'], ENT_QUOTES);?>
 host<?php if($_v['data']['cluster']['left'] !== 1){?>
s are down or send<?php }else{?>
 is down or sends<?php }?> no measures, greyed in the legend and left out of the pools.</p>
<?php }?><?php if($_v['data']['cluster']['charts']){?>
	<div class="charts"><?php if(!empty($_v['data']['cluster']['charts']))foreach($_v['data']['cluster']['charts'] as $_v['chart']){?><?php $_v['data_line_chart'] = $_v['chart'];?>
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
<?php }?></div>
<?php }?><h3>Hosts</h3>
<div class="table-wrap">
<table class="table hosts">
	<thead><tr><th>Host</th><?php if($_v['data']['logins']){?>
<th>SSH</th><?php }?><th class="center">Cores</th><th>CPU</th><th colspan="2">Memory <small>GB</small></th><th colspan="2">Disk <small>GB</small></th><th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
<?php if(!empty($_v['data']['rows']))foreach($_v['data']['rows'] as $_v['row']){?>		<tr<?php if($_v['row']['open']){?>
 class="host-row" data-host-row="<?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
" data-host-url="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
" tabindex="0" aria-expanded="false"<?php }?>>
			<td class="key nowrap" title="<?php echo htmlspecialchars($_v['row']['host'], ENT_QUOTES);?>
, <?php echo htmlspecialchars($_v['row']['status'], ENT_QUOTES);?>
"><?php if($_v['row']['open']){?>
<i class="ic ic-chev chev" aria-hidden="true"></i><?php }?><i class="ic ic-server host-state <?php if($_v['row']['up']){?>
up<?php }else{?>
down<?php }?>" role="img" aria-label="<?php echo htmlspecialchars($_v['row']['status'], ENT_QUOTES);?>
"></i> <a href="<?php echo htmlspecialchars($_v['row']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
</a></td>
<?php if($_v['data']['logins']){?>
			<td class="nowrap"><?php if($_v['row']['login']){ $_v['data_copy'] = $_v['row']['copy'];?>
<span class="copyable<?php if($_v['data_copy']['fill']){?>
 fill<?php }?>"><code<?php if($_v['data_copy']['fill'] || $_v['data_copy']['shown'] !== $_v['data_copy']['value']){?>
 title="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_copy']['shown'], ENT_QUOTES);?>
</code><button type="button" class="btn-icon" data-copy="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
" title="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
" aria-label="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
"><i class="ic ic-copy" aria-hidden="true"></i></button></span>
<?php }else{?>
<span class="faint">-</span><?php }?></td>
<?php }?>			<td class="center"><?php if($_v['row']['cpus']){ echo htmlspecialchars($_v['row']['cpus'], ENT_QUOTES); }else{?>
<span class="faint">-</span><?php }?></td>
			<td><?php $_v['data_meter'] = $_v['row']['cpu'];?>
<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
			</td>
			<td><?php $_v['data_meter'] = $_v['row']['memory'];?>
<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
			</td>
			<td class="size faint nowrap"><?php if($_v['row']['measured']){ echo htmlspecialchars($_v['row']['memoryUsed'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['row']['memoryTotal'], ENT_QUOTES); }?></td>
			<td><?php $_v['data_meter'] = $_v['row']['disk'];?>
<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
			</td>
			<td class="size faint nowrap"><?php if($_v['row']['measured']){ echo htmlspecialchars($_v['row']['diskUsed'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['row']['diskTotal'], ENT_QUOTES); }?></td>
			<td class="actions"><?php if($_v['row']['drop']){?>
<form class="inline inventory-drop" method="post" action="<?php echo htmlspecialchars($_v['row']['drop'], ENT_QUOTES);?>
" data-confirm="<?php echo htmlspecialchars($_v['row']['asked'], ENT_QUOTES);?>
"><button class="btn-icon danger" type="submit" title="Remove from the inventory" aria-label="Remove <?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
 from the inventory"><i class="ic ic-trash" aria-hidden="true"></i></button></form><?php }?><?php $_v['data_menu'] = $_v['row']['menu']; if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><i class="ic ic-more" aria-hidden="true"></i></summary>
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
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
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
		<tr class="empty"><td colspan="9">No hosts yet.<?php if($_v['data']['add']){?>
 Add one above.<?php }?></td></tr>
<?php }?>	</tbody>
</table>
</div>
<p class="faint live-stamp">Updated <?php echo htmlspecialchars($_v['data']['now'], ENT_QUOTES);?>
</p>
</section>
