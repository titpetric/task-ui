{{/* promptbox is the prompt box of a folder page, of a directory browsed
     inside a folder and of a run's page, for a user with prompt/edit and
     a Claude token: Claude Code on a prompt in that directory, a run of
     @claude-prompt there, and the threads asked there under it, each
     question with Claude's answer once its run ends. A run's box asks
     about the run, with its output, and lists the threads about it; one
     who only reads prompts sees those threads without the box, a running
     run's box hidden until a thread about it comes. Open on a
     workspace's page, a run's page and once a prompt was asked, folded
     elsewhere. */}}
{{- define "promptbox" }}
{{- with .Prompt }}
<details class="promptbox"{{ if or .Open .Threads }} open{{ end }}{{ if and .Watch (not .Threads) }} hidden{{ end }}>
	{{- if .Run }}
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>Prompt <small>Claude Code on this run and its output, in {{ .Where }}</small></h3></summary>
	{{- else }}
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>Prompt <small>Claude Code in {{ .Where }}, with the CLAUDE.md and .claude/ there and above</small></h3></summary>
	{{- end }}
	{{- if not .Watch }}
	<form class="note-form" data-api="{{ .API }}" data-refresh-notes>
		{{- if .Run }}
		<input type="hidden" name="run" value="{{ .Run }}">
		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="why did this run end the way it did" aria-label="The prompt"></textarea>
		<div class="note-form-foot"><span class="faint">Runs @claude-prompt as you in {{ .Where }}, with this run's output; a fix it makes goes to main. Claude's answer shows under the question.</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
		{{- else }}
		{{- if .Folder }}
		<input type="hidden" name="folder" value="{{ .Folder }}">
		<input type="hidden" name="dir" value="{{ .Dir }}">
		{{- end }}
		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="what does this directory build, and how" aria-label="The prompt"></textarea>
		<div class="note-form-foot"><span class="faint">Runs @claude-prompt as you, one at a time in {{ .Where }}; Claude's answer shows under the question.</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
		{{- end }}
	</form>
	{{- end }}
	<div class="notes prompt-threads" data-notes data-dir="{{ .Place }}">
		{{- range .Threads }}
		{{- template "thread" (list $ .) }}
		{{- end }}
	</div>
</details>
{{- end }}
{{- end }}
