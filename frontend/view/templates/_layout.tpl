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
<body data-roles='{{ json .Roles }}' data-workspaces='{{ json .Workspaces }}'{{ with .User }} data-me="{{ .Name }}"{{ end }} data-user-menu="{{ if .Can "user/edit" }}edit{{ else if .Can "user/view" }}view{{ end }}">
{{ template "icons" . }}
{{- if not .User }}
	<header class="top login-top">
		<h1><a href="/"><span>TaskUI</span>{{ template "logo" }}</a></h1>
	</header>
	{{ template "content" . }}
{{- else }}
	{{- if .Can "prompt/edit" }}
	{{- $ask := .HeaderPrompt }}
	<header class="top" data-prompt-folder="{{ $ask.Folder }}" data-prompt-dir="{{ $ask.Dir }}" data-prompt-run="{{ $ask.Run }}" data-prompt-label="{{ $ask.Label }}" data-prompt-href="{{ $ask.Href }}">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		{{- /* The logo toggles the prompt's panel; the name links home. */}}
		<h1><a href="/"><span>TaskUI</span></a><button type="button" class="logo-toggle" id="header-toggle" tabindex="-1" aria-label="Prompt" aria-controls="header-answer" aria-expanded="false">{{ template "logo" }}</button></h1>
	{{- else }}
	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="/"><span>TaskUI</span>{{ template "logo" }}</a></h1>
	{{- end }}
		{{- if .Can "prompt/edit" }}
		<input type="text" class="header-prompt" id="header-prompt" tabindex="-1" autocomplete="off" spellcheck="false">
		<div class="header-answer" id="header-answer" hidden></div>
		{{- else }}
		<span class="spacer"></span>
		{{- end }}
		<details class="user" id="user-menu">
			<summary aria-label="{{ .User.Name }}, {{ .User.EffectiveRole }}" title="{{ .User.Name }}"><span class="avatar role-{{ .User.EffectiveRole }}">{{ .User.Initials }}</span></summary>
			<div class="user-menu">
				<div class="who"><b>{{ .User.Name }}</b><span>{{ .User.EffectiveRole }}</span></div>
				<a href="/users/{{ .User.Name }}"><svg class="ic"><use href="#i-users"/></svg>Profile</a>
				<a href="/keys"><svg class="ic"><use href="#i-key"/></svg>SSH keys</a>
				{{- if .CanSettings }}
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
				<a href="/inventory" class="{{ if .Under "/inventory" }}active{{ end }}"><svg class="ic"><use href="#i-server"/></svg><span>Inventory</span></a>
				<a href="/workflows" class="{{ if .Under "/workflows" }}active{{ end }}"><svg class="ic"><use href="#i-flow"/></svg><span>Workflows</span></a>
				{{- if .SideWorkflows }}
				<div class="flows">
					{{- range .SideWorkflows }}{{ if not .SidebarHidden }}
					<div class="flow{{ if $.IsPath (print "/workflows/" .ID) }} active{{ end }}">
						<i class="shebang{{ if .Running }} running{{ end }}" aria-hidden="true">#!</i>
						<a class="name" href="/workflows/{{ .ID }}" title="{{ range $i, $l := $.Links .Steps }}{{ if $i }} → {{ end }}{{ $l.Text }}{{ end }}">{{ .Name }}</a>
						{{- if .Running }}
						<a class="play" href="/workflows/{{ .ID }}" title="{{ .Name }}, running" aria-label="Watch {{ .Name }}"><svg class="ic"><use href="#i-eye"/></svg></a>
						{{- else if .CanRun }}
						<button type="button" class="play" data-workflow-run="{{ .ID }}"{{ if .Prompted }} data-workflow-prompt="{{ .Name }}"{{ end }} title="Run {{ .Name }}" aria-label="Run {{ .Name }}"><svg class="ic"><use href="#i-play"/></svg></button>
						{{- end }}
					</div>
					{{- end }}{{ end }}
				</div>
				{{- end }}
				<a href="/runs" class="{{ if .Under "/runs" }}active{{ end }}"><svg class="ic"><use href="#i-list"/></svg><span>Runs</span><i class="dot{{ if .Running }} running{{ end }}" data-any-running></i></a>
				<a href="/artifacts" class="{{ if .Under "/artifacts" }}active{{ end }}"><svg class="ic"><use href="#i-box"/></svg><span>Artifacts</span></a>
				<a href="/schedules" class="{{ if .Under "/schedules" }}active{{ end }}"><svg class="ic"><use href="#i-clock"/></svg><span>Scheduled</span></a>
				<a href="/notes" class="{{ if .Under "/notes" }}active{{ end }}"><svg class="ic"><use href="#i-note"/></svg><span>Notes</span></a>
			</nav>
			<nav class="folders">
				{{- $total := 0 }}{{ range .SideFolders }}{{ $total = len .Jobs | add $total }}{{ end }}
				{{- range $g := .Groups }}
				{{- if $g.Workspace }}{{ template "sideworkspace" (list $ $g $total) }}{{ else }}{{ template "sideproject" (list $ $g $total) }}{{ end }}
				{{- else }}
				<div class="empty">No folders yet. Add a git repository in Settings, or put a folder with jobs under the root.</div>
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
			{{- template "setup-trail" .SetupTrail }}
		</main>
		{{- /* The right rail, on wide screens: the latest runs and workflow
		     runs by anyone. Narrower, Runs in the sidebar lists them. */}}
		<aside class="rail" aria-label="Recents">
			<div class="rail-inner">
				<div class="label side-label"><a href="/runs">Recents</a></div>
				<nav class="recents" id="recents">
					{{- range $g := .Recents }}
					<div class="recent-group">{{ $g.Label }}</div>
					{{- range $e := $g.Entries }}
					{{- with $e.Flow }}
					<a href="/workflows/{{ .WorkflowID }}" class="recent flow" title="Workflow {{ .Name }}, {{ .UserName }}, {{ .State }}">
						<i class="dot {{ .State }}"></i>
						<span class="what"><b>{{ .Name }}</b><small><svg class="ic"><use href="#i-flow"/></svg>workflow</small></span>
						{{- if not (ended .State) }}<svg class="ic act"><use href="#i-eye"/></svg>{{ end }}
						{{ template "avatar" (list $ .UserName) }}
					</a>
					{{- end }}
					{{- with $e.Run }}
					{{- $can := $.CanRunJob .Folder .Dir .Job }}
					<a href="/runs/{{ .ID }}" class="recent" title="Open {{ .Name }}, {{ .UserName }}, {{ .State }}">
						<i class="dot {{ .State }}"></i>
						<span class="what"><b>{{ .Job }}</b><small>{{ .Place }}</small></span>
						{{- if not (ended .State) }}<svg class="ic act"><use href="#i-{{ if $can }}play{{ else }}eye{{ end }}"/></svg>{{ end }}
						{{ template "avatar" (list $ .UserName) }}
					</a>
					{{- end }}
					{{- end }}
					{{- else }}
					<div class="empty">Runs started by anyone show up here.</div>
					{{- end }}
					{{- /* The last slot when the rail can't show every entry:
					     the script shows it in place of the last one that fits. */}}
					<a href="/runs" class="recent more" data-recents-more hidden><svg class="ic"><use href="#i-list"/></svg><span class="what"><b>All runs</b><small></small></span></a>
				</nav>
			</div>
		</aside>
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
	<symbol id="i-prompt" viewBox="0 0 16 16"><path d="M2.5 3.5a1 1 0 0 1 1-1h9a1 1 0 0 1 1 1v6.5a1 1 0 0 1-1 1H7l-3 2.5v-2.5h-.5a1 1 0 0 1-1-1z" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round"/><path d="M5.5 5.5 7 7 5.5 8.5M8.5 8.5h2" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-server" viewBox="0 0 16 16"><rect x="2.5" y="2.5" width="11" height="4.5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="2.5" y="9" width="11" height="4.5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="5" cy="4.75" r=".8" fill="currentColor"/><circle cx="5" cy="11.25" r=".8" fill="currentColor"/></symbol>
	<symbol id="i-plug" viewBox="0 0 16 16"><path d="M6 1.5v3M10 1.5v3M4 4.5h8v3a4 4 0 0 1-8 0zM8 11.5v3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-note" viewBox="0 0 16 16"><path d="M3 2.5h10v7.5l-3.5 3.5H3zM13 10H9.5v3.5M5.5 5.5h5M5.5 8h3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round" stroke-linecap="round"/></symbol>
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
	<symbol id="i-terminal" viewBox="0 0 16 16"><rect x="1.5" y="2.5" width="13" height="11" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M4.5 6l2 2-2 2M8 10.5h3.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-play" viewBox="0 0 16 16"><path d="M4.5 2.5v11l8-5.5z" fill="currentColor"/></symbol>
	<symbol id="i-stop" viewBox="0 0 16 16"><rect x="3.5" y="3.5" width="9" height="9" rx="1" fill="currentColor"/></symbol>
	<symbol id="i-eye" viewBox="0 0 16 16"><path d="M1.5 8s2.5-4.5 6.5-4.5S14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8z" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="8" cy="8" r="2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-menu" viewBox="0 0 16 16"><path d="M2.5 4h11M2.5 8h11M2.5 12h11" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-expand" viewBox="0 0 16 16"><path d="M9.5 2.5h4v4M13.5 2.5 9 7M6.5 13.5h-4v-4M2.5 13.5 7 9" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-collapse" viewBox="0 0 16 16"><path d="M13 7H9V3M9 7l4.5-4.5M3 9h4v4M7 9l-4.5 4.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-more" viewBox="0 0 16 16"><circle cx="3.5" cy="8" r="1.3" fill="currentColor"/><circle cx="8" cy="8" r="1.3" fill="currentColor"/><circle cx="12.5" cy="8" r="1.3" fill="currentColor"/></symbol>
	<symbol id="i-chev" viewBox="0 0 16 16"><path d="M6 3.5 10.5 8 6 12.5" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-download" viewBox="0 0 16 16"><path d="M8 2.5v8M4.5 7 8 10.5 11.5 7M2.5 13.5h11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
</svg>
{{ end }}

{{- /* The replay button of a finished run with a recording: it plays
     the recording in the run window at its recorded pace, and reads
     nothing from the server but the recording. */}}
{{ define "replay" }}{{ if and .CastPath (ended .State) }}<button type="button" class="btn btn-sm replay" data-replay="{{ .ID }}" data-replay-title="{{ .Name }}" data-replay-who="{{ .UserName }}" title="Replay the recording" aria-label="Replay the recording">|&gt;</button>{{ end }}{{ end }}

{{ define "dialog" }}
<dialog class="run" id="run-dialog" aria-labelledby="run-title">
	<div class="run-head">
		<span class="label" id="run-kind">Run</span>
		<h3 id="run-title"></h3>
		<span class="run-subtitle" id="run-subtitle" hidden></span>
		<span class="tag rec" id="run-rec" hidden>Recording</span>
		<span class="tag input-on" id="run-input" hidden>Input on</span>
		<span class="spacer"></span>
		<span class="dim" id="run-who"></span>
		<span class="elapsed" id="run-elapsed"></span>
		<span class="state" id="run-state"></span>
		<button type="button" class="btn btn-sm" id="run-replay" title="Replay the recording" aria-label="Replay the recording" hidden>|&gt;</button>
		<a class="btn btn-sm" id="run-page" href="/runs" hidden>Open run page</a>
	</div>
	<div class="run-body">
		<div class="terminal-host" id="run-terminal"></div>
	</div>
	<p class="run-notice" id="run-notice" role="status"></p>
	<div class="run-reason" id="run-reason" hidden>
		<label for="run-reason-text">Why does this need to run outside its hours? task-ui saves your reason with the run.</label>
		<input type="text" id="run-reason-text" placeholder="customer is waiting on the hotfix" autocomplete="off" maxlength="200">
	</div>
	<div class="run-foot">
		<nav class="run-chain" id="run-chain" aria-label="The runs of the chain" hidden>
			<button type="button" class="btn btn-sm" id="run-prev" title="The run before this one">‹ Previous</button>
			<span class="run-pos" id="run-pos"></span>
			<button type="button" class="btn btn-sm" id="run-next" title="The run after this one">Next ›</button>
		</nav>
		<a class="btn btn-sm" id="run-report" href="#" title="Claude's answer or summary, as a document" hidden>Report</a>
		<span class="hint" id="run-hint"></span>
		<button type="button" class="btn" id="run-attach" hidden>Attach</button>
		<button type="button" class="btn btn-primary" id="run-force" hidden>Run force</button>
		<button type="button" class="btn" id="run-again" hidden>Run again</button>
		<button type="button" class="btn" id="run-take" aria-pressed="false" title="Send your keystrokes to the run, to answer a prompt it waits on" hidden>Take input</button>
		<button type="button" class="btn" id="run-stop" hidden>Stop</button>
		<button type="button" class="btn btn-danger" id="run-kill" hidden>Kill</button>
		<button type="button" class="btn" id="run-pause" hidden>Pause</button>
		<button type="button" class="btn" id="run-restart" hidden>Restart</button>
		<button type="button" class="btn" id="run-close" aria-label="Close">Close</button>
	</div>
</dialog>
{{- /* The image window: an image or an svg of the artifacts, scaled to
     fit, in place of a download. A file of a checkout that runs carries
     the Run or Attach of its row in the foot. */}}
<dialog class="run preview" id="preview-dialog" aria-labelledby="preview-title">
	<div class="run-head">
		<span class="label" id="preview-label">Artifact</span>
		<h3 id="preview-title"></h3>
	</div>
	<div class="run-body preview-body"><img id="preview-image" alt=""><div id="preview-doc" class="preview-doc" hidden></div></div>
	<div class="run-foot">
		<span class="hint"><select id="preview-version" aria-label="Version" hidden></select></span>
		<a class="btn" id="preview-download" href="#" download>Download</a>
		<span class="preview-actions" id="preview-actions"></span>
		<button type="button" class="btn" id="preview-close" aria-label="Close">Close</button>
	</div>
</dialog>
{{- /* The pipeline window: one workflow run's pipeline, fetched from its
     page. */}}
<dialog class="run pipeline-window" id="pipeline-dialog" aria-labelledby="pipeline-title">
	<div class="run-head">
		<span class="label">Workflow run</span>
		<h3 id="pipeline-title"></h3>
	</div>
	<div class="run-body pipeline-body" id="pipeline-body"></div>
	<div class="run-foot">
		<span class="hint"></span>
		<a class="btn" id="pipeline-page" href="#">Open page</a>
		<button type="button" class="btn" id="pipeline-close" aria-label="Close">Close</button>
	</div>
</dialog>
{{- /* The summary window: Claude's summary of a run's or a workflow
     run's output. */}}
<dialog class="run preview run-summary" id="summary-dialog" aria-labelledby="summary-title">
	<div class="run-head">
		<span class="label">Summary</span>
		<h3 id="summary-title"></h3>
	</div>
	<div class="run-body preview-body"><div id="summary-text" class="preview-doc markdown"></div></div>
	<div class="run-foot">
		<span class="hint" id="summary-meta"></span>
		<a class="btn" id="summary-run" href="#">View run</a>
		<button type="button" class="btn" id="summary-close" aria-label="Close">Close</button>
	</div>
</dialog>
{{- /* The source window: a file a folder's pipeline reads, coloured, as
     large as the run window. */}}
<dialog class="run source" id="source-dialog" aria-labelledby="source-title">
	<div class="run-head">
		<span class="label">Source</span>
		<h3 id="source-title"></h3>
	</div>
	<div class="run-body"><pre class="code" id="source-code"></pre></div>
	<div class="run-foot">
		<span class="hint"></span>
		<button type="button" class="btn" id="source-close" aria-label="Close">Close</button>
	</div>
</dialog>
{{ end }}

{{ define "confirm" }}
<dialog class="confirm" id="workflow-input-dialog" aria-labelledby="workflow-input-title">
	<form method="dialog" data-workflow-input>
		<h3 id="workflow-input-title">Run</h3>
		<p>Optional input for the workflow's prompt. Claude reads it after the prompt, under the line <code>Input:</code>.</p>
		<textarea name="input" rows="3" spellcheck="false" aria-label="Input for the prompt"></textarea>
		<div class="confirm-foot">
			<button type="button" class="btn" data-workflow-input-cancel>Cancel</button>
			<button type="submit" class="btn btn-primary">Run</button>
		</div>
	</form>
</dialog>
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
<span class="state"><i class="dot {{ .State }}"></i>{{ .State }}{{ if and (ne .State "ok") (ended .State) }} <span class="faint">({{ .ExitCode }})</span>{{ end }}</span>
{{- else -}}
<span class="faint">never run</span>
{{- end -}}
{{ end }}

{{- /* The pages either side of a page of runs, from a view.Pager. */}}
{{ define "pager" }}{{ with . }}
<nav class="pager" aria-label="Pages">
	{{- if .Prev }}<a class="btn btn-sm" href="{{ .Prev }}" rel="prev">Previous</a>{{ end }}
	{{- if .Next }}<a class="btn btn-sm" href="{{ .Next }}" rel="next">Next</a>{{ end }}
</nav>
{{- end }}{{ end }}

{{- /* The footer of a step of the settings, from a view.SetupTrail: the
     step before, the steps in order with this one in bold, the step after. */}}
{{ define "setup-trail" }}{{ with . }}
<nav class="setup-trail" aria-label="Setup steps">
	{{- with .Previous }}<a class="btn btn-sm" href="{{ .Href }}" rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: {{ .Name }}</a>{{ end }}
	<ol class="steps">
		{{- range .Steps }}
		<li>{{ if .Current }}<b aria-current="step">{{ .Name }}</b>{{ else }}<a href="{{ .Href }}">{{ .Name }}</a>{{ end }}</li>
		{{- end }}
	</ol>
	{{- with .Next }}<a class="btn btn-sm" href="{{ .Href }}" rel="next">Next: {{ .Name }}<svg class="ic"><use href="#i-chev"/></svg></a>{{ end }}
</nav>
{{- end }}{{ end }}

{{- /* The page size select of a form that lists runs. */}}
{{ define "pagesize" }}
	<label class="page-size">Page size
		<select name="n">
			{{- range list "15" "50" "100" }}<option value="{{ . }}"{{ if eq (index $.Form "n") . }} selected{{ end }}>{{ . }}</option>{{ end }}
		</select>
	</label>
{{- end }}

{{/* avatar takes (list $ name): the initials of a user in the colour of their role. */}}
{{ define "avatar" }}{{ $role := index (index . 0).Roles (index . 1) }}<span class="avatar sm{{ with $role }} role-{{ . }}{{ end }}" title="{{ index . 1 }}">{{ initials (index . 1) }}</span>{{ end }}

{{/* who takes (list $ run): who started a run, and how. */}}
{{ define "who" }}{{ $p := index . 0 }}{{ with index . 1 }}<span class="who-ran">{{ template "user" (list $p .UserName) }}{{ if .IsScheduled }} <span class="badge">scheduled</span>{{ end }}{{ if .Reason }} <span class="badge bad" title="{{ .Reason }}">out of hours</span>{{ else if .IsForced }} <span class="badge signal">forced</span>{{ end }}</span>{{ end }}{{ end }}

{{ define "hours" }}{{ $text := index . 0 }}{{ $page := index . 1 }}{{ if $text }}{{ hoursLine $text }} {{ if $page.InHours $text }}<span class="badge ok">open now</span>{{ else }}<span class="badge bad">out of hours now</span>{{ end }}{{ else }}<span class="faint">always</span>{{ end }}{{ end }}

{{/* stepeditor is the workflow form: a name, the steps in order, each
     with its toggle, and for a user with prompt/edit the prompt steps,
     the prompt Claude gets once they end and where it runs. list: page, workflow (nil for
     a new one), API path, method, where to go after, the submit label. A
     step lists every job the user sees. */}}
{{ define "stepeditor" }}
{{- $p := index . 0 }}{{ $w := index . 1 }}{{ $api := index . 2 }}{{ $method := index . 3 }}{{ $next := index . 4 }}{{ $label := index . 5 -}}
<form class="form workflow-form" data-workflow data-api="{{ $api }}{{ if $w }}{{ $w.ID }}{{ end }}" data-method="{{ $method }}" data-next="{{ $next }}">
	<div class="field">
		<label for="wf-name{{ if $w }}-{{ $w.ID }}{{ end }}">Name</label>
		<input type="text" id="wf-name{{ if $w }}-{{ $w.ID }}{{ end }}" name="name" value="{{ if $w }}{{ $w.Name }}{{ end }}" placeholder="deploy" autocomplete="off" required>
	</div>
	<div class="field">
		<label class="check"><input type="checkbox" id="wf-sidebar{{ if $w }}-{{ $w.ID }}{{ end }}" name="sidebar"{{ if or (not $w) (not $w.SidebarHidden) }} checked{{ end }}> Show in sidebar <span class="faint">turn off for a workflow you only use as a step of others</span></label>
	</div>
	<ol class="steps" data-steps>
		{{- if $w }}{{ range $w.Steps }}{{ if .Prompt }}{{ template "promptrow" (list $p .Folder .Dir .Prompt .Disabled) }}{{ else }}{{ template "steprow" (list $p .Folder .Job .WorkflowID .Disabled .Args) }}{{ end }}{{ end }}{{ else }}{{ template "steprow" (list $p "" "" "" false "") }}{{ end }}
	</ol>
	<template data-step-template>{{ template "steprow" (list $p "" "" "" false "") }}</template>
	{{- if $p.Can "prompt/edit" }}
	<template data-prompt-template>{{ template "promptrow" (list $p "_" "" "" false) }}</template>
	{{- end }}
	{{- $id := "" }}{{ if $w }}{{ $id = print "-" $w.ID }}{{ end }}
	{{- if $p.Can "prompt/edit" }}
	{{- $on := and $w $w.Prompted }}
	<section class="switchable prompt-section{{ if not $on }} off{{ end }}" data-switchable>
		<div class="switch-head">
			{{- template "switch" (list $on "Ask Claude once the steps end") }}
			<label class="label" for="wf-prompt{{ $id }}">Prompt</label>
			<span class="faint">Claude gets the prompt after the steps finish, with their output. With every step off, the workflow runs the prompt alone.</span>
		</div>
		<div class="switch-body">
			<div class="field">
				<textarea id="wf-prompt{{ $id }}" name="prompt" rows="4" spellcheck="false" placeholder="summarize the run, change nothing" data-workflow-prompt-text{{ if not $on }} disabled{{ end }}>{{ if $w }}{{ $w.Prompt }}{{ end }}</textarea>
				<div class="field-tools"><button type="button" class="btn btn-sm" data-workflow-preset="{{ summaryPrompt }}" title="Fill the prompt with the short summary of Run and summarize">Summarize</button></div>
				<p class="faint field-note">Claude also gets the last 3 answers of this workflow.</p>
			</div>
			<div class="field">
				<label for="wf-prompt-folder{{ $id }}">Run in <span class="faint">the folder Claude works in, as you</span></label>
				<select id="wf-prompt-folder{{ $id }}" name="prompt_folder"{{ if not $on }} disabled{{ end }}>
					<option value="">the root</option>
					{{- range $p.Folders }}
					<option value="{{ .ID }}"{{ if and $w (eq .ID $w.PromptFolder) }} selected{{ end }}>{{ .ID }}</option>
					{{- end }}
				</select>
			</div>
		</div>
	</section>
	{{- else if and $w $w.Prompted }}
	<input type="hidden" name="prompt" value="{{ $w.Prompt }}">
	<input type="hidden" name="prompt_folder" value="{{ $w.PromptFolder }}">
	{{- end }}
	<div class="form-foot">
		<button type="button" class="btn" data-step-add>Add step</button>
		{{- if $p.Can "prompt/edit" }}
		<button type="button" class="btn" data-step-add-prompt><svg class="ic"><use href="#i-prompt"/></svg>Add prompt step</button>
		{{- end }}
		<button class="btn btn-primary" type="submit">{{ $label }}</button>
	</div>
</form>
{{ end }}

{{/* switch takes (list on label): the leading switch of a part of a
     form that turns it on, with its settings, or off, collapsed to its
     heading. task-ui.js sets the off class of its [data-switchable] and
     the disabled state of the fields inside it. */}}
{{ define "switch" -}}
<label class="switch" title="{{ index . 1 }}"><input type="checkbox" data-switch aria-label="{{ index . 1 }}"{{ if index . 0 }} checked{{ end }}><i aria-hidden="true"></i></label>
{{- end }}

{{/* hostjobs lists the jobs that run a command on hosts in a job
     picker, under Hosts, for a user with the hosts permission. */}}
{{- define "hostjobs" }}{{ $job := "" }}{{ if gt (len .) 1 }}{{ $job = index . 1 }}{{ end }}{{ with index . 0 }}{{ if .HostJobs }}
			<optgroup label="Hosts">
				{{- range .HostJobs }}
				<option value="{{ .ID }}" data-folder="_" title="{{ .Description }}"{{ if eq .ID $job }} selected{{ end }}>{{ .Name }}</option>
				{{- end }}
			</optgroup>
{{- end }}{{ end }}{{ end }}

{{ define "steprow" }}
{{- $p := index . 0 }}{{ $folder := index . 1 }}{{ $job := index . 2 }}{{ $ref := index . 3 }}{{ $off := index . 4 }}{{ $args := index . 5 -}}
<li class="step switchable{{ if $off }} off{{ end }}" data-step data-switchable>
	{{- template "switch" (list (not $off) "Run this step") }}
	<span class="step-head" data-step-head>{{ if $ref }}{{ range $p.StepWorkflows }}{{ if eq .ID $ref }}{{ .Name }}{{ end }}{{ end }}{{ else }}{{ $folder }} / {{ $job }}{{ end }}</span>
	<select data-step-folder aria-label="Folder or workflow" required>
		{{- range $p.Folders }}{{ if .Jobs }}
		<option value="{{ .ID }}"{{ if eq .ID $folder }} selected{{ end }}>{{ .ID }}</option>
		{{- end }}{{ end }}
		{{- if $p.HostJobs }}
		<option value="_"{{ if eq "_" $folder }} selected{{ end }}>_</option>
		{{- end }}
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
		{{- template "hostjobs" (list $p $job) }}
	</select>
	<input type="text" data-step-args aria-label="Args, for a job that takes them" placeholder="args" value="{{ argLine $job $args }}" spellcheck="false" autocomplete="off">
	<span class="step-tools">
		<button type="button" class="btn-icon" data-step-up title="Move up" aria-label="Move step up">↑</button>
		<button type="button" class="btn-icon" data-step-down title="Move down" aria-label="Move step down">↓</button>
		<button type="button" class="btn-icon danger" data-step-remove title="Remove" aria-label="Remove step"><svg class="ic"><use href="#i-trash"/></svg></button>
	</span>
</li>
{{- end }}

{{/* promptrow is a prompt step of the workflow form: the prompt Claude
     gets with the output of the steps before it, and the folder it runs
     in; list: page, folder, dir, prompt, off. Its answer is its output
     for the prompts after it. */}}
{{ define "promptrow" }}
{{- $p := index . 0 }}{{ $folder := index . 1 }}{{ $dir := index . 2 }}{{ $prompt := index . 3 }}{{ $off := index . 4 -}}
<li class="step prompt-step switchable{{ if $off }} off{{ end }}" data-step data-step-prompt data-switchable>
	{{- template "switch" (list (not $off) "Run this step") }}
	<span class="step-head" data-step-head><svg class="ic"><use href="#i-prompt"/></svg>prompt in {{ $folder }}{{ with $dir }}/{{ . }}{{ end }}</span>
	<textarea data-step-prompt-text rows="3" spellcheck="false" aria-label="The prompt" placeholder="fetch the data, print a short table" required>{{ $prompt }}</textarea>
	<select data-step-place aria-label="Run in">
		<option value="_"{{ if eq $folder "_" }} selected{{ end }}>the root</option>
		{{- range $p.Folders }}
		<option value="{{ .ID }}"{{ if eq .ID $folder }} selected{{ end }}>{{ .ID }}</option>
		{{- end }}
	</select>
	<input type="hidden" data-step-dir value="{{ $dir }}">
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
{{- $ := index . 0 }}
{{- range $i, $l := $.Links (index . 1) }}{{ if $i }}<span class="faint then" aria-hidden="true"></span>{{ end }}<a class="nowrap" href="{{ $l.Href }}"{{ with $l.Prompt }} title="{{ . }}"{{ end }}>{{ if $l.State }}<i class="dot {{ $l.State }}"></i>{{ end }}{{ if $l.Workflow }}<svg class="ic"><use href="#i-flow"/></svg>{{ else if $l.Prompt }}<svg class="ic"><use href="#i-prompt"/></svg>{{ end }}{{ $l.Text }}</a>{{ end }}
{{- end }}

{{- /* sideproject is a project in the sidebar: its row, its jobs and
     the folders inside it. A project folds on its chevron; the row opens
     it. The project of the open page shows as active and stays open. Without its
     top folder listed, the folders stand on their own. */}}
{{- define "sideproject" }}
{{- $ := index . 0 }}{{ $g := index . 1 }}{{ $total := index . 2 }}
{{- with $p := $g.Top }}
				{{- $active := and $.Folder (or (eq $.Folder.ID $p.ID) (hasPrefix $.Folder.ID (print $p.ID "/"))) }}
				{{- $here := and $.Folder (eq $.Folder.ID $p.ID) }}
				<details class="project{{ if $active }} active{{ end }}" data-project="{{ $p.ID }}" open>
					<summary{{ if $here }} class="active"{{ end }} title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
						<a class="section{{ if $here }} active{{ end }}" href="{{ $.Href $p.ID }}" data-folder="{{ $p.ID }}">
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
					{{- range $g.Nested }}{{ template "sidefolder" (list $ . $total $g.ID) }}{{ end }}
				</details>
{{- else }}
{{- range $g.Nested }}{{ template "sidefolder" (list $ . $total $g.ID) }}{{ end }}
{{- end }}
{{- end }}

{{- /* sideworkspace is a workspace in the sidebar: its row folds the
     projects under it, each a project with its own fold. Without its top
     folder listed, the projects stand on their own. */}}
{{- define "sideworkspace" }}
{{- $ := index . 0 }}{{ $g := index . 1 }}{{ $total := index . 2 }}
{{- with $p := $g.Top }}
{{- $active := and $.Folder (or (eq $.Folder.ID $p.ID) (hasPrefix $.Folder.ID (print $p.ID "/"))) }}
{{- $here := and $.Folder (eq $.Folder.ID $p.ID) }}
<details class="workspace{{ if $active }} active{{ end }}" data-project="{{ $p.ID }}" open>
	<summary{{ if $here }} class="active"{{ end }} title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
		<svg class="ic chev"><use href="#i-chev"/></svg>
		<a class="section{{ if $here }} active{{ end }}" href="{{ $.Href $p.ID }}" data-folder="{{ $p.ID }}">
			<span>{{ $p.Name }}</span>
			{{- if $p.Error }}<small class="bad">!</small>{{ else }}<small>{{ len $g.Projects }}</small>{{ end }}
		</a>
	</summary>
	{{- if $p.Jobs }}
	<div class="jobs top">
		{{- template "sidejobs" (list $ $p) }}
	</div>
	{{- end }}
	{{- range $g.Projects }}{{ template "sideproject" (list $ . $total) }}{{ end }}
</details>
{{- else }}
{{- range $g.Projects }}{{ template "sideproject" (list $ . $total) }}{{ end }}
{{- end }}
{{- end }}

{{/* sidefolder is a folder inside a project in the sidebar: a disclosure
     of its jobs, open on its own page or when the tree is small. It
     names its path below its project. */}}
{{- define "sidefolder" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $total := index . 2 }}{{ $top := index . 3 }}
{{- $active := and $.Folder (eq $.Folder.ID $p.ID) }}
<details class="folder{{ if $active }} active{{ end }}" data-folder="{{ $p.ID }}"{{ if or $active (le $total 24) }} open{{ end }}>
	<summary title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
		<a href="{{ $.Href $p.ID }}"><svg class="ic"><use href="#i-folder{{ if $active }}-open{{ end }}"/></svg><span>{{ with folderDir $p.ID $top }}<i>{{ . }}/</i>{{ end }}<b>{{ $p.Name }}</b></span></a>
		<i class="dot"></i>
		{{- if $p.Error }}<small class="bad">!</small>{{ else }}<small>{{ len $p.Jobs }}</small>{{ end }}
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
		{{- template "sidejobs" (list $ $p) }}
	</div>
</details>
{{- end }}

{{/* tree is the Files panel of a folder page, a dashboard of what it
     holds: .. up a level, the directories first, linked into, then the
     files, each opening in the preview window. A directory that is a
     folder links to its page. A directory with jobs, a folder or one with
     a job source at its top, and a file that is a job source have Run and
     |>summarize in Jobs, with a menu of the targets (TreeMenus). A file
     that runs carries #! and exec, and Run in Jobs for those who may run
     it. A sealed entry shows plain and locked. */}}
{{- define "tree" }}
<div class="table-wrap">
<table class="table file-table col-hover">
	<thead><tr><th>Name</th><th>Action</th><th>Summary</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
	<tbody>
	{{- with $.Up }}
		<tr class="up"><td class="key"><svg class="ic dir"><use href="#i-folder"/></svg> <a href="{{ . }}" title="Up a level">..</a></td><td></td><td></td><td></td><td class="hide-sm"></td><td></td></tr>
	{{- end }}
	{{- range .Tree }}
		{{- if .Sealed }}
		<tr class="sealed" aria-disabled="true" title="task-ui lists {{ .Name }} but does not open it">
			<td class="key"><svg class="ic"><use href="#i-secret"/></svg> <span>{{ .Name }}</span></td>
			<td></td>
			<td></td>
			<td class="num nowrap">{{ if not .Dir }}{{ size .Size }}{{ end }}</td>
			<td class="hide-sm nowrap" title="{{ stamp .ModTime }}">{{ ago .ModTime }}</td>
			<td></td>
		</tr>
		{{- else }}
		<tr class="{{ if .Dir }}dir{{ else }}file{{ end }}">
			<td class="key">
				{{- if .Exec }}<i class="shebang" aria-hidden="true">#!</i>{{ else }}<svg class="ic{{ if .Dir }} dir{{ end }}"><use href="#i-{{ if .Dir }}folder{{ else }}file{{ end }}"/></svg>{{ end }}
				{{- if .Dir }}{{ if .Href }} <a href="{{ .Href }}">{{ .Name }}</a>{{ else }} <span>{{ .Name }}</span>{{ end }}
				{{- else if .Href }} <a href="{{ .Href }}" data-file="{{ .Path }}">{{ .Name }}</a>
				{{- else }} <span>{{ .Name }}</span>{{ end }}
				{{- if .Exec }} <span class="badge exec" title="Runs: executable, or a shebang on its first line">exec</span>{{ end }}</td>
			{{- $menu := index $.TreeMenus .Name }}{{ $exec := false }}{{ if and (not $menu) (not .Dir) .Exec .Path }}{{ $exec = $.Folder.Job (print "./" .Path) }}{{ end }}
			<td class="jobs" data-label="Action">
				{{- if $menu }}{{ template "runmenus" (list $ $menu) }}
				{{- else if $exec }}{{ with $exec }}
				{{- $run := $.RunningRun .Key }}
				{{- if $run }}<a class="btn btn-sm" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}"><svg class="ic"><use href="#i-eye"/></svg>Attach</a>
				{{- else if $.CanRun $.Folder . }}<div class="actions">{{ template "runButton" (list $ $.Folder .) }}{{ template "forceButton" (list $ $.Folder .) }}</div>{{ end }}
				{{- end }}{{ end }}
			</td>
			<td class="jobs summary" data-label="Summary"><div class="actions">
				{{- if $menu }}{{ with $.SummaryMenu $menu }}{{ if $.CanRunMenu . }}{{ template "runsplit" (list $ . true) }}{{ end }}{{ end }}
				{{- else if $exec }}{{ with $exec }}{{ if and (not ($.RunningRun .Key)) ($.CanRun $.Folder .) }}{{ template "summarizeButton" (list $ $.Folder .) }}{{ end }}{{ end }}{{ end }}
				{{- with $.TreeSummary .Name }}<a class="btn btn-sm summary-link" href="{{ .Href }}" title="{{ .Preview }}"><svg class="ic"><use href="#i-file"/></svg>Summary</a>{{ end -}}
			</div></td>
			<td class="num nowrap">{{ if not .Dir }}{{ size .Size }}{{ end }}</td>
			<td class="hide-sm nowrap" title="{{ stamp .ModTime }}">{{ ago .ModTime }}</td>
			<td class="actions">
				{{- if and (not .Dir) .Href }}
				<a class="btn-icon" href="{{ .Href }}?download" title="Download {{ .Name }}" aria-label="Download {{ .Name }}"><svg class="ic"><use href="#i-download"/></svg></a>
				{{- end }}
			</td>
		</tr>
		{{- end }}
	{{- else }}
		<tr class="empty"><td colspan="6">This directory is empty.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{- end }}

{{/* spacebar is the space a directory takes: its files, and the git
     history beside them, as one bar with a legend; the bar leaves out
     what git ignores, and the legend names it apart. */}}
{{- define "spacebar" }}
<div class="spacebar" role="img" aria-label="{{ size .Total }} in all: {{ size .Files }} of files, {{ size .History }} of history">
	<div class="spacebar-track"><i class="files" style="width: {{ printf "%.2f" .FilesPercent }}%"></i><i class="history" style="width: {{ printf "%.2f" .HistoryPercent }}%"></i></div>
	<div class="spacebar-legend">
		<span><i class="swatch files"></i>Files <b>{{ size .Files }}</b></span>
		<span><i class="swatch history"></i>History <b>{{ size .History }}</b></span>
		<span class="faint">{{ size .Total }} in all{{ if .Ignored }}, and {{ size .Ignored }} git ignores{{ end }}</span>
	</div>
</div>
{{- end }}

{{/* spacemini is the same in a table cell: a short bar and the total. */}}
{{- define "spacemini" }}
<span class="spacemini" title="{{ size .Files }} of files, {{ size .History }} of history{{ if .Ignored }}; {{ size .Ignored }} more that git ignores{{ end }}"><span class="spacebar-track"><i class="files" style="width: {{ printf "%.2f" .FilesPercent }}%"></i><i class="history" style="width: {{ printf "%.2f" .HistoryPercent }}%"></i></span><span class="nowrap">{{ size .Total }}</span></span>
{{- end }}

{{/* showsource is the Show button of a file a folder's pipeline reads:
     it opens the file, coloured, in the source window. */}}
{{- define "showsource" }}
{{- $folder := index . 0 }}{{ $file := index . 1 }}{{ $lines := $folder.Lines $file -}}
<button type="button" class="btn btn-sm show-source" data-source="{{ $folder.ID }}" data-file="{{ $file }}" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span>Show {{ $lines }} line{{ if ne $lines 1 }}s{{ end }}</span></button>
{{- end }}

{{/* sources are the job cards of a folder, a section per source: the
     folder's own, or those of a directory inside it (Folder.Sub), whose
     jobs run there and have their page under it. */}}
{{- define "sources" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}
{{- range $s := $p.Sources }}
<section class="block" data-live>
	<h3>{{ if eq (len $p.Sources) 1 }}Jobs{{ else }}{{ $s.Kind }}{{ end }} <small>{{ $s.File }}{{ if $s.Command }} · {{ join $s.Command " " }}{{ end }}</small></h3>
	<div class="jobcards">
	{{- range $r := $p.SourceJobs $s }}
		{{- $run := $.RunningRun $r.Key }}{{ $can := $.CanRun $p $r }}
		{{- /* A card is four rows the cards of a grid row share: the name, its kind
		     and Run; what it does, unless that repeats the name or the first
		     step; how it last went; what it runs, numbered, the first three
		     lines and the rest behind "show N more". */}}
		<article class="jobcard{{ if $run }} running{{ end }}">
			<header>
				<a class="name" href="{{ $.JobHref $p $r }}">{{ $r.Name }}</a>
				<span class="badge">{{ $r.Kind }}</span>
				{{- if $r.Interactive }}<span class="badge">interactive</span>{{ end }}
				<span class="actions">
					{{- if $run }}
					<a class="btn btn-sm" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}"><svg class="ic"><use href="#i-eye"/></svg>Attach</a>
					{{- if $can }}<button type="button" class="btn-icon danger" data-kill="{{ $run.ID }}" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-stop"/></svg></button>{{ end }}
					{{- else if $can }}{{ template "runButtons" (list $ $p $r) }}
					{{- else }}<a class="btn btn-sm" href="{{ $.JobHref $p $r }}">History</a>{{ end -}}
				</span>
			</header>
			<p class="desc">{{ if and $r.Description (ne $r.Description $r.Name) (not (and $r.Steps (eq $r.Description (index $r.Steps 0)))) }}{{ $r.Description }}{{ end }}</p>
			<div class="meta">
				{{- with $run }}<i class="dot running"></i><span class="state-label">running</span><span class="when">{{ ago .StartedAt }}</span><span class="who-ran">{{ template "user" (list $ .UserName) }}</span>
				{{- else }}{{ with $.LastRun $r.Key }}<i class="dot {{ .State }}"></i><span class="state-label">{{ .State }}{{ if ne .State "ok" }} <span class="faint">({{ .ExitCode }})</span>{{ end }}</span><span class="when">{{ ago .StartedAt }}</span><span class="who-ran">{{ template "user" (list $ .UserName) }}</span>
				{{- else }}<i class="dot pending"></i><span class="state-label faint">never run</span>{{ end }}{{ end -}}
			</div>
			<div class="code">
				{{- if $r.DependsOn }}<div class="comment"># after {{ join $r.DependsOn ", " }}</div>{{ end }}
				<ol>
					{{- range $i, $step := $r.Steps }}{{ if lt $i 3 }}<li>{{ $step }}</li>{{ end }}{{ else }}<li>{{ join $r.Command " " }}</li>{{ end -}}
				</ol>
				{{- if gt (len $r.Steps) 3 }}
				<details>
					<summary><span class="more">show {{ sub (len $r.Steps) 3 }} more</span><span class="less">show less</span></summary>
					<ol start="4">
						{{- range $i, $step := $r.Steps }}{{ if ge $i 3 }}<li>{{ $step }}</li>{{ end }}{{ end -}}
					</ol>
				</details>
				{{- end }}
			</div>
		</article>
	{{- else }}
		<div class="alert">{{ if $s.Error }}{{ $s.Error }}{{ else }}No targets in {{ $s.File }}.{{ end }}</div>
	{{- end }}
	</div>
</section>
{{- end }}
{{- end }}

{{- /* runmenus holds Run of a Files list row's Action column, from its
     RunMenu, and Run force beside it for a default with hours; for a
     user who may run none of the jobs, the default's History. The
     Summary column holds |>summarize, from Page.SummaryMenu. (list $
     menu) */}}
{{- define "runmenus" }}
{{- $ := index . 0 }}{{ $m := index . 1 }}{{ $f := $m.Folder }}
{{- if $.CanRunMenu $m }}
<div class="actions">
	{{- template "runsplit" (list $ $m false) }}
	{{- with $m.Main }}{{ if and ($f.HoursOf .) ($.CanRun $f .) (not ($.RunningRun .Key)) }}
	<button type="button" class="btn btn-danger btn-sm" data-run data-force data-folder="{{ $f.ID }}"{{ with .Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ .ID }}" data-title="{{ $f.ID }}{{ with .Dir }}/{{ . }}{{ end }} / {{ .Name }}" title="Run out of hours, with a reason">Run force</button>
	{{- end }}{{ end }}
</div>
{{- else }}{{ with $m.Main }}
<a class="btn btn-sm" href="{{ $.JobHref $f . }}" title="No permission to run {{ .ID }}"><i class="dot {{ $.LastState . }}"></i>History</a>
{{- end }}{{ end }}
{{- end }}

{{- /* runsplit holds Run, or |>summarize with summarize, as a split
     button: the button runs the menu's Main, an eye on it while it runs,
     and the dot on Run is how it last ended; the chevron beside it opens
     the menu, a table of the targets, each row runs one, its play glyph
     in the colour of its last run. With no Main the button opens the
     menu. (list $ menu summarize) */}}
{{- define "runsplit" }}
{{- $ := index . 0 }}{{ $m := index . 1 }}{{ $sum := index . 2 }}{{ $f := $m.Folder }}{{ $main := $m.Main }}
{{- $shown := false }}
<span class="runsplit"{{ if and $main (not $sum) }} data-key="{{ $main.Key }}"{{ end }}>
	{{- with $main }}{{ $run := $.RunningRun .Key }}
	{{- if $run }}{{ if not $sum }}{{ $shown = true }}<a class="btn btn-sm" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}" title="Attach to {{ .ID }}, running"><i class="dot running"></i>Attach</a>{{ end }}
	{{- else if $.CanRun $f . }}{{ $shown = true }}<button type="button" class="btn btn-sm{{ if not $sum }} btn-primary{{ end }}" data-run{{ if $sum }} data-summarize{{ end }} data-folder="{{ $f.ID }}"{{ with .Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ .ID }}" data-title="{{ $f.ID }}{{ with .Dir }}/{{ . }}{{ end }} / {{ .Name }}"{{ with .ArgsCommand }} data-args="{{ . }}"{{ end }} title="Run {{ .ID }}{{ if $sum }} and summarize{{ end }}">
		{{- if $sum }}|&gt;summarize{{ else }}<svg class="ic"><use href="#i-play"/></svg>Run{{ if .TakesArgs }}...{{ end }}<i class="dot {{ $.LastState . }}"></i>{{ end }}</button>
	{{- end }}{{ end }}
	{{- if $m.Menu }}
	<details class="cc-more run-menu">
		<summary class="btn btn-sm{{ if not $sum }} btn-primary{{ end }}" title="Pick the target to run{{ if $sum }} and summarize{{ end }}" aria-label="Pick the target to run{{ if $sum }} and summarize{{ end }}">
			{{- if not $shown }}{{ if $sum }}|&gt;summarize{{ else }}<svg class="ic"><use href="#i-play"/></svg>Run{{ end }}{{ end }}<svg class="ic chev"><use href="#i-chev"/></svg></summary>
		<div class="cc-menu" role="menu">
			<table class="menu-table"><tbody>
			{{- range $m.Jobs }}{{ $run := $.RunningRun .Key }}
				{{- if $run }}
				<tr role="menuitem" tabindex="0" data-attach="{{ $run.ID }}" title="Attach to {{ .ID }}, running"><td class="glyph"><svg class="ic"><use href="#i-eye"/></svg></td><td class="name">{{ .ID }}</td><td class="desc">{{ $m.Describe . }}</td></tr>
				{{- else if $.CanRun $f . }}
				<tr role="menuitem" tabindex="0" data-run{{ if $sum }} data-summarize{{ end }} data-folder="{{ $f.ID }}"{{ with .Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ .ID }}" data-title="{{ $f.ID }}{{ with .Dir }}/{{ . }}{{ end }} / {{ .Name }}"{{ with .ArgsCommand }} data-args="{{ . }}"{{ end }}><td class="glyph {{ $.LastState . }}" title="Last run: {{ $.LastState . }}"><svg class="ic"><use href="#i-play"/></svg></td><td class="name">{{ .ID }}</td><td class="desc">{{ $m.Describe . }}</td></tr>
				{{- else }}
				<tr class="idle" aria-disabled="true" title="No permission to run {{ .ID }}"><td class="glyph"><i class="shebang" aria-hidden="true">#!</i></td><td class="name">{{ .ID }}</td><td class="desc">{{ $m.Describe . }}</td></tr>
				{{- end }}
			{{- end }}
			</tbody></table>
		</div>
	</details>
	{{- end }}
</span>
{{- end }}

{{- /* runButtons hold Run, |>summarize, and Run force for a job with
     hours. Run of a job that takes args opens a field for them under its
     card or head first, with the command they follow. |>summarize runs
     the job the same way, and Claude summarises the run on its page when
     it ends. */}}
{{ define "runButtons" }}
{{- template "runButton" . }}{{ template "summarizeButton" . }}{{ template "forceButton" . }}
{{- end }}

{{- /* runButton holds Run of a job. (list $ folder job) */}}
{{ define "runButton" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $r := index . 2 -}}
<button type="button" class="btn btn-primary btn-sm" data-run data-folder="{{ $p.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $p.ID }}{{ with $r.Dir }}/{{ . }}{{ end }} / {{ $r.Name }}"{{ with $r.ArgsCommand }} data-args="{{ . }}" title="Run with arguments after {{ . }}"{{ end }}><svg class="ic"><use href="#i-play"/></svg>Run{{ if $r.TakesArgs }}...{{ end }}</button>
{{- end }}

{{- /* summarizeButton is |>summarize of a job Claude summarizes for the
     user, Page.Summarizes. (list $ folder job) */}}
{{ define "summarizeButton" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $r := index . 2 -}}
{{- if $.Summarizes $r }}
<button type="button" class="btn btn-sm" data-run data-summarize data-folder="{{ $p.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $p.ID }}{{ with $r.Dir }}/{{ . }}{{ end }} / {{ $r.Name }}"{{ with $r.ArgsCommand }} data-args="{{ . }}"{{ end }} title="Run and summarize">|&gt;summarize</button>
{{- end }}
{{- end }}

{{- /* forceButton holds Run force, beside Run only for a job with hours,
     where task-ui refuses a plain run outside them. To force a run that
     failed, use its window. (list $ folder job) */}}
{{ define "forceButton" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $r := index . 2 -}}
{{- if $p.HoursOf $r }}
<button type="button" class="btn btn-danger btn-sm" data-run data-force data-folder="{{ $p.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $p.ID }}{{ with $r.Dir }}/{{ . }}{{ end }} / {{ $r.Name }}" title="Run out of hours, with a reason">Run force</button>
{{- end }}
{{- end }}

{{- define "sidejobs" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}
{{- range $r := $p.Jobs }}
{{- $run := $.RunningRun $r.Key }}
<div class="job{{ if and $.Job (eq $.Job.Key $r.Key) }} active{{ end }}" data-job-key="{{ $r.Key }}">
	<i class="shebang{{ if $run }} running{{ end }}" aria-hidden="true">#!</i>
	<a href="{{ $.Href $p.ID }}/{{ $r.ID }}">{{ if ne $r.ID $r.Name }}<i>{{ index (split $r.ID ":") 0 }}:</i>{{ end }}{{ $r.Name }}</a>
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

{{- /* keyname is the name field of a key form: the name before the
     algorithm, which task-ui adds. */}}
{{- define "keyname" }}
<div class="field">
	<label for="key-name">Name <span class="faint">lower case letters, digits, dashes and underscores; the algorithm follows, as deploy_ed25519</span></label>
	<input type="text" id="key-name" name="name" value="{{ . }}" required pattern="[a-z0-9][a-z0-9_\-]{0,63}" spellcheck="false" autocomplete="off">
</div>
{{- end }}

{{- /* keypurpose is the purpose field of a key form. */}}
{{- define "keypurpose" }}
<div class="field">
	<span class="label">Purpose</span>
	<div class="checks">
		<label class="check"><input type="radio" name="purpose" value="ssh" checked> SSH key <span class="faint">runs, git and logins authenticate with it</span></label>
		<label class="check"><input type="radio" name="purpose" value="encrypt"> Encryption key <span class="faint">ed25519 or RSA; encrypts the secrets, never handed out</span></label>
	</div>
</div>
{{- end }}

{{- /* whopicker is a "who may run" field as boxes: the roles that run,
     then the users, each a value of the set the name holds. */}}
{{- define "whopicker" }}
{{- $name := index . 0 }}{{ $choices := index . 1 }}
<div class="checks who">
	<span class="faint">Roles</span>
	{{- range $choices }}{{ if .Role }}
	<label class="check"><input type="checkbox" name="{{ $name }}" value="{{ .Name }}" data-type="set"{{ if .Checked }} checked{{ end }}> {{ .Name }}</label>
	{{- end }}{{ end }}
</div>
<div class="checks who">
	<span class="faint">Users</span>
	{{- range $choices }}{{ if not .Role }}
	<label class="check"><input type="checkbox" name="{{ $name }}" value="{{ .Name }}" data-type="set"{{ if .Checked }} checked{{ end }}> {{ .Name }}</label>
	{{- end }}{{ end }}
</div>
{{- end }}

{{- /* workspacedelete removes a workspace: the workspace, the folders
     of the projects inside it that go with it, and where to go after. */}}
{{- define "workspacedelete" }}
{{- $ws := index . 0 }}{{ $nested := index . 1 }}
<button type="button" class="btn-icon danger" data-delete="/api/workspaces/{{ $ws.Folder }}" data-confirm="Remove the workspace {{ $ws.Folder }} with its checkout, keys and secrets{{ with $nested }}, and the projects inside it: {{ join . ", " }}{{ end }}? You can clone it again from {{ $ws.SourceURL }}." data-confirm-ok="Remove" data-next="{{ index . 2 }}" title="Remove the workspace" aria-label="Remove {{ $ws.Folder }}"><svg class="ic"><use href="#i-trash"/></svg></button>
{{- end }}

{{- define "keypolicies" }}
<div class="field">
	<span class="label">Who gets it <span class="faint">their runs get the key through the ssh agent; check any roles and users</span></span>
	{{- template "keychecks" . }}
</div>
{{- end }}

{{- /* keygrants is who gets a workspace's SSH key: the roles and the
     users checked, or every run in the workspace with none checked. */}}
{{- define "keygrants" }}
<div class="field">
	<span class="label">Who gets it <span class="faint">with nothing checked, every run in the workspace gets it; check roles or users to limit it to their runs</span></span>
	{{- template "keychecks" . }}
</div>
{{- end }}

{{- /* keychecks are a key's roles and users as boxes: the roles and the
     users it goes to, and every user to choose from. */}}
{{- define "keychecks" }}
{{- $roles := index . 0 }}{{ $users := index . 1 }}{{ $all := index . 2 }}
	<div class="checks who">
		<span class="faint">Roles</span>
		{{- range roles }}
		{{- $role := . }}{{ $on := false }}{{ range $roles }}{{ if eq . $role }}{{ $on = true }}{{ end }}{{ end }}
		<label class="check"><input type="checkbox" name="policies" value="{{ $role }}" data-type="set"{{ if $on }} checked{{ end }}> {{ $role }}</label>
		{{- end }}
	</div>
	<div class="checks who">
		<span class="faint">Users</span>
		{{- range $all }}
		{{- $name := .Name }}{{ $on := false }}{{ range $users }}{{ if eq . $name }}{{ $on = true }}{{ end }}{{ end }}
		<label class="check"><input type="checkbox" name="users" value="{{ $name }}" data-type="set"{{ if $on }} checked{{ end }}> {{ $name }} <span class="faint">{{ .Role }}</span></label>
		{{- end }}
	</div>
{{- end }}

{{- /* When runs ran: a bar per day and per hour of the last 24, each split
     into the runs that ended ok, failed and the rest, and for a job or a
     workflow the duration of its latest runs. Every bar and every run
     on the line hovers its counts. */}}
{{- define "activity" }}
<div class="activity">
	<p class="activity-sum">{{ .Total }} run{{ if ne .Total 1 }}s{{ end }} in {{ len .Days }} days{{ if .Failed }} · <span class="bad">{{ .Failed }} failed</span>{{ end }}{{ if .Median }} · median {{ duration .Median }}{{ end }}<span class="activity-key"><i class="ok"></i>runs <i class="line avg"></i>average <i class="line max"></i>slowest <i class="line min"></i>fastest</span></p>
	<div class="activity-charts">
		<figure class="activity-chart">
			<figcaption class="activity-title"><span>Last {{ len .Days }} days</span>{{ template "activity-stats" .DayStats }}</figcaption>
			<div class="activity-bars" style="--n: {{ len .Days }}" role="img" aria-label="Runs per day, the last {{ len .Days }} days">
				{{- $last := sub (len .Days) 4 }}
				{{- range $i, $b := .Days }}{{ with $b }}
				<span class="tip{{ if lt $i 4 }} tip-start{{ else if ge $i $last }} tip-end{{ end }}" data-tip="{{ .Title }}" aria-label="{{ .Title }}">{{ if .OtherStyle }}<i class="other" style="{{ .OtherStyle }}"></i>{{ end }}{{ if .FailedStyle }}<i class="bad" style="{{ .FailedStyle }}"></i>{{ end }}{{ if .OKStyle }}<i class="ok" style="{{ .OKStyle }}"></i>{{ end }}</span>
				{{- end }}{{ end }}
			</div>
			{{- with $.DayLines }}{{ template "activity-lines" (list . $.Days) }}{{ end }}
			<div class="activity-axis">{{ range .DayTicks }}<span style="left: {{ .At }}%">{{ .Label }}</span>{{ end }}</div>
		</figure>
		<figure class="activity-chart hours">
			<figcaption class="activity-title"><span>Last 24 hours</span>{{ template "activity-stats" .HourStats }}</figcaption>
			<div class="activity-bars" style="--n: 24" role="img" aria-label="Runs per hour, the last 24 hours">
				{{- range $i, $b := .Hours }}{{ with $b }}
				<span class="tip{{ if lt $i 4 }} tip-start{{ else if ge $i 20 }} tip-end{{ end }}" data-tip="{{ .Title }}" aria-label="{{ .Title }}">{{ if .OtherStyle }}<i class="other" style="{{ .OtherStyle }}"></i>{{ end }}{{ if .FailedStyle }}<i class="bad" style="{{ .FailedStyle }}"></i>{{ end }}{{ if .OKStyle }}<i class="ok" style="{{ .OKStyle }}"></i>{{ end }}</span>
				{{- end }}{{ end }}
			</div>
			{{- with $.HourLines }}{{ template "activity-lines" (list . $.Hours) }}{{ end }}
			<div class="activity-axis">{{ range .HourTicks }}<span style="left: {{ .At }}%">{{ .Label }}</span>{{ end }}</div>
		</figure>
	</div>
	{{- with .Durations }}
	<figure class="activity-chart durations">
		<div class="activity-line">
			<svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="Duration of the latest {{ len .Points }} runs, the longest {{ duration .Longest }}">
				<path class="wash" d="{{ .Area }}"></path>
				{{- range .Points }}{{ if .Failed }}
				<line class="bad" x1="{{ .At }}" y1="0" x2="{{ .At }}" y2="100" vector-effect="non-scaling-stroke"></line>
				{{- end }}{{ end }}
				<path class="line" d="{{ .Line }}" vector-effect="non-scaling-stroke"></path>
			</svg>
			{{- $n := len .Points }}
			{{- range $i, $pt := .Points }}
			<span class="hit tip{{ if lt $i 4 }} tip-start{{ else if ge $i (sub $n 4) }} tip-end{{ end }}" style="left:{{ $pt.X }}%;width:{{ $pt.Width }}%" data-tip="{{ $pt.Title }}" aria-label="{{ $pt.Title }}"></span>
			{{- end }}
			<span class="activity-longest">{{ duration .Longest }}</span>
		</div>
		<figcaption><span>duration of the latest {{ len .Points }} runs</span></figcaption>
	</figure>
	{{- end }}
</div>
{{- end }}

{{/* pipenode is a step of the pipeline: its name, state and, for a run's
     step, its terminal's last screen, polled while it runs (data-frame).
     A step that ran a workflow holds that run's chain. */}}
{{- define "pipenode" }}
{{- $l := index . 0 }}{{ $i := index . 1 }}
<li class="node{{ with $l.State }} {{ . }}{{ end }}{{ if $l.Disabled }} disabled{{ end }}" data-step="{{ $i }}">
	<span class="rail"><i class="dot{{ with $l.State }} {{ . }}{{ end }}"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="{{ $l.Href }}"{{ if $l.RunID }} data-run-link="{{ $l.RunID }}"{{ end }}>{{ if $l.Workflow }}<svg class="ic"><use href="#i-flow"/></svg>{{ else if $l.Prompt }}<svg class="ic"><use href="#i-prompt"/></svg>{{ end }}{{ $l.Text }}</a>
			{{- if and $l.RunID (not $l.Complete) }}<button type="button" class="btn btn-sm" data-expand="{{ $l.RunID }}" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand{{ with $l.Lines }} {{ . }} lines{{ end }}</span></button>{{ end }}
		</div>
		<small>{{ if $l.State }}{{ $l.State }}{{ else if $l.Disabled }}step {{ add $i 1 }}, off: runs skip it{{ else }}step {{ add $i 1 }}{{ end }}</small>
		{{- with $l.Prompt }}
		<p class="step-prompt">{{ . }}</p>
		{{- end }}
		{{- if $l.RunID }}<a class="frame" href="/runs/{{ $l.RunID }}" data-attach="{{ $l.RunID }}"{{ if eq $l.State "running" }} data-frame="{{ $l.RunID }}"{{ end }} role="img" aria-label="Terminal of {{ $l.Text }}">{{ $l.FrameSVG }}</a>{{ end }}
		{{- if $l.WorkflowRunID }}
		<ol class="pipeline sub" data-sub="{{ $l.WorkflowRunID }}" data-workflow="{{ $l.WorkflowID }}">
			{{- range $j, $s := $l.Sub }}{{ template "pipenode" (list $s $j) }}{{ end }}
		</ol>
		{{- end }}
	</div>
</li>
{{- end }}

{{- /* activity-stats are a chart's finished runs in one line: the
     fastest, the mean, the slowest and their time together. */}}
{{- define "activity-stats" }}{{ if .Count }}<span class="faint" title="{{ .Count }} finished run{{ if ne .Count 1 }}s{{ end }}: the fastest, the mean, the slowest and their time together">min/avg/max/total: {{ duration .Fastest }} {{ duration .Mean }} {{ duration .Slowest }} {{ duration .Spent }}</span>{{ end }}{{ end }}

{{- /* The durations under a count chart: the average, the slowest and the
     fastest run of each bucket as lines, a point in the middle of its
     column; each column hovers its numbers. */}}
{{- define "activity-lines" }}
{{- $l := index . 0 }}{{ $bars := index . 1 }}{{ $n := len $bars }}{{ $last := sub $n 4 }}
<div class="activity-line durations" style="--n: {{ $n }}">
	<svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="Average, slowest and fastest run per column, the slowest {{ duration $l.Slowest }}">
		<path class="max" d="{{ $l.Max }}" vector-effect="non-scaling-stroke"></path>
		<path class="min" d="{{ $l.Min }}" vector-effect="non-scaling-stroke"></path>
		<path class="avg" d="{{ $l.Average }}" vector-effect="non-scaling-stroke"></path>
	</svg>
	<div class="activity-hits">
		{{- range $i, $b := $bars }}
		<span class="tip{{ if lt $i 4 }} tip-start{{ else if ge $i $last }} tip-end{{ end }}" data-tip="{{ $b.AverageTitle }}" aria-label="{{ $b.AverageTitle }}"></span>
		{{- end }}
	</div>
</div>
{{- end }}
