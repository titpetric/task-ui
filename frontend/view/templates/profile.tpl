{{ define "content" }}
{{- $u := .Account }}{{ $pr := .Profile }}
{{- $all := not $u.FolderList }}
<div class="page-head profile-head">
	<div class="profile-who">
		<span class="avatar lg role-{{ $u.Role }}" aria-hidden="true">{{ $u.Initials }}</span>
		<div>
			<h2>{{ $u.Name }} <span class="badge{{ if $u.CanAdmin }} signal{{ end }}">{{ $u.Role }}</span></h2>
			<p>{{ if $pr.Own }}Your recent runs and notes, your folders and your permissions.{{ else }}Recent runs and notes by {{ $u.Name }} that you can see, their folders and their permissions.{{ end }} Added {{ ago $u.CreatedAt }}.</p>
		</div>
	</div>
	<div class="actions">
		<a class="btn" href="/runs?user={{ $u.Name }}"><svg class="ic"><use href="#i-list"/></svg>Runs</a>
		{{- if and .CanAddUsers (.Can "user/delete") (not $u.CanAdmin) .User (ne $u.ID .User.ID) }}
		<button type="button" class="btn btn-danger" data-delete="/api/users/{{ $u.ID }}" data-confirm="Delete the user {{ $u.Name }}? This removes them from the password file and ends their sessions and tokens right away. Their runs, notes and history keep the name." data-confirm-ok="Delete" data-next="/settings/users?notice=User+deleted"><svg class="ic"><use href="#i-trash"/></svg>Delete user</button>
		{{- end }}
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
				<td class="actions"><a class="btn btn-sm" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">{{ if eq .State "running" }}Attach{{ else }}View{{ end }}</a>{{ template "replay" . }}</td>
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
	<h3>Permissions <small>{{ $u.Name }}'s own settings, on top of the {{ $u.Role }} role</small></h3>
	<p>The role comes from <code>TASK_UI_ADMINS</code> and <code>TASK_UI_VIEWERS</code>; everyone else is a runner. {{ if $.Can "user/view" }}<a href="/settings/permissions#role-{{ $u.Role }}">The {{ $u.Role }} role's settings</a>{{ else }}The {{ $u.Role }} role's settings{{ end }} apply first. A box here adds an allow or a deny for {{ $u.Name }} alone, and a deny in either place wins. The mark after each box shows the result.{{ if .Fixed }} Admins can do everything; you can't change that.{{ end }}</p>
	{{- template "matrix" (list $ .) }}
</section>
{{- end }}

{{ template "mcp-access" (list $ $u) }}
{{ end }}
