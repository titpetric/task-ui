{{/* promptbox is the prompt box of a folder page, of a directory browsed
     inside a folder, of the Workflows page and of a run's page, for a user with prompt/edit and
     a Claude token: Claude Code on a prompt in that directory, a run of
     @claude-prompt there, and the threads asked there under it, each
     question with Claude's answer once its run ends. A run's box asks
     about the run, with its output, and lists the threads about it; one
     who only reads prompts sees those threads without the box, a running
     run's box hidden until a thread about it comes. A workflow run's box
     lists the threads of its prompt and summary steps alone. Open on a
     workspace's page, a run's page and where a thread exists, folded
     elsewhere. */}}
{{- define "promptbox" }}
{{- with .Prompt }}
<details class="promptbox"{{ if or .Open .Threads }} open{{ end }}{{ if and .Watch (not .Threads) }} hidden{{ end }}>
	{{- if .Area }}
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>Prompt <small>Claude Code builds or changes a workflow with the task tools, runs it and answers with its link</small></h3></summary>
	{{- else if .WorkflowRun }}
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>Prompts <small>the run's prompt and summary steps, each with the output of the steps before it</small></h3></summary>
	{{- else if .Run }}
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>Prompt <small>ask Claude Code about this run and its output, in {{ .Where }}</small></h3></summary>
	{{- else }}
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>Prompt <small>ask Claude Code in {{ .Where }}; it reads the CLAUDE.md and .claude/ there and in the folders above</small></h3></summary>
	{{- end }}
	{{- if not .Watch }}
	<form class="note-form" data-api="{{ .API }}" data-refresh-notes>
		{{- if .Run }}
		<input type="hidden" name="run" value="{{ .Run }}">
		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="why did this run end the way it did" aria-label="The prompt"></textarea>
		<div class="note-form-foot"><span class="faint">Runs @claude-prompt as you in {{ .Where }}, with this run's output. If Claude makes a fix, it commits to main. Claude replies under your question.</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
		{{- else if .Area }}
		<input type="hidden" name="area" value="{{ .Area }}">
		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="Describe a workflow to build or change" aria-label="The prompt"></textarea>
		<div class="note-form-foot"><span class="faint run-in">Runs @claude-prompt as you in <select name="folder" aria-label="Run in"><option value="_">the root</option>{{ range $.Folders }}<option value="{{ .ID }}">{{ .ID }}</option>{{ end }}</select> with the task tools: it saves the workflow and runs it up to 3 times.</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
		{{- else }}
		{{- if .Folder }}
		<input type="hidden" name="folder" value="{{ .Folder }}">
		<input type="hidden" name="dir" value="{{ .Dir }}">
		{{- end }}
		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="what does this directory build, and how" aria-label="The prompt"></textarea>
		<div class="note-form-foot"><span class="faint">Runs @claude-prompt as you in {{ .Where }}, one prompt at a time. Claude replies under your question.</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
		{{- end }}
	</form>
	{{- end }}
	<div class="notes prompt-threads" data-notes data-dir="{{ .Place }}"{{ with .Area }} data-area="{{ . }}"{{ end }}>
		{{- range .Threads }}
		{{- template "thread" (list $ .) }}
		{{- end }}
	</div>
</details>
{{- end }}
{{- end }}
