{{- /* The tooling panel: what atkins recorded the tree running, a chart
     per window with the tools behind it, and the commands behind a
     toggle. The toggle is a details element, so it needs no script and
     the browser remembers nothing to go wrong. */}}
{{- define "tooling" }}
<div class="tooling">
	<p class="tooling-sum">{{ .Runs }} command{{ if ne .Runs 1 }}s{{ end }} in {{ .Days }} days · {{ duration .Total }} of tooling<span class="tooling-db" title="{{ .Database }}">{{ .Database }}</span></p>
	<div class="tooling-charts">
		{{- range .Windows }}
		<figure class="tooling-chart">
			<figcaption class="tooling-title"><span>Last {{ .Days }} days</span><small>{{ .Runs }} command{{ if ne .Runs 1 }}s{{ end }} · {{ duration .Total }}{{ if .Busiest }} · busiest day {{ duration .Busiest }}{{ end }}</small></figcaption>
			<div class="tooling-bars" style="--n: {{ len .Bars }}" role="img" aria-label="Time spent per day, the last {{ .Days }} days">
				{{- $last := sub (len .Bars) 4 }}
				{{- range $i, $b := .Bars }}
				<span class="tip{{ if lt $i 4 }} tip-start{{ else if ge $i $last }} tip-end{{ end }}" data-tip="{{ .Title }}" aria-label="{{ .Title }}">{{ with .Style }}<i style="{{ . }}"></i>{{ end }}</span>
				{{- end }}
			</div>
			<ul class="tooling-tools">
				{{- range .Tools }}
				<li><span class="key">{{ .Tool }}</span><span class="tooling-bar"><i style="{{ .Style }}"></i></span><span class="nowrap">{{ duration .Total }}</span><small class="faint nowrap">{{ .Runs }} run{{ if ne .Runs 1 }}s{{ end }}</small></li>
				{{- else }}
				<li class="empty">Nothing ran in this window.</li>
				{{- end }}
			</ul>
		</figure>
		{{- end }}
	</div>
	{{- with .Commands }}
	<details class="tooling-details">
		<summary>{{ len . }} command{{ if ne (len .) 1 }}s{{ end }} by the time they have taken</summary>
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
</div>
{{- end }}
