{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Secrets</h2>
		<p>Variables added to the environment of every run, kept per workspace. A workspace's shared secret reaches every folder in that workspace and no other; a secret for one folder reaches that folder's jobs and wins over the shared one. Before a run, they are written into the env files the pipeline reads.</p>
	</div>
	<div class="actions">
		<form data-api="/api/secrets/rekey" data-method="POST" data-next="/settings/secrets?notice=Secrets+written+again+for+the+current+keys"><button class="btn" type="submit" title="Write every secret again for its workspace's current key, after a rotation">Re-encrypt</button></form>
		<a class="btn btn-primary" href="/settings/secrets/new">Add secret</a>
	</div>
</div>

<div class="alert"><svg class="ic"><use href="#i-secret"/></svg> Encrypted at rest with each workspace's own key, which its <a href="/settings/workspaces">Keys</a> section rotates; a rotation keeps the old key as a backup and writes the workspace's secrets again. The instance SSH key{{ with .Encryption }} <b>{{ join . ", " }}</b>{{ end }} encrypts only the settings of the integrations.</div>

{{- range .SecretGroups }}
{{- $g := . }}
<section class="block secret-group">
	<div class="block-head">
		<h3>{{ .Title }} {{ if ne .Title .Workspace }}<span class="badge">{{ .Workspace }}</span>{{ end }}</h3>
		<div class="actions">
			<a class="btn btn-sm" href="/settings/secrets/new?scope={{ .Shared }}">Add secret</a>
		</div>
	</div>
	{{- with .Key }}
	<p class="dim">Encrypted with <span class="mono-break">{{ .Fingerprint }}</span>.</p>
	{{- else }}
	<p class="dim"><span class="badge bad">no key</span> The workspace has no encryption key yet: <a href="/settings/workspaces/{{ $g.Workspace }}">generate one</a> before adding a secret.</p>
	{{- end }}
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Scope</th><th>Name</th><th>Value</th><th>Updated</th><th>Actions</th></tr></thead>
		<tbody>
		{{- range .Secrets }}
			<tr>
				<td class="nowrap">{{ if eq .Scope $g.Shared }}<span class="badge signal">every folder</span>{{ else }}<a href="{{ $.Href .Scope }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .Scope }}</a>{{ end }}</td>
				<td class="key">{{ .Name }}</td>
				<td class="faint"><svg class="ic"><use href="#i-secret"/></svg> {{ len .Content }} characters</td>
				<td class="nowrap" title="{{ stamp .UpdatedAt }}">{{ ago .UpdatedAt }}</td>
				<td class="actions">
					<a class="btn-icon" href="/settings/secrets/edit?scope={{ .Scope }}&name={{ .Name }}" title="Replace" aria-label="Replace {{ .Name }}"><svg class="ic"><use href="#i-edit"/></svg></a>
					<button type="button" class="btn-icon danger" data-delete="/api/secrets/{{ .ID }}" data-confirm="Delete {{ .Name }} from {{ if eq .Scope $g.Shared }}every folder in {{ $g.Workspace }}{{ else }}{{ .Scope }}{{ end }}? Runs stop receiving it." data-next="reload" title="Delete" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">No secrets in {{ .Workspace }} yet. Add one for every folder in it, or for a single folder.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
{{ end }}
