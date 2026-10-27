{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Permissions</h2>
		<p>What each role may do. Click a box to cycle it: inherit, ✓ allow, ✕ deny. A user's own allow or deny, on their profile, decides first.</p>
	</div>
</div>

<div class="alert">
	Admins can do everything; you can't change that. <b>Allow All</b> sets every rule in its row, and a rule you set on its own overrides it. Folder access and each job's allow list still limit what a user can see and run.
</div>

{{- range .Matrices }}
<section class="block matrix-role" id="{{ .ID }}">
	<h3><span class="avatar sm role-{{ .Title }}" aria-hidden="true">{{ initials .Title }}</span> {{ .Title }} {{ if .Fixed }}<small>can do everything</small>{{ else if not .Editable }}<small>read only</small>{{ end }}</h3>
	{{- $members := index $.RoleMembers .Title }}
	<p class="role-members">
		<span class="label">Users</span>
		{{- range $members.Users }} <a class="key-chip" href="/users/{{ . }}">{{ . }}</a>{{ else }} <span class="faint">none</span>{{ end }}
		{{- if $.Can "user/edit" }}{{ if $.CanAddUsers }} <a href="/settings/users/new">Add user</a>{{ end }}{{ end }}
	</p>
	{{- if $.Can "key/view" }}
	<p class="role-members">
		<span class="label">Keys</span>
		{{- range $members.Keys }} <a class="key-chip" href="/settings/keys/{{ . }}">{{ . }}</a>{{ else }} <span class="faint">none</span>{{ end }}
		{{- if $.Can "key/edit" }} <a href="/settings/keys/generate">Generate key</a>{{ end }}
	</p>
	{{- end }}
	{{- template "matrix" (list $ .) }}
</section>
{{- end }}
{{ end }}
