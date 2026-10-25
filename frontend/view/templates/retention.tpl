{{ define "content" }}
{{- $admin := and .User .User.CanAdmin }}
<div class="page-head">
	<div>
		<h2>Data retention</h2>
		<p>How long task-ui keeps what it records, one row per area. <code>@cleanup</code> applies the rows in this order every midnight, oldest first: run history takes a run's recording, summary and threads with it.{{ if not $admin }} An admin changes the rules.{{ end }}</p>
	</div>
	{{- if $admin }}
	<div class="actions">
		<button type="button" class="btn" data-run data-folder="_" data-job="@cleanup" data-title="_ / cleanup" title="Run @cleanup now, with the saved rules">Clean up now</button>
	</div>
	{{- end }}
</div>

<form class="form retention" data-api="/api/retention" data-method="PUT" data-next="/settings/retention?notice=Data+retention+saved">
<section class="block">
	<h3>What the instance keeps <small>a row per area, the oldest items go first</small></h3>
	<div class="table-wrap">
	<table class="table retention">
		<thead><tr><th>Area</th><th>What goes</th><th>Keep</th><th>Size limit</th><th>Holds now</th><th>Last cleanup</th></tr></thead>
		<tbody>
		{{- range .Retention }}
			<tr>
				<td class="key">{{ .Area.Title }}</td>
				<td class="desc">{{ .Area.What }}</td>
				{{- if .Area.Fixed }}
				<td colspan="2" class="dim">{{ .Area.Fixed }}</td>
				{{- else }}
				<td><select name="{{ .Area.Kind }}.age" aria-label="Keep {{ .Area.Title }} for"{{ if not $admin }} disabled{{ end }}>
					{{- range .Ages }}<option value="{{ .Value }}"{{ if .Selected }} selected{{ end }}>{{ .Label }}</option>{{ end }}
				</select></td>
				<td>{{ if .Area.Limited }}<select name="{{ .Area.Kind }}.limit" aria-label="Size limit of {{ .Area.Title }}"{{ if not $admin }} disabled{{ end }}>
					{{- range .Limits }}<option value="{{ .Value }}"{{ if .Selected }} selected{{ end }}>{{ .Label }}</option>{{ end }}
				</select>{{ else }}<span class="faint">none</span>{{ end }}</td>
				{{- end }}
				<td class="nowrap">{{ .Holds }}</td>
				<td class="nowrap">{{ with $.Cleanup }}<a href="/runs/{{ .ID }}" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</a> {{ template "state" . }}{{ else }}<span class="faint">never</span>{{ end }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- if $admin }}
	<div class="form-foot">
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
	{{- end }}
</section>
</form>
{{ end }}
