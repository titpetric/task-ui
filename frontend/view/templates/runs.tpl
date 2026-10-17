{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Runs</h2>
		<p>Every run, by anyone, newest first. A run in progress can be attached to from here; closing the window leaves it running.</p>
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
	<button class="btn" type="submit">Filter</button>
</form>

<section class="block" data-live>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Folder</th><th>Job</th><th>By</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
		{{- range .Runs }}
			<tr>
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td data-label="Folder"><a href="/f/{{ .Folder }}">{{ .Folder }}</a></td>
				<td class="key"><a href="/f/{{ .Folder }}/{{ .Job }}">{{ .Job }}</a></td>
				<td data-label="By">{{ template "who" . }}</td>
				<td class="nowrap" data-label="Took">{{ duration .Elapsed }}</td>
				<td data-label="Result">{{ template "state" . }}{{ if .Attached }} <span class="faint">{{ .Attached }} watching</span>{{ end }}</td>
				<td class="actions">
					<a class="btn btn-sm" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">{{ if eq .State "running" }}Attach{{ else }}View{{ end }}</a>
					{{- if and .Live (eq .State "running") }}
					<button type="button" class="btn-icon" data-stop="{{ .ID }}" title="Stop the run" aria-label="Stop the run"><svg class="ic"><use href="#i-stop"/></svg></button>
					<button type="button" class="btn-icon danger" data-kill="{{ .ID }}" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-trash"/></svg></button>
					{{- end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="7">No runs match. Run a job from its folder and it shows up here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
