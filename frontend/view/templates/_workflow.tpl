{{ define "stepeditor" }}
{{- $p := index . 0 }}{{ $w := index . 1 }}{{ $api := index . 2 }}{{ $method := index . 3 }}{{ $next := index . 4 }}{{ $label := index . 5 -}}
{{- $id := "new" }}{{ if $w }}{{ $id = $w.ID }}{{ end }}{{ $asks := $p.Can "prompt/edit" }}{{ $shell := $p.Can "shell" -}}
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
			<label class="check"><input type="checkbox" name="sidebar"{{ if or (not $w) (not $w.SidebarHidden) }} checked{{ end }}> Show in sidebar</label>
			<small class="faint">Off for a workflow you only use as a step of others.</small>
		</div>
		{{- $runAs := $p.User.Name }}{{ if $w }}{{ $runAs = $w.RunAs }}{{ end }}
		<div class="field">
			{{- if $p.Users }}
			<label for="wf-user-{{ $id }}">Runs as</label>
			<select id="wf-user-{{ $id }}" name="user">
				<option value=""{{ if not $runAs }} selected{{ end }}>Whoever starts it</option>
				{{- template "user-options" (list $p.Users $runAs $p.UserKeys) }}
			</select>
			{{- else }}
			<span class="label">Runs as</span>
			<span class="wf-runas">{{ or $runAs "Whoever starts it" }}</span>
			{{- if not $w }}<input type="hidden" name="user" value="{{ $runAs }}">{{ end }}
			{{- end }}
		</div>
		{{- $over := "" }}{{ if $w }}{{ $over = $w.RunsOver }}{{ end }}
		<div class="field">
			<label for="wf-over-{{ $id }}">Runs over</label>
			<select id="wf-over-{{ $id }}" name="runs_over">
				<option value=""{{ if not $over }} selected{{ end }}>No workspace: the steps run once</option>
				{{- range $p.StepSpaces }}
				<option value="{{ .ID }}"{{ if eq .ID $over }} selected{{ end }}>Every project of {{ .Title }}</option>
				{{- end }}
			</select>
		</div>
	</div>

	<section class="wf-steps" aria-labelledby="wf-steps-{{ $id }}">
		<div class="wf-steps-head">
			<div>
				<h3 class="section-title" id="wf-steps-{{ $id }}">Steps</h3>
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
						{{- if $shell }}
						<button type="button" role="menuitem" data-step-add="script"><svg class="ic"><use href="#i-script"/></svg><span><b>Script</b><small>A shell command you write, run in a folder, like go get -u ./...</small></span></button>
						{{- end }}
						{{- if $asks }}
						<button type="button" role="menuitem" data-step-add="prompt"><svg class="ic"><use href="#i-prompt"/></svg><span><b>Prompt</b><small>Claude answers a prompt with the output of the steps before it.</small></span></button>
						<button type="button" role="menuitem" data-step-add="summary"><svg class="ic"><use href="#i-summary"/></svg><span><b>Summary</b><small>Claude summarises the steps before it, after a failure too.</small></span></button>
						{{- end }}
						<button type="button" role="menuitem" data-step-add="check"><svg class="ic"><use href="#i-rotate"/></svg><span><b>Check</b><small>While the step before it writes Next:, the steps run again.</small></span></button>
					</div>
				</details>
			</div>
		</div>
		<template data-step-template="job">{{ template "wfstep" (list $p (newStep "job") true) }}</template>
		{{- if $shell }}
		<template data-step-template="script">{{ template "wfstep" (list $p (newStep "script") true) }}</template>
		{{- end }}
		{{- if $asks }}
		<template data-step-template="prompt">{{ template "wfstep" (list $p (newStep "prompt") true) }}</template>
		<template data-step-template="summary">{{ template "wfstep" (list $p (newStep "summary") true) }}</template>
		{{- end }}
		<template data-step-template="check">{{ template "wfstep" (list $p (newStep "check") true) }}</template>
		<p class="wf-live" aria-live="polite" data-steps-live></p>
	</section>

	<div class="form-foot">
		<a class="btn" href="{{ $next }}">Cancel</a>
		<button class="btn btn-primary" type="submit">{{ $label }}</button>
	</div>
</form>
{{ end }}

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
				{{- template "stepspace" (list $p $space true false) }}
				<div class="field">
					<span class="label" data-step-folder-label>{{ if eq $kind "workflow" }}Workflow{{ else }}Folder{{ end }}</span>
					<select data-step-folder aria-label="Folder or workflow">
						<option value="" disabled{{ if and (not $s.Folder) (not $s.WorkflowID) }} selected{{ end }}>pick one</option>
						{{- range $p.StepFolders }}{{ if .Jobs }}
						<option value="{{ .ID }}" data-workspace="{{ $p.StepSpaceOf .ID }}" data-rel="{{ $p.StepRel .ID }}"{{ if eq .ID $s.Folder }} selected{{ end }}>{{ $p.StepLabel .ID }}</option>
						{{- end }}{{ end }}
						{{- if $p.HostJobs }}
						<option value="internal" data-workspace="internal" data-rel="."{{ if eq "internal" $s.Folder }} selected{{ end }}>internal hosts</option>
						{{- end }}
						{{- range $p.StepWorkflows }}
						<option value="workflow:{{ .ID }}" data-workspace="workflow:"{{ if eq .ID $s.WorkflowID }} selected{{ end }}>{{ .Name }}</option>
						{{- end }}
					</select>
				</div>
				<div class="field" data-step-job-field>
					<span class="label">Job</span>
					<select data-step-job aria-label="Job">
						{{- range $f := $p.StepFolders }}{{ range $f.Jobs }}
						<option value="{{ .ID }}" data-folder="{{ $f.ID }}"{{ if and (eq $f.ID $s.Folder) (eq .ID $s.Job) }} selected{{ end }}>{{ .Name }}</option>
						{{- end }}{{ end }}
						{{- template "hostjobs" (list $p $s.Job) }}
					</select>
				</div>
				<div class="field wf-wide" data-step-args-field>
					<span class="label">Arguments</span>
					<input type="text" data-step-args aria-label="Arguments" placeholder="none" value="{{ argLine $s.Job $s.Args }}" spellcheck="false" autocomplete="off">
					<small class="faint">After --, for a job that takes them; the command, for a host.</small>
				</div>
			</div>
			{{- else if eq $kind "script" }}
			<div class="field">
				<span class="label">Command</span>
				<textarea data-step-script rows="3" spellcheck="false" aria-label="Command" placeholder="go get -u ./...">{{ $s.Command }}</textarea>
				<small class="faint">Runs with bash -l -c in the folder, as the workflow's user. A non-zero exit fails the step.</small>
			</div>
			<div class="wf-fields wf-cascade">
				{{- template "stepspace" (list $p ($p.StepSpaceOf $s.Folder) false false) }}
				<div class="field">
					<span class="label">Runs in</span>
					<select data-step-place aria-label="Folder the command runs in">
						<option value="" disabled{{ if not $s.Folder }} selected{{ end }}>pick one</option>
						<option value="internal" data-workspace="internal" data-rel="."{{ if eq $s.Folder "internal" }} selected{{ end }}>the root</option>
						{{- range $p.StepFolders }}
						<option value="{{ .ID }}" data-workspace="{{ $p.StepSpaceOf .ID }}" data-rel="{{ $p.StepRel .ID }}"{{ if eq .ID $s.Folder }} selected{{ end }}>{{ $p.StepLabel .ID }}</option>
						{{- end }}
					</select>
				</div>
			</div>
			{{- else if eq $kind "check" }}
			<div class="wf-fields">
				<div class="field wf-wide">
					<span class="label">Looks for <span class="faint">in the step before it</span></span>
					<input type="text" data-step-markers aria-label="What the check looks for" placeholder="Next:" value="{{ $s.Markers }}" spellcheck="false" autocomplete="off">
					<small class="faint">One or more, comma separated, case ignored. While the step before the check writes one of them, the workflow's steps run again from the first.</small>
				</div>
				<div class="field">
					<span class="label">Passes <span class="faint">at the most</span></span>
					<input type="number" data-step-tries aria-label="Passes at the most" min="2" max="20" step="1" value="{{ $s.Passes }}">
					<small class="faint">The first pass counted, so 3 lets the steps run twice more. The check runs nothing itself and ends ok whatever it finds.</small>
				</div>
			</div>
			{{- else }}
			{{- if eq $kind "prompt" }}
			<div class="field">
				<span class="label">Prompt</span>
				<textarea data-step-prompt-text rows="5" spellcheck="false" aria-label="Prompt" placeholder="Fetch the forecast for the next 3 days and print a short table.">{{ $s.Prompt }}</textarea>
				<small class="faint">Claude gets it with the output of the steps before it, the run's input and the last 3 answers of this workflow. Its answer is the step's output.</small>
			</div>
			{{- end }}
			<div class="wf-fields wf-cascade">
				{{- $space := $p.StepSpaceOf $s.Folder }}{{ if eq $s.Folder "" }}{{ $space = "" }}{{ end }}
				{{- template "stepspace" (list $p $space false $p.StepNone) }}
				<div class="field">
					<span class="label">Asks in <span class="faint">the folder Claude works in</span></span>
					{{- $found := eq $s.Folder "internal" }}
					<select data-step-place aria-label="Folder Claude works in">
						{{- if $p.StepNone }}
						<option value="" data-workspace="" data-rel="."{{ if eq $s.Folder "" }} selected{{ end }}>none (the workspace root)</option>
						{{- end }}
						<option value="internal" data-workspace="internal" data-rel="."{{ if eq $s.Folder "internal" }} selected{{ end }}>the root</option>
						{{- range $p.StepFolders }}{{ if eq .ID $s.Folder }}{{ $found = true }}{{ end }}
						<option value="{{ .ID }}" data-workspace="{{ $p.StepSpaceOf .ID }}" data-rel="{{ $p.StepRel .ID }}"{{ if eq .ID $s.Folder }} selected{{ end }}>{{ $p.StepLabel .ID }}</option>
						{{- end }}
						{{- if and $s.Folder (not $found) }}
						<option value="{{ $s.Folder }}" data-workspace="{{ $p.StepSpaceOf $s.Folder }}" data-rel="{{ $p.StepRel $s.Folder }}" selected>{{ $p.StepLabel $s.Folder }}</option>
						{{- end }}
					</select>
				</div>
				<div class="field">
					<span class="label">Directory</span>
					<input type="text" data-step-dir aria-label="Directory inside the folder" placeholder="none" value="{{ $s.Dir }}" spellcheck="false" autocomplete="off">
				</div>
			</div>
			{{- if eq $kind "summary" }}
			<div class="wf-builtin">
				<span class="label">Claude gets</span>
				<blockquote><p class="wf-question">{{ workflowSummaryQuestion }}</p><p>{{ summaryPrompt }}</p></blockquote>
			</div>
			{{- end }}
			{{- end }}
			<p class="wf-hint" data-step-hint role="status"{{ if not $s.Missing }} hidden{{ end }}>{{ $s.Missing }}</p>
			<p class="wf-error" data-step-error role="alert" hidden></p>
		</div>
	</div>
</li>
{{- end }}

{{ define "stepspace" }}
{{- $p := index . 0 }}{{ $sel := index . 1 }}{{ $flows := index . 2 }}{{ $none := index . 3 -}}
<div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
		{{- if $none }}
		<option value="" data-workspace="" data-rel="."{{ if eq $sel "" }} selected{{ end }}>none (the workspace root)</option>
		{{- end }}
		{{- range $p.StepSpaces }}
		<option value="{{ .ID }}"{{ if eq .ID $sel }} selected{{ end }}>{{ .Title }}</option>
		{{- end }}
		{{- if and $flows $p.StepWorkflows }}
		<option value="workflow:"{{ if eq "workflow:" $sel }} selected{{ end }}>Workflows</option>
		{{- end }}
	</select>
</div>
{{- end }}

{{- define "wfthread" }}{{ $p := index . 0 }}{{ with index . 1 }}
<div class="notes wf-thread" data-workflow-thread="{{ .ID }}"{{ if .Running }} data-follow="{{ .WorkflowID }}" data-page="/workflows/{{ .WorkflowID }}/runs/{{ .ID }}"{{ end }}>
	<article class="note instruction {{ .NoteState }}">
		<header>
			<span class="author">{{ template "user" (list $p .UserName) }}</span>
			<span class="faint" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</span>
			<span class="spacer"></span>
		</header>
		{{- if .Question }}
		<div class="note-body">{{ $p.Message .Question }}</div>
		{{- else }}
		<div class="note-body"><p class="faint">Ran {{ .Name }}, with no input.</p></div>
		{{- end }}
		<ol class="replies">
			<li>
				<header>
					{{- if .Agent }}<span class="author">{{ template "user" (list $p .Agent) }}</span>{{ else }}<span class="author">{{ .Name }}</span>{{ end }}
					{{- if .Running }} <span class="badge note-state working">working</span>
					{{- else }} <span class="faint" title="{{ stamp .FinishedAt }}">{{ ago .FinishedAt }}</span> <span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span>{{ end }}
				</header>
				{{- if .Answer }}
				<div class="note-body">{{ $p.Message .Answer }}</div>
				{{- else if .Running }}
				<div class="note-body"><p class="faint">{{ if .Agent }}Works through the steps; Claude replies here when the run ends.{{ else }}Works through the steps; how the run ended shows here.{{ end }}</p></div>
				{{- else }}
				<div class="note-body"><p>{{ if .Agent }}No answer: the run{{ else }}The run{{ end }} ended {{ .State }}{{ with .Message }}, {{ . }}{{ end }}{{ if .FinishedAt }}, in {{ duration .Took }}{{ end }}.</p></div>
				{{- end }}
			</li>
		</ol>
	</article>
</div>
{{- end }}{{ end }}

{{- define "wffold" }}{{ $open := false }}{{ if gt (len .) 1 }}{{ $open = index . 1 }}{{ end }}{{ with index . 0 }}{{ if .Answer }}<button type="button" class="btn-icon wf-fold" data-answer-toggle aria-expanded="{{ $open }}" aria-controls="wf-answer-{{ .ID }}" title="{{ if $open }}Hide{{ else }}Show{{ end }}" aria-label="{{ if $open }}Hide{{ else }}Show{{ end }}"><svg class="ic chev"><use href="#i-chev"/></svg></button>{{ else }}<span class="wf-fold" aria-hidden="true"></span>{{ end }}{{ end }}{{ end }}

{{- define "wfanswer" }}{{ $p := index . 0 }}{{ $cols := index . 2 }}{{ $open := false }}{{ if gt (len .) 3 }}{{ $open = index . 3 }}{{ end }}{{ with index . 1 }}{{ if .Answer }}
<tr class="wf-answer-row" id="wf-answer-{{ .ID }}" data-answer-of="{{ .ID }}"{{ if not $open }} hidden{{ end }}><td colspan="{{ $cols }}"><div class="note-body">{{ $p.Message .Answer }}</div></td></tr>
{{- end }}{{ end }}{{ end }}