{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Scheduled</h2>
		<p>Jobs that run in the background on a cron expression. Each schedule runs as its user: whoever created it, or the user an admin picks under Run as. The scheduler skips a job if it's still running when its time comes, or if its user can't run it at that time.</p>
	</div>
</div>

{{- $any := false }}{{ range $p := .Folders }}{{ range $p.Jobs }}{{ if $.CanRun $p . }}{{ $any = true }}{{ end }}{{ end }}{{ end }}{{ if .HostJobs }}{{ $any = true }}{{ end }}
{{- if $any }}
<section class="block" id="add-schedule">
<h3>Add schedule</h3>
<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload">
	<div class="field">
		<label for="sch-folder">Folder</label>
		<select id="sch-folder" name="folder" data-narrows="sch-job" required>
			{{- range $p := .Folders }}{{ if $.CanRunAny $p }}
			<option value="{{ $p.ID }}"{{ if eq (index $.Form "folder") $p.ID }} selected{{ end }}>{{ $p.ID }}</option>
			{{- end }}{{ end }}
			{{- if .HostJobs }}
			<option value="_"{{ if eq (index $.Form "folder") "_" }} selected{{ end }}>_</option>
			{{- end }}
		</select>
	</div>
	<div class="field">
		<label for="sch-job">Job</label>
		<select id="sch-job" name="job" required>
			{{- range $p := .Folders }}{{ range $p.Jobs }}{{ if $.CanRun $p . }}
			<option value="{{ .ID }}" data-folder="{{ $p.ID }}">{{ .Name }}</option>
			{{- end }}{{ end }}{{ end }}
			{{- template "hostjobs" (list $ "") }}
		</select>
	</div>
	<div class="field">
		<label for="sch-args">Args <span class="faint">for a job that takes them; the command of a host</span></label>
		<input type="text" id="sch-args" name="args" data-type="args" placeholder="df -h /" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label for="sch-expr">Expression</label>
		<input type="text" id="sch-expr" name="expression" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required>
	</div>
	{{- template "runas" (list $ "sch-user") }}
	<div class="form-foot">
		<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-plus"/></svg>Add schedule</button>
	</div>
</form>
</section>
{{- end }}

<section class="block">
	{{- template "blockhead" (head "Schedules" "" (and (not .Schedules) (print "No schedules yet." (or (and $any " Pick a job and a cron expression above.") "")))) }}
	{{- with .Schedules }}
	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Folder</th><th>Job</th><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
		{{- range . }}
			<tr>
				<td data-label="Folder"><a href="{{ $.Href .Folder }}">{{ .Folder }}</a></td>
				<td class="key" data-label="Job">{{ if and (eq .Folder "_") (hasPrefix .Job "@") }}{{ .Job }}{{ else }}<a href="{{ $.Href .Folder }}/{{ .Job }}">{{ .Job }}</a>{{ end }}{{ with argLine .Job .Args }} <code>{{ . }}</code>{{ end }}</td>
				<td data-label="Expression"><code class="nowrap">{{ .Expression }}</code></td>
				<td data-label="Next run">{{ template "nextrun" . }}</td>
				<td class="nowrap hide-narrow" data-label="Last run">{{ if .LastRunAt }}{{ ago .LastRunAt }}{{ else }}<span class="faint">never</span>{{ end }}</td>
				<td data-label="Runs as">{{ template "runas-row" (list $ .) }}</td>
				<td data-label="State">{{ if .IsEnabled }}<span class="badge ok">enabled</span>{{ else }}<span class="badge">disabled</span>{{ end }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on the schedule " .Expression)
					(item (or (and .IsEnabled "stop") "play") (or (and .IsEnabled "Disable") "Enable") "data-api" (print "/api/schedules/" .ID) "data-method" "PUT" "data-body" (or (and .IsEnabled `{"enabled":false}`) `{"enabled":true}`) "data-next" "reload")
					(danger "trash" "Delete" "data-delete" (print "/api/schedules/" .ID) "data-confirm" (print "Delete the schedule " .Expression " for " .Folder " / " .Job "?") "data-confirm-ok" "Delete" "data-next" "reload")
				) }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{ end }}
