{{ define "content" }}
{{- $u := .Account }}{{ $pr := .Profile }}
<div class="page-head profile-head">
	<div class="profile-who">
		<span class="avatar lg role-{{ $u.EffectiveRole }}" aria-hidden="true">{{ $u.Initials }}</span>
		<div>
			<h2>{{ $u.Name }}</h2>
			<p>{{ if $pr.Own }}Your latest runs, notes and prompts.{{ else }}The latest runs, notes and prompts by {{ $u.Name }} that you can see.{{ end }}{{ with $u.CreatedAt }} Joined {{ ago . }}.{{ end }}</p>
		</div>
	</div>
	<div class="actions">
		{{- if $pr.CanEdit }}
		<a class="btn btn-primary" href="/users/{{ $u.Name }}/edit"><svg class="ic"><use href="#i-edit"/></svg>Edit</a>
		{{- end }}
		{{ template "menu" (menu (print "Actions on " $u.Name)
			(item "list" "Runs by this user" "href" (print "/runs?user=" (urlquery $u.Name)))
			(and .CanAddUsers (.Can "user/delete") (not $u.CanAdmin) .User (ne $u.ID .User.ID) (danger "trash" "Delete" "data-delete" (print "/api/users/" $u.ID) "data-confirm" (print "Delete the user " $u.Name "? This removes them from the password file and ends their sessions and tokens right away. Their runs, notes and history keep the name.") "data-confirm-ok" "Delete" "data-next" "/settings/users?notice=User+deleted"))
		) }}
	</div>
</div>

<section class="block" id="runs">
	{{- template "blockhead" (head "Runs" "the latest, in folders you can see" (and (not $pr.Runs) "No runs yet.")) }}
	{{- with $pr.Runs }}
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Job</th><th>Duration</th><th>Result</th><th class="actions" aria-label="Actions"></th></tr></thead>
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
	<p class="more-link"><a href="/runs?user={{ urlquery $u.Name }}">Every run by {{ $u.Name }}</a></p>
	{{- end }}
</section>

<section class="block" id="workflow-runs">
	{{- template "blockhead" (head "Workflow runs" "the latest you can see" (and (not $pr.FlowRuns) "No workflow runs yet.")) }}
	{{- with $pr.FlowRuns }}
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Workflow</th><th>Result</th><th class="actions" aria-label="Actions"></th></tr></thead>
		<tbody>
		{{- range . }}
			<tr>
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td class="key"><a href="/workflows/{{ .WorkflowID }}"><svg class="ic lead"><use href="#i-flow"/></svg>{{ .Name }}</a></td>
				<td data-label="Result"><span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span></td>
				<td class="actions"><a class="btn btn-sm" href="/workflows/{{ .WorkflowID }}/runs/{{ .ID }}">View</a></td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>

<section class="block" id="notes">
	{{- template "blockhead" (head "Notes" (print "the latest " $u.Name " wrote") (and (not $pr.Notes) "No notes yet.")) }}
	{{- with $pr.Notes }}
	<ul class="profile-list">
		{{- range . }}
		<li><svg class="ic"><use href="#i-note"/></svg><a class="excerpt" href="/notes#note-{{ .ID }}">{{ excerpt .Body }}</a>{{ with .State }} <span class="badge note-state {{ . }}">{{ . }}</span>{{ end }}<span class="faint when" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span></li>
		{{- end }}
	</ul>
	{{- end }}
</section>

<section class="block" id="prompts">
	{{- template "blockhead" (head "Prompts" (print "the latest " $u.Name " asked, in places you can see") (and (not $pr.Prompts) "No prompts yet.")) }}
	{{- with $pr.Prompts }}
	<ul class="profile-list">
		{{- range . }}
		{{- $href := $.PlaceHref .Dir }}{{ if .AboutRunID }}{{ $href = printf "/runs/%s" .AboutRunID }}{{ end }}
		<li><svg class="ic"><use href="#i-terminal"/></svg><a class="excerpt" href="{{ $href }}#note-{{ .ID }}">{{ excerpt .Body }}</a> <span class="faint mono">{{ .Dir }}/</span>{{ with .State }} <span class="badge note-state {{ . }}">{{ . }}</span>{{ end }}<span class="faint when" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span></li>
		{{- end }}
	</ul>
	{{- end }}
</section>

{{- if or $pr.Own (.Can "key/view") }}
{{ template "userkeys" (list . $u (print "/users/" $u.Name)) }}
{{- end }}
{{ end }}
