{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Settings</h2>
		<p>What the instance brings to every run: the projects it keeps up to date, the SSH keys it authenticates with, the secrets it adds to the environment, and who may sign in.</p>
	</div>
</div>

<div class="table-wrap">
<table class="table settings">
	<tbody>
		{{- $keys := 0 }}{{ range .Keys }}{{ if not .Backup }}{{ $keys = add $keys 1 }}{{ end }}{{ end }}
		<tr>
			<td class="key nowrap"><a href="/settings/projects"><svg class="ic"><use href="#i-folder"/></svg> Projects</a></td>
			<td class="desc">The folders under the root. Add one from a git repository; a checkout is pulled before every run.</td>
			<td class="nowrap dim">{{ len .Entries }} {{ if eq (len .Entries) 1 }}folder{{ else }}folders{{ end }}</td>
			<td class="actions"><a class="btn btn-sm" href="/settings/projects">Open</a></td>
		</tr>
		<tr>
			<td class="key nowrap"><a href="/settings/keys"><svg class="ic"><use href="#i-key"/></svg> SSH keys</a></td>
			<td class="desc">Key pairs the jobs use to reach hosts, git remotes and services. Generate, rotate or upload them.</td>
			<td class="nowrap dim">{{ $keys }} {{ if eq $keys 1 }}key{{ else }}keys{{ end }} <span class="faint">{{ .KeysDir }}</span></td>
			<td class="actions"><a class="btn btn-sm" href="/settings/keys">Open</a></td>
		</tr>
		<tr>
			<td class="key nowrap"><a href="/settings/secrets"><svg class="ic"><use href="#i-secret"/></svg> Secrets</a></td>
			<td class="desc">Variables added to the environment of every run, for all folders or for one, and written into the env files the pipelines read.</td>
			<td class="nowrap dim">{{ len .Secrets }} {{ if eq (len .Secrets) 1 }}secret{{ else }}secrets{{ end }}</td>
			<td class="actions"><a class="btn btn-sm" href="/settings/secrets">Open</a></td>
		</tr>
		<tr>
			<td class="key nowrap"><a href="/settings/integrations"><svg class="ic"><use href="#i-plug"/></svg> Integrations</a></td>
			<td class="desc">The services task-ui talks to, and their API keys and tokens, encrypted with the instance keys: Claude Code, Beszel.</td>
			<td class="nowrap dim"></td>
			<td class="actions"><a class="btn btn-sm" href="/settings/integrations">Open</a></td>
		</tr>
		<tr>
			<td class="key nowrap"><a href="/settings/users"><svg class="ic"><use href="#i-users"/></svg> Users</a></td>
			<td class="desc">Who may sign in, from the password file, with the role the lists give, and which folders they see.</td>
			<td class="nowrap dim">{{ len .Users }} {{ if eq (len .Users) 1 }}user{{ else }}users{{ end }}</td>
			<td class="actions"><a class="btn btn-sm" href="/settings/users">Open</a></td>
		</tr>
	</tbody>
</table>
</div>
{{ end }}
