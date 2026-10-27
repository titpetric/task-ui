{{ define "content" }}
{{- $s := .Secret }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>A variable for the environment of runs. A secret's value is encrypted with the instance's encryption key and never shown again; an env value is listed on the Secrets page. One applies to every workspace, or to one workspace. Saving an existing name replaces its value.</p>
	</div>
</div>

<form class="form" data-api="/api/secrets" data-method="POST" data-next="/settings/secrets?notice=Variable+{{ if $s.ID }}saved{{ else }}added{{ end }}">
	{{- template "secretfields" (list $s .SecretGroups) }}
	<div class="form-foot">
		<a class="btn" href="/settings/secrets">Cancel</a>
		<button class="btn btn-primary" type="submit">{{ if $s.ID }}Save{{ else }}Add variable{{ end }}</button>
	</div>
</form>
{{ end }}
