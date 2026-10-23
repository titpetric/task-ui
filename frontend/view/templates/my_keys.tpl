{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Your SSH keys</h2>
		<p>The keys your runs get through their ssh agent: keys made for you, and keys granted to your role{{ if .User }}, {{ .User.EffectiveRole }}{{ end }}. Copy a public key to a git host or a server that needs to trust it. The private key stays on the task-ui server.</p>
	</div>
</div>

<section class="block">
	<h3>Keys</h3>
	<div class="table-wrap">
	<table class="table keys key-table cards">
		<colgroup><col class="c-name"><col class="c-algo"><col class="c-size"><col class="c-pub"><col class="c-purpose"><col class="c-actions"></colgroup>
		<thead><tr><th>Name</th><th>Algorithm</th><th>Size</th><th>Public key</th><th>For</th><th class="actions" aria-label="Actions"></th></tr></thead>
		<tbody>
		{{- range .Keys }}
			{{- $users := index $.KeyUsers .Name }}{{ $mine := false }}{{ range $users }}{{ if and $.User (eq . $.User.Name) }}{{ $mine = true }}{{ end }}{{ end }}
			<tr>
				<td class="key"><span class="mono">{{ .Name }}</span>{{ with .Comment }}<div class="faint">{{ . }}</div>{{ end }}</td>
				<td class="mono" data-label="Algorithm">{{ .Algo }}</td>
				<td class="mono nowrap" data-label="Size">{{ .Size }} bits</td>
				<td class="pubkey" data-label="Public key"><div class="pub-line"><span class="mono" title="{{ .PublicKey }}">{{ .ShortPublic }}</span><button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy the public key of {{ .Name }}"><svg class="ic"><use href="#i-copy"/></svg></button></div></td>
				<td data-label="For">{{ if $mine }}you{{ else }}your role{{ end }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on " .Name)
					(detail "Fingerprint" .Fingerprint "mono" "")
					(and (not .CreatedAt.IsZero) (detail "Created" (ago .CreatedAt) "title" (stamp .CreatedAt)))
					(item "copy" "Copy public key" "data-copy" .PublicKey)
					(item "download" "Download public key" "data-download" .PublicKey "data-filename" (print .Name ".pub"))
				) }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="6">No keys for you or your role yet. Someone with Allow Edits on Instance Keys can grant one under Settings, Keys.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
