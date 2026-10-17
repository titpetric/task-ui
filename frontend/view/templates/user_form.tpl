{{ define "content" }}
{{- $u := .Account }}
<div class="page-head">
	<div>
		<h2>{{ .Title }}</h2>
		<p>{{ if $u }}Change the role, the folders or the password. An empty password keeps the one there is.{{ else }}A name, a password and a role. Limit the folders to show the user a part of the instance.{{ end }}</p>
	</div>
</div>

<form class="form" data-api="{{ if $u }}/api/users/{{ $u.ID }}{{ else }}/api/users{{ end }}" data-method="{{ if $u }}PUT{{ else }}POST{{ end }}" data-next="/settings/users?notice=User+saved">
	<div class="field">
		<label for="u-name">Name</label>
		<input type="text" id="u-name" name="name" value="{{ if $u }}{{ $u.Name }}{{ end }}" autocomplete="off" spellcheck="false" required{{ if $u }} readonly{{ end }}>
	</div>
	<div class="field">
		<label for="u-password">Password{{ if $u }} <span class="faint">leave empty to keep</span>{{ end }}</label>
		<input type="password" id="u-password" name="password" autocomplete="new-password"{{ if not $u }} required{{ end }}>
	</div>
	<div class="field">
		<label for="u-role">Role</label>
		<select id="u-role" name="role">
			{{- range roles }}<option value="{{ . }}"{{ if and $u (eq $u.Role .) }} selected{{ end }}>{{ . }}</option>{{ end }}
		</select>
	</div>
	<div class="field">
		<label for="u-folders">Folders <span class="faint">comma separated; empty shows all</span></label>
		<input type="text" id="u-folders" name="folders" data-type="list" value="{{ if $u }}{{ join $u.FolderList ", " }}{{ end }}" placeholder="infra, website" spellcheck="false" autocomplete="off">
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/users">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
