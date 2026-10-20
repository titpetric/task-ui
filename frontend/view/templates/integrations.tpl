{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Integrations</h2>
		<p>The services task-ui talks to. Their keys and tokens are encrypted with the instance SSH {{ if eq (len .Encryption) 1 }}key{{ else }}keys{{ end }} {{ if .Encryption }}<b>{{ join .Encryption ", " }}</b>{{ else }}<span class="badge bad">none yet</span>{{ end }}, like the secrets, and never shown again; leave a secret empty to keep it.</p>
	</div>
</div>

{{- range .Integrations }}
<section class="block integration" id="{{ .ID }}">
	<div class="block-head">
		<h3>{{ .Title }} {{ if .Configured }}<span class="badge ok">set up</span>{{ else }}<span class="badge">not set up</span>{{ end }}</h3>
		<div class="actions">
			<button type="button" class="btn btn-sm btn-danger" data-delete="/api/integrations/{{ .ID }}" data-confirm="Remove every setting of {{ .Title }}?" data-confirm-ok="Remove" data-next="reload">Remove</button>
		</div>
	</div>
	<p class="dim">{{ message .Description }}</p>
	<form class="form" data-api="/api/integrations/{{ .ID }}" data-method="PUT" data-next="/settings/integrations?notice={{ .Title }}+saved">
		{{- $i := . }}
		{{- range .Fields }}
		<div class="field">
			<label for="{{ $i.ID }}-{{ .Name }}">{{ .Label }}{{ if .Secret }} {{ if index $i.Set .Name }}<span class="badge ok">set</span>{{ else }}<span class="badge">empty</span>{{ end }}{{ end }}</label>
			{{- if .Secret }}
			<input type="password" id="{{ $i.ID }}-{{ .Name }}" name="values.{{ .Name }}" placeholder="{{ if index $i.Set .Name }}leave empty to keep it{{ else }}{{ .Placeholder }}{{ end }}" autocomplete="new-password" spellcheck="false">
			{{- else }}
			<input type="text" id="{{ $i.ID }}-{{ .Name }}" name="values.{{ .Name }}" value="{{ index $i.Values .Name }}" placeholder="{{ .Placeholder }}" autocomplete="off" spellcheck="false">
			{{- end }}
			{{- with .Help }}<small class="faint">{{ . }}</small>{{ end }}
		</div>
		{{- end }}
		<div class="form-foot"><button class="btn btn-primary" type="submit">Save</button></div>
	</form>
</section>
{{- end }}
{{ end }}
