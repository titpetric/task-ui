{{ define "content" }}
{{- $keyed := false }}{{ range .SecretGroups }}{{ if .Key }}{{ $keyed = true }}{{ end }}{{ end }}
<div class="page-head">
	<div>
		<h2>Secrets</h2>
		<p>Variables for the environment of runs, per workspace or per project; a project secret replaces a workspace one of the same name. Values are never shown again.</p>
	</div>
	<div class="actions">
		{{- if $keyed }}
		<a class="btn btn-primary" href="/settings/secrets/new{{ with .SecretWorkspace }}?scope=workspace:{{ . }}{{ end }}"><svg class="ic"><use href="#i-plus"/></svg>Add secret</a>
		{{- else }}
		<button class="btn btn-primary" type="button" disabled title="The instance has no encryption key yet"><svg class="ic"><use href="#i-plus"/></svg>Add secret</button>
		{{- end }}
	</div>
</div>

{{- with .SecretWorkspace }}
<p class="dim secret-filter">Showing the workspace <b>{{ . }}</b>. <a href="/settings/secrets">Every workspace</a></p>
{{- end }}

{{- range .SecretGroups }}{{ template "secretgroup" (list $ . false) }}{{ end }}
{{ end }}
