{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Your SSH keys</h2>
		<p>The keys your runs get through their ssh agent: keys made for you, and keys granted to your role{{ if .User }}, {{ .User.EffectiveRole }}{{ end }}. Copy a public key to a git host or a server that needs to trust it. The private key stays on the task-ui server.</p>
	</div>
</div>

<section class="block">
	{{- template "blockhead" (head "Keys" "" (and (not .Keys) (or (and (.Can "key/edit") "No keys for you or your role yet. Generate or upload one for yourself.") "No keys for you or your role yet. Someone with Allow Edits on Instance Keys can grant one under Settings, Keys."))
		(and .User (.Can "key/edit") (item "key" "Generate key" "href" (userKeyHref "generate" .User.Name)))
		(and .User (.Can "key/edit") (item "upload" "Upload key" "href" (userKeyHref "upload" .User.Name)))) }}
	{{- with .Keys }}
	<div class="table-wrap fit">
	<table class="table keys key-table cards">
		{{ template "keyhead" }}
		<tbody>
		{{- range . }}{{ template "keyrow" (list $ . "/keys") }}{{ end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{ end }}
