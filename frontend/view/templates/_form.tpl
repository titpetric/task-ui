{{- define "formblock" -}}
{{- if .Err -}}
<div class="alert bad" role="alert">{{ .Err }}</div>
{{- else -}}
<form class="note-form run-form" data-run-form{{ with .Folder }} data-folder="{{ . }}"{{ end }}{{ with .Dir }} data-dir="{{ . }}"{{ end }}{{ with .Job }} data-job="{{ . }}"{{ end }}{{ with .Prompt }} data-prompt="{{ . }}"{{ end }}>
{{ with .Title }}<p class="form-title">{{ . }}</p>
{{ end }}{{ range .Fields }}{{ template "formfield" . }}
{{ end }}<div class="form-foot"><span class="faint form-state"></span><button class="btn btn-primary" type="submit">{{ .Label }}</button></div>
</form>
{{- end -}}
{{- end }}

{{- define "formfield" -}}
{{- if eq .Component "hidden" -}}
<input type="hidden" name="{{ .Key }}" value="{{ .Value }}" data-key="{{ .Key }}" data-component="hidden"{{ with .When }} data-when="{{ . }}"{{ end }}{{ if .Required }} data-required{{ end }}>
{{- else -}}
<div class="field" data-key="{{ .Key }}" data-component="{{ .Component }}"{{ with .When }} data-when="{{ . }}"{{ end }}{{ if .Required }} data-required{{ end }}{{ if .Hidden }} hidden{{ end }}>
{{ template "formcontrol" . }}
{{ with .Help }}<small class="faint">{{ . }}</small>
{{ end }}</div>
{{- end -}}
{{- end }}

{{- define "formcontrol" -}}
{{- $key := .Key -}}
{{- if eq .Component "bool" -}}
<label class="check"><input type="checkbox" name="{{ .Key }}" data-type="toggle"{{ if .Checked }} checked{{ end }}> {{ .Label }}</label>
{{- else if eq .Component "multi" -}}
<span class="label">{{ .Label }}</span>
<div class="checks">{{ range .Options }}<label class="check"><input type="checkbox" name="{{ $key }}" value="{{ .Value }}" data-type="set"{{ if .Selected }} checked{{ end }}{{ with .Next }} data-next="{{ . }}"{{ end }}> {{ .Label }}</label>{{ end }}</div>
{{- else if eq .Component "textarea" -}}
<span class="label">{{ .Label }}</span>
<textarea name="{{ .Key }}" rows="4" spellcheck="false" aria-label="{{ .Label }}"{{ with .Placeholder }} placeholder="{{ . }}"{{ end }}{{ if .Required }} required{{ end }}>{{ .Value }}</textarea>
{{- else if eq .Component "number" -}}
<span class="label">{{ .Label }}</span>
<input type="number" name="{{ .Key }}" value="{{ .Value }}" aria-label="{{ .Label }}"{{ with .Placeholder }} placeholder="{{ . }}"{{ end }}{{ if .Required }} required{{ end }}>
{{- else if eq .Component "text" -}}
<span class="label">{{ .Label }}</span>
<input type="text" name="{{ .Key }}" value="{{ .Value }}" aria-label="{{ .Label }}"{{ with .Placeholder }} placeholder="{{ . }}"{{ end }}{{ if .Required }} required{{ end }}>
{{- else -}}
<span class="label">{{ .Label }}</span>
<select name="{{ .Key }}" aria-label="{{ .Label }}"{{ if .Required }} required{{ end }}>{{ if not .Required }}<option value=""></option>{{ end }}{{ range .Options }}<option value="{{ .Value }}"{{ with .Next }} data-next="{{ . }}"{{ end }}{{ if .Selected }} selected{{ end }}>{{ .Label }}</option>{{ end }}</select>
{{- end -}}
{{- end }}
