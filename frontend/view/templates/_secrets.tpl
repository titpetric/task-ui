{{- /* secretfields are the fields of a variable's form, (list secret
     groups): the scope it applies to, every workspace or one, its name,
     and its value. A secret's value never shows again, so an edit leaves
     the field empty and keeps the value; an env value is filled in and
     may be empty. The is_secret box decides which it is. */}}
{{- define "secretfields" }}
{{- $s := index . 0 }}{{ $groups := index . 1 }}
{{- if $s.ID }}<input type="hidden" name="id" value="{{ $s.ID }}">{{ end }}
	<div class="field">
		<label for="s-scope">Applies to</label>
		<select id="s-scope" name="scope">
			<option value="all"{{ if eq $s.Scope "all" }} selected{{ end }}>All workspaces</option>
			{{- range $groups }}
			<option value="{{ .Shared }}"{{ if eq $s.Scope .Shared }} selected{{ end }}>{{ .Title }}</option>
			{{- end }}
		</select>
	</div>
	<div class="field">
		<label for="s-name">Name</label>
		<input type="text" id="s-name" name="name" value="{{ $s.Name }}" pattern="[A-Za-z_][A-Za-z0-9_]*" spellcheck="false" autocomplete="off" required>
	</div>
	<div class="field">
		<label for="s-content">Value <span class="faint">{{ if and $s.ID $s.IsSecret }}blank keeps the current value{{ else }}what runs get; may be empty{{ end }}</span></label>
		{{- if and $s.ID $s.IsSecret }}
		<textarea id="s-content" name="content" rows="3" spellcheck="false" autocomplete="new-password" placeholder="unchanged"></textarea>
		{{- else }}
		<textarea id="s-content" name="content" rows="3" spellcheck="false" autocomplete="off">{{ $s.Content }}</textarea>
		{{- end }}
	</div>
	<label class="check"><input type="checkbox" name="is_secret"{{ if or (not $s.ID) $s.IsSecret }} checked{{ end }}> Secret <span class="faint">the value is encrypted and never shown again; an env value is listed on the page</span></label>
{{- end }}
