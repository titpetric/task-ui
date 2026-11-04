{* input scope: data_editor. See stepeditor.md. *}<form class="form workflow-form" data-workflow data-api="{data_editor.api}" data-method="{data_editor.method}" data-next="{base}{data_editor.next}" novalidate>
	<div class="wf-meta">
		<div class="field">
			<label for="wf-name-{data_editor.id}">Name</label>
			<input type="text" id="wf-name-{data_editor.id}" name="name" value="{data_editor.name}" placeholder="nightly deploy" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="wf-desc-{data_editor.id}">Description <span class="faint">optional</span></label>
			<textarea id="wf-desc-{data_editor.id}" name="description" rows="2" placeholder="What the workflow is for, in a line or two">{data_editor.description}</textarea>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="sidebar"{if $data_editor.sidebar} checked{/if}> Show in sidebar</label>
			<small class="faint">Off for a workflow you only use as a step of others.</small>
		</div>
		<div class="field">
{eval $data_options = $data_editor.users}{if $data_options.users}			<label for="wf-user-{data_editor.id}">Runs as</label>
			<select id="wf-user-{data_editor.id}" name="user">
				<option value=""{if !$data_editor.runAsValue} selected{/if}>Whoever starts it</option>
{include components/user-options.tpl}			</select>
{else}			<span class="label">Runs as</span>
			<span class="wf-runas">{data_editor.runAs}</span>
{if $data_editor.isNew}			<input type="hidden" name="user" value="{data_editor.runAsValue}">
{/if}{/if}			<small class="faint">Every step runs as this user: with their permissions, the Claude token granted to their role or user, and the keys and secrets of the workspace each step's folder is in. A step that runs another workflow follows that workflow's user when it has one. Whoever starts it: the run is the user's who presses Run, or the schedule's. The keys are the instance keys the user gets; a step's workspace adds its own keys.</small>
		</div>
		<div class="field">
			<label for="wf-over-{data_editor.id}">Runs over</label>
			<select id="wf-over-{data_editor.id}" name="runs_over">
				<option value=""{if $data_editor.overNone} selected{/if}>No workspace: the steps run once</option>
{foreach $data_editor.overs as $over}				<option value="{over.id}"{if $over.selected} selected{/if}>Every project of {over.title}</option>
{/foreach}			</select>
			<small class="faint">The steps run once per project of the workspace, a project at a time: a step in a project of it runs in each project in turn, at the same path, and a project without the step's job skips it. A prompt or a summary whose place is the workspace root asks in the project's root, so Claude works in the project of its pass. A failed step ends that project's pass, and the next project starts. Steps outside the workspace run in every pass as they are. A workflow step may not run over projects.</small>
		</div>
	</div>

	<section class="wf-steps" aria-labelledby="wf-steps-{data_editor.id}">
		<div class="wf-steps-head">
			<div>
				<h3 class="section-title" id="wf-steps-{data_editor.id}">Steps</h3>
				<p class="faint">Top to bottom, each once the one before it ends ok. A prompt or a summary reads the output of the steps before it.</p>
			</div>
			<button type="button" class="btn btn-sm" data-steps-fold aria-pressed="false">Expand all</button>
		</div>
		<ol class="wf-rail" data-steps>
{foreach $data_editor.steps as $data_step}{include components/wfstep.tpl}
{/foreach}		</ol>
		<div class="wf-add">
			<span class="wf-node wf-node-add" aria-hidden="true"><svg class="ic"><use href="#i-plus"/></svg></span>
			<div class="wf-add-body">
				<p class="wf-empty faint" data-steps-empty{if $data_editor.emptyHidden} hidden{/if}>No steps yet. Start with a command{if $data_editor.asks}, a prompt for Claude, or a summary{/if}.</p>
				<details class="wf-add-menu">
					<summary class="btn" aria-haspopup="menu"><svg class="ic"><use href="#i-plus"/></svg>Add step</summary>
					<div class="wf-add-list" role="menu" aria-label="Kind of step">
						<button type="button" role="menuitem" data-step-add="job"><svg class="ic"><use href="#i-terminal"/></svg><span><b>Command</b><small>A job of a folder, a command on hosts, or another workflow.</small></span></button>
{if $data_editor.shell}						<button type="button" role="menuitem" data-step-add="script"><svg class="ic"><use href="#i-script"/></svg><span><b>Script</b><small>A shell command you write, run in a folder, like go get -u ./...</small></span></button>
{/if}{if $data_editor.asks}						<button type="button" role="menuitem" data-step-add="prompt"><svg class="ic"><use href="#i-prompt"/></svg><span><b>Prompt</b><small>Claude answers a prompt with the output of the steps before it.</small></span></button>
						<button type="button" role="menuitem" data-step-add="summary"><svg class="ic"><use href="#i-summary"/></svg><span><b>Summary</b><small>Claude summarises the steps before it, after a failure too.</small></span></button>
{/if}						<button type="button" role="menuitem" data-step-add="check"><svg class="ic"><use href="#i-rotate"/></svg><span><b>Check</b><small>While the step before it writes Next:, the steps run again.</small></span></button>
					</div>
				</details>
			</div>
		</div>
		<template data-step-template="job">{eval $data_step = $data_editor.blankJob}{include components/wfstep.tpl}</template>
{if $data_editor.shell}		<template data-step-template="script">{eval $data_step = $data_editor.blankScript}{include components/wfstep.tpl}</template>
{/if}{if $data_editor.asks}		<template data-step-template="prompt">{eval $data_step = $data_editor.blankPrompt}{include components/wfstep.tpl}</template>
		<template data-step-template="summary">{eval $data_step = $data_editor.blankSummary}{include components/wfstep.tpl}</template>
{/if}		<template data-step-template="check">{eval $data_step = $data_editor.blankCheck}{include components/wfstep.tpl}</template>
		<p class="wf-live" aria-live="polite" data-steps-live></p>
	</section>

	<div class="form-foot">
		<a class="btn" href="{base}{data_editor.next}">Cancel</a>
		<button class="btn btn-primary" type="submit">{data_editor.label}</button>
	</div>
</form>
