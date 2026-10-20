{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Hosts</h2>
		<p>The hosts Beszel monitors: whether each is up, who task-ui signs in to it as over ssh, its cores, and how busy its CPU, memory and disk are, with their sizes. Read from the hub at most once a minute.</p>
	</div>
</div>

{{- if .HostsOff }}
<p class="empty-state">The Beszel integration is not set up.{{ if .User.CanAdmin }} Add its URL and token under <a href="/settings/integrations#beszel">Settings, Integrations</a>.{{ end }}</p>
{{- else }}
<div class="table-wrap">
<table class="table hosts">
	<thead><tr><th>Host</th><th>State</th>{{ if .Logins }}<th>SSH</th>{{ end }}<th class="num">Cores</th><th>CPU</th><th>Memory</th><th>Disk</th></tr></thead>
	<tbody>
	{{- range .Hosts }}
		<tr>
			<td class="key nowrap" title="{{ .Host }}"><svg class="ic"><use href="#i-server"/></svg> {{ .Name }}</td>
			<td><span class="badge {{ if .Up }}ok{{ else if eq .Status "down" }}bad{{ end }}">{{ .Status }}</span></td>
			{{- if $.Logins }}
			<td class="nowrap">{{ with index $.Logins .Host }}<code>{{ . }}</code> <button type="button" class="btn-icon" data-copy="ssh {{ . }}" title="Copy ssh {{ . }}" aria-label="Copy ssh {{ . }}"><svg class="ic"><use href="#i-copy"/></svg></button>{{ else }}<span class="faint">-</span>{{ end }}</td>
			{{- end }}
			<td class="num">{{ with .CPUs }}{{ . }}{{ else }}<span class="faint">-</span>{{ end }}</td>
			<td>{{ template "meter" .Info.CPU }}</td>
			<td>{{ template "meter" .Info.Memory }}{{ with .Stats }} <span class="faint nowrap">{{ printf "%.1f" .MemoryUsed }} / {{ printf "%.1f" .Memory }} GB</span>{{ end }}</td>
			<td>{{ template "meter" .Info.Disk }}{{ with .Stats }} <span class="faint nowrap">{{ printf "%.0f" .DiskUsed }} / {{ printf "%.0f" .Disk }} GB</span>{{ end }}</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="7">Beszel monitors no hosts.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{- end }}
{{ end }}

{{/* meter is a share, 0 to 100, as a short bar and its number. */}}
{{ define "meter" }}<span class="meter{{ if ge . 90.0 }} high{{ else if ge . 75.0 }} warm{{ end }}"><span class="spacebar-track"><i class="files" style="width: {{ printf "%.1f" . }}%"></i></span><span class="nowrap">{{ printf "%.0f" . }}%</span></span>{{ end }}
