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
	<article class="note{{ if .State }} instruction {{ .State }}{{ end }}" id="note-{{ .ID }}">
		<header>
			<span class="author">{{ template "avatar" (list $ .UserName) }} <b>{{ .UserName }}</b></span>
			<span class="faint" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span>
			{{- if .State }} <span class="badge note-state {{ .State }}">{{ .State }}</span>{{ end }}
			{{- if .RunID }} <a class="faint mono" href="/runs/{{ .RunID }}" title="The run the agent worked in">run {{ slice .RunID 0 8 }}</a>{{ end }}
			<span class="spacer"></span>
			{{- if or (eq .UserName $.User.Name) $.User.CanAdmin }}<button type="button" class="btn-icon danger" data-delete="/api/notes/{{ .ID }}" data-confirm="Delete the note of {{ .UserName }} and its replies?" data-next="reload" title="Delete" aria-label="Delete the note"><svg class="ic"><use href="#i-trash"/></svg></button>{{ end }}
		</header>
		<div class="note-body">{{ message .Body }}</div>
		{{- if .Replies }}
		<ol class="replies">
			{{- range .Replies }}
			<li>
				<header><span class="author">{{ template "avatar" (list $ .UserName) }} <b>{{ .UserName }}</b></span> <span class="faint" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span>{{ if .RunID }} <a class="faint mono" href="/runs/{{ .RunID }}" title="The run this reply comes from">run {{ slice .RunID 0 8 }}</a>{{ end }}<span class="spacer"></span>{{ if or (eq .UserName $.User.Name) $.User.CanAdmin }}<button type="button" class="btn-icon danger" data-delete="/api/notes/{{ .ID }}" data-confirm="Delete this reply of {{ .UserName }}?" data-next="reload" title="Delete the reply" aria-label="Delete the reply"><svg class="ic"><use href="#i-trash"/></svg></button>{{ end }}</header>
				<div class="note-body">{{ message .Body }}</div>
			</li>
			{{- end }}
		</ol>
		{{- end }}
		<details class="reply">
			<summary>Reply</summary>
			<form class="note-form" data-note-form data-parent="{{ .ID }}">
				<textarea name="body" rows="2" required aria-label="A reply"></textarea>
				<div class="note-form-foot"><span></span><button class="btn" type="submit">Reply</button></div>
			</form>
		</details>
	</article>
	{{- else }}
	<p class="empty-state">No notes yet.</p>
	{{- end }}
</div>
{{ end }}
