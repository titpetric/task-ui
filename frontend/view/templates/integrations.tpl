{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Integrations</h2>
		<p>The services task-ui talks to. task-ui encrypts their tokens with the instance key {{ with .InstanceKey }}<b class="mono-break">{{ .Fingerprint }}</b>{{ else }}<span class="badge bad">none yet</span>{{ end }}; leave a secret field empty to keep it.</p>
	</div>
</div>

{{- $users := .Users }}
{{- $admin := and .User .User.CanAdmin }}
<section class="block integration-cards">
	<div class="wf-steps-head">
		<div>
			<h3 class="section-title">The services</h3>
			<p class="faint">A card per integration, its settings and the switches that leave a variable out of a run, then how to get its key.</p>
		</div>
		<button type="button" class="btn btn-sm" data-cards-fold aria-pressed="true">Fold all</button>
	</div>
	<ol class="wf-rail">
		{{- range $n, $int := .Integrations }}
		<li class="wf-step integration-step kind-{{ $int.ID }} open" id="{{ $int.ID }}">
			<span class="wf-node" aria-hidden="true">{{ add $n 1 }}</span>
			<div class="wf-card">
				<div class="wf-card-head">
					<button type="button" class="wf-toggle" data-card-toggle aria-expanded="true">
						<span class="wf-kind"><svg class="ic" aria-hidden="true"><use href="#i-plug"/></svg>Integration</span>
						<span class="wf-line">{{ $int.Title }}</span>
						{{- if $int.Configured }}<span class="badge ok">set up</span>{{ else }}<span class="badge">not set up</span>{{ end }}
						<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
					</button>
					{{ template "menu" (menu (print "Actions on " $int.Title)
						(danger "trash" "Clear" "data-delete" (print "/api/integrations/" $int.ID) "data-confirm" (print "Clear every setting of " $int.Title "?") "data-confirm-ok" "Clear" "data-next" "reload")
					) }}
				</div>
				<div class="wf-card-body">
					<div class="dim desc integration-desc">{{ message $int.Description }}</div>
					<form class="form integration-form" data-api="/api/integrations/{{ $int.ID }}" data-method="PUT" data-next="/settings/integrations?notice={{ $int.Title }}+saved">
						<div class="wf-fields">
							{{- range $int.Fields }}
							{{- if .Who }}
							<div class="field wf-wide">
								<span class="label">{{ .Label }}</span>
								{{- template "whofield" (who (print "values." .Name) "joined" $int.Checked $int.Checked $users "every run gets the token") }}
								{{- with .Help }}<small class="faint">{{ . }}</small>{{ end }}
							</div>
							{{- else if or .Empty .Source }}
							<div class="field wf-wide env-flag">
								<div class="field-head">
									<span class="label">{{ .Label }}</span>
									{{- template "envswitch" (list $int .) }}
								</div>
								<small class="faint">{{ with .Help }}{{ . }} {{ end }}<code>{{ .Env }}</code></small>
							</div>
							{{- else }}
							<div class="field{{ if .Secret }} wf-wide{{ end }}">
								{{- if .Toggle }}
								<div class="field-head">
									<label for="{{ $int.ID }}-{{ .Name }}">{{ .Label }}{{ if .Secret }} {{ if index $int.Set .Name }}<span class="badge ok">set</span>{{ else }}<span class="badge">empty</span>{{ end }}{{ end }}</label>
									{{- template "envswitch" (list $int .) }}
								</div>
								{{- else }}
								<label for="{{ $int.ID }}-{{ .Name }}">{{ .Label }}{{ if .Secret }} {{ if index $int.Set .Name }}<span class="badge ok">set</span>{{ else }}<span class="badge">empty</span>{{ end }}{{ end }}</label>
								{{- end }}
								{{- if .Secret }}
								<input type="password" id="{{ $int.ID }}-{{ .Name }}" name="values.{{ .Name }}" placeholder="{{ if index $int.Set .Name }}leave empty to keep it{{ else }}{{ .Placeholder }}{{ end }}" autocomplete="new-password" spellcheck="false">
								{{- else if .Options }}
								{{- $value := index $int.Values .Name }}{{ if not $value }}{{ $value = index .Options 0 }}{{ end }}
								<select id="{{ $int.ID }}-{{ .Name }}" name="values.{{ .Name }}">
									{{- range .Options }}
									<option value="{{ . }}"{{ if eq . $value }} selected{{ end }}>{{ . }}</option>
									{{- end }}
								</select>
								{{- else }}
								{{- $value := index $int.Values .Name }}{{ if and (not $value) .Default }}{{ $value = .Default }}{{ end }}
								<input type="text" id="{{ $int.ID }}-{{ .Name }}" name="values.{{ .Name }}" value="{{ $value }}" placeholder="{{ .Placeholder }}" autocomplete="off" spellcheck="false">
								{{- end }}
								<small class="faint">{{ with .Help }}{{ . }} {{ end }}{{ with .Env }}<code>{{ . }}</code>{{ end }}</small>
							</div>
							{{- end }}
							{{- end }}
						</div>
						<details class="integration-key">
							<summary>Get the {{ $int.Credential }}</summary>
							<ol class="steps">
								{{- range $int.Steps }}
								<li>{{ message . }}</li>
								{{- end }}
							</ol>
							{{- if eq $int.ID "forgejo" }}
							<p class="dim">
								{{- $url := "" }}{{ with $.Instance }}{{ $url = .URL }}{{ end }}
								{{- if $url }}Sync installs the push hook <code>{{ $url }}/api/webhooks/forgejo</code> on every repository it syncs.{{ else }}The instance has no public URL (<code>TASK_UI_URL</code>), so sync installs no push hook.{{ end }}
							</p>
							{{- end }}
						</details>
						<div class="form-foot">
							{{- if and (eq $int.ID "forgejo") $int.Configured $admin }}
							<button type="button" class="btn" data-run data-folder="internal" data-job="@forgejo-sync" data-run-args='["--dry-run"]' data-title="internal / forgejo-sync --dry-run" title="Run @forgejo-sync --dry-run: the report alone, nothing changes">Dry run</button>
							<button type="button" class="btn" data-run data-folder="internal" data-job="@forgejo-sync" data-title="internal / forgejo-sync" title="Run @forgejo-sync now: workspaces, homes, projects and hooks">Sync now</button>
							{{- end }}
							<button class="btn btn-primary" type="submit">Save</button>
						</div>
					</form>
					{{- if eq $int.ID "forgejo" }}
					<p class="dim integration-sync">
						{{- with $.LastRun "internal/@forgejo-sync" }}Last sync <span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span> {{ ago .StartedAt }} by {{ .UserName }}: <a href="/runs/{{ .ID }}">the report</a>.{{ else }}No sync ran yet: <b>Dry run</b> shows what one would do.{{ end }}
					</p>
					{{- end }}
				</div>
			</div>
		</li>
		{{- end }}
	</ol>
</section>
{{ end }}

{{/* envswitch is a setting's switch to send its variable or leave it out:
     (list integration field), a checkbox under `values.<field>.enabled`
     whose stored row is `<field>_enabled`. */}}
{{ define "envswitch" }}
{{- $int := index . 0 }}{{ $f := index . 1 -}}
<label class="switch env-switch" title="Send {{ $f.Env }} in a run">
	<input type="checkbox" name="values.{{ $f.Name }}.enabled" value="1" data-type="toggle"{{ if index $int.Enabled $f.Name }} checked{{ end }} aria-label="Send {{ $f.Env }}">
	<i aria-hidden="true"></i>
</label>
{{ end }}
