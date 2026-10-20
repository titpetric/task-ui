{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Inventory</h2>
		<p>The hosts Beszel monitors, largest first: whether each is up, who task-ui signs in to it as over ssh, its cores, and how busy its CPU, memory and disk are, with their sizes. An admin runs a command on them all at once, like pssh.</p>
	</div>
</div>

{{- if .HostsOff }}
<p class="empty-state">The Beszel integration is not set up.{{ if .User.CanAdmin }} Add its URL and token under <a href="/settings/integrations#beszel">Settings, Integrations</a>.{{ end }}</p>
{{- else }}
{{- if and .User.CanAdmin .Logins }}
<section class="block inventory-run">
	<h3>Run on the hosts</h3>
	<form data-inventory-run>
		<div class="run-commands">
			{{- range list "uptime" "df -h /" "free -h" "docker ps --format '{{.Names}}\t{{.Status}}'" "uname -r" "cat /etc/os-release | grep PRETTY" }}
			<button type="button" class="btn btn-sm" data-command="{{ . }}">{{ . }}</button>
			{{- end }}
		</div>
		<textarea name="command" rows="2" required spellcheck="false" placeholder="a shell command, run on each host at once" aria-label="The command"></textarea>
		<div class="run-hosts">
			{{- range $h := .Hosts }}{{ with index $.Logins $h.Host }}
			<label class="check"><input type="checkbox" name="hosts" value="{{ $h.Host }}" checked> <code>{{ . }}</code></label>
			{{- end }}{{ end }}
		</div>
		<div class="note-form-foot"><span class="faint">Each host runs it over its ssh login, at once, for at most two minutes. Ctrl+Enter runs.{{ if .Summarize }} Run and summarize hands every host's output to Claude as one report once all are done.{{ end }}</span><span class="actions">{{ if .Summarize }}<button class="btn" type="submit" name="summarize" value="1"><svg class="ic"><use href="#i-note"/></svg>Run and summarize</button>{{ end }}<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run</button></span></div>
	</form>
	<div class="inventory-results" data-inventory-results></div>
</section>
{{- end }}
<div class="table-wrap">
<table class="table hosts">
	<thead><tr><th>Host</th><th>State</th>{{ if .Logins }}<th>SSH</th>{{ end }}<th class="center">Cores</th><th>CPU</th><th colspan="2">Memory</th><th colspan="2">Disk</th></tr></thead>
	<tbody>
	{{- range .Hosts }}
		<tr>
			<td class="key nowrap" title="{{ .Host }}"><svg class="ic"><use href="#i-server"/></svg> {{ .Name }}</td>
			<td><span class="badge {{ if .Up }}ok{{ else if eq .Status "down" }}bad{{ end }}">{{ .Status }}</span></td>
			{{- if $.Logins }}
			<td class="nowrap">{{ with index $.Logins .Host }}<code>{{ . }}</code> <button type="button" class="btn-icon" data-copy="ssh {{ . }}" title="Copy ssh {{ . }}" aria-label="Copy ssh {{ . }}"><svg class="ic"><use href="#i-copy"/></svg></button>{{ else }}<span class="faint">-</span>{{ end }}</td>
			{{- end }}
			<td class="center">{{ with .CPUs }}{{ . }}{{ else }}<span class="faint">-</span>{{ end }}</td>
			<td>{{ template "meter" .Info.CPU }}</td>
			<td>{{ template "meter" .Info.Memory }}</td>
			<td class="size faint nowrap">{{ with .Stats }}{{ printf "%.1f" .MemoryUsed }} / {{ printf "%.1f" .Memory }} GB{{ end }}</td>
			<td>{{ template "meter" .Info.Disk }}</td>
			<td class="size faint nowrap">{{ with .Stats }}{{ printf "%.0f" .DiskUsed }} / {{ printf "%.0f" .Disk }} GB{{ end }}</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="9">Beszel monitors no hosts.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{- end }}
{{ end }}

{{/* meter is a share, 0 to 100, as a short bar and its number. */}}
{{ define "meter" }}<span class="meter{{ if ge . 90.0 }} high{{ else if ge . 75.0 }} warm{{ end }}"><span class="spacebar-track"><i class="files" style="width: {{ printf "%.1f" . }}%"></i></span><span class="nowrap">{{ printf "%.0f" . }}%</span></span>{{ end }}
