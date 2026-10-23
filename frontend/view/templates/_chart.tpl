{{/* chart is a month of measures: its title, the drawing, and a row a
     series of its least, mean and most, the stack's total last. */}}
{{ define "chart" }}
<figure class="chart">
	<figcaption class="chart-title">{{ .Title }}</figcaption>
	{{- if .Empty }}
	<p class="faint chart-empty">No records in the last 30 days.</p>
	{{- else }}
	{{ .SVG }}
	<table class="chart-legend">
		<thead><tr><th></th><th>min</th><th>avg</th><th>max</th></tr></thead>
		<tbody>
		{{- range .Legend }}
			<tr><td><i class="swatch s{{ .Color }}"></i>{{ .Name }}</td><td>{{ .Min }}</td><td>{{ .Avg }}</td><td>{{ .Max }}</td></tr>
		{{- end }}
		{{- with .Total }}
			<tr class="total"><td>{{ .Name }}</td><td>{{ .Min }}</td><td>{{ .Avg }}</td><td>{{ .Max }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	{{- end }}
</figure>
{{ end }}
