{{ define "content" }}
{{- $s := .Secret }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>The value is encrypted with the key of the workspace its scope is in and never shown again in full. A secret for a workspace reaches every folder in it and no other workspace. Saving a name that exists replaces its value.</p>
	</div>
</div>

<form class="form" data-api="/api/secrets" data-method="POST" data-next="/settings/secrets?notice=Secret+saved">
	<div class="field">
		<label for="s-scope">Scope</label>
		<select id="s-scope" name="scope">
			{{- range .SecretGroups }}
			<optgroup label="{{ .Title }}">
				<option value="{{ .Shared }}"{{ if eq $s.Scope .Shared }} selected{{ end }}>every folder in {{ .Workspace }}</option>
				{{- range .Folders }}
				<option value="{{ . }}"{{ if eq $s.Scope . }} selected{{ end }}>{{ . }}</option>
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
