{{ define "content" }}
{{- $s := .Secret }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>The value is encrypted with the instance SSH {{ if eq (len .Encryption) 1 }}key{{ else }}keys{{ end }} {{ if .Encryption }}<b>{{ join .Encryption ", " }}</b>{{ else }}<span class="badge bad">none yet</span>{{ end }} and never shown again in full. Saving a name that exists replaces its value.</p>
	</div>
</div>

<form class="form" data-api="/api/secrets" data-method="POST" data-next="/settings/secrets?notice=Secret+saved">
	<div class="field">
		<label for="s-scope">Folder</label>
		<select id="s-scope" name="scope">
			<option value="all"{{ if eq $s.Scope "all" }} selected{{ end }}>all folders</option>
			{{- range .Folders }}
			<option value="{{ .ID }}"{{ if eq $s.Scope .ID }} selected{{ end }}>{{ .ID }}</option>
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
