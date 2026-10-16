{{ define "content" }}
{{ template "crumbs" (list (crumb "Settings" "/settings") (crumb "Projects" "/settings/projects") (crumb "Add" "")) }}
<div class="page-head">
	<div>
		<h2>Add project</h2>
		<p>Clones the repository into the root. The clone is the test: it either succeeds and the folder appears, or git's output is shown here and nothing is kept. Private repositories are reached with the instance SSH keys.</p>
	</div>
</div>

{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}
{{- if .Notice }}<pre class="summary">{{ .Notice }}</pre>{{ end }}

<form method="post" action="/settings/projects/clone" class="form">
	<div class="field">
		<label for="p-url">Repository URL</label>
		<input type="text" id="p-url" name="url" value="{{ index .Form "url" }}" placeholder="git@github.com:org/infra.git" spellcheck="false" autocomplete="off" required autofocus>
	</div>
	<div class="field">
		<label for="p-id">Folder <span class="faint">optional, the repository name by default</span></label>
		<input type="text" id="p-id" name="id" value="{{ index .Form "id" }}" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label for="p-branch">Branch <span class="faint">optional, the default branch otherwise</span></label>
		<input type="text" id="p-branch" name="branch" value="{{ index .Form "branch" }}" spellcheck="false" autocomplete="off">
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/projects">Cancel</a>
		<button class="btn btn-primary" type="submit">Clone</button>
	</div>
</form>
{{ end }}
