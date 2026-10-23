{{ define "content" }}
{{- $keyed := false }}{{ range .SecretGroups }}{{ if .Key }}{{ $keyed = true }}{{ end }}{{ end }}
<div class="page-head">
	<div>
		<h2>Secrets</h2>
		<p>Variables for the environment of runs, per workspace or per project; a project secret replaces a workspace one of the same name. Values are never shown again.</p>
	</div>
	<div class="actions">
		{{- if $keyed }}
		<a class="btn btn-primary" href="/settings/secrets/new{{ with .SecretWorkspace }}?scope=workspace:{{ . }}{{ end }}">Add secret</a>
		{{- else }}
		<button class="btn btn-primary" type="button" disabled title="The instance has no encryption key yet">Add secret</button>
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
			{{- else }}
			<button class="btn btn-sm" type="button" disabled title="The instance has no encryption key yet">Add secret</button>
			{{- end }}
		</div>
	</div>
	{{- if not .Key }}
	<div class="alert signal secret-nokey"><svg class="ic"><use href="#i-secret"/></svg> The instance has no encryption key, so it stores no secret. <a href="/settings/instance#keys">Instance</a></div>
	{{- end }}
	{{- if or .Key .Secrets }}
	<div class="table-wrap">
	<table class="table secrets">
		<thead><tr><th>Name</th><th>Scope</th><th>Updated</th><th>Reads</th><th>Last read</th><th class="actions" aria-label="Actions"></th></tr></thead>
		{{- with .SharedSecrets }}
		<tbody>
			<tr class="secret-scope"><th colspan="6">Every project in {{ $g.Workspace }}</th></tr>
			{{- range . }}{{ template "secretrow" (list $ $g .) }}{{ end }}
		</tbody>
		{{- end }}
		{{- range .Projects }}
		<tbody>
			<tr class="secret-scope"><th colspan="6"><a href="{{ $.Href .Folder }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .Folder }}</a></th></tr>
			{{- range .Secrets }}{{ template "secretrow" (list $ $g .) }}{{ end }}
		</tbody>
		{{- end }}
		{{- if not .Secrets }}
		<tbody>
			<tr class="empty"><td colspan="6">No secrets yet. Add one for every project in {{ .Workspace }}, or for a single project.</td></tr>
		</tbody>
		{{- end }}
	</table>
	</div>
	{{- end }}
</section>
{{- end }}
{{ end }}

{{- /* secretrow is a secret's row: its name, its scope, when it changed,
     how many runs read it, linking to its read log, the last run that
     did, and replace and delete. The value never shows. */}}
{{- define "secretrow" }}
{{- $p := index . 0 }}{{ $g := index . 1 }}{{ $s := index . 2 }}
			<tr>
				<td class="key">{{ $s.Name }}</td>
				<td class="scope faint">{{ $s.Scope }}</td>
				<td class="nowrap" title="{{ stamp $s.UpdatedAt }}">{{ ago $s.UpdatedAt }}</td>
				<td class="nowrap">{{ with index $p.SecretReadCounts $s.ID }}<a href="/settings/secrets/{{ $s.ID }}/reads" title="The runs that got {{ $s.Name }}">{{ . }}</a>{{ else }}<span class="faint">0</span>{{ end }}</td>
				{{- $last := index $p.SecretLastReads $s.ID }}
				<td>{{ if $last.RunID }}<span title="{{ stamp $last.CreatedAt }}">{{ ago $last.CreatedAt }}</span> <a href="/runs/{{ $last.RunID }}">{{ $last.Folder }}/{{ with $last.Dir }}{{ . }}/{{ end }}{{ $last.Job }}</a>{{ else }}<span class="faint">never</span>{{ end }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on " $s.Name)
					(detail "Changed" (stamp $s.UpdatedAt))
					(item "edit" "Replace the value" "href" (print "/settings/secrets/edit?scope=" (urlquery $s.Scope) "&name=" (urlquery $s.Name)))
					(item "list" "Reads" "href" (print "/settings/secrets/" $s.ID "/reads"))
					(danger "trash" "Delete" "data-delete" (print "/api/secrets/" $s.ID) "data-confirm" (print "Delete " $s.Name " from " (or (and (eq $s.Scope $g.Shared) (print "every project in " $g.Workspace)) $s.Scope) "? Runs stop getting it.") "data-confirm-ok" "Delete" "data-next" "reload")
				) }}</td>
			</tr>
{{- end }}
