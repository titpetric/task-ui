{{ define "content" }}
{{- $s := .Secret }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>The value is encrypted with the key of its workspace and never shown again. A secret for a workspace reaches every project in it and no other workspace. Saving a name that exists replaces its value. A workspace without an encryption key stores no secrets: add one under <a href="/settings/keys">Keys</a>.</p>
	</div>
</div>

<form class="form" data-api="/api/secrets" data-method="POST" data-next="/settings/secrets?notice=Secret+saved">
	<div class="field">
		<label for="s-scope">Scope</label>
		<select id="s-scope" name="scope">
			{{- range .SecretGroups }}
			{{- $off := not .Key }}
			<optgroup label="{{ .Title }}{{ if $off }}: no encryption key{{ end }}"{{ if $off }} disabled{{ end }}>
				<option value="{{ .Shared }}"{{ if eq $s.Scope .Shared }} selected{{ end }}{{ if $off }} disabled{{ end }}>every project in {{ .Workspace }}</option>
				{{- range .Folders }}
				<option value="{{ . }}"{{ if eq $s.Scope . }} selected{{ end }}{{ if $off }} disabled{{ end }}>{{ . }}</option>
				{{- end }}
			</optgroup>
			{{- end }}
		</select>
	</div>
	<div class="field">
		<label for="s-name">Name</label>
		<input type="text" id="s-name" name="name" value="{{ $s.Name }}" pattern="[A-Za-z_][A-Za-z0-9_]*" spellcheck="false" autocomplete="off" required{{ if $s.Name }} readonly{{ end }}>
	</div>
	<div class="field">
		<label for="s-content">Value</label>
		<textarea id="s-content" name="content" rows="3" spellcheck="false" autocomplete="off" required></textarea>
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/secrets">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
