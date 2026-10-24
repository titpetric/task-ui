{{/* actions is the git line of a folder page and of a directory browsed
     inside a folder: the checkout the place is in, its branch, what it
     has to push and pull and what changed. The jobs, the skills' and the
     built-ins among them, are rows of the Jobs table. */}}
{{- define "actions" }}
{{- with .Actions }}{{ with .Git }}
<div class="actionsbar">
	<div class="gitline" title="The checkout {{ if .Checkout }}{{ .Checkout }}{{ else }}at the root{{ end }}{{ with .Upstream }}, tracking {{ . }}{{ end }}">
		<span class="branch"><span class="glyph" aria-hidden="true">⎇</span> {{ if .Branch }}{{ .Branch }}{{ else }}<span class="faint">detached</span> {{ .Detached }}{{ end }}</span>
		{{- if .Ahead }}<span class="ahead" title="{{ .Ahead }} commits to push">↑{{ .Ahead }}</span>{{ end }}
		{{- if .Behind }}<span class="behind" title="{{ .Behind }} commits to pull">↓{{ .Behind }}</span>{{ end }}
		{{- if .Dirty }}<span class="dirty" title="{{ .Dirty }} files changed or untracked">●{{ .Dirty }}</span>{{ end }}
		{{- if and .Upstream (not .Ahead) (not .Behind) (not .Dirty) }}<span class="clean">up to date</span>{{ end }}
		<span class="where">· {{ if .Checkout }}{{ .Checkout }}{{ else }}the root{{ end }}</span>
	</div>
</div>
{{- end }}{{ end }}
{{- end }}
