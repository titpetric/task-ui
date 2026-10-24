{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Dashboard</h2>
		<p>Every running terminal, live, for everyone to see. Open one to attach, or pick a folder on the left to start a run.</p>
	</div>
</div>

{{- /* The dashboard follows the runs topic: a run that starts adds its
     tile, one that ends marks it, and the count and the control centre's
     buttons keep up; no reload. */}}
{{- template "pipelinenote" (list $ ($.PipelineAlert "")) }}

<div data-sse="runs" data-dashboard>
{{- $jobs := 0 }}{{ range .Folders }}{{ $jobs = len .Jobs | add $jobs }}{{ end }}
<div class="metrics">
	<div><b>Folders</b><span>{{ len .Folders }}</span></div>
	<div><b>Jobs</b><span>{{ $jobs }}</span></div>
	<div><b>Running</b><span data-running-count>{{ len .Running }}</span></div>
	<div><b>Scheduled</b><span>{{ len .Schedules }}</span></div>
	<div><b>Your role</b><span>{{ .User.EffectiveRole }}</span></div>
</div>

<section class="block">
	<h3>Running now <small>one run per job, up to a run per CPU; terminals take no turn</small></h3>
	<div class="tiles" data-tiles>
	{{- range .Runs }}
		<a class="tile" href="/runs/{{ .ID }}" data-attach="{{ .ID }}" data-tile="{{ .ID }}">
			<div class="tile-head">
				<i class="dot {{ .State }}"></i>
				<b>{{ .Name }}</b>
				<span class="term-title" data-term-title{{ with .Title }} title="{{ . }}"{{ end }}>{{ .Title }}</span>
				<span class="spacer"></span>
				<span class="faint">{{ if and (eq .State "queued") .Message }}{{ .Message }}{{ else if .Attached }}{{ .Attached }} watching{{ end }}</span>
				{{ template "avatar" (list $ .UserName) }}
			</div>
			<div class="frame"{{ if eq .State "running" }} data-frame="{{ .ID }}"{{ end }} role="img" aria-label="Terminal of {{ .Name }}">{{ .FrameSVG }}</div>
		</a>
	{{- end }}
	</div>
	<div class="alert" data-tiles-empty{{ if .Runs }} hidden{{ end }}>Nothing is running. Open a folder and press Run, or run a workflow, and its terminal appears here for everyone.</div>
</section>

<section class="block">
	<h3>Control centre <small>job names that many folders share get a column; the rest go under other</small></h3>
	{{- with $c := .Control }}
	{{- template "controltable" (list $ $c false) }}
	{{- end }}
</section>
</div>
{{ end }}
