{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Hosts</h2>
		<p>The hosts Beszel monitors: whether each is up, its load, memory and disk, and the machine. Read from the hub at most once a minute.</p>
	</div>
</div>

{{- if .HostsOff }}
<p class="empty-state">The Beszel integration is not set up.{{ if .User.CanAdmin }} Add its URL and token under <a href="/settings/integrations#beszel">Settings, Integrations</a>.{{ end }}</p>
{{- else }}
<div class="table-wrap">
<table class="table hosts">
	<thead><tr><th>Host</th><th>State</th><th>CPU</th><th>Memory</th><th>Disk</th><th class="hide-sm">Uptime</th><th class="hide-sm">Machine</th><th class="hide-sm">Agent</th></tr></thead>
	<tbody>
	{{- range .Hosts }}
		<tr>
			<td class="key nowrap"><svg class="ic"><use href="#i-server"/></svg> {{ .Name }}<div class="faint mono">{{ .Host }}</div></td>
			<td><span class="badge {{ if .Up }}ok{{ else if eq .Status "down" }}bad{{ end }}">{{ .Status }}</span></td>
			<td>{{ template "meter" .Info.CPU }}</td>
			<td>{{ template "meter" .Info.Memory }}</td>
			<td>{{ template "meter" .Info.Disk }}</td>
			<td class="hide-sm nowrap">{{ if .Info.Uptime }}{{ duration .UptimeDuration }}{{ else }}<span class="faint">-</span>{{ end }}</td>
			<td class="hide-sm"><div>{{ .Info.CPUModel }}</div><div class="faint">{{ if .Info.Cores }}{{ .Info.Cores }} cores{{ end }}{{ with .Info.Kernel }} · {{ . }}{{ end }}</div></td>
			<td class="hide-sm faint mono">{{ .Info.Agent }}</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="8">Beszel monitors no hosts.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{- end }}
{{ end }}

{{/* meter is a share, 0 to 100, as a short bar and its number. */}}
{{ define "meter" }}<span class="meter{{ if ge . 90.0 }} high{{ else if ge . 75.0 }} warm{{ end }}"><span class="spacebar-track"><i class="files" style="width: {{ printf "%.1f" . }}%"></i></span><span class="nowrap">{{ printf "%.0f" . }}%</span></span>{{ end }}
