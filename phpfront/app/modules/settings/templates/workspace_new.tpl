{* The add workspace form. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="form" data-api="/api/workspaces" data-method="POST" data-next="{data.next}">
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
{eval $data_recursive = $data.recursive}{include components/recursive-field.tpl}
	<div class="form-foot">
		<a class="btn" href="{data.cancel}">Cancel</a>
		<button class="btn btn-primary" type="submit">Add workspace</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
