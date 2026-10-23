{{ define "content" }}
{{- with .HostDetail }}{{ $d := . }}
<div class="page-head">
	<div>
		<h2>{{ $d.System.Name }}</h2>
		<p>A host of the <a href="/inventory">inventory</a>, as its line on the Inventory page opens: its containers, its durable filesystems and its month in charts.</p>
	</div>
</div>
{{- /* The Inventory page fetches this page and puts what follows under the
     host's line, its expanded row. */}}
<div class="host-detail" data-host-detail="{{ $d.System.Name }}">
	<p class="host-facts faint">
		<span><code>{{ $d.System.Host }}</code></span>
		{{- with $d.Login }}<span>ssh <code>{{ . }}</code></span>{{ else }}<span>no ssh login</span>{{ end }}
		{{- with $d.System.CPUs }}<span>{{ . }} cores</span>{{ end }}
		{{- if $d.System.Up }}<span>up {{ $d.Uptime }}</span>{{ else }}<span class="bad">{{ $d.System.Status }}</span>{{ end }}
		{{- with $d.System.Info.Agent }}<span>Beszel agent {{ . }}</span>{{ end }}
		{{- if not $d.Read.IsZero }}<span>read over ssh {{ ago $d.Read }}</span>{{ end }}
	</p>

	<h4 class="section-title">Containers <small>{{ len $d.Containers }}{{ if $d.Measured }}, measures from Beszel{{ end }}{{ if $d.Sized }}, sizes from docker ps{{ end }}</small></h4>
	{{- if $d.Containers }}
	<div class="table-wrap">
	<table class="table host-containers" data-sortable>
		<thead><tr>
			<th data-sort="text">Container</th>
			<th data-sort="text">Status</th>
			<th class="num" data-sort="num" aria-sort="descending">CPU</th>
			<th class="num" data-sort="num">Memory</th>
			<th class="num" data-sort="num">Size</th>
			<th class="num" data-sort="num">Net</th>
		</tr></thead>
		<tbody>
		{{- range $d.Containers }}
			<tr>
				<td data-v="{{ .Name }}"><span class="key">{{ .Name }}</span>{{ with .Image }}<span class="faint under">{{ . }}</span>{{ end }}</td>
				<td data-v="{{ .Status }}" class="nowrap{{ if and .State (ne .State "running") }} faint{{ end }}">{{ with .Status }}{{ . }}{{ else }}<span class="faint">-</span>{{ end }}</td>
				{{- if .Measured }}
				<td class="num" data-v="{{ .CPU }}">{{ printf "%.1f" .CPU }}%</td>
				<td class="num nowrap" data-v="{{ .Memory }}">{{ megabytes .Memory }}{{ if .MemoryShare }}<span class="faint under">{{ printf "%.1f" .MemoryShare }}% of RAM</span>{{ end }}</td>
				{{- else }}
				<td class="num faint" data-v="-1" title="{{ $d.MeasureNote }}">-</td>
				<td class="num faint" data-v="-1" title="{{ $d.MeasureNote }}">-</td>
				{{- end }}
				{{- if .Sized }}
				<td class="num nowrap" data-v="{{ .SizeRw }}">{{ bytes .SizeRw }}<span class="faint under">of {{ bytes .SizeVirtual }}</span></td>
				{{- else }}
				<td class="num faint" data-v="-1" title="{{ $d.SizeNote }}">-</td>
				{{- end }}
				{{- if and .Measured .Net }}
				<td class="num nowrap" data-v="{{ .Net }}">{{ rate .Net }}</td>
				{{- else }}
				<td class="num faint" data-v="-1" title="{{ if .Measured }}Beszel reports no traffic of the container{{ else }}{{ $d.MeasureNote }}{{ end }}">-</td>
				{{- end }}
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- if not $d.Sized }}<p class="faint host-note">No sizes: {{ $d.Unsized }}.</p>{{ end }}
	{{- else }}
	<p class="empty-state">{{ if or $d.Measured $d.Sized }}The host runs no containers.{{ else }}Neither Beszel nor docker ps lists containers on the host: {{ $d.Unsized }}.{{ end }}</p>
	{{- end }}

	<h4 class="section-title">Filesystems <small>durable storage{{ if $d.FromDF }}, from df over ssh{{ else if $d.Filesystems }}, the ones Beszel watches{{ end }}</small></h4>
	{{- if $d.Filesystems }}
	<div class="table-wrap">
	<table class="table host-filesystems">
		<thead><tr><th>Mount</th>{{ if $d.FromDF }}<th>Device</th><th>Type</th>{{ end }}<th class="num">Size</th><th class="num">Used</th><th class="num">Free</th><th>Use</th></tr></thead>
		<tbody>
		{{- range $d.Filesystems }}
			<tr>
				<td class="key"><code>{{ .Mount }}</code></td>
				{{- if $d.FromDF }}<td class="faint"><code>{{ .Device }}</code></td><td class="faint">{{ .Type }}</td>{{ end }}
				<td class="num nowrap">{{ bytes .Size }}</td>
				<td class="num nowrap">{{ bytes .Used }}</td>
				<td class="num nowrap">{{ bytes .Free }}</td>
				<td>{{ template "usebar" .Percent }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- else }}
	<p class="empty-state">Neither df over ssh nor Beszel reports the host's filesystems{{ if not $d.Login }}: task-ui has no ssh login to it, and Beszel no measures{{ end }}.</p>
	{{- end }}

	{{- if or $d.Charts $d.ContainerCharts }}
	<h4 class="section-title">The last 30 days <small>Beszel's records</small></h4>
	<div class="charts">
		{{- range $d.Charts }}{{ template "chart" . }}{{ end }}
		{{- range $d.ContainerCharts }}{{ template "chart" . }}{{ end }}
	</div>
	{{- else }}
	<p class="empty-state">Beszel holds no records of the host in the last 30 days.</p>
	{{- end }}
</div>
{{- end }}
{{ end }}

{{/* usebar is a share, 0 to 100, as a short bar and its number. */}}
{{ define "usebar" }}<span class="meter{{ if ge . 90.0 }} high{{ else if ge . 75.0 }} warm{{ end }}"><span class="spacebar-track"><i class="files" style="width: {{ printf "%.1f" . }}%"></i></span><span class="nowrap">{{ printf "%.0f" . }}%</span></span>{{ end }}
