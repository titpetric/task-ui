{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Your SSH keys</h2>
		<p>The keys your runs get through their ssh agent: keys made for you, and keys granted to your role{{ if .User }}, {{ .User.EffectiveRole }}{{ end }}. Copy a public key to a git host or a server that needs to trust it. The private key stays on the task-ui server.</p>
	</div>
</div>

<section class="block">
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Key</th><th>Fingerprint</th><th>For</th><th>Actions</th></tr></thead>
		<tbody>
		{{- range .Keys }}
			{{- $users := index $.KeyUsers .Name }}
			<tr>
				<td class="key nowrap">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="fingerprint"><span class="kind dim">{{ .Type }}</span>{{ .Fingerprint }}</td>
				<td class="nowrap">{{ $mine := false }}{{ range $users }}{{ if and $.User (eq . $.User.Name) }}{{ $mine = true }}{{ end }}{{ end }}{{ if $mine }}you{{ else }}your role{{ end }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">No keys for you or your role yet. Someone with Allow Edits on Instance Keys can grant one under Settings, Keys.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
