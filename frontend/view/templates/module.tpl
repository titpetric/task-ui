{{ define "content" }}
{{- with .Module }}
<div class="page-head">
	<div>
		<h2>{{ .Label }}</h2>
		{{- with .Description }}<p>{{ . }}</p>{{ end }}
	</div>
	<div class="actions">
		<span class="badge mono" title="The module's name, which decides its link">{{ .Name }}</span>
		{{- with .Version }}<span class="badge" title="What the module calls its own version">{{ . }}</span>{{ end }}
	</div>
</div>

{{/* What index.php echoed, as HTML. A module is code the instance runs, so
     its output is markup and not a value to escape; view.Module.Body says
     so and says what that means for anything a user typed. */}}
<div class="module-body">{{ .Body }}</div>

<p class="module-foot faint">Rendered in <span class="mono">{{ .Render }}</span> by phpscript.</p>
{{- end }}
{{ end }}
