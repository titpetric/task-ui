{{ define "content" }}
{{- $p := .Project }}{{ $f := .EnvFile }}
{{ template "crumbs" (list (crumb $p.ID (print "/f/" $p.ID)) (crumb $f.Path "")) }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-file"/></svg> {{ $f.Path }}</h2>
		<p>What the pipeline reads with env.include. Saving writes the file into the folder with the secrets below appended; the same happens before every run.</p>
	</div>
	<div class="dim nowrap">{{ if $f.Managed }}<span class="badge ok">managed</span>{{ else if $f.Exists }}<span class="badge">from the folder</span>{{ else }}<span class="badge bad">missing</span>{{ end }}</div>
</div>

{{- if .Notice }}<div class="alert ok" role="status">{{ .Notice }}</div>{{ end }}
{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}

<div class="grid-2 wide">
<form method="post" action="/f/{{ $p.ID }}/env/{{ $f.Path }}" class="form">
	<div class="field">
		<label for="env-content">Content <span class="faint">KEY=value, one per line, # comments</span></label>
		<textarea id="env-content" name="content" rows="18" spellcheck="false" autocomplete="off">{{ $f.Content }}</textarea>
	</div>
	<div class="form-foot">
		<a class="btn" href="/f/{{ $p.ID }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Save and generate</button>
	</div>
</form>

<section class="block">
	<h3>Appended secrets <small>at generation time</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Scope</th><th>Key</th></tr></thead>
		<tbody>
		{{- range .Secrets }}
			<tr><td class="nowrap">{{ if eq .Scope "all" }}<span class="badge signal">all folders</span>{{ else }}{{ .Scope }}{{ end }}</td><td class="key">{{ .Key }}</td></tr>
		{{- else }}
			<tr class="empty"><td colspan="2">No secrets reach this folder. <a href="/settings/secrets/new?scope={{ $p.ID }}">Add one</a>.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
</div>
{{ end }}
