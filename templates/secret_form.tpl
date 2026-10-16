{{ define "content" }}
{{- $s := .Secret }}
{{ template "crumbs" (list (crumb "Settings" "/settings") (crumb "Secrets" "/settings/secrets") (crumb .Title "")) }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>The value is stored under {{ .EnvPath }}{{ if not .EnvPath }}the env path{{ end }} and never shown again in full.</p>
	</div>
</div>

{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}

<form method="post" action="/settings/secrets/save" class="form">
	<div class="field">
		<label for="s-scope">Folder</label>
		<select id="s-scope" name="scope">
			<option value="all"{{ if eq $s.Scope "all" }} selected{{ end }}>all folders</option>
			{{- range .Projects }}
			<option value="{{ .ID }}"{{ if eq $s.Scope .ID }} selected{{ end }}>{{ .ID }}</option>
			{{- end }}
		</select>
	</div>
	<div class="field">
		<label for="s-key">Key</label>
		<input type="text" id="s-key" name="key" value="{{ $s.Key }}" pattern="[A-Za-z_][A-Za-z0-9_]*" spellcheck="false" autocomplete="off" required{{ if $s.Key }} readonly{{ end }}>
	</div>
	<div class="field">
		<label for="s-value">Value</label>
		<textarea id="s-value" name="value" rows="3" spellcheck="false" autocomplete="off">{{ $s.Value }}</textarea>
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/secrets">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
