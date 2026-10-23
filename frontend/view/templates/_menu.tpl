{{/* menu takes a Menu, from (menu label parts...): the [...] button of
     a row, or of a block, and the list it opens on a click. On top, the
     details, what the row holds that its columns leave out; then the
     items, an icon and a label each, a line between their groups, the
     danger group last. An item with href is a link, any other a button
     whose data-* attributes say what a click does. task-ui.js opens it
     over the page, moves through it with the arrow keys and closes it on
     Escape, a click outside or a pick; on a phone it opens as a sheet.
     An empty menu draws nothing. docs/development.md, Row actions, says
     what goes in one. */}}
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
