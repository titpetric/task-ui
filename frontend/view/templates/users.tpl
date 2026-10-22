{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Users</h2>
		<p>Who can sign in, from the password file. Viewers read history and watch runs. Runners also start runs where the allow lists permit. Admins manage the instance. On a user's profile, you set which folders they see and their own permissions on top of their role's.</p>
	</div>
	{{- if .CanAddUsers }}
	<div class="actions">
		<a class="btn btn-primary" href="/settings/users/new">Create user</a>
	</div>
	{{- end }}
</div>

<div class="alert">
	Users live in the password file (htpasswd format), and task-ui reads it again whenever it changes. <b>Create user</b> adds a user, and so does <code>task-ui passwd &lt;user&gt;</code>, which also resets a password. To remove a user, delete their line from the file, or use <b>Delete</b> here (not for admins). Roles come from <code>TASK_UI_ADMINS</code> and <code>TASK_UI_VIEWERS</code>; everyone else is a runner. <a href="/settings/permissions">Permissions</a> sets what each role can do.
</div>

<div class="table-wrap">
<table class="table cards">
	<thead><tr><th>User</th><th>Role</th><th>Folders</th><th>Since</th><th></th></tr></thead>
	<tbody>
	{{- range .Users }}
		<tr>
			<td class="key nowrap">{{ template "user" (list $ .Name) }}</td>
			<td data-label="Role"><span class="badge{{ if .CanAdmin }} signal{{ end }}">{{ .Role }}</span></td>
			<td class="desc" data-label="Folders">{{ with .FolderList }}{{ join . ", " }}{{ else }}<span class="faint">all</span>{{ end }}</td>
			<td class="nowrap" data-label="Since" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
			<td class="actions">
				<a class="btn-icon" href="/users/{{ .Name }}" title="Profile of {{ .Name }}" aria-label="Profile of {{ .Name }}"><svg class="ic"><use href="#i-edit"/></svg></a>
				{{- if and $.CanAddUsers ($.Can "user/delete") (not .CanAdmin) $.User (ne .ID $.User.ID) }}
				<button type="button" class="btn-icon danger" data-delete="/api/users/{{ .ID }}" data-confirm="Delete the user {{ .Name }}? This removes them from the password file and ends their sessions and tokens right away. Their runs, notes and history keep the name." data-confirm-ok="Delete" data-next="/settings/users?notice=User+deleted" title="Delete {{ .Name }}" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>
				{{- end }}
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">No users yet. Add one with <code>task-ui passwd &lt;user&gt;</code> and it appears here.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>

{{ template "mcp-connect" . }}
{{ end }}
