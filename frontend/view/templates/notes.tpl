{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Notes</h2>
		<p>Leave a note for someone with <code>@name</code>. If you can run jobs, the note is an instruction to the people it names. It stays open until they mark it working, and then done or failed, here or over MCP.{{ if .Agent }} Name <span class="mention">@{{ .Agent }}</span> in a note or a reply to start the agent on the thread. It works on the request and replies here when it finishes, or explains why it can't.{{ end }}</p>
	</div>
</div>

<form class="note-form" data-note-form>
	<textarea name="body" rows="3" required data-mentions placeholder="{{ if .Agent }}@{{ .Agent }} update the weather job to …{{ else }}@name …{{ end }}" aria-label="A new note"></textarea>
	<div class="note-form-foot"><span class="faint">Ctrl+Enter posts. Markdown: <code>**bold**</code>, <code>- lists</code>, <code>`code`</code> and <code>```</code> blocks.</span><button class="btn btn-primary" type="submit">Post note</button></div>
</form>

<div class="notes" data-notes>
	{{- range .Threads }}
	{{- template "thread" (list $ .) }}
	{{- else }}
	<p class="empty-state">No notes yet.</p>
	{{- end }}
</div>
{{ end }}
