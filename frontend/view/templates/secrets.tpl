{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Secrets and env</h2>
		<p>Variables for the environment of runs, secret or plain env, applying to every workspace or to one. A secret's value is never shown again; an env value is. Runs record what they get, and <code>task-ui get-secret NAME</code> reads either.</p>
	</div>
	<div class="actions">
		{{- if .SecretKey }}
		<a class="btn btn-primary" href="/settings/secrets/new{{ with .SecretWorkspace }}?scope=workspace:{{ . }}{{ end }}"><svg class="ic"><use href="#i-plus"/></svg>Add variable</a>
		{{- else }}
		<button class="btn btn-primary" type="button" disabled title="The instance has no encryption key yet"><svg class="ic"><use href="#i-plus"/></svg>Add variable</button>
		{{- end }}
	</div>
</div>

{{- with .SecretWorkspace }}
<p class="dim secret-filter">Showing what reaches the workspace <b>{{ . }}</b>. <a href="/settings/secrets">Every workspace</a></p>
{{- end }}

<section class="block">
	{{- $note := print (len .Secrets) " variable" (or (and (ne (len .Secrets) 1) "s") "") }}
	{{- template "blockhead" (head "Variables" $note (and (not .Secrets) "No variables yet.")) }}
	{{- if .Error }}
	<div class="alert bad">{{ .Error }}</div>
	{{- end }}
	{{- with .Secrets }}
	<div class="table-wrap">
	<table class="table secrets">
		<thead><tr><th>Name</th><th>Value</th><th>Applies to</th><th>Reads</th><th class="actions" aria-label="Actions"></th></tr></thead>
		<tbody>
		{{- range . }}{{ $title := index $.SecretScopeTitles .Scope }}
			<tr>
				<td class="key">{{ .Name }}{{ if .IsSecret }} <span class="badge" title="The value is encrypted and never shown">secret</span>{{ else }} <span class="badge" title="The value is shown on this page">env</span>{{ end }}</td>
				<td>{{ if .IsSecret }}<span class="faint" title="The value is never shown">••••••</span>{{ else if .Content }}<code>{{ .Content }}</code>{{ else }}<span class="faint">empty</span>{{ end }}</td>
				<td class="scope faint">{{ or $title .Scope }}</td>
				{{- $id := .ID }}{{ $last := index $.SecretLastReads .ID }}
				<td class="nowrap">{{ with index $.SecretReadCounts .ID }}<a href="/settings/secrets/{{ $id }}/reads" title="The runs that got this variable">{{ . }}</a>{{ else }}<span class="faint">0</span>{{ end }} {{ if $last.RunID }}<a class="faint" href="/runs/{{ $last.RunID }}" title="Last read {{ stamp $last.CreatedAt }}: {{ $last.Folder }}/{{ with $last.Dir }}{{ . }}/{{ end }}{{ $last.Job }}">{{ ago $last.CreatedAt }}</a>{{ else }}<span class="faint" title="No run has got this variable">never</span>{{ end }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on " .Name)
					(detail "Changed" (stamp .UpdatedAt))
					(and ($.Can "secret/edit") (item "edit" "Edit" "href" (print "/settings/secrets/edit?id=" (urlquery .ID))))
					(item "list" "Reads" "href" (print "/settings/secrets/" .ID "/reads"))
					(and ($.Can "secret/delete") (danger "trash" "Delete" "data-delete" (print "/api/secrets/" .ID) "data-confirm" (print "Delete " .Name " from " (or $title .Scope) "? Runs stop getting it.") "data-confirm-ok" "Delete" "data-next" "reload"))
				) }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>

{{- with .IntegrationEnv }}
<section class="block" id="integration-env">
	{{- template "blockhead" (head "Integration environment" "read only, computed from the integrations for your runs" "" ) }}
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Name</th><th>Value</th></tr></thead>
		<tbody>
		{{- range . }}
			<tr><td class="key">{{ .Name }}</td><td>{{ if .Redacted }}<span class="faint mono">REDACTED</span>{{ else if .Value }}<code>{{ .Value }}</code>{{ else }}<span class="faint">empty</span>{{ end }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
{{ end }}
