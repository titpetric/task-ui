{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Add workspace</h2>
		<p>Clones the repository into the root, under the folder you name. The clone is the test: it either succeeds and the workspace appears, or git's output is shown here and nothing is kept. Projects then clone inside it from Add project.</p>
	</div>
</div>

<form class="form" data-api="/api/workspaces" data-method="POST" data-next="/settings/workspaces?notice=Workspace+added">
	<div class="field">
		<label for="w-url">Repository URL</label>
		<input type="text" id="w-url" name="url" placeholder="git@github.com:org/workspace.git" spellcheck="false" autocomplete="off" required autofocus>
	</div>
	<div class="field">
		<label for="w-folder">Folder <span class="faint">the namespace of its projects; one name, no slashes</span></label>
		<input type="text" id="w-folder" name="folder" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off" required>
	</div>
	<div class="field">
		<label for="w-title">Title <span class="faint">optional</span></label>
		<input type="text" id="w-title" name="title" autocomplete="off">
	</div>
	<div class="field">
		<label for="w-description">Description <span class="faint">optional</span></label>
		<input type="text" id="w-description" name="description" autocomplete="off">
	</div>
	<div class="field">
		<label for="w-branch">Branch <span class="faint">optional, the default branch otherwise</span></label>
		<input type="text" id="w-branch" name="branch" spellcheck="false" autocomplete="off">
	</div>
	{{- template "recursivefield" (list "w-recursive" true) }}
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces">Cancel</a>
		<button class="btn btn-primary" type="submit">Clone</button>
	</div>
</form>
{{ end }}
