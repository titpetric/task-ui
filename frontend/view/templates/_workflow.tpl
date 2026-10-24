{{/* stepeditor is the workflow form, new and edit alike: the name, a
     description and the sidebar toggle, then the steps in order on a
     rail, each a card of its kind that folds to one line, and Add step
     with the kinds to pick from. A prompt or a summary takes prompt/edit.
     list: page, workflow (nil for a new one), API path, method, where to
     go after, the submit label. task-ui.js reads the cards into the
     request's steps, moves them by their grip or their menu, and puts an
     error the API names a step by under that step. */}}
{{ define "stepeditor" }}
{{- $p := index . 0 }}{{ $w := index . 1 }}{{ $api := index . 2 }}{{ $method := index . 3 }}{{ $next := index . 4 }}{{ $label := index . 5 -}}
{{- $id := "new" }}{{ if $w }}{{ $id = $w.ID }}{{ end }}{{ $asks := $p.Can "prompt/edit" -}}
<form class="form workflow-form" data-workflow data-api="{{ $api }}{{ if $w }}{{ $w.ID }}{{ end }}" data-method="{{ $method }}" data-next="{{ $next }}" novalidate>
	<div class="wf-meta">
		<div class="field">
			<label for="wf-name-{{ $id }}">Name</label>
			<input type="text" id="wf-name-{{ $id }}" name="name" value="{{ if $w }}{{ $w.Name }}{{ end }}" placeholder="nightly deploy" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="wf-desc-{{ $id }}">Description <span class="faint">optional</span></label>
			<textarea id="wf-desc-{{ $id }}" name="description" rows="2" placeholder="What the workflow is for, in a line or two">{{ if $w }}{{ $w.Description }}{{ end }}</textarea>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="sidebar"{{ if or (not $w) (not $w.SidebarHidden) }} checked{{ end }}> Show in sidebar <span class="faint">off for a workflow you only use as a step of others</span></label>
		</div>
	</div>

	<section class="wf-steps" aria-labelledby="wf-steps-{{ $id }}">
		<div class="wf-steps-head">
			<div>
				<h3 class="section-title" id="wf-steps-{{ $id }}">Steps</h3>
				<p class="faint">Top to bottom, each once the one before it ends ok. A prompt or a summary reads the output of the steps before it.</p>
			</div>
			<button type="button" class="btn btn-sm" data-steps-fold aria-pressed="false">Expand all</button>
		</div>
		<ol class="wf-rail" data-steps>
			{{- if $w }}{{ range $w.Steps }}{{ template "wfstep" (list $p . false) }}{{ end }}{{ end }}
		</ol>
		<div class="wf-add">
			<span class="wf-node wf-node-add" aria-hidden="true"><svg class="ic"><use href="#i-plus"/></svg></span>
			<div class="wf-add-body">
				<p class="wf-empty faint" data-steps-empty{{ if and $w $w.Steps }} hidden{{ end }}>No steps yet. Start with a command{{ if $asks }}, a prompt for Claude, or a summary{{ end }}.</p>
				<details class="wf-add-menu">
					<summary class="btn" aria-haspopup="menu"><svg class="ic"><use href="#i-plus"/></svg>Add step</summary>
					<div class="wf-add-list" role="menu" aria-label="Kind of step">
						<button type="button" role="menuitem" data-step-add="job"><svg class="ic"><use href="#i-terminal"/></svg><span><b>Command</b><small>A job of a folder, a command on hosts, or another workflow.</small></span></button>
						{{- if $asks }}
						<button type="button" role="menuitem" data-step-add="prompt"><svg class="ic"><use href="#i-prompt"/></svg><span><b>Prompt</b><small>Claude answers a prompt with the output of the steps before it.</small></span></button>
						<button type="button" role="menuitem" data-step-add="summary"><svg class="ic"><use href="#i-summary"/></svg><span><b>Summary</b><small>Claude summarises the steps before it, after a failure too.</small></span></button>
						{{- end }}
					</div>
				</details>
			</div>
		</div>
		<template data-step-template="job">{{ template "wfstep" (list $p (newStep "job") true) }}</template>
		{{- if $asks }}
		<template data-step-template="prompt">{{ template "wfstep" (list $p (newStep "prompt") true) }}</template>
		<template data-step-template="summary">{{ template "wfstep" (list $p (newStep "summary") true) }}</template>
		{{- end }}
		<p class="wf-live" aria-live="polite" data-steps-live></p>
	</section>

	<div class="form-foot">
		<a class="btn" href="{{ $next }}">Cancel</a>
		<button class="btn btn-primary" type="submit">{{ $label }}</button>
	</div>
</form>
{{ end }}

{{/* wfstep is one card of the workflow form, (list page step open): its
     node on the rail, its head, the grip, its kind and the step in one
     line, with the [...] menu, and its fields under it, folded unless
     open. A command picks a workspace, then a folder of it and its job,
     or Workflows, then a workflow; a prompt its text and place, a
     summary its place, each a workspace and then a folder of it. */}}
{{ define "wfstep" }}
{{- $p := index . 0 }}{{ $s := index . 1 }}{{ $open := index . 2 }}{{ $kind := $s.StepKind -}}
{{- $off := "Disable" }}{{ $offIcon := "stop" }}{{ if $s.Disabled }}{{ $off = "Enable" }}{{ $offIcon = "play" }}{{ end -}}
<li class="wf-step kind-{{ $kind }}{{ if $s.Disabled }} off{{ end }}{{ if $open }} open{{ end }}" data-step data-kind="{{ $kind }}"{{ if $s.Disabled }} data-off{{ end }}>
	<span class="wf-node" aria-hidden="true" data-step-num></span>
	<div class="wf-card">
		<div class="wf-card-head">
			<button type="button" class="wf-grip" data-step-grip title="Drag to move the step, or focus it and press the up and down arrows" aria-label="Move the step"><svg class="ic"><use href="#i-grip"/></svg></button>
			<button type="button" class="wf-toggle" data-step-toggle aria-expanded="{{ if $open }}true{{ else }}false{{ end }}">
				<span class="wf-kind"><svg class="ic" data-step-icon><use href="#i-{{ $s.Icon }}"/></svg><span data-step-kind>{{ $s.KindLabel }}</span></span>
				<span class="wf-line" data-step-line>{{ $s.Line }}</span>
				<span class="badge wf-off" data-step-off-badge>off</span>
				<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
			</button>
			{{- template "menu" (menu "Actions on the step"
				(item "up" "Move up" "data-step-move" "up")
				(item "down" "Move down" "data-step-move" "down")
				(sep)
				(item $offIcon $off "data-step-off" "1")
				(item "copy" "Duplicate" "data-step-duplicate" "1")
				(danger "trash" "Remove" "data-step-remove" "1")
			) }}
		</div>
		<div class="wf-card-body" data-step-body>
			{{- if or (eq $kind "job") (eq $kind "workflow") }}
			{{- $space := $p.StepSpaceOf $s.Folder }}{{ if eq $kind "workflow" }}{{ $space = "workflow:" }}{{ end }}
			<div class="wf-fields wf-cascade">
				{{- template "stepspace" (list $p $space true) }}
				<div class="field">
					<span class="label" data-step-folder-label>{{ if eq $kind "workflow" }}Workflow{{ else }}Folder{{ end }}</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled{{ if and (not $s.Folder) (not $s.WorkflowID) }} selected{{ end }}>pick one</option>
						{{- range $p.Folders }}{{ if .Jobs }}
						<option value="{{ .ID }}" data-workspace="{{ $p.StepSpaceOf .ID }}"{{ if eq .ID $s.Folder }} selected{{ end }}>{{ .ID }}</option>
						{{- end }}{{ end }}
						{{- if $p.HostJobs }}
						<option value="_" data-workspace="_"{{ if eq "_" $s.Folder }} selected{{ end }}>_ hosts</option>
						{{- end }}
						{{- range $p.StepWorkflows }}
						<option value="workflow:{{ .ID }}" data-workspace="workflow:"{{ if eq .ID $s.WorkflowID }} selected{{ end }}>{{ .Name }}</option>
						{{- end }}
					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
						{{- range $f := $p.Folders }}{{ range $f.Jobs }}
						<option value="{{ .ID }}" data-folder="{{ $f.ID }}"{{ if and (eq $f.ID $s.Folder) (eq .ID $s.Job) }} selected{{ end }}>{{ .Name }}</option>
						{{- end }}{{ end }}
						{{- template "hostjobs" (list $p $s.Job) }}
					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments <span class="faint">after --, for a job that takes them; the command, for a host</span></span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="{{ argLine $s.Job $s.Args }}" spellcheck="false" autocomplete="off">
				</div>
			</div>
			{{- else }}
			{{- if eq $kind "prompt" }}
			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table.">{{ $s.Prompt }}</textarea>
				<p class="faint field-note">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</p>
			</div>
			{{- end }}
			<div class="wf-fields wf-cascade">
				{{- template "stepspace" (list $p ($p.StepSpaceOf $s.Folder) false) }}
				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in, as you</span></span>
					{{- $found := eq $s.Folder "_" }}
					<select data-step-place aria-label="Folder Claude works in">
						<option value="_" data-workspace="_"{{ if eq $s.Folder "_" }} selected{{ end }}>the root</option>
						{{- range $p.Folders }}{{ if eq .ID $s.Folder }}{{ $found = true }}{{ end }}
						<option value="{{ .ID }}" data-workspace="{{ $p.StepSpaceOf .ID }}"{{ if eq .ID $s.Folder }} selected{{ end }}>{{ .ID }}</option>
						{{- end }}
						{{- if and $s.Folder (not $found) }}
						<option value="{{ $s.Folder }}" data-workspace="{{ $p.StepSpaceOf $s.Folder }}" selected>{{ $s.Folder }}</option>
						{{- end }}
					</select>
				</div>
				<div class="field">
					<span class="label">Directory <span class="faint">inside it, optional</span></span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="{{ $s.Dir }}" spellcheck="false" autocomplete="off">
				</div>
			</div>
			{{- if eq $kind "summary" }}
			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question">{{ workflowSummaryQuestion }}</p><p>{{ summaryPrompt }}</p></blockquote>
				<p class="faint field-note">With the output of every step before it, the run's input and the last 3 answers of this workflow. A summary runs after a failed step too, to say why; the last answer is the run's.</p>
			</div>
			{{- end }}
			{{- end }}
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
{{- end }}

{{/* stepspace is the Workspace field of a step card, (list page
     selected workflows): the default workspace, the named ones the user
     sees a folder of, and Workflows for a command. The folder select
     next to it lists the options whose data-workspace matches;
     task-ui.js narrows it when the workspace changes. */}}
{{ define "stepspace" }}
{{- $p := index . 0 }}{{ $sel := index . 1 }}{{ $flows := index . 2 -}}
<div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
		{{- range $p.StepSpaces }}
		<option value="{{ .ID }}"{{ if eq .ID $sel }} selected{{ end }}>{{ .Title }}</option>
		{{- end }}
		{{- if and $flows $p.StepWorkflows }}
		<option value="workflow:"{{ if eq "workflow:" $sel }} selected{{ end }}>Workflows</option>
		{{- end }}
	</select>
</div>
{{- end }}
