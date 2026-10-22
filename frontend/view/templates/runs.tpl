{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Runs</h2>
		<p>Every run, by anyone, newest first. Attach to a running job from here; closing the window doesn't stop it.</p>
	</div>
</div>

<form method="get" action="/runs" class="filters">
	<label>Folder
		<select name="folder">
			<option value="">all folders</option>
			{{- range .Folders }}<option value="{{ .ID }}"{{ if eq (index $.Form "folder") .ID }} selected{{ end }}>{{ .ID }}</option>{{ end }}
		</select>
	</label>
	<label>Job <input type="text" name="job" value="{{ index .Form "job" }}" spellcheck="false" autocomplete="off"></label>
	<label>State
		<select name="state">
			<option value="">any state</option>
			{{- range states }}<option value="{{ . }}"{{ if eq (index $.Form "state") . }} selected{{ end }}>{{ . }}</option>{{ end }}
		</select>
	</label>
	<label>By
		<select name="user">
			<option value="">anyone</option>
			{{- range .UserNames }}<option value="{{ . }}"{{ if eq (index $.Form "user") . }} selected{{ end }}>{{ . }}</option>{{ end }}
		</select>
	</label>
	{{- template "pagesize" . }}
	<button class="btn" type="submit">Filter</button>
</form>

{{- with .Activity }}
<section class="block">
	<h3>Activity <small>in the server's time zone</small></h3>
	{{- template "activity" . }}
</section>
{{- end }}

{{- /* The table follows the runs topic: a run that starts goes on top,
     one that ends changes in place, no reload. */}}
<section class="block" data-sse="runs">
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Folder</th><th>Job</th><th>By</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody data-runs-table{{ if .Summarize }} data-summarize{{ end }}>
		{{- range .Runs }}{{ $r := . }}
			<tr data-run-id="{{ .ID }}">
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td data-label="Folder">{{ with $.RunFolderHref .Folder .Dir }}<a href="{{ . }}">{{ $r.Place }}</a>{{ else }}{{ .Place }}{{ end }}</td>
				<td class="key">{{ with $.RunJobHref .Folder .Dir .Job }}<a href="{{ . }}">{{ $r.Job }}</a>{{ else }}{{ .Job }}{{ end }}</td>
				<td data-label="By">{{ template "who" (list $ .) }}</td>
				<td class="nowrap" data-label="Took" data-took>{{ duration .Elapsed }}</td>
				<td data-label="Result" data-result>{{ template "state" . }}{{ if .Attached }} <span class="faint">{{ .Attached }} watching</span>{{ end }}</td>
				<td class="actions" data-actions>
					<a class="btn btn-sm" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">{{ if eq .State "running" }}Attach{{ else }}View{{ end }}</a>
					{{- if and $.Summarize (ne .State "running") }}
					<button type="button" class="btn btn-sm" data-summary="/api/runs/{{ .ID }}/summary" data-summary-page="/runs/{{ .ID }}" data-summary-title="{{ .Name }}" title="Claude's summary of the output">Summary</button>
					{{- end }}
					{{- if and .Live (eq .State "running") }}
					<button type="button" class="btn-icon" data-stop="{{ .ID }}" title="Stop the run" aria-label="Stop the run"><svg class="ic"><use href="#i-stop"/></svg></button>
					<button type="button" class="btn-icon danger" data-kill="{{ .ID }}" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-trash"/></svg></button>
					{{- end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="7">No runs match. Run a job from its folder and it appears here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- template "pager" .Pager }}
</section>
{{ end }}
