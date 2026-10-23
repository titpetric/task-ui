{{ define "content" }}
{{- $u := .Account }}{{ $pr := .Profile }}
{{- $all := not $u.FolderList }}
<div class="page-head profile-head">
	<div class="profile-who">
		<span class="avatar lg role-{{ $u.EffectiveRole }}" aria-hidden="true">{{ $u.Initials }}</span>
		<div>
			<h2>{{ $u.Name }} <span class="badge{{ if $u.CanAdmin }} signal{{ end }}">{{ $u.EffectiveRole }}</span></h2>
			<p>{{ if $pr.Own }}Your recent runs and notes, your folders and your permissions.{{ else }}Recent runs and notes by {{ $u.Name }} that you can see, their folders and their permissions.{{ end }} Added {{ ago $u.CreatedAt }}.</p>
		</div>
	</div>
	<div class="actions">
		<a class="btn" href="/runs?user={{ $u.Name }}"><svg class="ic"><use href="#i-list"/></svg>Runs</a>
		{{ template "menu" (menu (print "Actions on " $u.Name)
			(and .CanAddUsers (.Can "user/delete") (not $u.CanAdmin) .User (ne $u.ID .User.ID) (danger "trash" "Delete" "data-delete" (print "/api/users/" $u.ID) "data-confirm" (print "Delete the user " $u.Name "? This removes them from the password file and ends their sessions and tokens right away. Their runs, notes and history keep the name.") "data-confirm-ok" "Delete" "data-next" "/settings/users?notice=User+deleted"))
		) }}
	</div>
</div>

<section class="block profile-activity">
	<h3>Activity <small>the latest, in folders you can see</small></h3>
	{{- if not $pr.Activity }}
	<p class="faint">Nothing you can see yet.</p>
	{{- end }}
	{{- with $pr.Runs }}
	<h4>Runs</h4>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Job</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
		{{- range . }}{{ $r := . }}
			<tr>
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td class="key"><a href="{{ with $.RunJobHref .Folder .Dir .Job }}{{ . }}{{ else }}/runs/{{ $r.ID }}{{ end }}">{{ .Name }}</a></td>
				<td class="nowrap" data-label="Took">{{ duration .Elapsed }}</td>
				<td data-label="Result">{{ template "state" . }}</td>
				<td class="actions"><a class="btn btn-sm" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">{{ if not (ended .State) }}Attach{{ else }}View{{ end }}</a>{{ template "menu" (menu (print "Actions on the run of " .Job)
					(detail "Run" .ID "mono" "")
					(detail "Started" (stamp .StartedAt))
					(item "file" "Open the run's page" "href" (print "/runs/" .ID))
					(and .CastPath (ended .State) (item "play" "Replay the recording" "data-replay" .ID "data-replay-title" .Name "data-replay-who" .UserName))
				) }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	<p class="more-link"><a href="/runs?user={{ $u.Name }}">Every run by {{ $u.Name }}</a></p>
	{{- end }}
	{{- with $pr.FlowRuns }}
	<h4>Workflow runs</h4>
	<ul class="profile-list">
		{{- range . }}
		<li><i class="dot {{ .State }}"></i><a href="/workflows/{{ .WorkflowID }}/runs/{{ .ID }}"><svg class="ic"><use href="#i-flow"/></svg>{{ .Name }}</a> <span class="faint">{{ .State }}</span><span class="faint when" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</span></li>
		{{- end }}
	</ul>
	{{- end }}
	{{- with $pr.Notes }}
	<h4>Notes</h4>
	<ul class="profile-list">
		{{- range . }}
		<li><svg class="ic"><use href="#i-note"/></svg><a class="excerpt" href="/notes#note-{{ .ID }}">{{ .Body }}</a>{{ with .State }} <span class="badge note-state {{ . }}">{{ . }}</span>{{ end }}<span class="faint when" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span></li>
		{{- end }}
	</ul>
	{{- end }}
	{{- with $pr.Prompts }}
	<h4>Prompts</h4>
	<ul class="profile-list">
		{{- range . }}
		{{- $href := $.PlaceHref .Dir }}{{ if .AboutRunID }}{{ $href = printf "/runs/%s" .AboutRunID }}{{ end }}
		<li><svg class="ic"><use href="#i-terminal"/></svg><a class="excerpt" href="{{ $href }}#note-{{ .ID }}">{{ .Body }}</a> <span class="faint mono">{{ .Dir }}/</span>{{ with .State }} <span class="badge note-state {{ . }}">{{ . }}</span>{{ end }}<span class="faint when" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span></li>
		{{- end }}
	</ul>
	{{- end }}
</section>

<section class="block" id="roles">
	<h3>Roles <small>what {{ $u.Name }} is</small></h3>
	{{- if .User.CanAdmin }}
	<form class="form" data-api="/api/users/{{ $u.ID }}" data-method="PUT" data-next="/users/{{ $u.Name }}?notice=Roles+saved#roles">
		<label class="check"><input type="checkbox" name="admin"{{ if $u.Admin }} checked{{ end }}> Administrative privileges (all) <span class="faint">every permission, whatever the role; the last admin keeps it</span></label>
		<div class="field">
			<label for="u-role">Role <span class="faint">without the admin flag</span></label>
			<select id="u-role" name="role">
				<option value="runner"{{ if eq $u.Role "runner" }} selected{{ end }}>runner</option>
				<option value="viewer"{{ if eq $u.Role "viewer" }} selected{{ end }}>viewer</option>
			</select>
		</div>
		<label class="check"><input type="checkbox" name="mcp"{{ if $u.Mcp }} checked{{ end }}> MCP agent user <span class="faint">a token reaches the MCP endpoint; an agent learns of work from the notes that name the user</span></label>
		<label class="check"><input type="checkbox" name="agent_reply"{{ if $u.AgentReply }} checked{{ end }}> Run instance Claude when mentioned <span class="faint">a note naming {{ $u.Name }} starts the Notes agent as them; needs MCP agent user</span></label>
		<div class="form-foot">
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
	<form class="form" data-api="/api/users/{{ $u.ID }}" data-method="PUT" data-next="/users/{{ $u.Name }}?notice=Folders+saved#folders">
		{{- template "folder-pick" (list .FolderChoices $all false) }}
		<div class="form-foot">
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
		{{- range . }}{{ template "keyrow" (list $ . (print "/users/" $u.Name)) }}{{ end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{- end }}

{{ template "mcp-access" (list $ $u) }}
{{ end }}
