{{ define "content" }}
{{- $keyed := false }}{{ range .SecretGroups }}{{ if .Key }}{{ $keyed = true }}{{ end }}{{ end }}
<div class="page-head">
	<div>
		<h2>Secrets</h2>
		<p>Variables added to the environment of the runs in a workspace. A workspace's secret reaches every project in it; a project's own secret reaches that project and wins over the workspace's of the same name. A value is never shown again once saved.</p>
	</div>
	<div class="actions">
		{{- if $keyed }}
		<a class="btn btn-primary" href="/settings/secrets/new{{ with .SecretWorkspace }}?scope=workspace:{{ . }}{{ end }}">Add secret</a>
		{{- else }}
		<button class="btn btn-primary" type="button" disabled title="No workspace has an encryption key yet">Add secret</button>
		{{- end }}
	</div>
</div>

{{- with .SecretWorkspace }}
<p class="dim secret-filter">Showing the workspace <b>{{ . }}</b>. <a href="/settings/secrets">Every workspace</a></p>
{{- end }}

{{- range .SecretGroups }}
{{- $g := . }}
<section class="block secret-group" id="ws-{{ .Workspace }}">
	<div class="block-head">
		<h3>{{ .Title }} {{ if ne .Title .Workspace }}<span class="badge">{{ .Workspace }}</span>{{ end }} <small>{{ len .Secrets }} secret{{ if ne (len .Secrets) 1 }}s{{ end }}</small></h3>
		<div class="actions">
			{{- if .Key }}
			<a class="btn btn-sm" href="/settings/secrets/new?scope={{ .Shared }}">Add secret</a>
			<details class="more-menu">
				<summary class="btn-icon" title="More for {{ .Workspace }}" aria-label="More for {{ .Workspace }}"><svg class="ic"><use href="#i-more"/></svg></summary>
				<div class="more-list" role="menu">
					<form data-api="/api/workspaces/{{ .Workspace }}/keys/generate" data-method="POST" data-next="/settings/secrets?rekeyed={rekeyed}{{ if $.SecretWorkspace }}&workspace={{ .Workspace }}{{ end }}#ws-{{ .Workspace }}" data-confirm="Rotate the encryption key of {{ .Workspace }}? task-ui generates a new key, keeps the current one as a backup, and re-encrypts the workspace's {{ len .Secrets }} secret{{ if ne (len .Secrets) 1 }}s{{ end }}. Runs see no change." data-confirm-ok="Rotate">
						<input type="hidden" name="encrypt" value="true" data-type="bool">
						<input type="hidden" name="comment" value="{{ .Key.Comment }}">
						<button class="more-item" type="submit" role="menuitem"><svg class="ic"><use href="#i-rotate"/></svg> Rotate encryption key</button>
					</form>
				</div>
			</details>
			{{- else }}
			<button class="btn btn-sm" type="button" disabled title="Add an encryption key to {{ .Workspace }} first">Add secret</button>
			{{- end }}
		</div>
	</div>
	{{- if not .Key }}
	<div class="alert signal secret-nokey"><svg class="ic"><use href="#i-secret"/></svg> Add an encryption key to store secrets in {{ .Workspace }}. <a href="/settings/keys#ws-{{ .Workspace }}">Keys</a></div>
	{{- end }}
	{{- if or .Key .Secrets }}
	<div class="table-wrap">
	<table class="table secrets">
		<thead><tr><th>Name</th><th>Scope</th><th>Updated</th><th>Actions</th></tr></thead>
		{{- with .SharedSecrets }}
		<tbody>
			<tr class="secret-scope"><th colspan="4">Every project in {{ $g.Workspace }}</th></tr>
			{{- range . }}{{ template "secretrow" (list $ $g .) }}{{ end }}
		</tbody>
		{{- end }}
		{{- range .Projects }}
		<tbody>
			<tr class="secret-scope"><th colspan="4"><a href="{{ $.Href .Folder }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .Folder }}</a></th></tr>
			{{- range .Secrets }}{{ template "secretrow" (list $ $g .) }}{{ end }}
		</tbody>
		{{- end }}
		{{- if not .Secrets }}
		<tbody>
			<tr class="empty"><td colspan="4">No secrets yet. Add one for every project in {{ .Workspace }}, or for a single project.</td></tr>
		</tbody>
		{{- end }}
	</table>
	</div>
	{{- end }}
</section>
{{- end }}
{{ end }}

{{- /* secretrow is a secret's row: its name, its scope, when it changed,
     and replace and delete. The value never shows. */}}
{{- define "secretrow" }}
{{- $g := index . 1 }}{{ $s := index . 2 }}
			<tr>
				<td class="key">{{ $s.Name }}</td>
				<td class="scope faint">{{ $s.Scope }}</td>
				<td class="nowrap" title="{{ stamp $s.UpdatedAt }}">{{ ago $s.UpdatedAt }}</td>
				<td class="actions">
					<a class="btn-icon" href="/settings/secrets/edit?scope={{ $s.Scope }}&name={{ $s.Name }}" title="Replace the value" aria-label="Replace {{ $s.Name }}"><svg class="ic"><use href="#i-edit"/></svg></a>
					<button type="button" class="btn-icon danger" data-delete="/api/secrets/{{ $s.ID }}" data-confirm="Delete {{ $s.Name }} from {{ if eq $s.Scope $g.Shared }}every project in {{ $g.Workspace }}{{ else }}{{ $s.Scope }}{{ end }}? Runs stop receiving it." data-next="reload" title="Delete" aria-label="Delete {{ $s.Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>
				</td>
			</tr>
{{- end }}
