{{/* promptbox is the prompt box of a folder page and of a directory
     browsed inside a folder, for an admin with a Claude token: Claude Code
     on a prompt in that directory, a run of internal/claude-prompt whose
     window opens. Open on a workspace's page, folded elsewhere. */}}
{{- define "promptbox" }}
{{- with .Prompt }}
<details class="promptbox"{{ if .Open }} open{{ end }}>
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>Prompt <small>Claude Code in {{ .Where }}, with the CLAUDE.md and .claude/ there and above</small></h3></summary>
	<form class="note-form" data-api="{{ .API }}" data-attach-result>
		{{- if .Folder }}
		<input type="hidden" name="folder" value="{{ .Folder }}">
		<input type="hidden" name="dir" value="{{ .Dir }}">
		{{- end }}
		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="what does this directory build, and how" aria-label="The prompt"></textarea>
		<div class="note-form-foot"><span class="faint">Runs internal/claude-prompt as you, one at a time in {{ .Where }}, and opens its window.</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
	</form>
</details>
{{- end }}
{{- end }}
