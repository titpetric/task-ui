{* The add project form. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="form" data-api="/api/projects" data-method="POST" data-next="{data.next}">
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
{foreach $data.options as $option}			<option value="{option.value}"{if $option.selected} selected{/if}>{option.folder} · {option.title}</option>
{/foreach}		</select>
		<small class="faint">internal, the default workspace while there is one, clones the project at the top of the root; any other clones it inside that workspace, as workspace/folder.</small>
	</div>
	<div class="field">
		<label for="p-branch">Branch <span class="faint">optional, defaults to the repository's default branch</span></label>
		<input type="text" id="p-branch" name="branch" spellcheck="false" autocomplete="off">
	</div>
{eval $data_recursive = $data.recursive}{include components/recursive-field.tpl}
	<div class="field">
		<label class="check"><input type="checkbox" id="p-sidebar" name="sidebar" checked> Show in sidebar</label>
		<small class="faint">Off hides the project from the sidebar; its pages still work.</small>
	</div>
	<div class="form-foot">
		<a class="btn" href="{data.back}">Cancel</a>
		<button class="btn btn-primary" type="submit">Add project</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
