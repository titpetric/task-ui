{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Users</h2>
		<p>Who may sign in, from the password file. Viewers read history and watch runs, runners start them where the allow lists permit, admins manage the instance. What is set here is which folders a user sees.</p>
	</div>
	{{- if .CanAddUsers }}
	<div class="actions">
		<a class="btn btn-primary" href="/settings/users/new">Create user</a>
	</div>
	{{- end }}
</div>

<div class="alert">
	Users live in the password file, htpasswd style, read again whenever it changes. <b>Create user</b> adds one, as does <code>task-ui passwd &lt;user&gt;</code>, which also resets a password; removing the line removes the user. The roles come from <code>TASK_UI_ADMINS</code> and <code>TASK_UI_VIEWERS</code>; everyone else is a runner.
</div>

<div class="table-wrap">
<table class="table cards">
	<thead><tr><th>User</th><th>Role</th><th>Folders</th><th>Since</th><th></th></tr></thead>
	<tbody>
	{{- range .Users }}
		<tr>
			<td class="key nowrap"><span class="avatar sm">{{ .Initials }}</span> {{ .Name }}</td>
			<td data-label="Role"><span class="badge{{ if .CanAdmin }} signal{{ end }}">{{ .Role }}</span></td>
			<td class="desc" data-label="Folders">{{ with .FolderList }}{{ join . ", " }}{{ else }}<span class="faint">all</span>{{ end }}</td>
			<td class="nowrap" data-label="Since" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
			<td class="actions">
				<a class="btn-icon" href="/settings/users/{{ .ID }}" title="Folders of {{ .Name }}" aria-label="Folders of {{ .Name }}"><svg class="ic"><use href="#i-edit"/></svg></a>
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">No users. Add one with <code>task-ui passwd &lt;user&gt;</code> and it shows up here.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{ end }}
