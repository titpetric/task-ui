{{/* actions is the git line of a folder page and of a directory browsed
     inside a folder: the checkout the place is in, its branch, what it
     has to push and pull and what changed, the changed files a table
     under it that the click on their count shows and hides. The jobs,
     the skills' and the built-ins among them, are rows of the Jobs
     table. */}}
{{- define "actions" }}
{{- with .Actions }}{{ with .Git }}
<div class="actionsbar">
	<div class="gitline" title="The checkout {{ if .Checkout }}{{ .Checkout }}{{ else }}at the root{{ end }}{{ with .Upstream }}, tracking {{ . }}{{ end }}">
		<span class="branch"><span class="glyph" aria-hidden="true">⎇</span> {{ if .Branch }}{{ .Branch }}{{ else }}<span class="faint">detached</span> {{ .Detached }}{{ end }}</span>
		{{- if .Ahead }}<span class="ahead" title="{{ .Ahead }} commits to push">↑{{ .Ahead }}</span>{{ end }}
		{{- if .Behind }}<span class="behind" title="{{ .Behind }} commits to pull">↓{{ .Behind }}</span>{{ end }}
		{{- if .Dirty }}<button type="button" class="dirty" data-git-status aria-expanded="false" aria-controls="git-status" title="Show the changed files"><span class="glyph" aria-hidden="true">●</span> {{ .Dirty }} file{{ if ne .Dirty 1 }}s{{ end }} changed or untracked<svg class="ic chev" aria-hidden="true"><use href="#i-chev"/></svg></button>{{ end }}
		{{- if and .Upstream (not .Ahead) (not .Behind) (not .Dirty) }}<span class="clean">up to date</span>{{ end }}
		<span class="where">· {{ if .Checkout }}{{ .Checkout }}{{ else }}the root{{ end }}</span>
	</div>
	{{ if .Dirty }}
	<div class="gitstatus" id="git-status" hidden>
		<div class="table-wrap">
			<table class="table gitstatus-table">
				<thead><tr><th>Status</th><th>File</th></tr></thead>
				<tbody>
					{{- range .Changes }}
					<tr><td class="gitstate"><code>{{ .Code }}</code> <span class="faint">{{ .State }}</span></td><td class="gitpath">{{ .Path }}{{ with .From }} <span class="faint">from {{ . }}</span>{{ end }}</td></tr>
					{{- end }}
				</tbody>
			</table>
		</div>
	</div>
	{{- end }}
</div>
{{- end }}{{ end }}
{{- end }}
