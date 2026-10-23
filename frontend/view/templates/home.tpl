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
	<div class="table-wrap">
	<table class="table control col-hover">
		<thead><tr><th>Folder</th>{{ range $c.Columns }}<th class="col"><span>{{ . }}</span></th>{{ end }}<th class="col"><span>other</span></th></tr></thead>
		<tbody>
		{{- range $row := $c.Rows }}
			<tr>
				<td class="key nowrap"><a href="{{ $.Href $row.Folder.ID }}">{{ $row.Folder.ID }}</a></td>
				{{- range $i, $cell := $row.Cells }}
				{{- /* One job is a bare button; jobs that share a name carry their kinds. */}}
				<td class="col" data-label="{{ index $c.Columns $i }}">{{ if $cell }}<div>{{ range $cell }}{{ template "controljob" (list $ $row.Folder . (gt (len $cell) 1)) }}{{ end }}</div>{{ end }}</td>
				{{- end }}
				<td class="col other" data-label="Other">
					{{- template "controlmore" (list $ $row.Folder $row.Extra) }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="{{ add (len $c.Columns) 2 }}">No jobs yet. Add a project from a git repository in Settings, or put a folder with jobs under the root.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
</div>
{{ end }}
