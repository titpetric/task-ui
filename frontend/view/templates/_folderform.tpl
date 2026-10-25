{{- /* recursivefield is the Recursive toggle of the workspace and project
     forms, one field for both: (list id checked). Off, discovery lists
     the folder and its top-level folders alone. */}}
{{- define "recursivefield" }}
{{- $id := index . 0 }}{{ $checked := index . 1 }}
	<div class="field">
		<label class="check"><input type="checkbox" id="{{ $id }}" name="recursive"{{ if $checked }} checked{{ end }}> Recursive</label>
		<small class="faint">Lists every folder inside; off lists only the folder and its top-level folders.</small>
	</div>
{{- end }}
