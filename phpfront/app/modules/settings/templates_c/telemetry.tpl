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

<form class="form telemetry" <?php if($_v['data']['admin']){?>
data-api="/api/telemetry" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['data']['next'], ENT_QUOTES);?>
"<?php }?> novalidate>
<section class="block">
	<h3>Recording <small>what a request leaves behind</small></h3>

	<div class="field">
		<label class="switch-row" for="telemetry-enabled">
			<span class="switch">
				<input type="checkbox" id="telemetry-enabled" name="enabled" value="1" data-type="toggle"<?php if($_v['data']['opts']['enabled']){?>
 checked<?php }?><?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
				<i aria-hidden="true"></i>
			</span>
			<span>Enabled</span>
		</label>
		<p class="faint">Off, a request passes through and records nothing. This is the one setting that takes effect on its own: everything else below rebuilds the recorder, which empties the traces it holds.</p>
	</div>

	<div class="field">
		<label for="telemetry-service">Service name <span class="faint">on every trace, and in the dashboard's heading</span></label>
		<input type="text" id="telemetry-service" name="service_name" value="<?php echo htmlspecialchars($_v['data']['opts']['service_name'], ENT_QUOTES);?>
"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
	</div>

	<div class="row">
		<div class="field">
			<label for="telemetry-rate">Sample rate <span class="faint">percent of requests traced</span></label>
			<input type="number" id="telemetry-rate" name="sample_rate" min="0" max="100" step="1" value="<?php echo htmlspecialchars($_v['data']['rate'], ENT_QUOTES);?>
" data-type="int"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
		</div>
		<div class="field">
			<label for="telemetry-ring">Retained traces <span class="faint">the oldest go first</span></label>
			<input type="number" id="telemetry-ring" name="ring_buffer_size" min="1" step="1" value="<?php echo htmlspecialchars($_v['data']['opts']['ring_buffer_size'], ENT_QUOTES);?>
" data-type="int"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
		</div>
	</div>

	<div class="row">
		<div class="field">
			<label for="telemetry-top">Top requests <span class="faint">groups the statistics keep</span></label>
			<input type="number" id="telemetry-top" name="top_requests" min="1" step="1" value="<?php echo htmlspecialchars($_v['data']['opts']['top_requests'], ENT_QUOTES);?>
" data-type="int"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
		</div>
		<div class="field">
			<label for="telemetry-spans">Spans per trace <span class="faint">the rest are counted, not kept; 0 is no cap</span></label>
			<input type="number" id="telemetry-spans" name="max_spans_per_trace" min="0" step="1" value="<?php echo htmlspecialchars($_v['data']['opts']['max_spans_per_trace'], ENT_QUOTES);?>
" data-type="int"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
		</div>
	</div>

	<div class="field">
		<label for="telemetry-ignore">Ignored paths <span class="faint">one per line; an entry ending in <code>/*</code> matches by prefix</span></label>
		<textarea id="telemetry-ignore" name="ignore_paths" rows="4"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>><?php echo htmlspecialchars($_v['data']['ignore'], ENT_QUOTES);?>
</textarea>
	</div>
</section>

<section class="block">
	<h3>The dashboard <small>how the live view reads, and who may open it</small></h3>

	<div class="field">
		<label class="switch-row" for="telemetry-stream">
			<span class="switch">
				<input type="checkbox" id="telemetry-stream" name="live_stream" value="1" data-type="toggle"<?php if($_v['data']['opts']['live_stream']){?>
 checked<?php }?><?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
				<i aria-hidden="true"></i>
			</span>
			<span>Live stream</span>
		</label>
		<p class="faint">Serves the live view over server sent events, so a trace appears as it happens instead of on a timer.</p>
	</div>

	<div class="field">
		<label for="telemetry-refresh">Refresh interval <span class="faint">seconds, where the browser cannot stream; 0 never reloads</span></label>
		<input type="number" id="telemetry-refresh" name="refresh_interval" min="0" step="1" value="<?php echo htmlspecialchars($_v['data']['opts']['refresh_interval'], ENT_QUOTES);?>
" data-type="int"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
	</div>

	<div class="field">
		<label for="telemetry-networks">Allowed networks <span class="faint">CIDR ranges, one per line; empty is every network</span></label>
		<textarea id="telemetry-networks" name="allowed_networks" rows="3"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>><?php echo htmlspecialchars($_v['data']['networks'], ENT_QUOTES);?>
</textarea>
		<p class="faint">The dashboard is an admin's either way: a trace holds request paths, query strings and the errors a handler recorded, and a request from outside these ranges gets a 404 rather than a refusal.</p>
	</div>

	<div class="field">
		<label for="telemetry-path">Path <span class="faint">where the dashboard is mounted</span></label>
		<input type="text" id="telemetry-path" name="path" value="<?php echo htmlspecialchars($_v['data']['opts']['option_path'], ENT_QUOTES);?>
"<?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
		<p class="faint">The router reads this once, at startup, so a path saved here takes a restart. It is <code><?php echo htmlspecialchars($_v['data']['path'], ENT_QUOTES);?>
</code> now.</p>
	</div>
</section>

<section class="block">
	<h3>What a trace holds <small>beside the spans</small></h3>

	<div class="field">
		<label class="switch-row" for="telemetry-logs">
			<span class="switch">
				<input type="checkbox" id="telemetry-logs" name="capture_logs" value="1" data-type="toggle"<?php if($_v['data']['opts']['capture_logs']){?>
 checked<?php }?><?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
				<i aria-hidden="true"></i>
			</span>
			<span>Capture logs</span>
		</label>
		<p class="faint">Records the log lines a trace writes. Off, an error still lands on its span.</p>
	</div>

	<div class="field">
		<label class="switch-row" for="telemetry-memory">
			<span class="switch">
				<input type="checkbox" id="telemetry-memory" name="track_memory_use" value="1" data-type="toggle"<?php if($_v['data']['opts']['track_memory_use']){?>
 checked<?php }?><?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
				<i aria-hidden="true"></i>
			</span>
			<span>Track memory use</span>
		</label>
		<p class="faint">Records the process's allocation change per trace.</p>
	</div>

	<div class="field">
		<label class="switch-row" for="telemetry-request-id">
			<span class="switch">
				<input type="checkbox" id="telemetry-request-id" name="trust_request_id" value="1" data-type="toggle"<?php if($_v['data']['opts']['trust_request_id']){?>
 checked<?php }?><?php if(!$_v['data']['admin']){?>
 disabled<?php }?>>
				<i aria-hidden="true"></i>
			</span>
			<span>Trust Request-Id</span>
		</label>
		<p class="faint">Reuses a client's <code>Request-Id</code> header as the trace identifier, so a request crossing two services reads as one trace. Behind a trusted proxy alone: a client that picks its own identifier can collide with another's.</p>
	</div>
<?php if($_v['data']['admin']){?>
	<div class="form-foot">
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
<?php }?></section>
</form>
<?php if($_v['data']['trail']){ $_v['data_trail'] = $_v['data']['trail'];?>

<nav class="setup-trail" aria-label="Setup steps">
<?php if($_v['data_trail']['previous']){?>
	<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data_trail']['previous']['href'], ENT_QUOTES);?>
" rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: <?php echo htmlspecialchars($_v['data_trail']['previous']['name'], ENT_QUOTES);?>
</a>
<?php }?>	<ol class="steps">
<?php if(!empty($_v['data_trail']['steps']))foreach($_v['data_trail']['steps'] as $_v['step']){?>		<li><?php if($_v['step']['current']){?>
<b aria-current="step"><?php echo htmlspecialchars($_v['step']['name'], ENT_QUOTES);?>
</b><?php }else{?>
<a href="<?php echo htmlspecialchars($_v['step']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['step']['name'], ENT_QUOTES);?>
</a><?php }?></li>
<?php }?>	</ol>
<?php if($_v['data_trail']['next']){?>
	<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data_trail']['next']['href'], ENT_QUOTES);?>
" rel="next">Next: <?php echo htmlspecialchars($_v['data_trail']['next']['name'], ENT_QUOTES);?>
<svg class="ic"><use href="#i-chev"/></svg></a>
<?php }?></nav>
<?php }?>
