{{ if $.Can "note/delete" }}delete{{ end }}
{{ if .Can "secret/view" }}{{ end }}{{ if can "shell" }}{{ end }}
