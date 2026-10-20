{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Chat</h2>
		<p>One channel for everyone signed in. Name someone with <code>@name</code>; the messages arrive as they are sent.</p>
	</div>
</div>

<section class="chat">
	<ol class="chat-log" data-notes data-chat-log>
		{{- range .Threads }}
		<li class="chat-line{{ if eq .UserName $.User.Name }} mine{{ end }}" id="note-{{ .ID }}">
			<span class="avatar sm">{{ .Initials }}</span>
			<div>
				<header><b>{{ .UserName }}</b> <span class="faint" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span></header>
				<div class="note-body">{{ message .Body }}</div>
			</div>
		</li>
		{{- else }}
		<li class="empty-state">Nothing said yet.</li>
		{{- end }}
	</ol>
	<form class="note-form chat-form" data-note-form data-channel="chat">
		<textarea name="body" rows="2" required placeholder="Message everyone, Enter sends" aria-label="A message"></textarea>
		<button class="btn btn-primary" type="submit">Send</button>
	</form>
</section>
{{ end }}
