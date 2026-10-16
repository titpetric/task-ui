{{ define "layout" -}}
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<title>{{ if .Title }}{{ .Title }} | {{ end }}TaskUI</title>
	<link rel="icon" href="/static/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="/static/vendor/xterm.css">
	<link rel="stylesheet" href="/static/task-ui.css">
</head>
<body>
{{ template "icons" . }}
{{- if not .User }}
	<header class="top login-top">
		<h1><a href="/">TaskUI</a></h1>
	</header>
	{{ template "content" . }}
{{- else }}
	<header class="top">
		<h1><a href="/">TaskUI</a></h1>
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
		<aside class="side">
			<nav class="folders">
				<a class="overview{{ if isPath "/" }} active{{ end }}" href="/"><svg class="ic"><use href="#i-home"/></svg><span>Overview</span></a>
				{{- range $p := .Projects }}
				<a href="/f/{{ $p.ID }}" data-folder="{{ $p.ID }}" class="folder{{ if and $.Project (eq $.Project.ID $p.ID) }} active{{ end }}" title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
					<svg class="ic"><use href="#i-folder{{ if and $.Project (eq $.Project.ID $p.ID) }}-open{{ end }}"/></svg>
					<span>{{ $p.ID }}</span>
					<i class="dot{{ if projectRunning $p.ID }} running{{ end }}"></i>
					{{- if $p.Error }}<small class="bad">!</small>{{ else }}<small>{{ len $p.Jobs }}</small>{{ end }}
				</a>
				{{- else }}
				<div class="empty">No folders. Add one from a git repository in Settings, or put a folder with something to run under the root.</div>
				{{- end }}
			</nav>
		</aside>
		<main>
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
	<script src="/static/vendor/xterm.js" defer></script>
	<script src="/static/vendor/addon-fit.js" defer></script>
	<script src="/static/task-ui.js" defer></script>
</body>
</html>
{{- end }}

{{ define "icons" }}
<svg xmlns="http://www.w3.org/2000/svg" style="display:none">
	<symbol id="i-folder" viewBox="0 0 16 16"><path d="M1.5 3.5A1.5 1.5 0 0 1 3 2h3.2l1.6 1.5H13a1.5 1.5 0 0 1 1.5 1.5v7A1.5 1.5 0 0 1 13 13.5H3a1.5 1.5 0 0 1-1.5-1.5z" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-folder-open" viewBox="0 0 16 16"><path d="M1.5 3.5A1.5 1.5 0 0 1 3 2h3.2l1.6 1.5H13a1.5 1.5 0 0 1 1.5 1.5V6" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M1.5 6.5h12.3a1 1 0 0 1 .96 1.27l-1.3 4.5a1.5 1.5 0 0 1-1.44 1.08H3a1.5 1.5 0 0 1-1.5-1.5z" fill="currentColor" opacity=".85"/></symbol>
	<symbol id="i-home" viewBox="0 0 16 16"><path d="M2 8.5 8 3l6 5.5M3.5 7.5V13h3V9.5h3V13h3V7.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/></symbol>
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
</svg>
{{ end }}

{{ define "dialog" }}
<dialog class="run" id="run-dialog" aria-labelledby="run-title">
	<div class="run-head">
		<span class="label" id="run-kind">Run</span>
		<h3 id="run-title"></h3>
		<span class="spacer"></span>
		<span class="elapsed" id="run-elapsed"></span>
		<span class="state" id="run-state"></span>
	</div>
	<div class="run-body">
		<div class="terminal-host" id="run-terminal"></div>
	</div>
	<p class="run-notice" id="run-notice" role="status"></p>
	<div class="run-foot">
		<span class="hint" id="run-hint">Closing this window stops the run.</span>
		<button type="button" class="btn btn-primary" id="run-force" hidden>Run force</button>
		<button type="button" class="btn btn-danger" id="run-stop" hidden>Stop</button>
		<button type="button" class="btn" id="run-again" hidden>Run again</button>
		<button type="button" class="btn" id="run-close">Close</button>
	</div>
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

{{ define "runButtons" }}
{{- $p := index . 0 }}{{ $r := index . 1 -}}
<button type="button" class="btn btn-primary btn-sm" data-run data-project="{{ $p.ID }}" data-job="{{ $r.ID }}" data-title="{{ $p.ID }}" data-interactive="{{ $r.Interactive }}">Run</button>
<button type="button" class="btn btn-danger btn-sm" data-run data-force data-project="{{ $p.ID }}" data-job="{{ $r.ID }}" data-title="{{ $p.ID }}" data-interactive="{{ $r.Interactive }}" title="Stop a run in progress, then run again">Run force</button>
{{- end }}

{{ define "crumbs" }}
<div class="crumbs"><a href="/">Overview</a>{{ range . }}<span>{{ if .Href }}<a href="{{ .Href }}">{{ .Text }}</a>{{ else }}{{ .Text }}{{ end }}</span>{{ end }}</div>
{{ end }}
