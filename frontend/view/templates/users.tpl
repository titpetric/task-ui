{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Users</h2>
		<p>Who can sign in, their role and folders. Click a user for their profile.</p>
	</div>
	{{- if .CanAddUsers }}
	<div class="actions">
		<a class="btn btn-primary" href="/settings/users/new">Create user</a>
	</div>
	{{- end }}
</div>

<div class="alert">
	Users come from the password file; <b>Create user</b> or <code>task-ui passwd &lt;user&gt;</code> adds one or resets a password. A new user starts as a viewer. An admin sets the admin flag, the role and MCP on the user's profile, or with <code>task-ui user</code>.
</div>

<div class="table-wrap">
<table class="table cards">
	<thead><tr><th>User</th><th>Role</th><th>Flags</th><th>Folders</th><th>Since</th><th></th></tr></thead>
	<tbody>
	{{- range .Users }}
		<tr class="row-link" data-href="/users/{{ .Name }}">
			<td class="key nowrap"><span class="user-ref">{{ template "avatar" (list $ .Name) }} <a class="user-name" href="/users/{{ .Name }}">{{ .Name }}</a>{{ template "user-menu" (list $ .Name) }}</span></td>
			<td data-label="Role"><span class="badge{{ if .CanAdmin }} signal{{ end }}">{{ .EffectiveRole }}</span></td>
			<td class="nowrap" data-label="Flags">{{ if .Mcp }}<span class="badge" title="MCP agent user">mcp</span>{{ end }}{{ if .AgentReply }} <span class="badge" title="Runs instance Claude when mentioned">agent</span>{{ end }}{{ if not .Mcp }}<span class="faint">none</span>{{ end }}</td>
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
		<tr class="empty"><td colspan="6">No users yet. Add one with <code>task-ui passwd &lt;user&gt;</code> and it appears here.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>

{{ template "mcp-connect" . }}
{{ end }}
