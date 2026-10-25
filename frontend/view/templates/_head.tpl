{{/* blockhead takes a Head, from (head title note empty buttons...): the
     head of a block, its title and note on the left and its buttons on
     the right. An empty list's line, "No keys added.", sits in the same
     row between them, and the block draws no table under it. */}}
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

{{/* buttons draws items as the buttons of a head, (list items small):
     a link for an item with href, else a button its data-* attributes
     drive, primary or danger in their colours, small in a block head. */}}
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

{{/* copyable is a value to copy, (list value label [shown [fill]]): the value,
     or shown in its place with the value in the tooltip, in code and
     the copy button right after it, "Copy <label>". fill makes it take
     the width it has and end the value with an ellipsis, the whole in
     the tooltip. */}}
{{- define "copyable" }}{{ $value := index . 0 }}{{ $label := index . 1 }}{{ $shown := $value }}{{ if gt (len .) 2 }}{{ $shown = index . 2 }}{{ end }}{{ $fill := false }}{{ if gt (len .) 3 }}{{ $fill = index . 3 }}{{ end -}}
<span class="copyable{{ if $fill }} fill{{ end }}"><code{{ if or $fill (ne $shown $value) }} title="{{ $value }}"{{ end }}>{{ $shown }}</code><button type="button" class="btn-icon" data-copy="{{ $value }}" title="Copy {{ $label }}" aria-label="Copy {{ $label }}"><svg class="ic"><use href="#i-copy"/></svg></button></span>
{{- end }}
