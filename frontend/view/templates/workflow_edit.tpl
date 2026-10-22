{{ define "content" }}
{{- with $w := .Workflow }}
<div class="page-head">
	<div>
		<h2>Edit {{ $w.Name }}</h2>
		<p>Change the name and the steps, in the order they run. Earlier runs keep the steps they ran with.</p>
	</div>
	<div class="actions">
		<button type="button" class="btn-icon danger" data-delete="/api/workflows/{{ $w.ID }}" data-confirm="Delete the workflow {{ $w.Name }}? Its runs stay in the history." data-next="/workflows" title="Delete" aria-label="Delete workflow"><svg class="ic"><use href="#i-trash"/></svg></button>
	</div>
</div>

{{ template "stepeditor" (list $ $w "/api/workflows/" "PUT" (print "/workflows/" $w.ID) "Save") }}
{{- end }}
{{ end }}
