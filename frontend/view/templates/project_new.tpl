{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Add project</h2>
		<p>Clones the repository into a workspace: the default one, _, is the root. The clone is the test: it either succeeds and the folder appears, or git's output is shown here and nothing is kept. Private repositories are reached with the instance SSH keys.</p>
	</div>
</div>

<form class="form" data-api="/api/projects" data-method="POST" data-next="/settings/projects?notice=Project+added">
	<div class="field">
		<label for="p-url">Repository URL</label>
		<input type="text" id="p-url" name="url" placeholder="git@github.com:org/infra.git" spellcheck="false" autocomplete="off" required autofocus>
	</div>
	<div class="field">
		<label for="p-id">Folder <span class="faint">optional, the repository name by default</span></label>
		<input type="text" id="p-id" name="id" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label for="p-workspace">Workspace <span class="faint">_ clones the project at the top of the root; a named one clones it inside, as workspace/folder</span></label>
		<select id="p-workspace" name="workspace">
			<option value="_" selected>_ · the root</option>
			{{- range .WorkspaceList }}{{ if ne .Folder "_" }}
			<option value="{{ .Folder }}">{{ .Folder }}{{ with .Title }} · {{ . }}{{ end }}</option>
			{{- end }}{{ end }}
		</select>
	</div>
	<div class="field">
		<label for="p-branch">Branch <span class="faint">optional, the default branch otherwise</span></label>
		<input type="text" id="p-branch" name="branch" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label class="check"><input type="checkbox" id="p-recursive" name="recursive" checked> Recursive <span class="faint">list the folders inside the checkout too; off lists its own jobs alone</span></label>
	</div>
	<div class="field">
		<label class="check"><input type="checkbox" id="p-sidebar" name="sidebar" checked> Show in sidebar <span class="faint">off keeps the project out of the sidebar; its pages stay</span></label>
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/projects">Cancel</a>
		<button class="btn btn-primary" type="submit">Clone</button>
	</div>
</form>
{{ end }}
