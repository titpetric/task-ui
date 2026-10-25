{{/* schedtable takes (list $ schedules can what): the schedules of a job
     or a workflow, Expression | Next run | Last run | Runs as | State,
     and with can each row's menu, Enable or Disable and Delete, whose
     confirm names what the schedule is for. Nothing for no schedules. */}}
{{ define "schedtable" }}{{ $p := index . 0 }}{{ $can := index . 2 }}{{ $what := index . 3 }}{{ with index . 1 }}
	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
		{{- range . }}
			<tr>
				<td class="key" data-label="Expression"><code>{{ .Expression }}</code></td>
				<td data-label="Next run">{{ template "nextrun" . }}</td>
				<td class="nowrap hide-narrow" data-label="Last run">{{ if .LastRunAt }}{{ ago .LastRunAt }}{{ else }}<span class="faint">never</span>{{ end }}</td>
				<td data-label="Runs as">{{ template "runas-row" (list $p .) }}</td>
				<td data-label="State">{{ if .IsEnabled }}<span class="badge ok">enabled</span>{{ else }}<span class="badge">disabled</span>{{ end }}{{ if .Triages }} <span class="badge signal" title="Claude triages a failed run">triage</span>{{ end }}</td>
				<td class="actions">{{ if $can }}{{ template "schedmenu" (list . .IsEnabled $what) }}{{ end }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
{{- end }}{{ end }}

{{/* schedmenu takes (list schedule on what): a schedule row's menu, Enable
     or Disable, by on, whether the schedule runs, and Delete, whose
     confirm names what the schedule is for. */}}
{{ define "schedmenu" }}{{ $on := index . 1 }}{{ $what := index . 2 }}{{ with index . 0 }}{{ template "menu" (menu (print "Actions on the schedule " .Expression)
	(item (or (and $on "stop") "play") (or (and $on "Disable") "Enable") "data-api" (print "/api/schedules/" .ID) "data-method" "PUT" "data-body" (or (and $on `{"enabled":false}`) `{"enabled":true}`) "data-next" "reload")
	(danger "trash" "Delete" "data-delete" (print "/api/schedules/" .ID) "data-confirm" (print "Delete the schedule " .Expression " for " $what "?") "data-confirm-ok" "Delete" "data-next" "reload")
) }}{{ end }}{{ end }}
