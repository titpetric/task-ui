{* input scope: data_step. See wfstep.md. *}<li class="wf-step kind-{data_step.kind}{if $data_step.disabled} off{/if}{if $data_step.open} open{/if}" data-step data-kind="{data_step.kind}"{if $data_step.disabled} data-off{/if}>
	<span class="wf-node" aria-hidden="true" data-step-num></span>
	<div class="wf-card">
		<div class="wf-card-head">
			<button type="button" class="wf-grip" data-step-grip title="Drag to move the step, or focus it and press the up and down arrows" aria-label="Move the step"><svg class="ic"><use href="#i-grip"/></svg></button>
			<button type="button" class="wf-toggle" data-step-toggle aria-expanded="{if $data_step.open}true{else}false{/if}">
				<span class="wf-kind"><svg class="ic" data-step-icon><use href="#i-{data_step.icon}"/></svg><span data-step-kind>{data_step.kindLabel}</span></span>
				<span class="wf-line" data-step-line>{data_step.line}</span>
				<span class="badge wf-off" data-step-off-badge>off</span>
				<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
			</button>
{eval $data_menu = $data_step.menu}{include components/menu.tpl}
		</div>
		<div class="wf-card-body" data-step-body>
{eval $data_space = $data_step.space}{if $data_step.runs}			<div class="wf-fields wf-cascade">{include components/stepspace.tpl}
				<div class="field">
					<span class="label" data-step-folder-label>{data_step.folderLabel}</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled{if $data_step.pickNone} selected{/if}>pick one</option>
{foreach $data_step.places as $place}						<option value="{place.id}" data-workspace="{place.workspace}" data-rel="{place.rel}"{if $place.selected} selected{/if}>{place.label}</option>
{/foreach}{if $data_step.hosts}						<option value="internal" data-workspace="internal" data-rel="."{if $data_step.rootSelected} selected{/if}>internal hosts</option>
{/if}{foreach $data_step.flows as $flow}						<option value="workflow:{flow.id}" data-workspace="workflow:"{if $flow.selected} selected{/if}>{flow.name}</option>
{/foreach}					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
{foreach $data_step.jobs as $one}						<option value="{one.id}" data-folder="{one.folder}"{if $one.selected} selected{/if}>{one.name}</option>
{/foreach}{eval $data_hostjobs = $data_step.hostJobs}{include components/host-jobs.tpl}					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments</span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="{data_step.args}" spellcheck="false" autocomplete="off">
					<small class="faint">After --, for a job that takes them; the command, for a host.</small>
				</div>
			</div>
{elseif $data_step.script}			<div class="field">
				<span class="label">Command</span>
				<textarea data-step-script rows="3" spellcheck="false" aria-label="Command" placeholder="go get -u ./...">{data_step.command}</textarea>
				<small class="faint">Runs with bash -l -c in the folder, as the workflow's user. A non-zero exit fails the step.</small>
			</div>
			<div class="wf-fields wf-cascade">{include components/stepspace.tpl}
				<div class="field">
					<span class="label">Runs in</span>
					<select data-step-place aria-label="Folder the command runs in">
						<option value="" disabled{if $data_step.noneSelected} selected{/if}>pick one</option>
						<option value="internal" data-workspace="internal" data-rel="."{if $data_step.rootSelected} selected{/if}>the root</option>
{foreach $data_step.places as $place}						<option value="{place.id}" data-workspace="{place.workspace}" data-rel="{place.rel}"{if $place.selected} selected{/if}>{place.label}</option>
{/foreach}					</select>
				</div>
			</div>
{else}{if $data_step.asks}			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table.">{data_step.prompt}</textarea>
				<small class="faint">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</small>
			</div>
{/if}			<div class="wf-fields wf-cascade">{include components/stepspace.tpl}
				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in</span></span>
					<select data-step-place aria-label="Folder Claude works in">
{if $data_step.none}						<option value="" data-workspace="" data-rel="."{if $data_step.noneSelected} selected{/if}>none (the workspace root)</option>
{/if}						<option value="internal" data-workspace="internal" data-rel="."{if $data_step.rootSelected} selected{/if}>the root</option>
{foreach $data_step.places as $place}						<option value="{place.id}" data-workspace="{place.workspace}" data-rel="{place.rel}"{if $place.selected} selected{/if}>{place.label}</option>
{/foreach}{eval $extra = $data_step.extra}{if $extra}						<option value="{extra.id}" data-workspace="{extra.workspace}" data-rel="{extra.rel}" selected>{extra.label}</option>
{/if}					</select>
				</div>
				<div class="field">
					<span class="label">Directory <span class="faint">inside it, optional</span></span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="{data_step.dir}" spellcheck="false" autocomplete="off">
				</div>
			</div>
{if $data_step.sums}			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question">{data_step.question}</p><p>{data_step.instruction}</p></blockquote>
				<small class="faint">With the output of every step before it, the run's input and the last 3 answers of this workflow. A summary runs after a failed step too, to say why; the last answer is the run's.</small>
			</div>
{/if}{/if}			<p class="wf-hint" data-step-hint role="status"{if !$data_step.missing} hidden{/if}>{data_step.missing}</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
