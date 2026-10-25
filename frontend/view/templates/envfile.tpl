{{ define "content" }}
{{- $p := .Folder }}{{ $f := .EnvFile }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-file"/></svg> {{ if $f.Path }}{{ $f.Path }}{{ else }}Add env file{{ end }}</h2>
		<p>A dotenv file task-ui writes into the folder, on save and before every run, with the secrets below added at the end. Saving it empty deletes it.</p>
	</div>
</div>

<div class="grid-2 wide">
{{- if $f.Path }}
<form class="form" data-api="/api/env/{{ $p.ID }}/{{ $f.Path }}" data-method="PUT" data-next="{{ $.Href $p.ID }}?notice={{ $f.Path }}+saved">
{{- else }}
<form class="form" data-api="/api/env/{{ $p.ID }}/{path}" data-method="PUT" data-next="{{ $.Href $p.ID }}?notice={path}+saved">
	<div class="field">
		<label for="env-path">Path <span class="faint">relative to the folder, such as .env</span></label>
		<input type="text" id="env-path" name="path" required spellcheck="false" autocomplete="off" placeholder=".env">
	</div>
{{- end }}
	<div class="field">
		<label for="env-content">Content <span class="faint">KEY=value, one per line, # comments</span></label>
		<textarea id="env-content" name="content" rows="18" spellcheck="false" autocomplete="off">{{ $f.Content }}</textarea>
	</div>
	<div class="form-foot">
		<a class="btn" href="{{ $.Href $p.ID }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>

<section class="block">
	<h3>Appended secrets <small>added each time task-ui writes the file</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Name</th></tr></thead>
		<tbody>
		{{- range .SecretNames }}
			<tr><td class="key">{{ . }}</td></tr>
		{{- else }}
			<tr class="empty"><td>No secrets reach this folder. <a href="/settings/secrets/new?scope={{ $p.ID }}">Add one</a>.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
</div>
{{ end }}
