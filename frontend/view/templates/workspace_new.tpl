{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Add workspace</h2>
		<p>Makes a directory under the root, the folder you name, with an empty <code>home/</code>, the <code>HOME</code> of the workspace's runs, and a <code>src/</code> for its projects. With a repository URL, it clones the repository into the home, where atkins finds <code>.atkins/skills</code> and Claude its <code>CLAUDE.md</code>. If the clone fails, you see git's output here and task-ui keeps nothing. Then clone projects into it with Add project.</p>
	</div>
</div>

<form class="form" data-api="/api/workspaces" data-method="POST" data-next="/settings/workspaces/{folder}?notice=Workspace+added">
	<div class="field">
		<label for="w-folder">Folder <span class="faint">one name, no slashes</span></label>
		<input type="text" id="w-folder" name="folder" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off" required autofocus>
		<small class="faint">The namespace of its projects.</small>
	</div>
	<div class="field">
		<label for="w-url">Home repository <span class="faint">optional, a repository to clone into the home</span></label>
		<input type="text" id="w-url" name="url" placeholder="git@github.com:org/workspace.git" spellcheck="false" autocomplete="off">
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
		<label for="w-branch">Home branch <span class="faint">optional, defaults to the repository's default branch</span></label>
		<input type="text" id="w-branch" name="branch" spellcheck="false" autocomplete="off">
	</div>
	{{- template "recursivefield" (list "w-recursive" true) }}
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces">Cancel</a>
		<button class="btn btn-primary" type="submit">Add workspace</button>
	</div>
</form>
{{ end }}
