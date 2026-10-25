{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Integrations</h2>
		<p>The services task-ui talks to. task-ui encrypts their tokens with the instance key {{ with .InstanceKey }}<b>{{ .Fingerprint }}</b>{{ else }}<span class="badge bad">none yet</span>{{ end }}; leave a secret field empty to keep it.</p>
	</div>
</div>

{{- $users := .Users }}
{{- $admin := and .User .User.CanAdmin }}
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
			{{- else if .Options }}
			{{- $value := index $i.Values .Name }}{{ if not $value }}{{ $value = index .Options 0 }}{{ end }}
			<select id="{{ $i.ID }}-{{ .Name }}" name="values.{{ .Name }}">
				{{- range .Options }}
				<option value="{{ . }}"{{ if eq . $value }} selected{{ end }}>{{ . }}</option>
				{{- end }}
			</select>
			{{- else }}
			<input type="text" id="{{ $i.ID }}-{{ .Name }}" name="values.{{ .Name }}" value="{{ index $i.Values .Name }}" placeholder="{{ .Placeholder }}" autocomplete="off" spellcheck="false">
			{{- end }}
			{{- end }}
			{{- with .Help }}<small class="faint">{{ . }}</small>{{ end }}
		</div>
		{{- end }}
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Save</button>
			{{- if and (eq .ID "forgejo") .Configured $admin }}
			<button type="button" class="btn" data-run data-folder="_" data-job="@forgejo-sync" data-title="_ / forgejo-sync" title="Run @forgejo-sync now: workspaces, homes, projects and hooks">Sync now</button>
			<button type="button" class="btn" data-run data-folder="_" data-job="@forgejo-sync" data-run-args='["--dry-run"]' data-title="_ / forgejo-sync --dry-run" title="Run @forgejo-sync --dry-run: the report alone, nothing changes">Dry run</button>
			{{- end }}
			<button type="button" class="btn btn-danger" data-delete="/api/integrations/{{ .ID }}" data-confirm="Clear every setting of {{ .Title }}?" data-confirm-ok="Clear" data-next="reload">Clear</button>
		</div>
	</form>
	{{- if eq .ID "forgejo" }}
	<p class="dim integration-sync">
		{{- $url := "" }}{{ with $.Instance }}{{ $url = .URL }}{{ end }}
		{{- if $url }}Sync installs the push hook <code>{{ $url }}/api/webhooks/forgejo</code> on every repository it syncs.{{ else }}The instance has no public URL (<code>TASK_UI_URL</code>), so sync installs no push hook.{{ end }}
		{{ with $.LastRun "_/@forgejo-sync" }}Last sync <span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span> {{ ago .StartedAt }} by {{ .UserName }}: <a href="/runs/{{ .ID }}">the report</a>.{{ else }}No sync ran yet: <b>Dry run</b> shows what one would do.{{ end }}
	</p>
	{{- end }}
	</div>
</section>
{{- end }}
{{ end }}
