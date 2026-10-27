{{ define "content" }}
{{- $ws := index .Form "workspace" }}{{ $back := "/settings/projects" }}{{ if $ws }}{{ $back = print "/settings/workspaces/" $ws }}{{ end }}
<div class="page-head">
	<div>
		<h2>Add project</h2>
		<p>Clones the repository into a workspace; the default workspace, _, is the root. If the clone works, the folder appears. If it fails, you see git's output here and task-ui keeps nothing. task-ui uses the instance SSH keys to reach private repositories.</p>
	</div>
</div>

<form class="form" data-api="/api/projects" data-method="POST" data-next="{{ $back }}?notice=Project+added">
	<div class="field">
		<label for="p-url">Repository URL</label>
		<input type="text" id="p-url" name="url" placeholder="git@github.com:org/infra.git" spellcheck="false" autocomplete="off" required autofocus>
	</div>
	<div class="field">
		<label for="p-id">Folder <span class="faint">optional, defaults to the repository name</span></label>
		<input type="text" id="p-id" name="id" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label for="p-workspace">Workspace</label>
		<select id="p-workspace" name="workspace" required>
			{{- range $i, $w := .WorkspaceList }}
			<option value="{{ .Folder }}"{{ if or (eq .Folder $ws) (and (not $ws) (eq $i 0)) }} selected{{ end }}>{{ .Folder }} · {{ if eq .Folder "internal" }}the root{{ else }}{{ or .Title .Folder }}{{ end }}</option>
			{{- end }}
		</select>
		<small class="faint">_, the default workspace while there is one, clones the project at the top of the root; any other clones it inside that workspace, as workspace/folder.</small>
	</div>
	<div class="field">
		<label for="p-branch">Branch <span class="faint">optional, defaults to the repository's default branch</span></label>
		<input type="text" id="p-branch" name="branch" spellcheck="false" autocomplete="off">
	</div>
	{{- template "recursivefield" (list "p-recursive" true) }}
	<div class="field">
		<label class="check"><input type="checkbox" id="p-sidebar" name="sidebar" checked> Show in sidebar</label>
		<small class="faint">Off hides the project from the sidebar; its pages still work.</small>
	</div>
	<div class="form-foot">
		<a class="btn" href="{{ $back }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Add project</button>
	</div>
</form>
{{ end }}
