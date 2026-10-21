{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Integrations</h2>
		<p>The services task-ui talks to. Their keys and tokens are encrypted with the instance SSH {{ if eq (len .Encryption) 1 }}key{{ else }}keys{{ end }} {{ if .Encryption }}<b>{{ join .Encryption ", " }}</b>{{ else }}<span class="badge bad">none yet</span>{{ end }}, which the workspaces' secrets do not use, and never shown again; leave a secret empty to keep it.</p>
	</div>
</div>

{{- $users := .Users }}
{{- range .Integrations }}
<section class="block integration" id="{{ .ID }}">
	<div class="block-head">
		<h3>{{ .Title }} {{ if .Configured }}<span class="badge ok">set up</span>{{ else }}<span class="badge">not set up</span>{{ end }}</h3>
		<label class="peek-toggle">Show<input type="checkbox" data-peek="integration.{{ .ID }}" checked aria-label="Show {{ .Title }}"><i aria-hidden="true"></i></label>
	</div>
	<div class="integration-body">
	<p class="dim">{{ message .Description }}</p>
	<form class="form" data-api="/api/integrations/{{ .ID }}" data-method="PUT" data-next="/settings/integrations?notice={{ .Title }}+saved">
		{{- $i := . }}
		{{- range .Fields }}
		<div class="field">
			{{- if .Who }}
			<span class="label">{{ .Label }}</span>
			<div class="checks">
				<span class="faint">Roles</span>
				{{- $name := .Name }}
				{{- range roles }}
				<label class="check"><input type="checkbox" name="values.{{ $name }}" value="{{ . }}" data-type="joined"{{ if index $i.Checked . }} checked{{ end }}> {{ . }}</label>
				{{- end }}
			</div>
			<div class="checks">
				<span class="faint">Users</span>
				{{- range $users }}
				<label class="check"><input type="checkbox" name="values.{{ $name }}" value="{{ .Name }}" data-type="joined"{{ if index $i.Checked .Name }} checked{{ end }}> {{ .Name }}</label>
				{{- end }}
			</div>
			{{- else }}
			<label for="{{ $i.ID }}-{{ .Name }}">{{ .Label }}{{ if .Secret }} {{ if index $i.Set .Name }}<span class="badge ok">set</span>{{ else }}<span class="badge">empty</span>{{ end }}{{ end }}</label>
			{{- if .Secret }}
			<input type="password" id="{{ $i.ID }}-{{ .Name }}" name="values.{{ .Name }}" placeholder="{{ if index $i.Set .Name }}leave empty to keep it{{ else }}{{ .Placeholder }}{{ end }}" autocomplete="new-password" spellcheck="false">
			{{- else }}
			<input type="text" id="{{ $i.ID }}-{{ .Name }}" name="values.{{ .Name }}" value="{{ index $i.Values .Name }}" placeholder="{{ .Placeholder }}" autocomplete="off" spellcheck="false">
			{{- end }}
			{{- end }}
			{{- with .Help }}<small class="faint">{{ . }}</small>{{ end }}
		</div>
		{{- end }}
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Save</button>
			<button type="button" class="btn btn-danger" data-delete="/api/integrations/{{ .ID }}" data-confirm="Clear every setting of {{ .Title }}?" data-confirm-ok="Clear" data-next="reload">Clear</button>
		</div>
	</form>
	</div>
</section>
{{- end }}
{{ end }}
