{{- define "blockhead" }}
<div class="block-head{{ if .Empty }} is-empty{{ end }}">
	<h3>{{ .Title }}{{ with .Note }} <small>{{ . }}</small>{{ end }}</h3>
	{{- with .Empty }}
	<p class="empty-line">{{ . }}</p>
	{{- end }}
	{{- with .Buttons }}
	{{ template "buttons" (list . true) }}
	{{- end }}
</div>
{{- end }}

{{- define "buttons" }}{{ $small := index . 1 -}}
<div class="actions">
	{{- range index . 0 }}
	{{- $class := "btn" }}{{ if $small }}{{ $class = print $class " btn-sm" }}{{ end }}{{ if .Primary }}{{ $class = print $class " btn-primary" }}{{ end }}{{ if .Danger }}{{ $class = print $class " btn-danger" }}{{ end }}
	{{- if .Href }}
	<a class="{{ $class }}" href="{{ .Href }}"{{ if .Download }} download{{ end }}{{ with .Attrs }} {{ . }}{{ end }}>{{ with .Icon }}<svg class="ic"><use href="#i-{{ . }}"/></svg>{{ end }}{{ .Label }}</a>
	{{- else }}
	<button type="button" class="{{ $class }}"{{ if .Disabled }} disabled{{ with .Note }} title="{{ . }}"{{ end }}{{ end }}{{ with .Attrs }} {{ . }}{{ end }}>{{ with .Icon }}<svg class="ic"><use href="#i-{{ . }}"/></svg>{{ end }}{{ .Label }}</button>
	{{- end }}
	{{- end }}
</div>
{{- end }}

{{- define "copyable" }}{{ $value := index . 0 }}{{ $label := index . 1 }}{{ $shown := $value }}{{ if gt (len .) 2 }}{{ $shown = index . 2 }}{{ end }}{{ $fill := false }}{{ if gt (len .) 3 }}{{ $fill = index . 3 }}{{ end -}}
<span class="copyable{{ if $fill }} fill{{ end }}"><code{{ if or $fill (ne $shown $value) }} title="{{ $value }}"{{ end }}>{{ $shown }}</code><button type="button" class="btn-icon" data-copy="{{ $value }}" title="Copy {{ $label }}" aria-label="Copy {{ $label }}"><svg class="ic"><use href="#i-copy"/></svg></button></span>
{{- end }}
