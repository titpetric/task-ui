{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Settings</h2>
		<p>Set up the instance in this order: each step builds on the ones before it. Every page ends with links to the step before and the step after.</p>
	</div>
</div>

<section class="block">
	<h3>Steps</h3>
	<div class="table-wrap">
	<table class="table settings">
		<tbody>
			{{- range .SetupSteps }}
			<tr>
				<td class="num dim">{{ .Number }}</td>
				<td class="key nowrap"><a href="{{ .Href }}"><svg class="ic"><use href="#i-{{ .Icon }}"/></svg> {{ .Name }}</a></td>
				<td class="desc">{{ .Desc }}</td>
				<td class="count dim"><span class="nowrap">{{ index $.StepCounts .Name }}</span></td>
				<td class="actions"><a class="btn btn-sm" href="{{ .Href }}">Open</a></td>
			</tr>
			{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
