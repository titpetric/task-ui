<?php $_v=&$this->vars;?>
<form class="form" data-api="<?php echo htmlspecialchars($_v['data']['api'], ENT_QUOTES);?>
" data-method="<?php echo htmlspecialchars($_v['data']['method'], ENT_QUOTES);?>
" data-next="<?php echo htmlspecialchars($_v['data']['next'], ENT_QUOTES);?>
" data-sched>
<?php if($_v['data']['edit']){?>
	<div class="field">
		<span class="label">Target</span>
		<p><?php if($_v['data']['target'] === "workflow"){?>
<i class="ic ic-flow lead" aria-hidden="true"></i><?php if($_v['data']['workflowHref']){?>
<a href="<?php echo htmlspecialchars($_v['data']['workflowHref'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['workflowName'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['data']['workflowName'], ENT_QUOTES); }?><?php }else{ if($_v['data']['targetHref']){?>
<a href="<?php echo htmlspecialchars($_v['data']['targetHref'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['data']['job'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['data']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['data']['job'], ENT_QUOTES); }?><?php }?></p>
		<small class="faint">What a schedule is for cannot change; delete it and add another.</small>
	</div>
<?php if($_v['data']['target'] === "workflow"){?>
	<input type="hidden" name="workflow" value="<?php echo htmlspecialchars($_v['data']['workflow'], ENT_QUOTES);?>
">
<?php if($_v['data']['pickUser']){ $_v['data_runas'] = $_v['data']['runAs']; if($_v['data_runas']['options']['users']){?>
	<div class="field">
		<label for="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
">Runs as</label>
		<select id="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
" name="user">
<?php $_v['data_options'] = $_v['data_runas']['options']; if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
		<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>
	</div>
<?php }?><?php } }else{?>
	<input type="hidden" name="folder" value="<?php echo htmlspecialchars($_v['data']['folder'], ENT_QUOTES);?>
">
	<input type="hidden" name="job" value="<?php echo htmlspecialchars($_v['data']['job'], ENT_QUOTES);?>
">
	<div class="field">
		<label for="sch-args">Args</label>
		<input type="text" id="sch-args" name="args" data-type="args" value="<?php echo htmlspecialchars($_v['data']['args'], ENT_QUOTES);?>
" placeholder="df -h /" spellcheck="false" autocomplete="off">
		<small class="faint">For a job that takes them; the command of a host.</small>
	</div>
	<div class="field whole">
		<label class="check"><input type="checkbox" id="sch-triage" name="triage"<?php if($_v['data']['triage']){?>
 checked<?php }?>> Triage failures with Claude</label>
		<small class="faint">When the run fails, Claude reads the run in its folder, fixes the cause when it can, replies on the run's page and posts a note.</small>
	</div>
<?php if($_v['data']['pickUser']){ $_v['data_runas'] = $_v['data']['runAs']; if($_v['data_runas']['options']['users']){?>
	<div class="field">
		<label for="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
">Runs as</label>
		<select id="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
" name="user">
<?php $_v['data_options'] = $_v['data_runas']['options']; if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
		<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>
	</div>
<?php }?><?php } } }else{?>
	<div class="field">
		<label for="sch-workflow">Workflow</label>
		<select id="sch-workflow" name="workflow" data-sched-workflow>
			<option value="">A job</option>
<?php if(!empty($_v['data']['workflows']))foreach($_v['data']['workflows'] as $_v['flow']){?>			<option value="<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
"<?php if($_v['flow']['id'] === $_v['data']['workflow']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
		<small class="faint">Pick a workflow to schedule its steps instead of one job.</small>
	</div>
	<div data-sched-job>
		<div class="field">
			<label for="sch-folder">Folder</label>
			<select id="sch-folder" name="folder" data-narrows="sch-job" required>
<?php if(!empty($_v['data']['folders']))foreach($_v['data']['folders'] as $_v['folder']){?>				<option value="<?php echo htmlspecialchars($_v['folder']['id'], ENT_QUOTES);?>
"<?php if($_v['folder']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['folder']['id'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
		</div>
		<div class="field">
			<label for="sch-job">Job</label>
			<select id="sch-job" name="job" required>
<?php if(!empty($_v['data']['jobs']))foreach($_v['data']['jobs'] as $_v['job']){?>				<option value="<?php echo htmlspecialchars($_v['job']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['data_hostjobs'] = $_v['data']['hosts']; if($_v['data_hostjobs']){?>
			<optgroup label="Hosts">
<?php if(!empty($_v['data_hostjobs']))foreach($_v['data_hostjobs'] as $_v['job']){?>				<option value="<?php echo htmlspecialchars($_v['job']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['job']['title'], ENT_QUOTES);?>
"<?php if($_v['job']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</option>
<?php }?>			</optgroup>
<?php }?>			</select>
		</div>
		<div class="field">
			<label for="sch-args">Args</label>
			<input type="text" id="sch-args" name="args" data-type="args" placeholder="df -h /" spellcheck="false" autocomplete="off">
			<small class="faint">For a job that takes them; the command of a host.</small>
		</div>
		<div class="field whole">
			<label class="check"><input type="checkbox" id="sch-triage" name="triage"> Triage failures with Claude</label>
			<small class="faint">When the run fails, Claude reads the run in its folder, fixes the cause when it can, replies on the run's page and posts a note.</small>
		</div>
<?php $_v['data_runas'] = $_v['data']['runAs']; if($_v['data_runas']['options']['users']){?>
	<div class="field">
		<label for="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
">Runs as</label>
		<select id="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
" name="user">
<?php $_v['data_options'] = $_v['data_runas']['options']; if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
		<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>
	</div>
<?php }?>	</div>
<?php }?>
	<div class="field">
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
		<input type="text" id="sch-expr" name="expression" value="<?php echo htmlspecialchars($_v['data']['expression'], ENT_QUOTES);?>
" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required data-sched-expr>
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

	<div class="field">
		<label class="check"><input type="checkbox" name="enabled"<?php if($_v['data']['enabled']){?>
 checked<?php }?>> Enabled</label>
	</div>
	<div class="form-foot">
		<a class="btn" href="<?php echo htmlspecialchars($_v['data']['back'], ENT_QUOTES);?>
">Cancel</a>
		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>
