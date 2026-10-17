{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Users</h2>
		<p>Who may sign in. Viewers read history and watch runs, runners start them where the allow lists permit, admins manage the instance. A user limited to folders sees only those.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/settings/users/new">Add user</a>
	</div>
</div>

<div class="table-wrap">
<table class="table">
	<thead><tr><th>User</th><th>Role</th><th>Folders</th><th>Created</th><th>Actions</th></tr></thead>
	<tbody>
	{{- range .Users }}
		<tr>
			<td class="key nowrap"><span class="avatar sm">{{ .Initials }}</span> {{ .Name }}</td>
			<td><span class="badge{{ if .CanAdmin }} signal{{ end }}">{{ .Role }}</span></td>
			<td class="desc">{{ with .FolderList }}{{ join . ", " }}{{ else }}<span class="faint">all</span>{{ end }}</td>
			<td class="nowrap" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
			<td class="actions">
				<a class="btn-icon" href="/settings/users/{{ .ID }}" title="Edit" aria-label="Edit {{ .Name }}"><svg class="ic"><use href="#i-edit"/></svg></a>
				{{- if ne .ID $.User.ID }}
				<button type="button" class="btn-icon danger" data-delete="/api/users/{{ .ID }}" data-confirm="Delete {{ .Name }}? Their runs stay in the history." data-next="reload" title="Delete" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>
				{{- end }}
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">No users. Add one, or list them in task-ui.yml.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{ end }}
