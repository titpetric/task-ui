{{ define "content" }}
{{- $admin := and .User .User.CanAdmin }}
<div class="page-head">
	<div>
		<h2>Data retention</h2>
		<p>How long the instance keeps what it records, and how much of it. Past either rule the oldest go first, until both hold. A save takes effect at the next run of <code>@cleanup</code>, which the schedules start every day at midnight.{{ if not $admin }} An admin changes these rules.{{ end }}</p>
	</div>
	{{- if $admin }}
	<div class="actions">
		<button type="button" class="btn" data-run data-folder="_" data-job="@cleanup" data-title="_ / cleanup" title="Run @cleanup now, with the saved rules">Clean up now</button>
	</div>
	{{- end }}
</div>

<form class="form retention" data-api="/api/retention" data-method="PUT" data-next="/settings/retention?notice=Data+retention+saved">
{{- range .Retention }}{{ $b := . }}
<section class="block retention-block" data-retention data-average="{{ .Average }}" data-one="{{ .One }}" data-many="{{ .Many }}">
	<h3>{{ .Title }}</h3>
	<p class="dim">{{ .Desc }}</p>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Holds</th><td>{{ .Usage.Says .One .Many }}</td></tr>
			<tr><th>Oldest</th><td>{{ with .Usage.Oldest }}<span title="{{ stamp . }}">{{ ago . }}</span>{{ else }}<span class="faint">none</span>{{ end }}</td></tr>
		</tbody>
	</table>
	</div>
	<fieldset class="field"{{ if not $admin }} disabled{{ end }}>
		<legend class="label">Keep for</legend>
		<div class="segmented">
			{{- range .Ages }}
			<label><input type="radio" name="{{ $b.Kind }}.age" value="{{ .Value }}"{{ if .Checked }} checked{{ end }}><span>{{ .Label }}</span></label>
			{{- end }}
		</div>
	</fieldset>
	<fieldset class="field"{{ if not $admin }} disabled{{ end }}>
		<legend class="label">Size limit</legend>
		<div class="size-limit">
			<input type="text" id="{{ .Kind }}-limit" name="{{ .Kind }}.limit" value="{{ .LimitText }}" placeholder="none" aria-label="Size limit of {{ .Title }}, such as 500MB or 2GB" autocomplete="off" spellcheck="false" data-size-input>
			{{- range .Sizes }}
			<button type="button" class="btn btn-sm" data-size="{{ . }}">{{ . }}</button>
			{{- end }}
			<button type="button" class="btn btn-sm" data-size="">None</button>
		</div>
		<small class="faint" data-estimate>{{ .Estimate }}</small>
	</fieldset>
</section>
{{- end }}
	{{- if $admin }}
	<div class="form-foot">
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
	{{- end }}
</form>
{{ end }}
