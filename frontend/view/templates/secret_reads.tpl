{{ define "content" }}
{{- $s := .Secret }}
<div class="page-head">
	<div>
		<h2>Reads of {{ $s.Name }} <span class="badge">{{ $s.Scope }}</span></h2>
		<p>The runs that got {{ $s.Name }} in their environment, newest first. task-ui keeps the last 1000 reads of each secret. The value never shows.</p>
	</div>
	<div class="actions">
		<a class="btn" href="/settings/secrets">Secrets</a>
	</div>
</div>

<section class="block">
	<h3>Reads <small>newest first</small></h3>
	<div class="table-wrap">
	<table class="table secret-reads">
		<thead><tr><th>When</th><th>Run</th><th>Job</th><th>By</th></tr></thead>
		<tbody>
		{{- range .SecretReads }}
			<tr>
				<td class="nowrap" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
				<td><a class="faint mono" href="/runs/{{ .RunID }}" title="{{ .RunID }}">{{ slice .RunID 0 8 }}</a></td>
				<td class="key">{{ .Folder }}/{{ with .Dir }}{{ . }}/{{ end }}{{ .Job }}</td>
				<td>{{ template "user" (list $ .UserName) }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">No run got {{ $s.Name }} yet.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- template "pager" .Pager }}
</section>
{{ end }}
