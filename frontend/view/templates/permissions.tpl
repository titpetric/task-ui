{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Permissions</h2>
		<p>What each role may do, by the rules of <code>config/permissions.yml</code>. A cell inherits (empty), allows (✓) or denies (✕); a click moves it on. A user's own policy, on their profile, adds to their role's: a deny in either wins, and what neither allows is denied.</p>
	</div>
</div>

<div class="alert">
	The admin role may do everything, and no policy changes that. A rule of <b>Allow All</b> sets every rule of its section; a rule set on its own overrides it. Folder visibility and the allow lists of jobs still narrow what a user sees and runs.
</div>

{{- range .Matrices }}
<section class="block matrix-role" id="{{ .ID }}">
	<h3><span class="avatar sm role-{{ .Title }}" aria-hidden="true">{{ initials .Title }}</span> {{ .Title }} {{ if .Fixed }}<small>everything, fixed</small>{{ else if not .Editable }}<small>read only</small>{{ end }}</h3>
	{{- template "matrix" (list $ .) }}
</section>
{{- end }}
{{ end }}
