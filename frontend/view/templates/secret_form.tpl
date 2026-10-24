{{ define "content" }}
{{- $s := .Secret }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>task-ui encrypts the value with the instance's encryption key and never shows it again. A workspace secret goes to every project in that workspace and no other. Saving an existing name replaces its value.</p>
	</div>
</div>

<form class="form" data-api="/api/secrets" data-method="POST" data-next="/settings/secrets?notice=Secret+saved">
	{{- template "secretfields" (list $s .SecretGroups) }}
	<div class="form-foot">
		<a class="btn" href="/settings/secrets">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
