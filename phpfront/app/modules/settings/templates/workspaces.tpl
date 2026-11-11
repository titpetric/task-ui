{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="block">
	<h3>Workspaces</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Folder</th><th>Title</th><th>Home repository</th><th class="hide-sm">Projects</th><th class="hide-sm">Secrets</th><th></th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr>
				<td class="key nowrap"><a href="{row.href}"><svg class="ic"><use href="#i-folder-open"/></svg> {row.folder}</a></td>
				<td class="desc">{row.title}{if $row.description} <span class="faint">{row.description}</span>{/if}</td>
				<td class="origin">{if $row.home}<span class="remote mono" title="{row.home}">{row.home}</span>{else}<span class="faint">empty</span>{/if}</td>
				<td class="hide-sm num">{if $row.projects}<a href="{row.edit}" title="{row.names}">{row.projects}</a>{else}<span class="faint">none</span>{if $row.adds} <a href="{row.add}" aria-label="Add a project to {row.folder}">add</a>{/if}{/if}</td>
				<td class="hide-sm num">{if $row.secrets}<a href="{row.variables}" title="{row.blocked}">{row.secrets}</a>{else}<span class="faint">none</span>{if $row.addsSecret} <a href="{row.variables}" aria-label="Add a secret to {row.folder}">add</a>{/if}{/if}</td>
				<td class="actions">{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
			</tr>
{/foreach}{if !$data.rows}			<tr class="empty"><td colspan="6">No workspaces yet. Add one, with its home from a git repository, then clone projects into it with Add project.</td></tr>
{/if}		</tbody>
	</table>
	</div>
</section>
<p class="faint" style="margin-top:0.6rem">Root: {data.root}</p>
{if $data.imports}
<dialog class="confirm workspace-import" id="import-dialog" aria-labelledby="import-title">
	<form class="form" data-import="/api/workspaces/import">
		<h3 id="import-title">Import a workspace</h3>
		<p>Adds a workspace from a bundle another task-ui exported, opened with its passphrase. Import makes the workspace's folder with its home and src, adds its SSH keys to the key folder, encrypts its secrets with this instance's encryption key, adds its projects, env files and grants, and then clones its home repository and each project. If a clone fails, it stays listed with git's output. If anything before the clones fails, task-ui keeps nothing. To import a bundle of the internal workspace, enter a folder name.</p>
		<div class="field">
			<label for="wi-bundle">Bundle <span class="faint">a .tui.age file</span></label>
			<input type="file" id="wi-bundle" name="bundle" accept=".age" required>
		</div>
		<div class="field">
			<label for="wi-passphrase">Passphrase</label>
			<input type="password" id="wi-passphrase" name="passphrase" autocomplete="off" spellcheck="false" required>
		</div>
		<div class="field">
			<label for="wi-folder">Folder <span class="faint">optional, one name, no slashes</span></label>
			<input type="text" id="wi-folder" name="folder" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off">
			<small class="faint">Defaults to the bundle's folder.</small>
		</div>
		<div class="form-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Import</button>
		</div>
	</form>
</dialog>
{/if}{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
