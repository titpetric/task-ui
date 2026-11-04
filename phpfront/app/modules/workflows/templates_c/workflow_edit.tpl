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


<?php $_v['data_editor'] = $_v['data']['editor'];?>

<form class="form workflow-form" data-workflow data-api="<?php echo htmlspecialchars($_v['data_editor']['api'], ENT_QUOTES);?>
" data-method="<?php echo htmlspecialchars($_v['data_editor']['method'], ENT_QUOTES);?>
" data-next="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_editor']['next'], ENT_QUOTES);?>
" novalidate>
	<div class="wf-meta">
		<div class="field">
			<label for="wf-name-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
">Name</label>
			<input type="text" id="wf-name-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
" name="name" value="<?php echo htmlspecialchars($_v['data_editor']['name'], ENT_QUOTES);?>
" placeholder="nightly deploy" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="wf-desc-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
">Description <span class="faint">optional</span></label>
			<textarea id="wf-desc-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
" name="description" rows="2" placeholder="What the workflow is for, in a line or two"><?php echo htmlspecialchars($_v['data_editor']['description'], ENT_QUOTES);?>
</textarea>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="sidebar"<?php if($_v['data_editor']['sidebar']){?>
 checked<?php }?>> Show in sidebar</label>
			<small class="faint">Off for a workflow you only use as a step of others.</small>
		</div>
		<div class="field">
<?php $_v['data_options'] = $_v['data_editor']['users']; if($_v['data_options']['users']){?>
			<label for="wf-user-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
">Runs as</label>
			<select id="wf-user-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
" name="user">
				<option value=""<?php if(!$_v['data_editor']['runAsValue']){?>
 selected<?php }?>>Whoever starts it</option>

<?php if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
<?php }else{?>
			<span class="label">Runs as</span>
			<span class="wf-runas"><?php echo htmlspecialchars($_v['data_editor']['runAs'], ENT_QUOTES);?>
</span>
<?php if($_v['data_editor']['isNew']){?>
			<input type="hidden" name="user" value="<?php echo htmlspecialchars($_v['data_editor']['runAsValue'], ENT_QUOTES);?>
">
<?php }?><?php }?>			<small class="faint">Every step runs as this user: with their permissions, the Claude token granted to their role or user, and the keys and secrets of the workspace each step's folder is in. A step that runs another workflow follows that workflow's user when it has one. Whoever starts it: the run is the user's who presses Run, or the schedule's. The keys are the instance keys the user gets; a step's workspace adds its own keys.</small>
		</div>
		<div class="field">
			<label for="wf-over-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
">Runs over</label>
			<select id="wf-over-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
" name="runs_over">
				<option value=""<?php if($_v['data_editor']['overNone']){?>
 selected<?php }?>>No workspace: the steps run once</option>
<?php if(!empty($_v['data_editor']['overs']))foreach($_v['data_editor']['overs'] as $_v['over']){?>				<option value="<?php echo htmlspecialchars($_v['over']['id'], ENT_QUOTES);?>
"<?php if($_v['over']['selected']){?>
 selected<?php }?>>Every project of <?php echo htmlspecialchars($_v['over']['title'], ENT_QUOTES);?>
</option>
<?php }?>			</select>
			<small class="faint">The steps run once per project of the workspace, a project at a time: a step in a project of it runs in each project in turn, at the same path, and a project without the step's job skips it. A prompt or a summary whose place is the workspace root asks in the project's root, so Claude works in the project of its pass. A failed step ends that project's pass, and the next project starts. Steps outside the workspace run in every pass as they are. A workflow step may not run over projects.</small>
		</div>
	</div>

	<section class="wf-steps" aria-labelledby="wf-steps-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
">
		<div class="wf-steps-head">
			<div>
				<h3 class="section-title" id="wf-steps-<?php echo htmlspecialchars($_v['data_editor']['id'], ENT_QUOTES);?>
">Steps</h3>
				<p class="faint">Top to bottom, each once the one before it ends ok. A prompt or a summary reads the output of the steps before it.</p>
			</div>
			<button type="button" class="btn btn-sm" data-steps-fold aria-pressed="false">Expand all</button>
		</div>
		<ol class="wf-rail" data-steps>
<?php if(!empty($_v['data_editor']['steps']))foreach($_v['data_editor']['steps'] as $_v['data_step']){?><li class="wf-step kind-<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES); if($_v['data_step']['disabled']){?>
 off<?php }?><?php if($_v['data_step']['open']){?>
 open<?php }?>" data-step data-kind="<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES);?>
"<?php if($_v['data_step']['disabled']){?>
 data-off<?php }?>>
	<span class="wf-node" aria-hidden="true" data-step-num></span>
	<div class="wf-card">
		<div class="wf-card-head">
			<button type="button" class="wf-grip" data-step-grip title="Drag to move the step, or focus it and press the up and down arrows" aria-label="Move the step"><svg class="ic"><use href="#i-grip"/></svg></button>
			<button type="button" class="wf-toggle" data-step-toggle aria-expanded="<?php if($_v['data_step']['open']){?>
true<?php }else{?>
false<?php }?>">
				<span class="wf-kind"><svg class="ic" data-step-icon><use href="#i-<?php echo htmlspecialchars($_v['data_step']['icon'], ENT_QUOTES);?>
"/></svg><span data-step-kind><?php echo htmlspecialchars($_v['data_step']['kindLabel'], ENT_QUOTES);?>
</span></span>
				<span class="wf-line" data-step-line><?php echo htmlspecialchars($_v['data_step']['line'], ENT_QUOTES);?>
</span>
				<span class="badge wf-off" data-step-off-badge>off</span>
				<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
			</button>
<?php $_v['data_menu'] = $_v['data_step']['menu'];if(!$_v['data_menu']['empty']){?>
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

		</div>
		<div class="wf-card-body" data-step-body>
<?php $_v['data_space'] = $_v['data_step']['space']; if($_v['data_step']['runs']){?>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label" data-step-folder-label><?php echo htmlspecialchars($_v['data_step']['folderLabel'], ENT_QUOTES);?>
</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled<?php if($_v['data_step']['pickNone']){?>
 selected<?php }?>>pick one</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_step']['hosts']){?>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>internal hosts</option>
<?php }?><?php if(!empty($_v['data_step']['flows']))foreach($_v['data_step']['flows'] as $_v['flow']){?>						<option value="workflow:<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" data-workspace="workflow:"<?php if($_v['flow']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
<?php if(!empty($_v['data_step']['jobs']))foreach($_v['data_step']['jobs'] as $_v['one']){?>						<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['one']['folder'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['name'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['data_hostjobs'] = $_v['data_step']['hostJobs'];if($_v['data_hostjobs']){?>
			<optgroup label="Hosts">
<?php if(!empty($_v['data_hostjobs']))foreach($_v['data_hostjobs'] as $_v['job']){?>				<option value="<?php echo htmlspecialchars($_v['job']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['job']['title'], ENT_QUOTES);?>
"<?php if($_v['job']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</option>
<?php }?>			</optgroup>
<?php }?>					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments</span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['args'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
					<small class="faint">After --, for a job that takes them; the command, for a host.</small>
				</div>
			</div>
<?php }elseif($_v['data_step']['script']){?>
			<div class="field">
				<span class="label">Command</span>
				<textarea data-step-script rows="3" spellcheck="false" aria-label="Command" placeholder="go get -u ./..."><?php echo htmlspecialchars($_v['data_step']['command'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Runs with bash -l -c in the folder, as the workflow's user. A non-zero exit fails the step.</small>
			</div>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Runs in</span>
					<select data-step-place aria-label="Folder the command runs in">
						<option value="" disabled<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>pick one</option>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
			</div>
<?php }else{ if($_v['data_step']['asks']){?>
			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table."><?php echo htmlspecialchars($_v['data_step']['prompt'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</small>
			</div>
<?php }?>			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in</span></span>
					<select data-step-place aria-label="Folder Claude works in">
<?php if($_v['data_step']['none']){?>
						<option value="" data-workspace="" data-rel="."<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?>						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['extra'] = $_v['data_step']['extra']; if($_v['extra']){?>
						<option value="<?php echo htmlspecialchars($_v['extra']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['extra']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['extra']['rel'], ENT_QUOTES);?>
" selected><?php echo htmlspecialchars($_v['extra']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field">
					<span class="label">Directory <span class="faint">inside it, optional</span></span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['dir'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
				</div>
			</div>
<?php if($_v['data_step']['sums']){?>
			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question"><?php echo htmlspecialchars($_v['data_step']['question'], ENT_QUOTES);?>
</p><p><?php echo htmlspecialchars($_v['data_step']['instruction'], ENT_QUOTES);?>
</p></blockquote>
				<small class="faint">With the output of every step before it, the run's input and the last 3 answers of this workflow. A summary runs after a failed step too, to say why; the last answer is the run's.</small>
			</div>
<?php }?><?php }?>			<p class="wf-hint" data-step-hint role="status"<?php if(!$_v['data_step']['missing']){?>
 hidden<?php }?>><?php echo htmlspecialchars($_v['data_step']['missing'], ENT_QUOTES);?>
</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>

<?php }?>		</ol>
		<div class="wf-add">
			<span class="wf-node wf-node-add" aria-hidden="true"><svg class="ic"><use href="#i-plus"/></svg></span>
			<div class="wf-add-body">
				<p class="wf-empty faint" data-steps-empty<?php if($_v['data_editor']['emptyHidden']){?>
 hidden<?php }?>>No steps yet. Start with a command<?php if($_v['data_editor']['asks']){?>
, a prompt for Claude, or a summary<?php }?>.</p>
				<details class="wf-add-menu">
					<summary class="btn" aria-haspopup="menu"><svg class="ic"><use href="#i-plus"/></svg>Add step</summary>
					<div class="wf-add-list" role="menu" aria-label="Kind of step">
						<button type="button" role="menuitem" data-step-add="job"><svg class="ic"><use href="#i-terminal"/></svg><span><b>Command</b><small>A job of a folder, a command on hosts, or another workflow.</small></span></button>
<?php if($_v['data_editor']['shell']){?>
						<button type="button" role="menuitem" data-step-add="script"><svg class="ic"><use href="#i-script"/></svg><span><b>Script</b><small>A shell command you write, run in a folder, like go get -u ./...</small></span></button>
<?php }?><?php if($_v['data_editor']['asks']){?>
						<button type="button" role="menuitem" data-step-add="prompt"><svg class="ic"><use href="#i-prompt"/></svg><span><b>Prompt</b><small>Claude answers a prompt with the output of the steps before it.</small></span></button>
						<button type="button" role="menuitem" data-step-add="summary"><svg class="ic"><use href="#i-summary"/></svg><span><b>Summary</b><small>Claude summarises the steps before it, after a failure too.</small></span></button>
<?php }?>					</div>
				</details>
			</div>
		</div>
		<template data-step-template="job"><?php $_v['data_step'] = $_v['data_editor']['blankJob'];?>
<li class="wf-step kind-<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES); if($_v['data_step']['disabled']){?>
 off<?php }?><?php if($_v['data_step']['open']){?>
 open<?php }?>" data-step data-kind="<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES);?>
"<?php if($_v['data_step']['disabled']){?>
 data-off<?php }?>>
	<span class="wf-node" aria-hidden="true" data-step-num></span>
	<div class="wf-card">
		<div class="wf-card-head">
			<button type="button" class="wf-grip" data-step-grip title="Drag to move the step, or focus it and press the up and down arrows" aria-label="Move the step"><svg class="ic"><use href="#i-grip"/></svg></button>
			<button type="button" class="wf-toggle" data-step-toggle aria-expanded="<?php if($_v['data_step']['open']){?>
true<?php }else{?>
false<?php }?>">
				<span class="wf-kind"><svg class="ic" data-step-icon><use href="#i-<?php echo htmlspecialchars($_v['data_step']['icon'], ENT_QUOTES);?>
"/></svg><span data-step-kind><?php echo htmlspecialchars($_v['data_step']['kindLabel'], ENT_QUOTES);?>
</span></span>
				<span class="wf-line" data-step-line><?php echo htmlspecialchars($_v['data_step']['line'], ENT_QUOTES);?>
</span>
				<span class="badge wf-off" data-step-off-badge>off</span>
				<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
			</button>
<?php $_v['data_menu'] = $_v['data_step']['menu'];if(!$_v['data_menu']['empty']){?>
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

		</div>
		<div class="wf-card-body" data-step-body>
<?php $_v['data_space'] = $_v['data_step']['space']; if($_v['data_step']['runs']){?>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label" data-step-folder-label><?php echo htmlspecialchars($_v['data_step']['folderLabel'], ENT_QUOTES);?>
</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled<?php if($_v['data_step']['pickNone']){?>
 selected<?php }?>>pick one</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_step']['hosts']){?>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>internal hosts</option>
<?php }?><?php if(!empty($_v['data_step']['flows']))foreach($_v['data_step']['flows'] as $_v['flow']){?>						<option value="workflow:<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" data-workspace="workflow:"<?php if($_v['flow']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
<?php if(!empty($_v['data_step']['jobs']))foreach($_v['data_step']['jobs'] as $_v['one']){?>						<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['one']['folder'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['name'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['data_hostjobs'] = $_v['data_step']['hostJobs'];if($_v['data_hostjobs']){?>
			<optgroup label="Hosts">
<?php if(!empty($_v['data_hostjobs']))foreach($_v['data_hostjobs'] as $_v['job']){?>				<option value="<?php echo htmlspecialchars($_v['job']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['job']['title'], ENT_QUOTES);?>
"<?php if($_v['job']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</option>
<?php }?>			</optgroup>
<?php }?>					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments</span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['args'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
					<small class="faint">After --, for a job that takes them; the command, for a host.</small>
				</div>
			</div>
<?php }elseif($_v['data_step']['script']){?>
			<div class="field">
				<span class="label">Command</span>
				<textarea data-step-script rows="3" spellcheck="false" aria-label="Command" placeholder="go get -u ./..."><?php echo htmlspecialchars($_v['data_step']['command'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Runs with bash -l -c in the folder, as the workflow's user. A non-zero exit fails the step.</small>
			</div>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Runs in</span>
					<select data-step-place aria-label="Folder the command runs in">
						<option value="" disabled<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>pick one</option>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
			</div>
<?php }else{ if($_v['data_step']['asks']){?>
			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table."><?php echo htmlspecialchars($_v['data_step']['prompt'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</small>
			</div>
<?php }?>			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in</span></span>
					<select data-step-place aria-label="Folder Claude works in">
<?php if($_v['data_step']['none']){?>
						<option value="" data-workspace="" data-rel="."<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?>						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['extra'] = $_v['data_step']['extra']; if($_v['extra']){?>
						<option value="<?php echo htmlspecialchars($_v['extra']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['extra']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['extra']['rel'], ENT_QUOTES);?>
" selected><?php echo htmlspecialchars($_v['extra']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field">
					<span class="label">Directory <span class="faint">inside it, optional</span></span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['dir'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
				</div>
			</div>
<?php if($_v['data_step']['sums']){?>
			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question"><?php echo htmlspecialchars($_v['data_step']['question'], ENT_QUOTES);?>
</p><p><?php echo htmlspecialchars($_v['data_step']['instruction'], ENT_QUOTES);?>
</p></blockquote>
				<small class="faint">With the output of every step before it, the run's input and the last 3 answers of this workflow. A summary runs after a failed step too, to say why; the last answer is the run's.</small>
			</div>
<?php }?><?php }?>			<p class="wf-hint" data-step-hint role="status"<?php if(!$_v['data_step']['missing']){?>
 hidden<?php }?>><?php echo htmlspecialchars($_v['data_step']['missing'], ENT_QUOTES);?>
</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
</template>
<?php if($_v['data_editor']['shell']){?>
		<template data-step-template="script"><?php $_v['data_step'] = $_v['data_editor']['blankScript'];?>
<li class="wf-step kind-<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES); if($_v['data_step']['disabled']){?>
 off<?php }?><?php if($_v['data_step']['open']){?>
 open<?php }?>" data-step data-kind="<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES);?>
"<?php if($_v['data_step']['disabled']){?>
 data-off<?php }?>>
	<span class="wf-node" aria-hidden="true" data-step-num></span>
	<div class="wf-card">
		<div class="wf-card-head">
			<button type="button" class="wf-grip" data-step-grip title="Drag to move the step, or focus it and press the up and down arrows" aria-label="Move the step"><svg class="ic"><use href="#i-grip"/></svg></button>
			<button type="button" class="wf-toggle" data-step-toggle aria-expanded="<?php if($_v['data_step']['open']){?>
true<?php }else{?>
false<?php }?>">
				<span class="wf-kind"><svg class="ic" data-step-icon><use href="#i-<?php echo htmlspecialchars($_v['data_step']['icon'], ENT_QUOTES);?>
"/></svg><span data-step-kind><?php echo htmlspecialchars($_v['data_step']['kindLabel'], ENT_QUOTES);?>
</span></span>
				<span class="wf-line" data-step-line><?php echo htmlspecialchars($_v['data_step']['line'], ENT_QUOTES);?>
</span>
				<span class="badge wf-off" data-step-off-badge>off</span>
				<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
			</button>
<?php $_v['data_menu'] = $_v['data_step']['menu'];if(!$_v['data_menu']['empty']){?>
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

		</div>
		<div class="wf-card-body" data-step-body>
<?php $_v['data_space'] = $_v['data_step']['space']; if($_v['data_step']['runs']){?>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label" data-step-folder-label><?php echo htmlspecialchars($_v['data_step']['folderLabel'], ENT_QUOTES);?>
</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled<?php if($_v['data_step']['pickNone']){?>
 selected<?php }?>>pick one</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_step']['hosts']){?>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>internal hosts</option>
<?php }?><?php if(!empty($_v['data_step']['flows']))foreach($_v['data_step']['flows'] as $_v['flow']){?>						<option value="workflow:<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" data-workspace="workflow:"<?php if($_v['flow']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
<?php if(!empty($_v['data_step']['jobs']))foreach($_v['data_step']['jobs'] as $_v['one']){?>						<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['one']['folder'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['name'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['data_hostjobs'] = $_v['data_step']['hostJobs'];if($_v['data_hostjobs']){?>
			<optgroup label="Hosts">
<?php if(!empty($_v['data_hostjobs']))foreach($_v['data_hostjobs'] as $_v['job']){?>				<option value="<?php echo htmlspecialchars($_v['job']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['job']['title'], ENT_QUOTES);?>
"<?php if($_v['job']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</option>
<?php }?>			</optgroup>
<?php }?>					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments</span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['args'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
					<small class="faint">After --, for a job that takes them; the command, for a host.</small>
				</div>
			</div>
<?php }elseif($_v['data_step']['script']){?>
			<div class="field">
				<span class="label">Command</span>
				<textarea data-step-script rows="3" spellcheck="false" aria-label="Command" placeholder="go get -u ./..."><?php echo htmlspecialchars($_v['data_step']['command'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Runs with bash -l -c in the folder, as the workflow's user. A non-zero exit fails the step.</small>
			</div>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Runs in</span>
					<select data-step-place aria-label="Folder the command runs in">
						<option value="" disabled<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>pick one</option>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
			</div>
<?php }else{ if($_v['data_step']['asks']){?>
			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table."><?php echo htmlspecialchars($_v['data_step']['prompt'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</small>
			</div>
<?php }?>			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in</span></span>
					<select data-step-place aria-label="Folder Claude works in">
<?php if($_v['data_step']['none']){?>
						<option value="" data-workspace="" data-rel="."<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?>						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['extra'] = $_v['data_step']['extra']; if($_v['extra']){?>
						<option value="<?php echo htmlspecialchars($_v['extra']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['extra']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['extra']['rel'], ENT_QUOTES);?>
" selected><?php echo htmlspecialchars($_v['extra']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field">
					<span class="label">Directory <span class="faint">inside it, optional</span></span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['dir'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
				</div>
			</div>
<?php if($_v['data_step']['sums']){?>
			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question"><?php echo htmlspecialchars($_v['data_step']['question'], ENT_QUOTES);?>
</p><p><?php echo htmlspecialchars($_v['data_step']['instruction'], ENT_QUOTES);?>
</p></blockquote>
				<small class="faint">With the output of every step before it, the run's input and the last 3 answers of this workflow. A summary runs after a failed step too, to say why; the last answer is the run's.</small>
			</div>
<?php }?><?php }?>			<p class="wf-hint" data-step-hint role="status"<?php if(!$_v['data_step']['missing']){?>
 hidden<?php }?>><?php echo htmlspecialchars($_v['data_step']['missing'], ENT_QUOTES);?>
</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
</template>
<?php }?><?php if($_v['data_editor']['asks']){?>
		<template data-step-template="prompt"><?php $_v['data_step'] = $_v['data_editor']['blankPrompt'];?>
<li class="wf-step kind-<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES); if($_v['data_step']['disabled']){?>
 off<?php }?><?php if($_v['data_step']['open']){?>
 open<?php }?>" data-step data-kind="<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES);?>
"<?php if($_v['data_step']['disabled']){?>
 data-off<?php }?>>
	<span class="wf-node" aria-hidden="true" data-step-num></span>
	<div class="wf-card">
		<div class="wf-card-head">
			<button type="button" class="wf-grip" data-step-grip title="Drag to move the step, or focus it and press the up and down arrows" aria-label="Move the step"><svg class="ic"><use href="#i-grip"/></svg></button>
			<button type="button" class="wf-toggle" data-step-toggle aria-expanded="<?php if($_v['data_step']['open']){?>
true<?php }else{?>
false<?php }?>">
				<span class="wf-kind"><svg class="ic" data-step-icon><use href="#i-<?php echo htmlspecialchars($_v['data_step']['icon'], ENT_QUOTES);?>
"/></svg><span data-step-kind><?php echo htmlspecialchars($_v['data_step']['kindLabel'], ENT_QUOTES);?>
</span></span>
				<span class="wf-line" data-step-line><?php echo htmlspecialchars($_v['data_step']['line'], ENT_QUOTES);?>
</span>
				<span class="badge wf-off" data-step-off-badge>off</span>
				<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
			</button>
<?php $_v['data_menu'] = $_v['data_step']['menu'];if(!$_v['data_menu']['empty']){?>
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

		</div>
		<div class="wf-card-body" data-step-body>
<?php $_v['data_space'] = $_v['data_step']['space']; if($_v['data_step']['runs']){?>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label" data-step-folder-label><?php echo htmlspecialchars($_v['data_step']['folderLabel'], ENT_QUOTES);?>
</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled<?php if($_v['data_step']['pickNone']){?>
 selected<?php }?>>pick one</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_step']['hosts']){?>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>internal hosts</option>
<?php }?><?php if(!empty($_v['data_step']['flows']))foreach($_v['data_step']['flows'] as $_v['flow']){?>						<option value="workflow:<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" data-workspace="workflow:"<?php if($_v['flow']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
<?php if(!empty($_v['data_step']['jobs']))foreach($_v['data_step']['jobs'] as $_v['one']){?>						<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['one']['folder'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['name'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['data_hostjobs'] = $_v['data_step']['hostJobs'];if($_v['data_hostjobs']){?>
			<optgroup label="Hosts">
<?php if(!empty($_v['data_hostjobs']))foreach($_v['data_hostjobs'] as $_v['job']){?>				<option value="<?php echo htmlspecialchars($_v['job']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['job']['title'], ENT_QUOTES);?>
"<?php if($_v['job']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</option>
<?php }?>			</optgroup>
<?php }?>					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments</span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['args'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
					<small class="faint">After --, for a job that takes them; the command, for a host.</small>
				</div>
			</div>
<?php }elseif($_v['data_step']['script']){?>
			<div class="field">
				<span class="label">Command</span>
				<textarea data-step-script rows="3" spellcheck="false" aria-label="Command" placeholder="go get -u ./..."><?php echo htmlspecialchars($_v['data_step']['command'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Runs with bash -l -c in the folder, as the workflow's user. A non-zero exit fails the step.</small>
			</div>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Runs in</span>
					<select data-step-place aria-label="Folder the command runs in">
						<option value="" disabled<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>pick one</option>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
			</div>
<?php }else{ if($_v['data_step']['asks']){?>
			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table."><?php echo htmlspecialchars($_v['data_step']['prompt'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</small>
			</div>
<?php }?>			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in</span></span>
					<select data-step-place aria-label="Folder Claude works in">
<?php if($_v['data_step']['none']){?>
						<option value="" data-workspace="" data-rel="."<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?>						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['extra'] = $_v['data_step']['extra']; if($_v['extra']){?>
						<option value="<?php echo htmlspecialchars($_v['extra']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['extra']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['extra']['rel'], ENT_QUOTES);?>
" selected><?php echo htmlspecialchars($_v['extra']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field">
					<span class="label">Directory <span class="faint">inside it, optional</span></span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['dir'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
				</div>
			</div>
<?php if($_v['data_step']['sums']){?>
			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question"><?php echo htmlspecialchars($_v['data_step']['question'], ENT_QUOTES);?>
</p><p><?php echo htmlspecialchars($_v['data_step']['instruction'], ENT_QUOTES);?>
</p></blockquote>
				<small class="faint">With the output of every step before it, the run's input and the last 3 answers of this workflow. A summary runs after a failed step too, to say why; the last answer is the run's.</small>
			</div>
<?php }?><?php }?>			<p class="wf-hint" data-step-hint role="status"<?php if(!$_v['data_step']['missing']){?>
 hidden<?php }?>><?php echo htmlspecialchars($_v['data_step']['missing'], ENT_QUOTES);?>
</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
</template>
		<template data-step-template="summary"><?php $_v['data_step'] = $_v['data_editor']['blankSummary'];?>
<li class="wf-step kind-<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES); if($_v['data_step']['disabled']){?>
 off<?php }?><?php if($_v['data_step']['open']){?>
 open<?php }?>" data-step data-kind="<?php echo htmlspecialchars($_v['data_step']['kind'], ENT_QUOTES);?>
"<?php if($_v['data_step']['disabled']){?>
 data-off<?php }?>>
	<span class="wf-node" aria-hidden="true" data-step-num></span>
	<div class="wf-card">
		<div class="wf-card-head">
			<button type="button" class="wf-grip" data-step-grip title="Drag to move the step, or focus it and press the up and down arrows" aria-label="Move the step"><svg class="ic"><use href="#i-grip"/></svg></button>
			<button type="button" class="wf-toggle" data-step-toggle aria-expanded="<?php if($_v['data_step']['open']){?>
true<?php }else{?>
false<?php }?>">
				<span class="wf-kind"><svg class="ic" data-step-icon><use href="#i-<?php echo htmlspecialchars($_v['data_step']['icon'], ENT_QUOTES);?>
"/></svg><span data-step-kind><?php echo htmlspecialchars($_v['data_step']['kindLabel'], ENT_QUOTES);?>
</span></span>
				<span class="wf-line" data-step-line><?php echo htmlspecialchars($_v['data_step']['line'], ENT_QUOTES);?>
</span>
				<span class="badge wf-off" data-step-off-badge>off</span>
				<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
			</button>
<?php $_v['data_menu'] = $_v['data_step']['menu'];if(!$_v['data_menu']['empty']){?>
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

		</div>
		<div class="wf-card-body" data-step-body>
<?php $_v['data_space'] = $_v['data_step']['space']; if($_v['data_step']['runs']){?>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label" data-step-folder-label><?php echo htmlspecialchars($_v['data_step']['folderLabel'], ENT_QUOTES);?>
</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled<?php if($_v['data_step']['pickNone']){?>
 selected<?php }?>>pick one</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_step']['hosts']){?>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>internal hosts</option>
<?php }?><?php if(!empty($_v['data_step']['flows']))foreach($_v['data_step']['flows'] as $_v['flow']){?>						<option value="workflow:<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" data-workspace="workflow:"<?php if($_v['flow']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
<?php if(!empty($_v['data_step']['jobs']))foreach($_v['data_step']['jobs'] as $_v['one']){?>						<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['one']['folder'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['name'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['data_hostjobs'] = $_v['data_step']['hostJobs'];if($_v['data_hostjobs']){?>
			<optgroup label="Hosts">
<?php if(!empty($_v['data_hostjobs']))foreach($_v['data_hostjobs'] as $_v['job']){?>				<option value="<?php echo htmlspecialchars($_v['job']['id'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['job']['title'], ENT_QUOTES);?>
"<?php if($_v['job']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</option>
<?php }?>			</optgroup>
<?php }?>					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments</span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['args'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
					<small class="faint">After --, for a job that takes them; the command, for a host.</small>
				</div>
			</div>
<?php }elseif($_v['data_step']['script']){?>
			<div class="field">
				<span class="label">Command</span>
				<textarea data-step-script rows="3" spellcheck="false" aria-label="Command" placeholder="go get -u ./..."><?php echo htmlspecialchars($_v['data_step']['command'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Runs with bash -l -c in the folder, as the workflow's user. A non-zero exit fails the step.</small>
			</div>
			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Runs in</span>
					<select data-step-place aria-label="Folder the command runs in">
						<option value="" disabled<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>pick one</option>
						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
			</div>
<?php }else{ if($_v['data_step']['asks']){?>
			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table."><?php echo htmlspecialchars($_v['data_step']['prompt'], ENT_QUOTES);?>
</textarea>
				<small class="faint">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</small>
			</div>
<?php }?>			<div class="wf-fields wf-cascade"><div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>

				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in</span></span>
					<select data-step-place aria-label="Folder Claude works in">
<?php if($_v['data_step']['none']){?>
						<option value="" data-workspace="" data-rel="."<?php if($_v['data_step']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?>						<option value="internal" data-workspace="internal" data-rel="."<?php if($_v['data_step']['rootSelected']){?>
 selected<?php }?>>the root</option>
<?php if(!empty($_v['data_step']['places']))foreach($_v['data_step']['places'] as $_v['place']){?>						<option value="<?php echo htmlspecialchars($_v['place']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['place']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['place']['rel'], ENT_QUOTES);?>
"<?php if($_v['place']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['place']['label'], ENT_QUOTES);?>
</option>
<?php }?><?php $_v['extra'] = $_v['data_step']['extra']; if($_v['extra']){?>
						<option value="<?php echo htmlspecialchars($_v['extra']['id'], ENT_QUOTES);?>
" data-workspace="<?php echo htmlspecialchars($_v['extra']['workspace'], ENT_QUOTES);?>
" data-rel="<?php echo htmlspecialchars($_v['extra']['rel'], ENT_QUOTES);?>
" selected><?php echo htmlspecialchars($_v['extra']['label'], ENT_QUOTES);?>
</option>
<?php }?>					</select>
				</div>
				<div class="field">
					<span class="label">Directory <span class="faint">inside it, optional</span></span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="<?php echo htmlspecialchars($_v['data_step']['dir'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
				</div>
			</div>
<?php if($_v['data_step']['sums']){?>
			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question"><?php echo htmlspecialchars($_v['data_step']['question'], ENT_QUOTES);?>
</p><p><?php echo htmlspecialchars($_v['data_step']['instruction'], ENT_QUOTES);?>
</p></blockquote>
				<small class="faint">With the output of every step before it, the run's input and the last 3 answers of this workflow. A summary runs after a failed step too, to say why; the last answer is the run's.</small>
			</div>
<?php }?><?php }?>			<p class="wf-hint" data-step-hint role="status"<?php if(!$_v['data_step']['missing']){?>
 hidden<?php }?>><?php echo htmlspecialchars($_v['data_step']['missing'], ENT_QUOTES);?>
</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
</template>
<?php }?>		<p class="wf-live" aria-live="polite" data-steps-live></p>
	</section>

	<div class="form-foot">
		<a class="btn" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_editor']['next'], ENT_QUOTES);?>
">Cancel</a>
		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_editor']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>


<section class="block" id="schedule">
<?php $_v['data_block'] = $_v['data']['block'];?>

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

<?php if($_v['data']['schedules']){?>
	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
<?php if(!empty($_v['data']['schedules']))foreach($_v['data']['schedules'] as $_v['row']){?>			<tr>
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
				<td class="actions"><?php if($_v['row']['can']){ $_v['data_menu'] = $_v['row']['menu'];if(!$_v['data_menu']['empty']){?>
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
<?php }?>	<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload" data-notice="Schedule added" data-sched>
		<input type="hidden" name="workflow" value="<?php echo htmlspecialchars($_v['data']['id'], ENT_QUOTES);?>
">
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
</section>
