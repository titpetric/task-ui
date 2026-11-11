<?php $_v=&$this->vars;?>
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
<?php $_v['data_menu'] = $_v['data_step']['menu']; if(!$_v['data_menu']['empty']){?>
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
<?php }?><?php $_v['data_hostjobs'] = $_v['data_step']['hostJobs']; if($_v['data_hostjobs']){?>
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
<?php }elseif($_v['data_step']['check']){?>
			<div class="wf-fields">
				<div class="field wf-wide">
					<span class="label">Looks for <span class="faint">in the step before it</span></span>
					<input type="text" data-step-markers aria-label="What the check looks for" placeholder="Next:" value="<?php echo htmlspecialchars($_v['data_step']['markers'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
					<small class="faint">One or more, comma separated, case ignored. While the step before the check writes one of them, the workflow's steps run again from the first.</small>
				</div>
				<div class="field">
					<span class="label">Passes <span class="faint">at the most</span></span>
					<input type="number" data-step-tries aria-label="Passes at the most" min="<?php echo htmlspecialchars($_v['data_step']['least'], ENT_QUOTES);?>
" max="<?php echo htmlspecialchars($_v['data_step']['most'], ENT_QUOTES);?>
" step="1" value="<?php echo htmlspecialchars($_v['data_step']['tries'], ENT_QUOTES);?>
">
					<small class="faint">The first pass counted, so 3 lets the steps run twice more. The check runs nothing itself and ends ok whatever it finds.</small>
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
					<span class="label">Directory</span>
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
			</div>
<?php }?><?php }?>			<p class="wf-hint" data-step-hint role="status"<?php if(!$_v['data_step']['missing']){?>
 hidden<?php }?>><?php echo htmlspecialchars($_v['data_step']['missing'], ENT_QUOTES);?>
</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
