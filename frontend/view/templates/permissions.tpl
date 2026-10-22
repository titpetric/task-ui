{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Permissions</h2>
		<p>Set what each role can do, using the rules in <code>config/permissions.yml</code>. Click a box to cycle it: empty (inherit), ✓ allow, ✕ deny. A user's own settings on their profile add to their role's. A deny always wins, and task-ui refuses anything not allowed.</p>
	</div>
</div>

<div class="alert">
	Admins can do everything; you can't change that. <b>Allow All</b> sets every rule in its row, and a rule you set on its own overrides it. Folder access and each job's allow list still limit what a user can see and run.
</div>

{{- range .Matrices }}
<section class="block matrix-role" id="{{ .ID }}">
	<h3><span class="avatar sm role-{{ .Title }}" aria-hidden="true">{{ initials .Title }}</span> {{ .Title }} {{ if .Fixed }}<small>can do everything</small>{{ else if not .Editable }}<small>read only</small>{{ end }}</h3>
	{{- template "matrix" (list $ .) }}
</section>
{{- end }}
{{ end }}
