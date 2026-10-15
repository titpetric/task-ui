{{ define "content" }}
{{- $p := .Project }}
<div class="crumbs"><a href="/">Overview</a><span>{{ $p.Title }}</span></div>
<div class="page-head">
	<div>
		<h2>{{ $p.Title }}</h2>
		{{- if $p.Description }}<p>{{ $p.Description }}</p>{{ end }}
	</div>
	<div class="dim">{{ $p.File }} <span class="badge">{{ $p.Kind }}</span></div>
</div>

{{- if $p.Error }}
<div class="alert bad"><b>Could not read the task file.</b> {{ $p.Error }}</div>
{{- end }}

<section class="block" data-live>
	<h3>Runbooks</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Runbook</th><th class="hide-sm">Description</th><th class="hide-sm">Last run</th><th></th></tr></thead>
		<tbody>
		{{- range $r := $p.Runbooks }}
			{{- $run := running $r.Key }}
			<tr>
				<td class="key"><a href="/p/{{ $p.ID }}/{{ $r.Name }}">{{ $r.Name }}</a>{{ if $r.Interactive }} <span class="badge">interactive</span>{{ end }}</td>
				<td class="desc hide-sm">{{ $r.Description }}</td>
				<td class="nowrap hide-sm">
					{{- if $run }}<span class="state"><i class="dot running"></i>running, {{ $run.User }}</span>
					{{- else }}{{ with last $p.ID $r.Name }}{{ template "state" . }} <span class="faint">{{ ago .Started }}</span>{{ else }}<span class="faint">never run</span>{{ end }}{{ end -}}
				</td>
				<td class="actions">
					{{- if canRun $p $r }}{{ template "runButtons" (list $p $r) }}{{ else }}<a class="btn btn-sm" href="/p/{{ $p.ID }}/{{ $r.Name }}">History</a>{{ end -}}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">No runbooks. Tasks with a desc and without internal: true show up here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
