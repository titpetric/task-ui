{{ define "content" }}
{{- $u := .Account }}{{ $pr := .Profile }}
{{- $all := not $u.FolderList }}{{ $view := print "/users/" $u.Name }}{{ $edit := print $view "/edit" }}
<div class="page-head">
	<div>
		<h2>Edit {{ $u.Name }}</h2>
		<p>The roles, folders, permissions, keys and MCP access of {{ $u.Name }}.</p>
	</div>
	<div class="actions">
		<a class="btn" href="{{ $view }}"><svg class="ic"><use href="#i-users"/></svg>View profile</a>
	</div>
</div>

<section class="block" id="roles">
	<h3>Roles <small>what {{ $u.Name }} is</small></h3>
	{{- if .User.CanAdmin }}
	<form class="form" data-api="/api/users/{{ $u.ID }}" data-method="PUT" data-next="{{ $edit }}?notice=Roles+saved#roles">
		<div class="field">
			<label class="check"><input type="checkbox" name="admin"{{ if $u.Admin }} checked{{ end }}> Administrative privileges (all) <span class="faint">every permission, whatever the role; the last admin keeps it</span></label>
		</div>
		<div class="field">
			<label for="u-role">Role <span class="faint">without the admin flag</span></label>
			<select id="u-role" name="role">
				<option value="runner"{{ if eq $u.Role "runner" }} selected{{ end }}>runner</option>
				<option value="viewer"{{ if eq $u.Role "viewer" }} selected{{ end }}>viewer</option>
			</select>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="mcp"{{ if $u.Mcp }} checked{{ end }}> MCP agent user <span class="faint">a token reaches the MCP endpoint; an agent learns of work from the notes that name the user</span></label>
		</div>
		<div class="field">
			<label class="check"><input type="checkbox" name="agent_reply"{{ if $u.AgentReply }} checked{{ end }}> Run instance Claude when mentioned <span class="faint">a note naming {{ $u.Name }} starts the Notes agent as them; needs MCP agent user</span></label>
		</div>
		<div class="form-foot">
			<a class="btn" href="{{ $view }}">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
	{{- else }}
	<p>{{ if $u.Admin }}Administrative privileges (all). {{ end }}Role: {{ $u.Role }}.{{ if $u.Mcp }} MCP agent user.{{ end }}{{ if $u.AgentReply }} Runs instance Claude when mentioned.{{ end }}</p>
	{{- end }}
</section>

<section class="block" id="folders">
	<h3>Folders <small>what {{ $u.Name }} can see</small></h3>
	{{- if .Can "user/edit" }}
	<form class="form" data-api="/api/users/{{ $u.ID }}" data-method="PUT" data-next="{{ $edit }}?notice=Folders+saved#folders">
		{{- template "folder-pick" (list .FolderChoices $all false) }}
		<div class="form-foot">
			<a class="btn" href="{{ $view }}">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
	{{- else }}
	<p>{{ with $u.FolderList }}{{ range $i, $f := . }}{{ if $i }}, {{ end }}<span class="mono">{{ $f }}</span>{{ end }}{{ else }}All projects.{{ end }}</p>
	{{- end }}
</section>

{{- with $pr.Matrix }}
<section class="block" id="permissions">
	<h3>Permissions <small>{{ $u.Name }}'s own rules, before the role's</small></h3>
	<p>Role: {{ if $.Can "user/view" }}<a href="/settings/permissions#role-{{ $u.EffectiveRole }}">{{ $u.EffectiveRole }}</a>{{ else }}{{ $u.EffectiveRole }}{{ end }}.{{ if .Fixed }} Admins can do everything; this can't change.{{ end }}</p>
	{{- if not .Fixed }}
	<p>An allow or a deny here is {{ $u.Name }}'s own and decides first, whatever the role says. Inherit, the empty box, leaves it to the role's rules, where a deny wins over an allow. The mark after each box shows the result.</p>
	{{- end }}
	{{- template "matrix" (list $ .) }}
</section>
{{- end }}

{{- if or $pr.Own (.Can "key/view") }}
<section class="block" id="keys">
	{{- template "blockhead" (head "Keys" (print "linked to " $u.Name ", in the ssh agent of their runs") (and (not .Keys) (print "No key linked to " $u.Name "; the role's keys come with the role."))
		(and (.Can "key/edit") (item "key" "Generate key" "href" (print "/settings/keys/generate?for=" $u.Name)))) }}
	{{- with .Keys }}
	<div class="table-wrap">
	<table class="table keys key-table cards">
		{{ template "keyhead" }}
		<tbody>
		{{- range . }}{{ template "keyrow" (list $ . $edit) }}{{ end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{- end }}

{{ template "mcp-access" (list $ $u) }}
{{ end }}
