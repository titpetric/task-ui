{{ define "content" }}
{{- $u := .Account }}
{{- $all := or (not $u) (not $u.FolderList) }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		{{- if $u }}
		<p>{{ template "avatar" (list $ $u.Name) }} <b>{{ $u.Name }}</b>, <span class="badge{{ if $u.CanAdmin }} signal{{ end }}">{{ $u.Role }}</span>. The name and the password come from the password file, the role from <code>TASK_UI_ADMINS</code> and <code>TASK_UI_VIEWERS</code>; here you pick the folders {{ $u.Name }} sees.</p>
		{{- else }}
		<p>Adds a user to the password file, who can sign in at once. They are a runner unless <code>TASK_UI_ADMINS</code> or <code>TASK_UI_VIEWERS</code> names them.</p>
		{{- end }}
	</div>
</div>

{{- if or (not $u) (.Can "user/view") }}
<form class="form" data-api="/api/users{{ with $u }}/{{ .ID }}{{ end }}" data-method="{{ if $u }}PUT{{ else }}POST{{ end }}" data-next="/settings/users?notice={{ if $u }}Folders+saved{{ else }}User+created{{ end }}">
	{{- if not $u }}
	<div class="field">
		<label for="u-name">Name <span class="faint">lowercase letters, digits and dashes; it also names their SSH keys</span></label>
		<input type="text" id="u-name" name="name" required pattern="[a-z0-9][a-z0-9\-]{0,31}" spellcheck="false" autocomplete="off" autofocus>
	</div>
	<div class="field">
		<label for="u-password">Password <span class="faint">at least 8 characters, stored as a bcrypt hash</span></label>
		<input type="password" id="u-password" name="password" required minlength="8" autocomplete="new-password">
	</div>
	{{- end }}
	<div class="field" data-folder-pick>
		<span class="label">Folders</span>
		<label class="check"><input type="radio" name="scope" value="all"{{ if $all }} checked{{ end }}> All projects</label>
		<label class="check"><input type="radio" name="scope" value="custom"{{ if not $all }} checked{{ end }}> Custom</label>
		<div class="folder-choices">
			{{- range .FolderChoices }}
			<label class="check"><input type="checkbox" name="folders" value="{{ .ID }}" data-type="set"{{ if .Checked }} checked{{ end }}{{ if $all }} disabled{{ end }}> <span class="mono">{{ .ID }}</span></label>
			{{- else }}
			<span class="faint">No folders yet.</span>
			{{- end }}
		</div>
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/users">Cancel</a>
		<button class="btn btn-primary" type="submit">{{ if $u }}Save{{ else }}Create user{{ end }}</button>
	</div>
</form>
{{- end }}
{{- with $u }}
{{ template "mcp-access" (list $ .) }}
{{- end }}
{{ end }}
