{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Scheduled</h2>
		<p>Jobs that run on a cron expression, in the background, as the user each schedule names: whoever made it, or the user an admin picks under Run as. The scheduler skips a job that is still running when its time comes, and a job its user may not run at that time.</p>
	</div>
</div>

{{- $any := false }}{{ range $p := .Folders }}{{ range $p.Jobs }}{{ if $.CanRun $p . }}{{ $any = true }}{{ end }}{{ end }}{{ end }}
{{- if $any }}
<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload">
	<div class="field">
		<label for="sch-folder">Folder</label>
		<select id="sch-folder" name="folder" data-narrows="sch-job" required>
			{{- range $p := .Folders }}{{ if $.CanRunAny $p }}
			<option value="{{ $p.ID }}"{{ if eq (index $.Form "folder") $p.ID }} selected{{ end }}>{{ $p.ID }}</option>
			{{- end }}{{ end }}
		</select>
	</div>
	<div class="field">
		<label for="sch-job">Job</label>
		<select id="sch-job" name="job" required>
			{{- range $p := .Folders }}{{ range $p.Jobs }}{{ if $.CanRun $p . }}
			<option value="{{ .ID }}" data-folder="{{ $p.ID }}">{{ .Name }}</option>
			{{- end }}{{ end }}{{ end }}
		</select>
	</div>
	<div class="field">
		<label for="sch-expr">Expression</label>
		<input type="text" id="sch-expr" name="expression" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required>
	</div>
	{{- template "runas" (list $ "sch-user") }}
	<div class="form-foot">
		<button class="btn btn-primary" type="submit">Schedule</button>
	</div>
</form>
{{- end }}

<div class="table-wrap fit">
<table class="table fit-cards">
	<thead><tr><th>Folder</th><th>Job</th><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
	<tbody>
	{{- range .Schedules }}
		<tr>
			<td data-label="Folder"><a href="{{ $.Href .Folder }}">{{ .Folder }}</a></td>
			<td class="key" data-label="Job"><a href="{{ $.Href .Folder }}/{{ .Job }}">{{ .Job }}</a></td>
			<td data-label="Expression"><code class="nowrap">{{ .Expression }}</code></td>
			<td data-label="Next run">{{ template "nextrun" . }}</td>
			<td class="nowrap hide-narrow" data-label="Last run">{{ if .LastRunAt }}{{ ago .LastRunAt }}{{ else }}<span class="faint">never</span>{{ end }}</td>
			<td data-label="Runs as">{{ template "runas-row" (list $ .) }}</td>
			<td data-label="State">{{ if .IsEnabled }}<span class="badge ok">enabled</span>{{ else }}<span class="badge">disabled</span>{{ end }}</td>
			<td class="actions">
				<form data-api="/api/schedules/{{ .ID }}" data-method="PUT" data-next="reload"><input type="hidden" name="enabled" data-type="bool" value="{{ if .IsEnabled }}false{{ else }}true{{ end }}"><button class="btn btn-sm" type="submit">{{ if .IsEnabled }}Disable{{ else }}Enable{{ end }}</button></form>
				<button type="button" class="btn-icon danger" data-delete="/api/schedules/{{ .ID }}" data-confirm="Delete the schedule {{ .Expression }} for {{ .Folder }} / {{ .Job }}?" data-next="reload" title="Delete" aria-label="Delete schedule"><svg class="ic"><use href="#i-trash"/></svg></button>
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="8">Nothing is scheduled.{{ if $any }} Pick a job and an expression above.{{ end }}</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{ end }}
