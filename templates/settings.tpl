{{ define "content" }}
{{ template "crumbs" (list (crumb "Settings" "")) }}
<div class="page-head">
	<div>
		<h2>Settings</h2>
		<p>What the instance brings to every run: the SSH keys it authenticates with, and the secrets it adds to the environment.</p>
	</div>
</div>

<div class="table-wrap">
<table class="table settings">
	<tbody>
		{{- $keys := 0 }}{{ range .Keys }}{{ if not .Backup }}{{ $keys = add $keys 1 }}{{ end }}{{ end }}
		<tr>
			<td class="key nowrap"><a href="/settings/keys"><svg class="ic"><use href="#i-key"/></svg> SSH keys</a></td>
			<td class="desc">Key pairs the jobs use to reach hosts, git remotes and services. Generate, rotate or upload them.</td>
			<td class="nowrap dim">{{ $keys }} {{ if eq $keys 1 }}key{{ else }}keys{{ end }} <span class="faint">{{ .KeysDir }}</span></td>
			<td class="actions"><a class="btn btn-sm" href="/settings/keys">Open</a></td>
		</tr>
		<tr>
			<td class="key nowrap"><a href="/settings/secrets"><svg class="ic"><use href="#i-secret"/></svg> Secrets</a></td>
			<td class="desc">Variables added to the environment of every run, for all folders or for one, and written into the env files the pipelines read.</td>
			<td class="nowrap dim">{{ len .Secrets }} {{ if eq (len .Secrets) 1 }}secret{{ else }}secrets{{ end }} <span class="faint">{{ .EnvPath }}</span></td>
			<td class="actions"><a class="btn btn-sm" href="/settings/secrets">Open</a></td>
		</tr>
	</tbody>
</table>
</div>
{{ end }}
