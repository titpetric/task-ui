{{ define "content" }}
{{- with $w := .Workflow }}
<div class="page-head">
	<div>
		<h2>{{ $w.Name }}</h2>
		<p class="chain">{{ range $i, $s := $w.Steps }}{{ if $i }}<span class="faint"> → </span>{{ end }}<a class="nowrap" href="/f/{{ $s.Folder }}/{{ $s.Job }}">{{ $s.Folder }} / {{ $s.Job }}</a>{{ end }}</p>
	</div>
	<div class="actions">
		{{- if $w.CanRun }}{{ if $w.Running }}
		<button type="button" class="btn" data-workflow-stop="{{ $w.Running }}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>
		{{- else }}
		<button type="button" class="btn btn-primary" data-workflow-run="{{ $w.ID }}"><svg class="ic"><use href="#i-play"/></svg>Run</button>
		{{- end }}{{ end }}
		{{- if $w.CanChange }}
		<button type="button" class="btn-icon danger" data-delete="/api/workflows/{{ $w.ID }}" data-confirm="Delete the workflow {{ $w.Name }}? Its runs stay in the history." data-next="/workflows" title="Delete" aria-label="Delete workflow"><svg class="ic"><use href="#i-trash"/></svg></button>
		{{- end }}
	</div>
</div>

<section class="block"{{ if $w.Running }} data-refresh="3000"{{ end }}>
	<h3>Runs</h3>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>By</th><th>Result</th><th>Steps</th></tr></thead>
		<tbody>
		{{- range $.WorkflowRuns }}
			<tr>
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td data-label="By">{{ .UserName }}</td>
				<td data-label="Result"><span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span>{{ if .Message }} <span class="faint">{{ .Message }}</span>{{ end }}</td>
				<td data-label="Steps"><span class="chain">{{ range $i, $s := .Steps }}{{ if $i }}<span class="faint"> → </span>{{ end }}<span class="nowrap">{{ if $s.RunID }}<a href="/runs/{{ $s.RunID }}" data-attach="{{ $s.RunID }}">{{ end }}<i class="dot {{ $s.State }}"></i>{{ $s.Folder }} / {{ $s.Job }}{{ if $s.RunID }}</a>{{ end }}</span>{{ end }}</span></td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">Not run yet.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

{{- if $w.CanChange }}
<section class="block">
	<h3>Change</h3>
	{{ template "stepeditor" (list $ $w "/api/workflows/" "PUT" "reload" "Save") }}
</section>
{{- end }}
{{- end }}
{{ end }}
