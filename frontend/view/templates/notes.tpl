{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Notes</h2>
		<p>Leave a note for someone with <code>@name</code>.{{ if .Agent }} A note or a reply naming <span class="mention">@{{ .Agent }}</span> starts it on the thread: it works until it is done, or says why it will not, and replies here.{{ end }}</p>
	</div>
</div>

<form class="note-form" data-note-form>
	<textarea name="body" rows="3" required placeholder="{{ if .Agent }}@{{ .Agent }} update the weather job to …{{ else }}@name …{{ end }}" aria-label="A new note"></textarea>
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
