{{/* thread takes (list $ thread): a note with its replies, each in
     markdown under its author, and a field to reply. A note to the agent
     or a prompt carries its state, a spinner while it works, and links
     its run. Its author, or an admin, clicks a note's text to edit it;
     the agent's answer that runs again shows working, its old text
     dimmed, until the new one takes its place. The notes board and the
     prompt box draw it. */}}
{{- define "thread" }}{{ $p := index . 0 }}{{ with index . 1 }}{{ $t := . }}
	<article class="note{{ if .State }} instruction {{ .State }}{{ end }}" id="note-{{ .ID }}">
		<header>
			<span class="author">{{ template "user" (list $p .UserName) }}</span>
			<span class="faint" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span>
			{{- if .EditedAt }} <span class="faint edited" title="Edited {{ stamp .EditedAt }}">edited</span>{{ end }}
			{{- if .State }} <span class="badge note-state {{ .State }}">{{ .State }}</span>{{ end }}
			{{- if .RunID }} <a class="faint mono" href="/runs/{{ .RunID }}" title="The run the agent worked in">run {{ slice .RunID 0 8 }}</a>{{ end }}
			<span class="spacer"></span>
			{{- template "menu" (menu (print "Actions on the note of " .UserName)
				(and ($p.Edits .Note $t) (item "edit" "Edit" "data-edit-open" .ID))
				(and .RunID (item "terminal" "Open its run" "href" (print "/runs/" .RunID)))
				(and ($p.Deletes .Note) (danger "trash" "Delete" "data-delete" (print "/api/notes/" .ID) "data-confirm" (print "Delete the note of " .UserName " and its replies?") "data-confirm-ok" "Delete" "data-next" "reload"))) }}
		</header>
		<div class="note-body"{{ if $p.Edits .Note $t }} data-edit="{{ .ID }}" data-body="{{ .Body }}" title="Click to edit"{{ end }}>{{ $p.Message .Body }}</div>
		{{- if .Replies }}
		<ol class="replies">
			{{- range .Replies }}
			<li{{ if eq .State "working" }} class="rerun"{{ end }}>
				<header><span class="author">{{ template "user" (list $p .UserName) }}</span> <span class="faint" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</span>{{ if .EditedAt }} <span class="faint edited" title="Edited {{ stamp .EditedAt }}">edited</span>{{ end }}{{ if eq .State "working" }} <span class="badge note-state working">working</span>{{ end }}{{ if .RunID }} <a class="faint mono" href="/runs/{{ .RunID }}" title="The run this reply comes from">run {{ slice .RunID 0 8 }}</a>{{ end }}<span class="spacer"></span>{{ template "menu" (menu (print "Actions on the reply of " .UserName)
					(and ($p.Edits . $t) (item "edit" "Edit" "data-edit-open" .ID))
					(and .RunID (item "terminal" "Open its run" "href" (print "/runs/" .RunID)))
					(and ($p.Deletes .) (danger "trash" "Delete" "data-delete" (print "/api/notes/" .ID) "data-confirm" (print "Delete this reply of " .UserName "?") "data-confirm-ok" "Delete" "data-next" "reload"))) }}</header>
				<div class="note-body"{{ if $p.Edits . $t }} data-edit="{{ .ID }}" data-body="{{ .Body }}" title="Click to edit"{{ end }}>{{ $p.Message .Body }}</div>
			</li>
			{{- end }}
		</ol>
		{{- end }}
		<details class="reply">
			<summary>Reply</summary>
			<form class="note-form" data-note-form data-parent="{{ .ID }}">
				<textarea name="body" rows="2" required data-mentions aria-label="A reply"></textarea>
				<div class="note-form-foot"><span></span><button class="btn" type="submit">Reply</button></div>
			</form>
		</details>
	</article>
{{- end }}{{ end }}
