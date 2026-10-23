{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Inventory</h2>
		<p>The hosts Beszel monitors, largest first. For each host: whether it's up, the ssh login task-ui uses, its cores, and how busy its CPU, memory and disk are. With Run on Hosts, you can run a command on many hosts at once, like pssh: each host gets a run of its own, with a recording, in Runs and the recents.</p>
	</div>
</div>

{{- if .HostsOff }}
<p class="empty-state">The Beszel integration isn't set up yet.{{ if .Can "integration/view" }} Add its URL and token under <a href="/settings/integrations#beszel">Settings, Integrations</a>.{{ end }}</p>
{{- else }}
{{- if and (.Can "hosts") .Logins }}
<section class="block inventory-run">
	<h3>Run on the hosts</h3>
	<form data-inventory-run>
		<div class="run-commands">
			{{- range list "uptime" "df -h /" "free -h" "docker ps --format '{{.Names}}\t{{.Status}}'" "uname -r" "cat /etc/os-release | grep PRETTY" }}
			<button type="button" class="btn btn-sm" data-command="{{ . }}">{{ . }}</button>
			{{- end }}
		</div>
		<textarea name="command" rows="2" required spellcheck="false" placeholder="a shell command to run on every picked host" aria-label="The command"></textarea>
		<div class="run-hosts" data-host-picker>
			<div class="host-tools">
				<span class="faint" data-host-count></span>
				<button type="button" class="btn btn-sm" data-hosts="all">All</button>
				<button type="button" class="btn btn-sm" data-hosts="none">None</button>
				<button type="button" class="btn btn-sm" data-hosts="invert">Invert</button>
			</div>
			<div class="host-picks">
				{{- range $h := .Hosts }}{{ with index $.Logins $h.Host }}
				<label class="host-pick"><input type="checkbox" name="hosts" value="{{ $h.Host }}" checked><span>{{ . }}</span></label>
				{{- end }}{{ end }}
			</div>
		</div>
		<div class="note-form-foot"><span></span><span class="actions">{{ if .Summarize }}<button class="btn" type="submit" name="summarize" value="1"><svg class="ic"><use href="#i-note"/></svg>Run and summarize</button>{{ end }}<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run</button></span></div>
	</form>
	{{- /* The user's last Run: a run of @host:<host> a host, one batch. While one
	     runs, the page fetches itself and puts this back. */}}
	<div class="inventory-results" data-inventory-results>
		{{- with .Batch }}
		{{- $first := index . 0 }}{{ $live := false }}{{ range . }}{{ if not (ended .State) }}{{ $live = true }}{{ end }}{{ end }}
		<div class="block-head"{{ if $live }} data-batch-live{{ end }}>
			<h4>Your last run <small><code>{{ range $first.ArgList }}{{ . }}{{ end }}</code> on {{ len . }} host{{ if ne (len .) 1 }}s{{ end }}, {{ ago $first.StartedAt }}</small></h4>
			{{- with $first.BatchID }}<a class="btn btn-sm" href="/runs?batch={{ . }}">In Runs</a>{{ end }}
		</div>
		<div class="table-wrap">
		<table class="table run-results">
			<thead><tr><th>Host</th><th>Result</th><th>Output</th></tr></thead>
			<tbody>
			{{- range . }}
				<tr>
					<td class="nowrap"><a href="/runs/{{ .ID }}"><code>{{ .Target }}</code></a></td>
					<td class="nowrap">{{ template "state" . }} <span class="faint">{{ duration .Elapsed }}</span> <button type="button" class="btn btn-sm" data-attach="{{ .ID }}">{{ if not (ended .State) }}Attach{{ else }}View{{ end }}</button></td>
					<td class="run-output">{{ if eq .State "queued" }}<span class="faint">Queued: waits for the run on {{ .Target }} before it.</span>{{ else if not (ended .State) }}<div class="frame" role="img" aria-label="Terminal of {{ .Target }}">{{ .FrameSVG }}</div>{{ else if .CastPath }}<pre class="output" data-output="{{ .ID }}" data-cols="{{ .TermCols }}" data-rows="{{ .TermRows }}"><span class="faint">Reading the recording…</span></pre>{{ else }}<span class="faint">No recording.</span>{{ end }}</td>
				</tr>
			{{- end }}
			</tbody>
		</table>
		</div>
		{{- end }}
	</div>
	{{- template "promptbox" . }}
</section>
{{- end }}
{{- /* The hosts follow Beszel: the page fetches itself once a minute and
     puts this section back, with the time of the fetch under the table. */}}
<section class="block" data-hosts-live>
{{- with .Cluster }}
<div class="cluster">
	<h3 class="section-title">Cluster <small>{{ len .Hosts }} host{{ if ne (len .Hosts) 1 }}s{{ end }}, {{ .Up }} up{{ with .Down }}, <span class="bad">{{ . }} down</span>{{ end }}</small></h3>
	<div class="pools">
		{{- range .Pools }}
		<div class="pool">
			<div class="pool-head">
				<span class="pool-name">{{ .Name }}</span>
				<span class="pool-figure">{{ .Figure }}</span>
				{{- if ne .Name "Containers" }}<span class="pool-pct">{{ printf "%.0f" .Percent }}%</span>{{ end }}
			</div>
			<div class="pool-bar" role="img" aria-label="{{ .Name }}: {{ .Figure }}">
				{{- range .Segments }}{{ if gt .Width 0.0 }}<span class="seg c{{ .Color }}" style="flex-grow: {{ printf "%.3f" .Width }}" title="{{ .Title }}"><i style="width: {{ printf "%.1f" .Fill }}%"></i></span>{{ end }}{{ end }}
			</div>
		</div>
		{{- end }}
	</div>
	<ul class="pool-legend">
		{{- range .Hosts }}
		<li{{ if .Left }} class="left"{{ end }}><i class="swatch{{ if not .Left }} c{{ .Color }}{{ end }}"></i>{{ .Name }}</li>
		{{- end }}
	</ul>
	{{- with .Left }}
	<p class="faint pool-note">{{ . }} host{{ if ne . 1 }}s are down or send{{ else }} is down or sends{{ end }} no measures, greyed in the legend and left out of the pools.</p>
	{{- end }}
	{{- with $.ClusterCharts }}
	<div class="charts">
		{{- range . }}{{ template "chart" . }}{{ end }}
	</div>
	{{- end }}
</div>
{{- end }}
<h3>Hosts</h3>
<div class="table-wrap">
<table class="table hosts">
	<thead><tr><th>Host</th>{{ if .Logins }}<th>SSH</th>{{ end }}<th class="center">Cores</th><th>CPU</th><th colspan="2">Memory <small>GB</small></th><th colspan="2">Disk <small>GB</small></th></tr></thead>
	<tbody>
	{{- $open := .Can "hosts" }}
	{{- range .Hosts }}
		{{- /* With hosts, a click on a host's line, or Enter or Space on it,
		     opens its expanded row: the page of the host, fetched. */}}
		<tr{{ if $open }} class="host-row" data-host-row="{{ .Name }}" tabindex="0" aria-expanded="false"{{ end }}>
			<td class="key nowrap" title="{{ .Host }}, {{ .Status }}">{{ if $open }}<svg class="ic chev" aria-hidden="true"><use href="#i-chev"/></svg>{{ end }}<svg class="ic host-state {{ if .Up }}up{{ else }}down{{ end }}" role="img" aria-label="{{ .Status }}"><use href="#i-server"/></svg> {{ .Name }}</td>
			{{- if $.Logins }}
			<td class="nowrap">{{ with index $.Logins .Host }}<code>{{ . }}</code> <button type="button" class="btn-icon" data-copy="ssh {{ . }}" title="Copy ssh {{ . }}" aria-label="Copy ssh {{ . }}"><svg class="ic"><use href="#i-copy"/></svg></button>{{ else }}<span class="faint">-</span>{{ end }}</td>
			{{- end }}
			<td class="center">{{ with .CPUs }}{{ . }}{{ else }}<span class="faint">-</span>{{ end }}</td>
			<td>{{ template "meter" .Info.CPU }}</td>
			<td>{{ template "meter" .Info.Memory }}</td>
			<td class="size faint nowrap">{{ with .Stats }}{{ printf "%.1f" .MemoryUsed }} / {{ printf "%.1f" .Memory }}{{ end }}</td>
			<td>{{ template "meter" .Info.Disk }}</td>
			<td class="size faint nowrap">{{ with .Stats }}{{ printf "%.0f" .DiskUsed }} / {{ printf "%.0f" .Disk }}{{ end }}</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="8">Beszel isn't monitoring any hosts.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
<p class="faint live-stamp">Updated {{ .Now.Format "15:04" }}</p>
</section>
{{- end }}
{{ end }}

{{/* meter is a share, 0 to 100, as a short bar and its number. */}}
{{ define "meter" }}<span class="meter{{ if ge . 90.0 }} high{{ else if ge . 75.0 }} warm{{ end }}"><span class="spacebar-track"><i class="files" style="width: {{ printf "%.1f" . }}%"></i></span><span class="nowrap">{{ printf "%.0f" . }}%</span></span>{{ end }}
