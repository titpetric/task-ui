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
	<link rel="stylesheet" href="/static/voice/voice.css?v={{ assetVersion }}">
</head>
<body data-roles='{{ json .Roles }}' data-workspaces='{{ json .Workspaces }}'{{ with .User }} data-me="{{ .Name }}"{{ end }} data-user-menu="{{ if .Can "user/edit" }}edit{{ else if .Can "user/view" }}view{{ end }}">
{{ template "icons" . }}
{{- if not .User }}
	<header class="top login-top">
		<h1><a href="/"><span>TaskUI</span>{{ template "logo" }}</a></h1>
	</header>
	{{ template "content" . }}
{{- else }}
	<a class="skip-link" href="#main">Skip to content</a>
	{{- if .Can "prompt/edit" }}
	{{- $ask := .HeaderPrompt }}
	<header class="top" data-prompt-folder="{{ $ask.Folder }}" data-prompt-dir="{{ $ask.Dir }}" data-prompt-run="{{ $ask.Run }}" data-prompt-label="{{ $ask.Label }}" data-prompt-href="{{ $ask.Href }}">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="/"><span>TaskUI</span></a><button type="button" class="logo-toggle" id="header-toggle" aria-label="Prompt" aria-controls="header-answer" aria-expanded="false">{{ template "logo" }}</button></h1>
	{{- else }}
	<header class="top">
		<button type="button" class="btn-icon menu-toggle" id="menu-toggle" aria-label="Menu" aria-controls="side" aria-expanded="false"><svg class="ic"><use href="#i-menu"/></svg></button>
		<h1><a href="/"><span>TaskUI</span>{{ template "logo" }}</a></h1>
	{{- end }}
		{{- if .Can "prompt/edit" }}
		<input type="text" class="header-prompt" id="header-prompt" aria-label="Ask Claude" autocomplete="off" spellcheck="false">
		{{- if .PromptUsers }}
		<span class="header-runas"><select class="header-prompt-user" id="header-prompt-user" aria-label="Runs as">{{ template "user-options" (list .PromptUsers .User.Name .UserKeys true) }}
		</select></span>
		{{- end }}
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
				{{- range .SideItems }}
				<a href="{{ .Href }}" class="{{ if .Active }}active{{ end }}"><svg class="ic"><use href="#i-{{ .Icon }}"/></svg><span>{{ .Title }}</span>{{ if .Pulse }}<i class="dot{{ if $.Running }} running{{ end }}" data-any-running></i>{{ end }}</a>
				{{- if and .Flows $.SideWorkflows }}
				<div class="flows">
					{{- range $.SideWorkflows }}{{ if not .SidebarHidden }}
					<div class="flow{{ if $.IsPath (print "/workflows/" .ID) }} active{{ end }}">
						<i class="shebang{{ if .Running }} running{{ end }}" aria-hidden="true">#!</i>
						<a class="name" href="/workflows/{{ .ID }}" title="{{ range $i, $l := $.Links .Steps }}{{ if $i }} -&gt; {{ end }}{{ $l.Text }}{{ end }}">{{ .Name }}</a>
						{{- if .Running }}
						<a class="play" href="/workflows/{{ .ID }}" title="{{ .Name }}, running" aria-label="Watch {{ .Name }}"><svg class="ic"><use href="#i-eye"/></svg></a>
						{{- else if .CanRun }}
						<button type="button" class="play" data-workflow-run="{{ .ID }}"{{ if .Prompted }} data-workflow-prompt="{{ .Name }}"{{ end }} title="Run {{ .Name }}" aria-label="Run {{ .Name }}"><svg class="ic"><use href="#i-play"/></svg></button>
						{{- end }}
					</div>
					{{- end }}{{ end }}
				</div>
				{{- end }}
				{{- end }}
			</nav>
			<nav class="folders">
				{{- $total := 0 }}{{ range .SideFolders }}{{ $total = len .Jobs | add $total }}{{ end }}
				{{- range $g := .Groups }}
				{{- if $g.Workspace }}{{ template "sideworkspace" (list $ $g $total) }}{{ else }}{{ template "sideproject" (list $ $g $total) }}{{ end }}
				{{- else }}
				<div class="empty">No folders yet. {{ if $.Can "project/edit" }}<a href="{{ addProjectHref "" }}">Add a git repository</a>{{ else }}Add a git repository in Settings{{ end }}, or put a folder with jobs under the root.</div>
				{{- end }}
			</nav>
		</aside>
		<main id="main" tabindex="-1">
			{{- if .Crumbs }}
			<div class="crumbs"><a href="/">Dashboard</a>{{ range .Crumbs }}<span>{{ if .Href }}<a href="{{ .Href }}">{{ .Text }}</a>{{ else }}{{ .Text }}{{ end }}</span>{{ end }}</div>
			{{- end }}
			{{- if .Notice }}<div class="alert ok" role="status">{{ .Notice }}</div>{{ end }}
			{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}
			{{ template "content" . }}
			{{- template "setup-trail" .SetupTrail }}
		</main>
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
						<span class="what"><b>{{ .Name }}</b><small><svg class="ic lead"><use href="#i-flow"/></svg>workflow</small></span>
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
					<a href="/runs" class="recent more" data-recents-more hidden><svg class="ic"><use href="#i-list"/></svg><span class="what"><b>All runs</b><small></small></span></a>
				</nav>
			</div>
		</aside>
	</div>
	{{ template "dialog" . }}
	{{ template "confirm" . }}
	{{ template "voice" . }}
{{- end }}
	<div class="toasts" id="toasts" role="status" aria-live="polite"></div>
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
	<script src="/static/voice/voice.js?v={{ assetVersion }}" defer></script>
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
	<symbol id="i-summary" viewBox="0 0 16 16"><path d="M11.5 2.5h-7l4 5.5-4 5.5h7" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-grip" viewBox="0 0 16 16"><circle cx="6" cy="3.5" r="1.1" fill="currentColor"/><circle cx="10" cy="3.5" r="1.1" fill="currentColor"/><circle cx="6" cy="8" r="1.1" fill="currentColor"/><circle cx="10" cy="8" r="1.1" fill="currentColor"/><circle cx="6" cy="12.5" r="1.1" fill="currentColor"/><circle cx="10" cy="12.5" r="1.1" fill="currentColor"/></symbol>
	<symbol id="i-plus" viewBox="0 0 16 16"><path d="M8 3v10M3 8h10" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-server" viewBox="0 0 16 16"><rect x="2.5" y="2.5" width="11" height="4.5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><rect x="2.5" y="9" width="11" height="4.5" rx="1" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="5" cy="4.75" r=".8" fill="currentColor"/><circle cx="5" cy="11.25" r=".8" fill="currentColor"/></symbol>
	<symbol id="i-plug" viewBox="0 0 16 16"><path d="M6 1.5v3M10 1.5v3M4 4.5h8v3a4 4 0 0 1-8 0zM8 11.5v3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-note" viewBox="0 0 16 16"><path d="M3 2.5h10v7.5l-3.5 3.5H3zM13 10H9.5v3.5M5.5 5.5h5M5.5 8h3" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linejoin="round" stroke-linecap="round"/></symbol>
	<symbol id="i-clock" viewBox="0 0 16 16"><circle cx="8" cy="8" r="6" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M8 4.5V8l2.5 1.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-pulse" viewBox="0 0 16 16"><path d="M1.5 8h3L6 4.5 8.5 11.5 10.5 8h4" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
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
	<symbol id="i-script" viewBox="0 0 16 16"><path d="M5 4 1.8 8 5 12M11 4l3.2 4L11 12M9.4 2.5l-2.8 11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-terminal" viewBox="0 0 16 16"><rect x="1.5" y="2.5" width="13" height="11" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M4.5 6l2 2-2 2M8 10.5h3.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-play" viewBox="0 0 16 16"><path d="M4.5 2.5v11l8-5.5z" fill="currentColor"/></symbol>
	<symbol id="i-stop" viewBox="0 0 16 16"><rect x="3.5" y="3.5" width="9" height="9" rx="1" fill="currentColor"/></symbol>
	<symbol id="i-pin" viewBox="0 0 16 16"><path d="M6 1.5h4M7 1.5v4.2L4.5 8.5h7L9 5.7V1.5M8 8.5v6" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-eye" viewBox="0 0 16 16"><path d="M1.5 8s2.5-4.5 6.5-4.5S14.5 8 14.5 8 12 12.5 8 12.5 1.5 8 1.5 8z" fill="none" stroke="currentColor" stroke-width="1.2"/><circle cx="8" cy="8" r="2" fill="none" stroke="currentColor" stroke-width="1.2"/></symbol>
	<symbol id="i-menu" viewBox="0 0 16 16"><path d="M2.5 4h11M2.5 8h11M2.5 12h11" stroke="currentColor" stroke-width="1.3" stroke-linecap="round"/></symbol>
	<symbol id="i-expand" viewBox="0 0 16 16"><path d="M9.5 2.5h4v4M13.5 2.5 9 7M6.5 13.5h-4v-4M2.5 13.5 7 9" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-collapse" viewBox="0 0 16 16"><path d="M13 7H9V3M9 7l4.5-4.5M3 9h4v4M7 9l-4.5 4.5" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-more" viewBox="0 0 16 16"><circle cx="3.5" cy="8" r="1.3" fill="currentColor"/><circle cx="8" cy="8" r="1.3" fill="currentColor"/><circle cx="12.5" cy="8" r="1.3" fill="currentColor"/></symbol>
	<symbol id="i-chev" viewBox="0 0 16 16"><path d="M6 3.5 10.5 8 6 12.5" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-up" viewBox="0 0 16 16"><path d="M8 13.5v-11M3.5 7 8 2.5 12.5 7" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-down" viewBox="0 0 16 16"><path d="M8 2.5v11M3.5 9 8 13.5 12.5 9" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-download" viewBox="0 0 16 16"><path d="M8 2.5v8M4.5 7 8 10.5 11.5 7M2.5 13.5h11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-upload" viewBox="0 0 16 16"><path d="M8 10.5v-8M4.5 6 8 2.5 11.5 6M2.5 13.5h11" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></symbol>
	<symbol id="i-mic" viewBox="0 0 16 16"><rect x="6" y="1.6" width="4" height="7.3" rx="2" fill="none" stroke="currentColor" stroke-width="1.2"/><path d="M3.8 7.2v1a4.2 4.2 0 0 0 8.4 0v-1M8 13v1.4M5.6 14.4h4.8" fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/></symbol>
	<symbol id="i-send" viewBox="0 0 16 16"><path d="M2.5 8h10M8.5 4 12.5 8l-4 4" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/></symbol>
</svg>
{{ end }}

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
		<button type="button" class="btn" id="run-close" aria-label="Close">Close</button>
		<span class="hint" id="run-hint"></span>
		<button type="button" class="btn" id="run-attach" hidden>Attach</button>
		<button type="button" class="btn" id="run-again" hidden>Run again</button>
		<button type="button" class="btn" id="run-take" aria-pressed="false" title="Send your keystrokes to the run, to answer a prompt it waits on" hidden>Take input</button>
		<button type="button" class="btn-icon" id="run-mic" aria-haspopup="dialog" title="Dictate into the run" aria-label="Dictate into the run" hidden><svg class="ic"><use href="#i-mic"/></svg></button>
		<button type="button" class="btn" id="run-stop" hidden>Stop</button>
		<button type="button" class="btn btn-danger" id="run-kill" hidden>Kill</button>
		<button type="button" class="btn" id="run-pause" hidden>Pause</button>
		<button type="button" class="btn" id="run-restart" hidden>Restart</button>
		<button type="button" class="btn btn-primary" id="run-force" hidden>Run force</button>
	</div>
</dialog>
<dialog class="run preview" id="preview-dialog" aria-labelledby="preview-title">
	<div class="run-head">
		<span class="label" id="preview-label">Artifact</span>
		<h3 id="preview-title"></h3>
	</div>
	<div class="run-body preview-body"><img id="preview-image" alt=""><div id="preview-doc" class="preview-doc" hidden></div><iframe id="preview-frame" class="preview-frame" title="The page the run wrote" sandbox referrerpolicy="no-referrer" hidden></iframe></div>
	<div class="run-foot">
		<span class="hint"><select id="preview-version" aria-label="Version" hidden></select></span>
		<a class="btn" id="preview-download" href="#" download>Download</a>
		<span class="preview-actions" id="preview-actions"></span>
		<button type="button" class="btn" id="preview-close" aria-label="Close">Close</button>
	</div>
</dialog>
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
		<div class="form-foot">
			<button type="button" class="btn" data-workflow-input-cancel>Cancel</button>
			<button type="submit" class="btn btn-primary">Run</button>
		</div>
	</form>
</dialog>
<dialog class="confirm" id="confirm-dialog" aria-labelledby="confirm-title">
	<h3 id="confirm-title">Are you sure?</h3>
	<p id="confirm-text"></p>
	<div class="form-foot">
		<button type="button" class="btn" id="confirm-cancel">Cancel</button>
		<button type="button" class="btn btn-danger" id="confirm-ok">Delete</button>
	</div>
</dialog>
{{ end }}

{{ define "voice" }}
<button type="button" class="voice-orb" id="voice-orb" aria-haspopup="dialog" title="Dictate" aria-label="Dictate" hidden></button>
<dialog class="voice" id="voice-overlay" aria-label="Dictate">
	<svg class="voice-avatar" viewBox="0 0 320 320" aria-hidden="true" focusable="false">
		<defs>
			<filter id="voice-fog" x="-30%" y="-30%" width="160%" height="160%" color-interpolation-filters="sRGB">
				<feTurbulence type="fractalNoise" baseFrequency="0.012 0.02" numOctaves="4" seed="7" result="noise"/>
				<feDisplacementMap in="SourceGraphic" in2="noise" scale="12" xChannelSelector="R" yChannelSelector="G" result="bent"/>
				<feGaussianBlur in="bent" stdDeviation="2.4"/>
			</filter>
			<filter id="voice-haze" x="-40%" y="-40%" width="180%" height="180%" color-interpolation-filters="sRGB">
				<feTurbulence type="fractalNoise" baseFrequency="0.008 0.014" numOctaves="3" seed="19" result="n2"/>
				<feDisplacementMap in="SourceGraphic" in2="n2" scale="26" xChannelSelector="R" yChannelSelector="B" result="b2"/>
				<feGaussianBlur in="b2" stdDeviation="9"/>
			</filter>
			<radialGradient id="voice-core">
				<stop offset="0%" stop-color="var(--voice-hot)" stop-opacity="0.95"/>
				<stop offset="45%" stop-color="var(--voice-warm)" stop-opacity="0.45"/>
				<stop offset="100%" stop-color="var(--voice-cool)" stop-opacity="0"/>
			</radialGradient>
			<linearGradient id="voice-arm-a" x1="0" y1="0" x2="1" y2="1">
				<stop offset="0%" stop-color="var(--voice-warm)" stop-opacity="0"/>
				<stop offset="35%" stop-color="var(--voice-warm)" stop-opacity="0.75"/>
				<stop offset="100%" stop-color="var(--voice-cool)" stop-opacity="0"/>
			</linearGradient>
			<linearGradient id="voice-arm-b" x1="1" y1="0" x2="0" y2="1">
				<stop offset="0%" stop-color="var(--voice-cool)" stop-opacity="0"/>
				<stop offset="40%" stop-color="var(--voice-cool)" stop-opacity="0.7"/>
				<stop offset="100%" stop-color="var(--voice-violet)" stop-opacity="0"/>
			</linearGradient>
			<linearGradient id="voice-arm-c" x1="0" y1="1" x2="1" y2="0">
				<stop offset="0%" stop-color="var(--voice-violet)" stop-opacity="0"/>
				<stop offset="45%" stop-color="var(--voice-violet)" stop-opacity="0.6"/>
				<stop offset="100%" stop-color="var(--voice-hot)" stop-opacity="0"/>
			</linearGradient>
		</defs>
		<g class="voice-haze" filter="url(#voice-haze)">
			<ellipse cx="146" cy="150" rx="104" ry="86"/>
			<ellipse cx="176" cy="168" rx="88" ry="104"/>
			<ellipse cx="160" cy="142" rx="118" ry="70"/>
		</g>
		<circle class="voice-core" cx="160" cy="160" r="86" fill="url(#voice-core)"/>
		<g class="voice-arms" filter="url(#voice-fog)">
			<g class="voice-whorl voice-whorl-1">
				<path d="M166 160C166 161 167 163 166 164C166 166 165 167 163 168C162 169 159 170 157 170C155 170 152 169 150 168C148 166 146 164 145 161C144 158 144 154 145 151C146 147 148 143 152 140C155 138 161 135 166 135C171 135 178 136 183 140C188 143 194 149 196 156C199 163 200 173 198 181C195 190 190 200 182 207C174 213 161 220 149 221C136 221 119 219 107 211C94 203 80 189 73 173C66 156 68 122 67 112" stroke="url(#voice-arm-a)" stroke-width="9"/>
				<path d="M155 163C154 163 153 161 153 160C152 158 152 156 153 155C154 153 155 151 157 150C159 149 162 148 164 148C167 148 170 149 172 151C175 153 177 157 178 160C179 164 179 168 177 172C176 176 173 181 168 184C164 187 157 189 151 189C145 189 137 187 131 182C126 178 119 170 117 162C114 153 113 141 117 132C120 122 128 110 138 102C148 95 163 89 178 89C193 89 213 93 227 103C242 114 258 144 264 152" stroke="url(#voice-arm-c)" stroke-width="6"/>
				<path d="M163 155C164 155 166 156 167 157C168 158 169 160 169 162C169 163 168 166 167 168C166 169 164 171 162 172C160 173 156 174 153 173C150 173 147 171 144 168C142 166 139 161 139 157C138 153 139 147 141 142C144 138 149 133 154 130C160 128 168 126 175 127C182 128 191 132 197 138C204 145 210 155 211 165C213 175 212 189 207 200C201 212 190 225 177 232C164 238 145 244 127 241C109 239 87 231 72 216C56 202 41 164 35 153" stroke="url(#voice-arm-b)" stroke-width="4"/>
			</g>
			<g class="voice-whorl voice-whorl-2">
				<path d="M163 165C162 166 161 167 159 167C158 168 156 167 154 167C152 166 151 164 150 162C149 160 148 158 148 155C149 153 150 149 152 147C154 145 158 143 161 142C165 142 170 142 174 144C178 145 182 149 185 154C187 158 189 165 188 171C188 177 185 185 180 190C175 196 167 201 159 203C150 205 138 205 129 201C119 197 107 189 101 178C95 167 89 152 90 137C91 122 97 103 108 89C120 75 151 61 159 55" stroke="url(#voice-arm-b)" stroke-width="8"/>
				<path d="M155 157C155 156 156 154 157 153C158 152 160 151 161 151C163 151 166 151 167 153C169 154 171 156 172 158C173 160 174 164 173 167C173 170 171 173 168 176C166 178 162 181 157 181C153 182 147 181 143 179C138 176 133 172 130 166C128 161 126 153 127 145C128 138 132 129 138 123C144 117 154 110 164 109C175 107 189 107 200 113C211 119 224 129 231 142C238 155 244 175 242 192C239 210 232 232 217 248C202 263 165 279 154 285" stroke="url(#voice-arm-a)" stroke-width="5"/>
			</g>
			<g class="voice-whorl voice-whorl-3">
				<path d="M159 166C158 166 156 166 154 165C153 164 152 163 151 161C151 159 150 157 151 155C151 153 153 150 155 149C157 147 160 145 163 145C166 145 170 146 173 148C176 150 180 153 181 157C183 161 184 167 183 172C181 177 179 183 174 187C169 191 162 195 155 196C147 197 138 196 130 191C122 187 114 179 109 170C105 160 102 146 104 134C106 122 113 106 124 96C134 86 151 75 169 73C187 70 219 80 229 81" stroke="url(#voice-arm-c)" stroke-width="7"/>
				<path d="M158 154C158 154 160 153 162 153C163 153 165 153 167 154C168 155 170 157 170 159C171 161 172 164 171 166C170 169 169 172 166 174C164 176 160 177 157 178C153 178 148 177 145 175C141 172 137 168 135 164C133 159 132 152 133 146C135 140 138 132 144 128C149 123 158 118 166 117C175 116 187 118 196 123C205 128 216 138 221 149C226 160 230 177 227 191C224 206 216 224 203 237C190 249 158 260 149 264" stroke="url(#voice-arm-a)" stroke-width="4"/>
			</g>
		</g>
	</svg>
	<p class="voice-said" id="voice-said" role="status" aria-live="polite"></p>
	<p class="voice-where" id="voice-where"></p>
	<div class="voice-acts">
		<button type="button" class="voice-act discard" id="voice-discard" title="Discard" aria-label="Discard"><svg class="ic"><use href="#i-trash"/></svg></button>
		<button type="button" class="voice-act send" id="voice-send" title="Send" aria-label="Send" autofocus><svg class="ic"><use href="#i-send"/></svg></button>
	</div>
</dialog>
{{ end }}

{{ define "state" }}
{{- if . -}}
<span class="state" data-run-id="{{ .ID }}"{{ if ended .State }} data-finished="true"{{ end }}><i class="dot {{ .State }}"></i>{{ .State }}{{ if and (ne .State "ok") (ended .State) }} <span class="faint">({{ .ExitCode }})</span>{{ end }}</span>
{{- else -}}
<span class="faint">never run</span>
{{- end -}}
{{ end }}

{{ define "pager" }}{{ with . }}
<nav class="pager" aria-label="Pages">
	{{- if .Prev }}<a class="btn btn-sm" href="{{ .Prev }}" rel="prev">Previous</a>{{ end }}
	{{- if .Next }}<a class="btn btn-sm" href="{{ .Next }}" rel="next">Next</a>{{ end }}
</nav>
{{- end }}{{ end }}

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

{{ define "pagesize" }}
	<label class="page-size" for="page-size">Page size
		<select id="page-size" name="n">
			{{- range list "15" "50" "100" }}<option value="{{ . }}"{{ if eq (index $.Form "n") . }} selected{{ end }}>{{ . }}</option>{{ end }}
		</select>
	</label>
{{- end }}

{{ define "avatar" }}{{ $role := index (index . 0).Roles (index . 1) }}<span class="avatar sm{{ with $role }} role-{{ . }}{{ end }}" title="{{ index . 1 }}">{{ initials (index . 1) }}</span>{{ end }}

{{ define "who" }}{{ $p := index . 0 }}{{ with index . 1 }}<span class="who-ran">{{ template "user" (list $p .UserName) }}{{ if .IsScheduled }} <span class="badge">scheduled</span>{{ end }}{{ if .Reason }} <span class="badge bad" title="{{ .Reason }}">out of hours</span>{{ else if .IsForced }} <span class="badge signal">forced</span>{{ end }}</span>{{ end }}{{ end }}

{{ define "hours" }}{{ $text := index . 0 }}{{ $page := index . 1 }}{{ if $text }}{{ hoursLine $text }} {{ if $page.InHours $text }}<span class="badge ok">open now</span>{{ else }}<span class="badge bad">out of hours now</span>{{ end }}{{ else }}<span class="faint">always</span>{{ end }}{{ end }}

{{- define "hostjobs" }}{{ $job := "" }}{{ if gt (len .) 1 }}{{ $job = index . 1 }}{{ end }}{{ with index . 0 }}{{ if .HostJobs }}
			<optgroup label="Hosts">
				{{- range .HostJobs }}
				<option value="{{ .ID }}" data-folder="internal" title="{{ .Description }}"{{ if eq .ID $job }} selected{{ end }}>{{ .Name }}</option>
				{{- end }}
			</optgroup>
{{- end }}{{ end }}{{ end }}

{{- define "sideproject" }}
{{- $ := index . 0 }}{{ $g := index . 1 }}{{ $total := index . 2 }}
{{- with $p := $g.Top }}
				{{- $active := and $.Folder (or (eq $.Folder.ID $p.ID) (hasPrefix $.Folder.ID (print $p.ID "/"))) }}
				{{- $here := and $.Folder (eq $.Folder.ID $p.ID) }}
				<details class="project{{ if $active }} active{{ end }}" data-project="{{ $p.ID }}" open>
					<summary{{ if $here }} class="active"{{ end }} title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
						<a class="section{{ if $here }} active{{ end }}" href="{{ $.Href $p.ID }}" data-folder="{{ $p.ID }}">
							{{- template "sideicon" (list "project" ($.Dirty $p.ID)) }}<span>{{ $p.Name }}</span>
							{{- if $p.Error }}{{ template "sidebad" $p }}{{ else if $p.Jobs }}<small>{{ len $p.Jobs }}</small>{{ end }}
						</a>
						<svg class="ic chev"><use href="#i-chev"/></svg>
					</summary>
					{{- if or $p.Jobs ($.PinnedJobs $p.ID) }}
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
			{{- if $p.Error }}{{ template "sidebad" $p }}{{ else }}<small>{{ len $g.Projects }}</small>{{ end }}
		</a>
	</summary>
	{{- if or $p.Jobs ($.PinnedJobs $p.ID) }}
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

{{- define "sidefolder" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $total := index . 2 }}{{ $top := index . 3 }}
{{- $active := and $.Folder (eq $.Folder.ID $p.ID) }}
{{- $icon := "folder" }}{{ if $active }}{{ $icon = "folder-open" }}{{ end }}
<details class="folder{{ if $active }} active{{ end }}" data-folder="{{ $p.ID }}"{{ if or $active (le $total 24) }} open{{ end }}>
	<summary title="{{ $p.ID }}/{{ range $i, $s := $p.Sources }}{{ if $i }}, {{ end }}{{ $s.File }}{{ end }}">
		<a href="{{ $.Href $p.ID }}">{{ template "sideicon" (list $icon ($.Dirty $p.ID)) }}<span>{{ with folderDir $p.ID $top }}<i>{{ . }}/</i>{{ end }}<b>{{ $p.Name }}</b></span></a>
		<i class="dot"></i>
		{{- if $p.Error }}{{ template "sidebad" $p }}{{ else }}<small>{{ add (len $p.Jobs) (len ($.PinnedJobs $p.ID)) }}</small>{{ end }}
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
		{{- template "sidejobs" (list $ $p) }}
	</div>
</details>
{{- end }}

{{- define "sideicon" }}
{{- $dirty := index . 1 -}}
<svg class="ic{{ if $dirty }} dirty{{ end }}">{{ if $dirty }}<title>{{ $dirty }} file{{ if ne $dirty 1 }}s{{ end }} changed or untracked</title>{{ end }}<use href="#i-{{ index . 0 }}"/></svg>
{{- end }}

{{- define "tree" }}{{ template "filetable" (list $.Up $.Tree) }}{{ end }}

{{- define "filetable" }}
{{- $up := index . 0 }}{{ $items := index . 1 }}
<div class="table-wrap">
<table class="table file-table col-hover">
	<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
	<tbody>
	{{- with $up }}
		<tr class="up"><td class="key"><svg class="ic dir"><use href="#i-folder"/></svg> <a href="{{ . }}" title="Up a level">..</a></td><td></td><td class="hide-sm"></td><td></td></tr>
	{{- end }}
	{{- range $items }}
		{{- if .Sealed }}
		<tr class="sealed" aria-disabled="true" title="task-ui lists {{ .Name }} but does not open it">
			<td class="key"><svg class="ic"><use href="#i-secret"/></svg> <span>{{ .Name }}</span></td>
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
			<td class="num nowrap">{{ if not .Dir }}{{ size .Size }}{{ end }}</td>
			<td class="hide-sm nowrap" title="{{ stamp .ModTime }}">{{ ago .ModTime }}</td>
			<td class="actions">{{ if and (not .Dir) .Href }}{{ template "menu" (menu (print "Actions on " .Name)
				(detail "Changed" (stamp .ModTime))
				(item "eye" "Preview" "href" .Href "data-file" .Path)
				(item "download" "Download" "href" (print .Href "?download"))) }}{{ end }}</td>
		</tr>
		{{- end }}
	{{- else }}
		<tr class="empty"><td colspan="4">This directory is empty.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{- end }}

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

{{- define "spacemini" }}
<span class="spacemini" title="{{ size .Files }} of files, {{ size .History }} of history{{ if .Ignored }}; {{ size .Ignored }} more that git ignores{{ end }}"><span class="spacebar-track"><i class="files" style="width: {{ printf "%.2f" .FilesPercent }}%"></i><i class="history" style="width: {{ printf "%.2f" .HistoryPercent }}%"></i></span><span class="nowrap">{{ size .Total }}</span></span>
{{- end }}

{{- define "jobtable" }}
{{- if $.JobGroups }}
<section class="block" data-live>
	<h3>Jobs <small>a row per job source, then files, skills and terminals</small></h3>
	{{- template "controltable" (list $ $.JobControl true) }}
</section>
{{- end }}
{{- end }}

{{- define "showsource" }}
{{- $folder := index . 0 }}{{ $file := index . 1 }}{{ $lines := $folder.Lines $file -}}
<button type="button" class="btn btn-sm show-source" data-source="{{ $folder.ID }}" data-file="{{ $file }}" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span>Show {{ $lines }} line{{ if ne $lines 1 }}s{{ end }}</span></button>
{{- end }}

{{ define "runButtons" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $r := index . 2 -}}
{{- $summarize := $.Summarizes $r -}}
{{- $force := $p.Hours -}}
<div class="run-split">
	{{- template "runButton" . }}
	{{- if or $summarize $force }}
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run {{ $r.Name }}" aria-label="More ways to run {{ $r.Name }}"><svg class="ic"><use href="#i-down"/></svg></summary>
		<div class="more-list">
			<div class="more-items" role="menu" aria-label="Run options">
				{{- if $summarize }}
				<button type="button" class="more-item" role="menuitem" data-run data-summarize data-folder="{{ $p.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $p.ID }}{{ with $r.Dir }}/{{ . }}{{ end }} / {{ $r.Name }}"{{ with $r.ArgsCommand }} data-args="{{ . }}"{{ end }} title="Run and summarize"><svg class="ic"><use href="#i-summary"/></svg><span>Run and summarize</span></button>
				{{- end }}
				{{- if $force }}
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="{{ $p.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $p.ID }}{{ with $r.Dir }}/{{ . }}{{ end }}" title="Run out of hours, with a reason"><svg class="ic"><use href="#i-play"/></svg><span>Run force</span></button>
				{{- end }}
			</div>
		</div>
	</details>
	{{- end }}
</div>
{{- end }}

{{ define "runButton" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $r := index . 2 -}}
<button type="button" class="btn btn-primary btn-sm" data-run data-folder="{{ $p.ID }}"{{ with $r.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}" data-title="{{ $p.ID }}{{ with $r.Dir }}/{{ . }}{{ end }} / {{ $r.Name }}"{{ with $r.ArgsCommand }} data-args="{{ . }}" title="Run with arguments after {{ . }}"{{ end }}><svg class="ic"><use href="#i-play"/></svg>Run{{ if $r.TakesArgs }}...{{ end }}</button>
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
{{- else }}{{ if not ($.PinnedJobs $p.ID) }}
<div class="empty">{{ if $p.Errors }}The pipeline has errors.{{ else }}Nothing to run.{{ end }}</div>
{{- end }}{{ end }}
{{- range $.PinnedJobs $p.ID }}{{ template "sidepin" (list $ $p .) }}{{ end }}
{{- end }}

{{- define "sidepin" }}
{{- $ := index . 0 }}{{ $p := index . 1 }}{{ $s := index . 2 }}{{ $r := $s.Job }}
{{- $run := $.RunningRun $r.Key }}
<div class="job pinned{{ if and $.Job (eq $.Job.Key $r.Key) }} active{{ end }}" data-job-key="{{ $r.Key }}">
	<i class="shebang{{ if $run }} running{{ end }}" aria-hidden="true">#!</i>
	<a href="{{ $.JobHref $p $r }}" title="Pinned: {{ $p.ID }} / {{ $s.Label }}">{{ $s.Label }}</a>
	{{- if $run }}
	<button type="button" class="play" data-attach="{{ $run.ID }}" title="Attach to {{ $r.Name }}, running" aria-label="Attach to {{ $r.Name }}"><svg class="ic"><use href="#i-eye"/></svg></button>
	{{- else if $.CanRun $p $r }}
	<button type="button" class="play" data-run data-folder="{{ $p.ID }}"{{ with $s.Pin.Dir }} data-dir="{{ . }}"{{ end }} data-job="{{ $r.ID }}"{{ with $s.ArgsJSON }} data-run-args="{{ . }}"{{ end }} data-title="{{ $p.ID }} / {{ $s.Label }}" title="Run {{ $s.Label }}" aria-label="Run {{ $s.Label }}"><svg class="ic"><use href="#i-play"/></svg></button>
	{{- end }}
</div>
{{- end }}

{{- define "logo" }}<svg class="logo" viewBox="0 0 32 32" aria-hidden="true"><rect width="32" height="32" rx="3"/><path class="prompt" d="M8 10l6 6-6 6"/><path class="cursor" d="M16 23h8"/></svg>{{ end }}

{{- define "keyname" }}
<div class="field">
	<label for="key-name">Name</label>
	<input type="text" id="key-name" name="name" value="{{ . }}" required pattern="[a-z0-9][a-z0-9_\-]{0,63}" spellcheck="false" autocomplete="off">
	<small class="faint">Lower case letters, digits, dashes and underscores; the algorithm follows, as deploy_ed25519.</small>
</div>
{{- end }}

{{- define "keypurpose" }}
<div class="field" role="radiogroup" aria-labelledby="key-purpose-label">
	<span class="label" id="key-purpose-label">Purpose</span>
	<label class="check"><input type="radio" name="purpose" value="ssh" checked> SSH key</label>
	<small class="faint">Runs, git and logins authenticate with it.</small>
	<label class="check"><input type="radio" name="purpose" value="encrypt"> Encryption key</label>
	<small class="faint">ed25519 or RSA; encrypts the secrets, never handed out.</small>
</div>
{{- end }}

{{- define "whofield" }}
<div class="whofield{{ if .Compact }} compact{{ end }}" data-who-empty="{{ .Empty }}">
	<div class="checks who-roles">
		<span class="faint">Roles</span>
		{{- range roles }}
		<label class="check"><input type="checkbox" name="{{ $.Name }}" value="{{ . }}" data-type="{{ $.Type }}"{{ if $.RoleChecked . }} checked{{ end }}> {{ . }} <small class="count">· {{ $.RoleCount . }}</small></label>
		{{- end }}
	</div>
	<div class="checks who-users">
		<span class="faint">Users</span>
		<ul class="who-list">
		{{- range .All }}
		<li><label class="check who-user" data-role="{{ .EffectiveRole }}"><input type="checkbox" name="{{ $.UserName }}" value="{{ .Name }}" data-type="{{ $.Type }}"{{ if $.UserChecked .Name }} checked{{ end }}> {{ .Name }} <small class="role">{{ .EffectiveRole }}</small> <small class="who-state"></small></label></li>
		{{- end }}
		{{- range .Extra }}
		<li><label class="check who-user" data-role=""><input type="checkbox" name="{{ $.UserName }}" value="{{ . }}" data-type="{{ $.Type }}" checked> {{ . }} <small class="role">no such user</small> <small class="who-state"></small></label></li>
		{{- end }}
		</ul>
	</div>
	{{- if not .Compact }}
	<p class="dim who-summary"></p>
	{{- end }}
</div>
{{- end }}

{{- define "keypolicies" }}
<div class="field">
	<span class="label">Who gets it</span>
	{{- template "whofield" ((who "policies" "set" (index . 0) (index . 1) (index . 2) "").UsersNamed "users") }}
	<small class="faint">Their runs get the key through the ssh agent; check any roles and users, or none to keep it from every run.</small>
</div>
{{- end }}

{{- define "activity" }}
<div class="activity">
	<p class="activity-sum">{{ .Total }} run{{ if ne .Total 1 }}s{{ end }} in {{ len .Days }} days{{ if .Failed }} · <span class="bad">{{ .Failed }} failed</span>{{ end }}{{ if .Median }} · median {{ duration .Median }}{{ end }}<span class="activity-key"><i class="ok"></i>runs <i class="line avg"></i>average <i class="line max"></i>slowest <i class="line min"></i>fastest</span></p>
	<div class="activity-charts">
		{{- template "activity-chart" (list (printf "Last %d days" (len .Days)) "day" .Days .DayLines .DayTicks .DayStats "") }}
		{{- template "activity-chart" (list "Last 24 hours" "hour" .Hours .HourLines .HourTicks .HourStats "hours") }}
		{{- template "activity-chart" (list "Last 7 days" "3 hours" .Week .WeekLines .WeekTicks .WeekStats "wide") }}
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

{{- define "pipenode" }}
{{- $l := index . 0 }}{{ $i := index . 1 }}
<li class="node{{ with $l.State }} {{ . }}{{ end }}{{ if $l.Disabled }} disabled{{ end }}" data-step="{{ $i }}">
	<span class="rail"><i class="dot{{ with $l.State }} {{ . }}{{ end }}"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="{{ $l.Href }}"{{ if $l.RunID }} data-run-link="{{ $l.RunID }}"{{ end }}>{{ if $l.Workflow }}<svg class="ic lead"><use href="#i-flow"/></svg>{{ else if $l.Summary }}<svg class="ic lead"><use href="#i-summary"/></svg>{{ else if $l.Check }}<svg class="ic lead"><use href="#i-rotate"/></svg>{{ else if $l.Prompt }}<svg class="ic lead"><use href="#i-prompt"/></svg>{{ end }}{{ $l.Text }}</a>
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

{{- define "activity-chart" }}
{{- $title := index . 0 }}{{ $unit := index . 1 }}{{ $bars := index . 2 }}{{ $lines := index . 3 }}{{ $ticks := index . 4 }}{{ $stats := index . 5 }}{{ $class := index . 6 }}
<figure class="activity-chart{{ with $class }} {{ . }}{{ end }}">
	<figcaption class="activity-title"><span>{{ $title }}</span>{{ template "activity-stats" $stats }}</figcaption>
	<div class="activity-bars" style="--n: {{ len $bars }}" role="img" aria-label="Runs per {{ $unit }}, {{ lower $title }}">
		{{- range $bars }}
		<span class="tip {{ .Edge }}" data-tip="{{ .Title }}" aria-label="{{ .Title }}">{{ if .OtherStyle }}<i class="other" style="{{ .OtherStyle }}"></i>{{ end }}{{ if .FailedStyle }}<i class="bad" style="{{ .FailedStyle }}"></i>{{ end }}{{ if .OKStyle }}<i class="ok" style="{{ .OKStyle }}"></i>{{ end }}</span>
		{{- end }}
	</div>
	{{- with $lines }}{{ template "activity-lines" (list . $bars) }}{{ end }}
	<div class="activity-axis">{{ range $ticks }}<span style="left: {{ .At }}%">{{ .Label }}</span>{{ end }}</div>
</figure>
{{- end }}

{{- define "activity-stats" }}{{ if .Count }}<span class="faint" title="{{ .Count }} finished run{{ if ne .Count 1 }}s{{ end }}: the fastest, the mean, the slowest and their time together">min/avg/max/total: {{ duration .Fastest }} {{ duration .Mean }} {{ duration .Slowest }} {{ duration .Spent }}</span>{{ end }}{{ end }}

{{- define "activity-lines" }}
{{- $l := index . 0 }}{{ $bars := index . 1 }}{{ $n := len $bars }}
<div class="activity-line durations" style="--n: {{ $n }}">
	<svg viewBox="0 0 100 100" preserveAspectRatio="none" role="img" aria-label="Average, slowest and fastest run per column, the slowest {{ duration $l.Slowest }}">
		<path class="max" d="{{ $l.Max }}" vector-effect="non-scaling-stroke"></path>
		<path class="min" d="{{ $l.Min }}" vector-effect="non-scaling-stroke"></path>
		<path class="avg" d="{{ $l.Average }}" vector-effect="non-scaling-stroke"></path>
	</svg>
	<div class="activity-hits">
		{{- range $i, $b := $bars }}
		<span class="tip {{ $b.Edge }}" data-tip="{{ $b.AverageTitle }}" aria-label="{{ $b.AverageTitle }}"></span>
		{{- end }}
	</div>
</div>
{{- end }}
