{{ define "layout" -}}
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="robots" content="noindex, nofollow">
	<title>{{ if .Title }}{{ .Title }} | {{ end }}task-ui</title>
	<link rel="icon" href="/static/icon.svg" type="image/svg+xml">
	<link rel="stylesheet" href="/static/vendor/xterm.css">
	<link rel="stylesheet" href="/static/task-ui.css">
</head>
<body>
{{- if not .User }}
	<header class="top login-top">
		<h1><a href="/">task-ui</a>{{ if .Instance }}<span>{{ .Instance }}</span>{{ else }}<span>runbooks</span>{{ end }}</h1>
	</header>
	{{ template "content" . }}
{{- else }}
	<header class="top">
		<h1><a href="/">task-ui</a>{{ if .Instance }}<span>{{ .Instance }}</span>{{ end }}</h1>
		<span class="spacer"></span>
		<div class="who">
			<span><b>{{ .User.Name }}</b></span>
			<span class="badge{{ if eq (print .User.Role) "admin" }} signal{{ end }}">{{ .User.Role }}</span>
			{{- if .User.CanAdmin }}
			<a class="btn btn-ghost btn-sm" href="/admin/keys">SSH keys</a>
			{{- end }}
			{{- if .LoginRequired }}
			<form method="post" action="/logout"><button class="btn btn-ghost btn-sm" type="submit">Sign out</button></form>
			{{- end }}
		</div>
	</header>
	<div class="shell">
		<aside class="side">
			<nav>
				<a class="overview{{ if isPath "/" }} active{{ end }}" href="/">Overview</a>
				{{- if .User.CanAdmin }}
				<a class="overview{{ if isPath "/admin/keys" }} active{{ end }}" href="/admin/keys">SSH keys</a>
				{{- end }}
			</nav>
			<span class="label">Projects</span>
			{{- range $p := .Projects }}
			<div class="project">
				<a href="/p/{{ $p.ID }}" class="{{ if and $.Project (eq $.Project.ID $p.ID) (not $.Runbook) }}active{{ end }}">
					{{ $p.Title }}<small>{{ len $p.Runbooks }}</small>
				</a>
				<ul class="runbooks">
				{{- range $r := $p.Runbooks }}
					{{- $run := running $r.Key }}
					<li><a href="/p/{{ $p.ID }}/{{ $r.Name }}" data-key="{{ $r.Key }}" class="{{ if and $.Runbook (eq $.Runbook.Key $r.Key) }}active{{ end }}" title="{{ $r.Description }}">
						<i class="dot{{ if $run }} running{{ end }}"></i><span>{{ $r.Name }}</span>
					</a></li>
				{{- end }}
				</ul>
				{{- if $p.Error }}<div class="empty">Could not read {{ $p.File }}</div>{{ end }}
			</div>
			{{- else }}
			<div class="empty">No projects. Add a folder with an atkins.yml or Taskfile.yml to the root folder.</div>
			{{- end }}
		</aside>
		<main>
			{{ template "content" . }}
		</main>
	</div>
	{{ template "dialog" . }}
{{- end }}
	<footer>
		<a class="wordmark" href="https://github.com/titpetric/task-ui">task-ui</a>
		<span class="sep">|</span>
		<span>runbooks by <a href="https://github.com/titpetric">{{ "@titpetric" }}</a></span>
	</footer>
	<script src="/static/vendor/xterm.js" defer></script>
	<script src="/static/vendor/addon-fit.js" defer></script>
	<script src="/static/task-ui.js" defer></script>
</body>
</html>
{{- end }}

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

{{ define "state" }}
{{- if . -}}
<span class="state"><i class="dot {{ .State }}"></i>{{ .State }}{{ if and (ne .State "ok") (ne .State "running") }} <span class="faint">({{ .ExitCode }})</span>{{ end }}</span>
{{- else -}}
<span class="faint">never run</span>
{{- end -}}
{{ end }}

{{ define "runButtons" }}
{{- $p := index . 0 }}{{ $r := index . 1 -}}
<button type="button" class="btn btn-primary btn-sm" data-run data-project="{{ $p.ID }}" data-runbook="{{ $r.Name }}" data-title="{{ $p.Title }}" data-interactive="{{ $r.Interactive }}">Run</button>
<button type="button" class="btn btn-danger btn-sm" data-run data-force data-project="{{ $p.ID }}" data-runbook="{{ $r.Name }}" data-title="{{ $p.Title }}" data-interactive="{{ $r.Interactive }}" title="Stop a run in progress, then run again">Run force</button>
{{- end }}
