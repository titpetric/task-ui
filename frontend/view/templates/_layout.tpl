{{ define "layout" -}}
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<title>{{ if .Title }}{{ .Title }} | {{ end }}TaskUI</title>
	<link rel="icon" href="/static/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="/static/vendor/xterm.css?v={{ assetVersion }}">
	<link rel="stylesheet" href="/static/task-ui.css?v={{ assetVersion }}">
</head>
<body>
{{ template "icons" . }}
{{- if not .User }}
	<header class="top login-top">
		<h1><a href="/"><span>TaskUI</span>{{ template "logo" }}</a></h1>
	</header>
	{{ template "content" . }}
{{- else }}
	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="/"><span>TaskUI</span>{{ template "logo" }}</a></h1>
		<span class="spacer"></span>
		<details class="user" id="user-menu">
			<summary aria-label="{{ .User.Name }}, {{ .User.Role }}" title="{{ .User.Name }}"><span class="avatar">{{ .User.Initials }}</span></summary>
			<div class="user-menu">
				<div class="who"><b>{{ .User.Name }}</b><span>{{ .User.Role }}</span></div>
				{{- if .User.CanAdmin }}
				<a href="/settings"><svg class="ic"><use href="#i-settings"/></svg>Settings</a>
				{{- end }}
				{{- if .LoginRequired }}
				<form method="post" action="/logout"><button type="submit"><svg class="ic"><use href="#i-logout"/></svg>Logout</button></form>
				{{- else }}
				<div class="who"><span>auto login, no logout</span></div>
				{{- end }}
			</div>
		</details>
	</header>
	<div class="shell">
		<div class="menu-backdrop" id="menu-backdrop"></div>
		<aside class="side" id="side">
			<nav class="menu">
				<a href="/" class="{{ if .IsPath "/" }}active{{ end }}"><svg class="ic"><use href="#i-grid"/></svg><span>Dashboard</span></a>
				<a href="/workflows" class="{{ if .Under "/workflows" }}active{{ end }}"><svg class="ic"><use href="#i-flow"/></svg><span>Workflows</span></a>
				{{- if .SideWorkflows }}
				<div class="flows">
					{{- range .SideWorkflows }}{{ if not .SidebarHidden }}
					<div class="flow{{ if $.IsPath (print "/workflows/" .ID) }} active{{ end }}">
						<i class="shebang{{ if .Running }} running{{ end }}" aria-hidden="true">#!</i>
						<a class="name" href="/workflows/{{ .ID }}" title="{{ range $i, $l := chain .Steps }}{{ if $i }} → {{ end }}{{ $l.Text }}{{ end }}">{{ .Name }}</a>
						{{- if .Running }}
						<a class="play" href="/workflows/{{ .ID }}" title="{{ .Name }}, running" aria-label="Watch {{ .Name }}"><svg class="ic"><use href="#i-eye"/></svg></a>
						{{- else if .CanRun }}
						<button type="button" class="play" data-workflow-run="{{ .ID }}" title="Run {{ .Name }}" aria-label="Run {{ .Name }}"><svg class="ic"><use href="#i-play"/></svg></button>
						{{- end }}
					</div>
					{{- end }}{{ end }}
				</div>
				{{- end }}
				<a href="/runs" class="{{ if .Under "/runs" }}active{{ end }}"><svg class="ic"><use href="#i-list"/></svg><span>Runs</span><i class="dot{{ if .Running }} running{{ end }}" data-any-running></i></a>
				<a href="/artifacts" class="{{ if .Under "/artifacts" }}active{{ end }}"><svg class="ic"><use href="#i-box"/></svg><span>Artifacts</span></a>
				<a href="/schedules" class="{{ if .Under "/schedules" }}active{{ end }}"><svg class="ic"><use href="#i-clock"/></svg><span>Scheduled</span></a>
			</nav>
			<nav class="folders">
				{{- $total := 0 }}{{ range .SideFolders }}{{ $total = len .Jobs | add $total }}{{ end }}
				{{- range $g := groups .SideFolders }}
				{{- with $p := $g.Top }}
				{{- /* A project folds on its chevron; the row opens it. The open page's project is marked and stays open. */}}
				{{- $active := and $.Folder (or (eq $.Folder.ID $p.ID) (hasPrefix $.Folder.ID (print $p.ID "/"))) }}
				{{- $here := and $.Folder (eq $.Folder.ID $p.ID) }}
				<details class="project{{ if $active }} active{{ end }}" data-project="{{ $p.ID }}" open>
					<summary{{ if $here }} class="active"{{ end }} title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
						<a class="section{{ if $here }} active{{ end }}" href="/f/{{ $p.ID }}" data-folder="{{ $p.ID }}">
							<svg class="ic"><use href="#i-project"/></svg><span>{{ $p.Name }}</span>
							{{- if $p.Error }}<small class="bad">!</small>{{ else if $p.Jobs }}<small>{{ len $p.Jobs }}</small>{{ end }}
						</a>
						<svg class="ic chev"><use href="#i-chev"/></svg>
					</summary>
					{{- if $p.Jobs }}
					<div class="jobs top">
						{{- template "sidejobs" (list $ $p) }}
					</div>
					{{- end }}
					{{- range $g.Nested }}{{ template "sidefolder" (list $ . $total) }}{{ end }}
				</details>
				{{- else }}
				{{- range $g.Nested }}{{ template "sidefolder" (list $ . $total) }}{{ end }}
				{{- end }}
				{{- else }}
				<div class="empty">No folders. Add one from a git repository in Settings, or put a folder with something to run under the root.</div>
				{{- end }}
			</nav>
			<div class="label side-label">Recents</div>
			<nav class="recents" id="recents">
				{{- range $e := .Recents }}
				{{- with $e.Flow }}
				<a href="/workflows/{{ .WorkflowID }}" class="recent flow" title="Workflow {{ .Name }}, {{ .UserName }}, {{ .State }}">
					<i class="dot {{ .State }}"></i>
					<span class="what"><b>{{ .Name }}</b><small><svg class="ic"><use href="#i-flow"/></svg>workflow</small></span>
					{{- if eq .State "running" }}<svg class="ic act"><use href="#i-eye"/></svg>{{ end }}
					<span class="avatar sm" title="{{ .UserName }}">{{ $e.Initials }}</span>
				</a>
				{{- end }}
				{{- with $e.Run }}
				{{- $can := $.CanRunJob .Folder .Job }}
				<a href="/runs/{{ .ID }}" data-attach="{{ .ID }}" class="recent" title="{{ if ne .State "running" }}View{{ else if $can }}Attach to{{ else }}Watch{{ end }} {{ .Folder }} / {{ .Job }}, {{ .UserName }}, {{ .State }}">
					<i class="dot {{ .State }}"></i>
					<span class="what"><b>{{ .Job }}</b><small>{{ .Folder }}</small></span>
					{{- if eq .State "running" }}<svg class="ic act"><use href="#i-{{ if $can }}play{{ else }}eye{{ end }}"/></svg>{{ end }}
					<span class="avatar sm" title="{{ .UserName }}">{{ .Initials }}</span>
				</a>
				{{- end }}
				{{- else }}
				<div class="empty">Runs started by anyone show up here.</div>
				{{- end }}
			</nav>
		</aside>
		<main>
			{{- if .Crumbs }}
			<div class="crumbs"><a href="/">Dashboard</a>{{ range .Crumbs }}<span>{{ if .Href }}<a href="{{ .Href }}">{{ .Text }}</a>{{ else }}{{ .Text }}{{ end }}</span>{{ end }}</div>
			{{- end }}
			{{- if .Notice }}<div class="alert ok" role="status">{{ .Notice }}</div>{{ end }}
			{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}
			{{ template "content" . }}
		</main>
	</div>
	{{ template "dialog" . }}
	{{ template "confirm" . }}
{{- end }}
	<footer>
		<span class="credit">
			<a class="wordmark" href="https://github.com/titpetric/task-ui">task-ui</a>
			<span class="sep" aria-hidden="true">|</span>
			<span>pipelines by <a href="https://github.com/titpetric">{{ "@titpetric" }}</a></span>
			<span class="sep" aria-hidden="true">|</span>
			<a href="https://blacksky.digital/products/task-ui">blacksky.digital</a>
		</span>
	</footer>
	<script src="/static/vendor/xterm.js?v={{ assetVersion }}" defer></script>
	<script src="/static/vendor/addon-fit.js?v={{ assetVersion }}" defer></script>
	<script src="/static/task-ui.js?v={{ assetVersion }}" defer></script>
</body>
</html>
{{- end }}

{{ define "icons" }}
<svg xmlns="http://www.w3.org/2000/svg" style="display:none">
	<symbol id="i-project" viewBox="0 0 16 16"><circle cx="4.5" cy="3.5" r="1.9" fill="none" stroke="currentColor" stroke-width="1.3"/><circle cx="4.5" cy="12.5" r="1.9" fill="none" stroke="currentColor" stroke-width="1.3"/><circle cx="11.5" cy="5.5" r="1.9" fill="none" stroke="currentColor" stroke-width="1.3"/><path d="M4.5 5.4v5.2M11.5 7.4c0 2.6-2.2 3.1-5.2 3.9" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-folder" viewBox="0 0 16 10"><path d="M.65 1.6A1 1 0 0 1 1.65.6h3.6l1.3 1.3h7.8a1 1 0 0 1 1 1v5.5a1 1 0 0 1-1 1H1.65a1 1 0 0 1-1-1z" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/></symbol>
	<symbol id="i-folder-open" viewBox="0 0 16 10"><path d="M.65 8.4V1.6A1 1 0 0 1 1.65.6h3.6l1.3 1.3h6.8a1 1 0 0 1 1 1v.7" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/><path d="M2.6 4.1h12.2a.6.6 0 0 1 .57.8l-1.2 3.9a1 1 0 0 1-.96.7H1.3a.6.6 0 0 1-.57-.8l1-3.9a1 1 0 0 1 .87-.7z" fill="currentColor" opacity=".85"/></symbol>
	<symbol id="i-grid" viewBox="0 0 16 16"><rect x="2" y="2" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="9" y="2" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="2" y="9" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="9" y="9" width="5" height="5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-list" viewBox="0 0 16 16"><path d="M5 4h9M5 8h9M5 12h9" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/><circle cx="2.5" cy="4" r="1" fill="currentColor"/><circle cx="2.5" cy="8" r="1" fill="currentColor"/><circle cx="2.5" cy="12" r="1" fill="currentColor"/></symbol>
	<symbol id="i-box" viewBox="0 0 16 16"><path d="M2 5.5 8 2.5l6 3v5l-6 3-6-3zM2 5.5l6 3 6-3M8 8.5v5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/></symbol>
	<symbol id="i-flow" viewBox="0 0 16 16"><circle cx="3" cy="4" r="1.6" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="13" cy="8" r="1.6" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="3" cy="12" r="1.6" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M4.6 4h3.4a2 2 0 0 1 2 2v.4M10 9.6v.4a2 2 0 0 1-2 2H4.6M10 6.4 11.4 8" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-clock" viewBox="0 0 16 16"><circle cx="8" cy="8" r="6" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M8 4.5V8l2.5 1.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-users" viewBox="0 0 16 16"><circle cx="6" cy="5.5" r="2.5" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M1.5 13.5c0-2.5 2-4 4.5-4s4.5 1.5 4.5 4M11 3.5a2.5 2.5 0 0 1 0 4.5M12 9.5c1.5.4 2.5 1.8 2.5 4" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-settings" viewBox="0 0 16 16"><circle cx="8" cy="8" r="2.2" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-logout" viewBox="0 0 16 16"><path d="M6 2.5H3.5a1 1 0 0 0-1 1v9a1 1 0 0 0 1 1H6M10 11l3-3-3-3M13 8H6" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-copy" viewBox="0 0 16 16"><rect x="5.5" y="5.5" width="8" height="8" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M10.5 5.5v-2a1 1 0 0 0-1-1h-6a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-rotate" viewBox="0 0 16 16"><path d="M13 8a5 5 0 1 1-1.5-3.6M13 2.5v3h-3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-trash" viewBox="0 0 16 16"><path d="M2.5 4h11M6 4V2.5h4V4M4 4l.7 9.5h6.6L12 4M6.5 6.5v5M9.5 6.5v5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-edit" viewBox="0 0 16 16"><path d="M11.5 2.5l2 2L6 12H4v-2zM10 4l2 2" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/><path d="M2.5 13.5h11" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-key" viewBox="0 0 16 16"><circle cx="5.5" cy="10.5" r="3" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M7.7 8.3 13.5 2.5M11 5l2 2M9.5 6.5l1.5 1.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-secret" viewBox="0 0 16 16"><rect x="3" y="7" width="10" height="7" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M5.5 7V5a2.5 2.5 0 0 1 5 0v2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-file" viewBox="0 0 16 16"><path d="M4 1.5h5l3.5 3.5v9.5h-8.5zM9 1.5V5h3.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/></symbol>
	<symbol id="i-play" viewBox="0 0 16 16"><path d="M4.5 2.5v11l8-5.5z" fill="currentColor"/></symbol>
	<symbol id="i-stop" viewBox="0 0 16 16"><rect x="3.5" y="3.5" width="9" height="9" rx="1" fill="currentColor"/></symbol>
	<symbol id="i-eye" viewBox="0 0 16 16"><path d="M1.5 8s2.5-4.5 6.5-4.5S14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8z" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="8" cy="8" r="2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-menu" viewBox="0 0 16 16"><path d="M2.5 4h11M2.5 8h11M2.5 12h11" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-expand" viewBox="0 0 16 16"><path d="M9.5 2.5h4v4M13.5 2.5 9 7M6.5 13.5h-4v-4M2.5 13.5 7 9" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-collapse" viewBox="0 0 16 16"><path d="M13 7H9V3M9 7l4.5-4.5M3 9h4v4M7 9l-4.5 4.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-chev" viewBox="0 0 16 16"><path d="M6 3.5 10.5 8 6 12.5" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-download" viewBox="0 0 16 16"><path d="M8 2.5v8M4.5 7 8 10.5 11.5 7M2.5 13.5h11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
</svg>
{{ end }}

{{ define "dialog" }}
<dialog class="run" id="run-dialog" aria-labelledby="run-title">
	<div class="run-head">
		<span class="label" id="run-kind">Run</span>
		<h3 id="run-title"></h3>
		<span class="spacer"></span>
		<span class="dim" id="run-who"></span>
		<span class="elapsed" id="run-elapsed"></span>
		<span class="state" id="run-state"></span>
	</div>
	<div class="run-body">
		<div class="terminal-host" id="run-terminal"></div>
	</div>
	<p class="run-notice" id="run-notice" role="status"></p>
	<div class="run-reason" id="run-reason" hidden>
		<label for="run-reason-text">Why does this have to run out of hours? The reason is kept with the run.</label>
		<input type="text" id="run-reason-text" placeholder="customer is waiting on the hotfix" autocomplete="off" maxlength="200">
	</div>
	<div class="run-foot">
		<span class="hint" id="run-hint">Closing this window leaves the run running.</span>
		<button type="button" class="btn" id="run-attach" hidden>Attach</button>
		<button type="button" class="btn btn-primary" id="run-force" hidden>Run force</button>
		<button type="button" class="btn" id="run-again" hidden>Run again</button>
		<button type="button" class="btn" id="run-stop" hidden>Stop</button>
		<button type="button" class="btn btn-danger" id="run-kill" hidden>Kill</button>
		<button type="button" class="btn" id="run-close" aria-label="Close">Close</button>
	</div>
</dialog>
{{- /* The source window: a file a folder's pipeline reads, coloured, as
     large as the run window. */}}
<dialog class="run source" id="source-dialog" aria-labelledby="source-title">
	<div class="run-head">
		<span class="label">Source</span>
		<h3 id="source-title"></h3>
		<span class="spacer"></span>
		<button type="button" class="btn" id="source-close" aria-label="Close">Close</button>
	</div>
	<div class="run-body"><pre class="code" id="source-code"></pre></div>
</dialog>
{{ end }}

{{ define "confirm" }}
<dialog class="confirm" id="confirm-dialog" aria-labelledby="confirm-title">
	<h3 id="confirm-title">Are you sure?</h3>
	<p id="confirm-text"></p>
	<div class="confirm-foot">
		<button type="button" class="btn" id="confirm-cancel">Cancel</button>
		<button type="button" class="btn btn-danger" id="confirm-ok">Delete</button>
	</div>
</dialog>
{{ end }}

{{ define "state" }}
{{- if . -}}
<span class="state"><i class="dot {{ .State }}"></i>{{ .State }}{{ if and (ne .State "ok") (ne .State "running") }} <span class="faint">({{ .ExitCode }})</span>{{ end }}</span>
{{- else -}}
<span class="faint">never run</span>
{{- end -}}
{{ end }}

{{ define "who" }}<span class="who-ran"><span class="avatar sm" title="{{ .UserName }}">{{ .Initials }}</span> {{ .UserName }}{{ if .IsScheduled }} <span class="badge">scheduled</span>{{ end }}{{ if .Reason }} <span class="badge bad" title="{{ .Reason }}">out of hours</span>{{ else if .IsForced }} <span class="badge signal">forced</span>{{ end }}</span>{{ end }}

{{ define "hours" }}{{ $text := index . 0 }}{{ $page := index . 1 }}{{ if $text }}{{ hoursLine $text }} {{ if $page.InHours $text }}<span class="badge ok">open now</span>{{ else }}<span class="badge bad">out of hours now</span>{{ end }}{{ else }}<span class="faint">always</span>{{ end }}{{ end }}

{{/* stepeditor is the workflow form: a name and the steps in order.
     list: page, workflow (nil for a new one), API path, method, where to
     go after, the submit label. A step offers every job the user sees. */}}
{{ define "stepeditor" }}
{{- $p := index . 0 }}{{ $w := index . 1 }}{{ $api := index . 2 }}{{ $method := index . 3 }}{{ $next := index . 4 }}{{ $label := index . 5 -}}
<form class="form workflow-form" data-workflow data-api="{{ $api }}{{ if $w }}{{ $w.ID }}{{ end }}" data-method="{{ $method }}" data-next="{{ $next }}">
	<div class="field">
		<label for="wf-name{{ if $w }}-{{ $w.ID }}{{ end }}">Name</label>
		<input type="text" id="wf-name{{ if $w }}-{{ $w.ID }}{{ end }}" name="name" value="{{ if $w }}{{ $w.Name }}{{ end }}" placeholder="deploy" autocomplete="off" required>
	</div>
	<div class="field">
		<label class="check"><input type="checkbox" id="wf-sidebar{{ if $w }}-{{ $w.ID }}{{ end }}" name="sidebar"{{ if or (not $w) (not $w.SidebarHidden) }} checked{{ end }}> Show in sidebar <span class="faint">off for a workflow that is a step of others</span></label>
	</div>
	<ol class="steps" data-steps>
		{{- if $w }}{{ range $w.Steps }}{{ template "steprow" (list $p .Folder .Job .WorkflowID) }}{{ end }}{{ else }}{{ template "steprow" (list $p "" "" "") }}{{ end }}
	</ol>
	<template data-step-template>{{ template "steprow" (list $p "" "" "") }}</template>
	<div class="form-foot">
		<button type="button" class="btn" data-step-add>Add step</button>
		<button class="btn btn-primary" type="submit">{{ $label }}</button>
	</div>
</form>
{{ end }}

{{ define "steprow" }}
{{- $p := index . 0 }}{{ $folder := index . 1 }}{{ $job := index . 2 }}{{ $ref := index . 3 -}}
<li class="step" data-step>
	<select data-step-folder aria-label="Folder or workflow" required>
		{{- range $p.Folders }}{{ if .Jobs }}
		<option value="{{ .ID }}"{{ if eq .ID $folder }} selected{{ end }}>{{ .ID }}</option>
		{{- end }}{{ end }}
		{{- if $p.StepWorkflows }}
		<optgroup label="Workflows">
			{{- range $p.StepWorkflows }}
			<option value="workflow:{{ .ID }}"{{ if eq .ID $ref }} selected{{ end }}>{{ .Name }}</option>
			{{- end }}
		</optgroup>
		{{- end }}
	</select>
	<select data-step-job aria-label="Job" required>
		{{- range $f := $p.Folders }}{{ range $f.Jobs }}
		<option value="{{ .ID }}" data-folder="{{ $f.ID }}"{{ if and (eq $f.ID $folder) (eq .ID $job) }} selected{{ end }}>{{ .Name }}</option>
		{{- end }}{{ end }}
	</select>
	<span class="step-tools">
		<button type="button" class="btn-icon" data-step-up title="Move up" aria-label="Move step up">↑</button>
		<button type="button" class="btn-icon" data-step-down title="Move down" aria-label="Move step down">↓</button>
		<button type="button" class="btn-icon danger" data-step-remove title="Remove" aria-label="Remove step"><svg class="ic"><use href="#i-trash"/></svg></button>
	</span>
</li>
{{- end }}

{{/* chain draws a workflow's steps, or a run's: a folder once for the
     steps that follow in it, a workflow step by its name, each linked. */}}
{{- define "chain" }}
{{- range $i, $l := chain . }}{{ if $i }}<span class="faint"> → </span>{{ end }}<a class="nowrap" href="{{ $l.Href }}"{{ if $l.RunID }} data-attach="{{ $l.RunID }}"{{ end }}>{{ if $l.State }}<i class="dot {{ $l.State }}"></i>{{ end }}{{ if $l.Workflow }}<svg class="ic"><use href="#i-flow"/></svg>{{ end }}{{ $l.Text }}</a>{{ end }}
{{- end }}

{{/* sidefolder is a folder inside a project in the sidebar: a disclosure
     of its jobs, open when it is the page's or the tree is small. */}}
{{- define "sidefolder" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $total := index . 2 }}
{{- $active := and $.Folder (eq $.Folder.ID $p.ID) }}
<details class="folder{{ if $active }} active{{ end }}" data-folder="{{ $p.ID }}"{{ if or $active (le $total 24) }} open{{ end }}>
	<summary title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
		<a href="/f/{{ $p.ID }}"><svg class="ic"><use href="#i-folder{{ if $active }}-open{{ end }}"/></svg><span>{{ with folderDir $p.ID }}<i>{{ . }}/</i>{{ end }}<b>{{ $p.Name }}</b></span></a>
		<i class="dot"></i>
		{{- if $p.Error }}<small class="bad">!</small>{{ else }}<small>{{ len $p.Jobs }}</small>{{ end }}
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
		{{- template "sidejobs" (list $ $p) }}
	</div>
</details>
{{- end }}

{{/* tree lists what the folder page's directory holds: the directories
     first, linked into, then the files, plain. */}}
{{- define "tree" }}
{{- $p := .Folder }}
<ul class="tree">
	{{- range .Tree }}
	<li class="{{ if .Dir }}dir{{ else }}file{{ end }}"><svg class="ic"><use href="#i-{{ if .Dir }}folder{{ else }}file{{ end }}"/></svg>{{ if .Dir }}<a href="/f/{{ $p.ID }}/{{ $.Browse }}{{ .Name }}">{{ .Name }}</a>{{ else }}<span>{{ .Name }}</span>{{ end }}</li>
	{{- else }}
	<li class="empty">Empty.</li>
	{{- end }}
</ul>
{{- end }}

{{/* showsource is the Show button of a file a folder's pipeline reads:
     it opens the file, coloured, in the source window. */}}
{{- define "showsource" }}
{{- $folder := index . 0 }}{{ $file := index . 1 -}}
<button type="button" class="btn btn-sm show-source" data-source="{{ $folder }}" data-file="{{ $file }}" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span>Show</span></button>
{{- end }}

{{ define "runButtons" }}
{{- $p := index . 0 }}{{ $r := index . 1 -}}
<button type="button" class="btn btn-primary btn-sm" data-run data-folder="{{ $p.ID }}" data-job="{{ $r.ID }}" data-title="{{ $p.ID }} / {{ $r.Name }}"><svg class="ic"><use href="#i-play"/></svg>Run</button>
{{- /* Run force stands beside Run only where a plain run can be refused: a job with hours. A run that failed is forced from its window. */}}
{{- if $p.HoursOf $r }}
<button type="button" class="btn btn-danger btn-sm" data-run data-force data-folder="{{ $p.ID }}" data-job="{{ $r.ID }}" data-title="{{ $p.ID }} / {{ $r.Name }}" title="Run out of hours, with a reason">Run force</button>
{{- end }}
{{- end }}

{{- define "sidejobs" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}
{{- range $r := $p.Jobs }}
{{- $run := $.RunningRun $r.Key }}
<div class="job{{ if and $.Job (eq $.Job.Key $r.Key) }} active{{ end }}" data-job-key="{{ $r.Key }}">
	<i class="shebang{{ if $run }} running{{ end }}" aria-hidden="true">#!</i>
	<a href="/f/{{ $p.ID }}/{{ $r.ID }}">{{ if ne $r.ID $r.Name }}<i>{{ index (split $r.ID ":") 0 }}:</i>{{ end }}{{ $r.Name }}</a>
	{{- if $run }}
	<button type="button" class="play" data-attach="{{ $run.ID }}" title="Attach to {{ $r.Name }}, running" aria-label="Attach to {{ $r.Name }}"><svg class="ic"><use href="#i-eye"/></svg></button>
	{{- else if $.CanRun $p $r }}
	<button type="button" class="play" data-run data-folder="{{ $p.ID }}" data-job="{{ $r.ID }}" data-title="{{ $p.ID }} / {{ $r.Name }}" title="Run {{ $r.Name }}" aria-label="Run {{ $r.Name }}"><svg class="ic"><use href="#i-play"/></svg></button>
	{{- end }}
</div>
{{- else }}
<div class="empty">Nothing to run.</div>
{{- end }}
{{- end }}

{{- define "logo" }}<svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="3"/><path class="prompt" d="M8 10l6 6-6 6"/><path class="cursor" d="M16 23h8"/></svg>{{ end }}
