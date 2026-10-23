{{ define "content" }}
{{- with $w := .Workflow }}
<div class="page-head">
	<div>
		<h2>Edit {{ $w.Name }}</h2>
		<p>Change the name and the steps: fold a step open to edit it, drag it by its grip to move it. Earlier runs keep the steps they ran with.</p>
	</div>
	<div class="actions">
		{{ template "menu" (menu (print "Actions on " $w.Name)
			(item "flow" "Open" "href" (print "/workflows/" $w.ID))
			(danger "trash" "Delete" "data-delete" (print "/api/workflows/" $w.ID) "data-confirm" (print "Delete the workflow " $w.Name "? Its runs stay in the history.") "data-confirm-ok" "Delete" "data-next" "/workflows")
		) }}
	</div>
</div>

{{ template "stepeditor" (list $ $w "/api/workflows/" "PUT" (print "/workflows/" $w.ID) "Save") }}
{{- end }}
{{ end }}
