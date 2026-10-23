{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Keys</h2>
		<p>There are two kinds of key, and a key can't be both. The <b>encryption key</b> protects secrets: the instance has one, and it encrypts every secret, the workspaces' and the integrations'. <b>SSH keys</b> let runs and git connect to hosts and remotes; git runs over ssh. Instance SSH keys go to the runs of the roles and users they link; a workspace's SSH keys go to that workspace's runs.</p>
	</div>
	{{- if .Can "key/edit" }}
	<div class="actions">
		<a class="btn" href="/settings/keys/upload"><svg class="ic"><use href="#i-upload"/></svg>Upload key</a>
		<a class="btn btn-primary" href="/settings/keys/generate"><svg class="ic"><use href="#i-key"/></svg>Generate key</a>
	</div>
	{{- end }}
</div>

<section class="block key-overview">
	<h3>Scopes <small>the keys of the instance and each workspace</small></h3>
	<div class="table-wrap">
	<table class="table key-purposes">
		<thead><tr><th>Scope</th><th>Secrets</th><th>Encryption key</th><th>SSH keys</th></tr></thead>
		<tbody>
			{{- $instance := 0 }}{{ range .Keys }}{{ if and (not .Encrypts) (not .Workspace) }}{{ $instance = add $instance 1 }}{{ end }}{{ end }}
			<tr class="key-instance" id="instance">
				<td><b>Instance</b><div class="faint">every run, by policy</div></td>
				<td><a href="/settings/integrations">integrations</a><div class="faint">their settings</div></td>
				<td>
					{{- with .InstanceKey }}
					<a class="badge ok" href="#key-{{ .Name }}" title="Its [...] in Every key rotates it">set</a>{{ with .Backup }} {{ template "backup-active" . }}{{ end }}
					<div class="key-fp faint" title="{{ .Name }}">{{ .Fingerprint }}</div>
					{{- else }}
					<span class="badge bad">missing</span> <span class="faint">task-ui creates it on start</span>
					{{- end }}
				</td>
				<td>
					{{- if $instance }}
					<a href="#instance-ssh">{{ $instance }} key{{ if ne $instance 1 }}s{{ end }}</a>
					<div class="key-names">{{ range .Keys }}{{ if and (not .Encrypts) (not .Workspace) }}<a class="key-chip" href="/settings/keys/{{ .Name }}" title="Links of {{ .Name }}">{{ .Name }}</a>{{ end }}{{ end }}</div>
					{{- else }}
					<span class="faint">none: runs get no SSH key from the instance</span>
					{{- end }}
				</td>
			</tr>
			{{- range .KeyWorkspaces }}
			{{- $ws := .Workspace }}{{ $api := printf "/api/workspaces/%s/keys" $ws.Folder }}{{ $base := printf "/settings/workspaces/%s" $ws.Folder }}
			<tr id="ws-{{ $ws.Folder }}">
				<td>
					<a href="{{ $base }}"><b>{{ or $ws.Title $ws.Folder }}</b></a>{{ if and $ws.Title (ne $ws.Title $ws.Folder) }} <span class="badge">{{ $ws.Folder }}</span>{{ end }}
					<div class="faint key-projects">{{ if .Projects }}{{ join .Projects ", " }}{{ else }}no projects{{ end }}</div>
				</td>
				<td>{{ if .Secrets }}<a href="/settings/secrets?workspace={{ $ws.Folder }}">{{ .Secrets }} secret{{ if ne .Secrets 1 }}s{{ end }}</a>{{ else }}<a class="faint" href="/settings/secrets?workspace={{ $ws.Folder }}">none</a>{{ end }}</td>
				<td>
					<span class="faint">the instance's</span>
				</td>
				<td>
					{{- if .SSH }}
					<a href="{{ $base }}#keys">{{ len .SSH }} key{{ if ne (len .SSH) 1 }}s{{ end }}</a>
					<div class="key-names">{{ range .SSH }}<a class="key-chip" href="{{ $base }}/keys/{{ .Name }}" title="Grants of {{ .Name }}">{{ .Name }}</a>{{ end }}</div>
					{{- else }}
					<a class="faint" href="{{ $base }}#keys">none</a>
					<div class="faint">Its runs get only the instance SSH keys their policies grant.</div>
					{{- end }}
				</td>
			</tr>
			{{- end }}
		</tbody>
	</table>
	</div>
</section>

<section class="block" id="instance-ssh">
	{{- template "blockhead" (head "Every key" .KeysDir (and (not .Keys) "No keys added.")) }}
	{{- with .Keys }}
	<p class="dim">Every key lives in the key folder as <code>&lt;name&gt;.key</code> and <code>&lt;name&gt;.pub</code>, the name ending in its algorithm and size. An instance SSH key goes to the runs of the roles and users it links, through their ssh agent, and the inventory signs in to hosts with it; a workspace's SSH key goes to the runs of that workspace. The git key, which a key's page sets, pulls and clones the checkouts. An encryption key never leaves task-ui.</p>
	<div class="table-wrap">
	<table class="table keys key-table cards">
		{{ template "keyhead" }}
		<tbody>
		{{- range . }}{{ template "keyrow" (list $ . "/settings/keys") }}{{ end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>

<section class="block">
	<h3>Who gets which key <small>through the ssh agent of their runs</small></h3>
	<div class="table-wrap">
	<table class="table key-holders">
		<thead><tr><th>Policy or user</th><th>Keys</th></tr></thead>
		<tbody>
		{{- range .KeyHolders }}
			<tr>
				{{- if .User }}
				<td class="nowrap">{{ template "user" (list $ .User.Name) }} <span class="faint">{{ .Role }}</span></td>
				{{- else }}
				<td class="nowrap"><span class="badge">{{ .Role }}</span> <span class="faint">every {{ .Role }}</span></td>
				{{- end }}
				<td>
					{{- $role := .Role }}
					{{- range .Direct }}<a class="key-chip" href="/settings/keys/{{ .Name }}" title="{{ .Name }}, {{ .Fingerprint }}">{{ .Name }}</a>{{ end }}
					{{- range .ViaRole }}<a class="key-chip faint" href="/settings/keys/{{ .Name }}" title="{{ .Name }}, from the {{ $role }} role">{{ .Name }} <small>via role</small></a>{{ end }}
					{{- if not (or .Direct .ViaRole) }}<span class="faint">none</span>{{ end }}
				</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

{{- range .KeyFolders }}
<div class="alert signal"><b>{{ .Project }}</b> has private keys in <b>{{ .Dir }}</b>: {{ range $i, $k := .Keys }}{{ if $i }}, {{ end }}{{ $k.Name }}{{ end }}. Its runs don't get them. Upload a key the project needs as an instance key, then link it to roles or users.</div>
{{- end }}

{{ end }}
