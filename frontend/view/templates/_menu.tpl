{{- define "menu" }}{{ if not .Empty -}}
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="{{ .Label }}" aria-label="{{ .Label }}"><svg class="ic"><use href="#i-more"/></svg></summary>
	<div class="more-list">
		{{- with .Details }}
		<dl class="more-details">
			{{- range . }}
			<dt>{{ .Label }}</dt><dd{{ if .Mono }} class="mono"{{ end }}{{ with .Title }} title="{{ . }}"{{ end }}>{{ .Value }}</dd>
			{{- end }}
		</dl>
		{{- end }}
		<div class="more-items" role="menu" aria-label="{{ .Label }}">
		{{- range $i, $g := .Groups }}{{ if $i }}
			<hr class="more-sep">{{ end }}
			{{- range $g }}
			{{- if .Href }}
			<a class="more-item{{ if .Danger }} danger{{ end }}" role="menuitem" href="{{ .Href }}"{{ if .Download }} download{{ end }}{{ with .Attrs }} {{ . }}{{ end }}><svg class="ic"><use href="#i-{{ .Icon }}"/></svg><span>{{ .Label }}</span>{{ with .Note }}<small class="faint">{{ . }}</small>{{ end }}</a>
			{{- else }}
			<button type="button" class="more-item{{ if .Danger }} danger{{ end }}" role="menuitem"{{ if .Disabled }} disabled{{ end }}{{ with .Attrs }} {{ . }}{{ end }}><svg class="ic"><use href="#i-{{ .Icon }}"/></svg><span>{{ .Label }}</span>{{ with .Note }}<small class="faint">{{ . }}</small>{{ end }}</button>
			{{- end }}
			{{- end }}
		{{- end }}
		</div>
	</div>
</details>
{{- end }}{{ end }}
