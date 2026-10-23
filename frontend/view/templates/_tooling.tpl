{{- /* The tooling panel: what atkins recorded the tree running, a
     section per database it records to. Each holds a chart per window,
     a group of bars a day, one per top command of the window, with a
     legend, and the commands behind a toggle. The toggle is a details
     element, so it needs no script and the browser remembers nothing
     to go wrong. */}}
{{- define "tooling" }}
<div class="tooling">
	{{- range .Scopes }}
	<section class="tooling-scope">
		<p class="tooling-sum"><span class="tooling-db" title="{{ .Database }}">&lrm;{{ .Database }}</span><span><span class="key">{{ .Name }}</span>{{ with .Repository }} <span class="faint">{{ . }}</span>{{ end }} · {{ .Runs }} command{{ if ne .Runs 1 }}s{{ end }} in {{ .Days }} days · {{ duration .Total }} of tooling</span></p>
		<div class="tooling-charts">
			{{- range .Windows }}
			<figure class="tooling-chart">
				<figcaption class="tooling-title"><span>Last {{ .Days }} days</span><small>{{ .Runs }} command{{ if ne .Runs 1 }}s{{ end }} · {{ duration .Total }}{{ if .Busiest }} · busiest day {{ duration .Busiest }}{{ end }}</small></figcaption>
				<div class="tooling-bars{{ if le .Days 7 }} week{{ end }}" style="--n: {{ len .Bars }}; --k: {{ len .Legend }}" role="img" aria-label="Time spent per day by {{ .By }}, the last {{ .Days }} days">
					{{- range $bar := .Bars }}
					{{- if .Segments }}
					<span aria-label="{{ .Title }}">{{ range .Segments }}{{ if .Title }}<b class="tip {{ $bar.Edge }} {{ .Color }}" style="{{ .Style }}" data-tip="{{ .Title }}"></b>{{ else }}<b></b>{{ end }}{{ end }}</span>
					{{- else }}
					<span class="tip {{ .Edge }}" data-tip="{{ .Title }}" aria-label="{{ .Title }}"></span>
					{{- end }}
					{{- end }}
				</div>
				<div class="activity-axis">{{ range .Ticks }}<span style="left: {{ .At }}%">{{ .Label }}</span>{{ end }}</div>
				<ul class="tooling-legend">
					{{- range .Legend }}
					<li><span class="swatch {{ .Color }}"></span><span class="key" title="{{ .Name }}">{{ .Name }}</span><span class="nowrap">{{ duration .Total }}</span><small class="faint nowrap">{{ .Runs }} run{{ if ne .Runs 1 }}s{{ end }}</small></li>
					{{- else }}
					<li class="empty">Nothing ran in this window.</li>
					{{- end }}
				</ul>
			</figure>
			{{- end }}
		</div>
		{{- with .Commands }}
		<details class="tooling-details">
			<summary><svg class="ic chev"><use href="#i-chev"/></svg>{{ len . }} command{{ if ne (len .) 1 }}s{{ end }} by the time they have taken</summary>
			<div class="table-wrap">
			<table class="table cards">
				<thead><tr><th>Command</th><th>Tool</th><th>Runs</th><th>Total</th><th>Average</th><th>Slowest</th><th>Failed</th><th>Last</th></tr></thead>
				<tbody>
				{{- range . }}
					<tr>
						<td class="key">{{ .Command }}</td>
						<td data-label="Tool">{{ .Tool }}</td>
						<td data-label="Runs">{{ .Runs }}</td>
						<td class="nowrap" data-label="Total">{{ duration .Total }}</td>
						<td class="nowrap" data-label="Average">{{ duration .Average }}</td>
						<td class="nowrap" data-label="Slowest">{{ duration .Max }}</td>
						<td data-label="Failed">{{ if .Failures }}<span class="bad">{{ .Failures }}</span>{{ else }}0{{ end }}</td>
						<td class="nowrap" data-label="Last">{{ if .LastRun }}{{ ago .LastRun }}{{ else }}-{{ end }}</td>
					</tr>
				{{- end }}
				</tbody>
			</table>
			</div>
		</details>
		{{- end }}
	</section>
	{{- end }}
</div>
{{- end }}
