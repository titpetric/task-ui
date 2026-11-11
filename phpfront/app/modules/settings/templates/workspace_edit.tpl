{eval $data_head = $data.head}
{include components/page-head.tpl}
{eval $set = $data.settings}
<section class="block">
	<h3>Settings <small class="mono-break">{set.dir}</small></h3>
	<form class="form" data-api="{set.action}" data-method="PUT" data-next="{set.next}">
		<div class="field">
			<label for="we-title">Title</label>
			<input type="text" id="we-title" name="title" value="{set.title}" autocomplete="off">
		</div>
		<div class="field">
			<label for="we-description">Description</label>
			<input type="text" id="we-description" name="description" value="{set.description}" autocomplete="off">
		</div>
{eval $data_recursive = $set.recursive}{include components/recursive-field.tpl}
		<div class="field">
			<label for="we-sidebar">Sidebar</label>
			<input type="number" id="we-sidebar" name="sidebar_limit" data-type="int" min="0" step="1" placeholder="all" value="{set.limit}" autocomplete="off">
			<small class="faint">How many projects the sidebar lists for this workspace, the most recently committed first. Empty or 0 lists them all, by name.</small>
		</div>
		<div class="field">
			<label for="we-hours">Hours</label>
			<textarea id="we-hours" name="hours" rows="3" spellcheck="false" autocomplete="off" placeholder="mon-fri 9-17&#10;fri 9-13">{set.hours}</textarea>
		</div>
		<div class="form-foot">
			<a class="btn" href="{set.cancel}">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
</section>
{eval $home = $data.home}
<section class="block workspace-home" id="home">
{eval $data_block = $home.block}{include components/block-head.tpl}{if !$home.empty}	<p><a class="mono-break" href="{home.href}"><svg class="ic"><use href="#i-folder-open"/></svg> {home.folder}</a> <span class="faint mono-break">{home.url}</span>{if $home.branch} <span class="badge">{home.branch}</span>{/if}</p>
{eval $git = $home.checkout}{if $git.found}	<p class="dim">{git.branch} <span class="faint">{git.commit}</span> {git.subject}{if $git.ago} <span class="faint">{git.ago}</span>{/if}; pulled before every run.</p>
{else}	<p class="dim">The home holds no checkout: the clone failed, or someone replaced the directory by hand. Replace clones it again.</p>
{/if}{/if}	<p class="dim">Every run in {home.workspace} gets <code>HOME</code> here: <code>bash -l</code> reads its <code>.bash_profile</code>, atkins its <code>.atkins/skills</code> and <code>.atkins/atkins.db</code>, Claude its <code>CLAUDE.md</code> and <code>.claude</code> skills. Registry logins, Claude Code's settings and the build caches stay in the server's home.</p>
</section>
{if $home.edit}
<dialog class="confirm" id="home-replace-dialog" aria-labelledby="home-replace-title">
	<form class="form" data-api="{home.action}" data-method="PUT" data-next="{home.next}">
		<h3 id="home-replace-title">{home.clone} the home repository</h3>
		<p>Removes <span class="mono-break">{home.folder}/</span> and clones the repository in its place. task-ui refuses while a run of {home.workspace} is active. If the clone fails, you see git's output and the home stays an empty directory.</p>
		<div class="field">
			<label for="hr-url">Repository URL</label>
			<input type="text" id="hr-url" name="url" value="{home.url}" placeholder="git@github.com:org/workspace.git" spellcheck="false" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="hr-branch">Branch <span class="faint">optional</span></label>
			<input type="text" id="hr-branch" name="branch" value="{home.branch}" spellcheck="false" autocomplete="off">
		</div>
		<div class="form-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Clone</button>
		</div>
	</form>
</dialog>
{/if}
<section class="block" id="projects">
{eval $data_block = $data.block}{include components/block-head.tpl}{eval $data_projects = $data.table}{if $data_projects.rows}{include components/project-table.tpl}
{/if}</section>
{if $data.keys}{eval $keys = $data.keys}
<section class="block workspace-keys" id="keys">
{eval $data_block = $keys.block}{include components/block-head.tpl}{if $keys.rows}	<p class="dim">A key goes to every run in the workspace, through the run's ssh agent.</p>
{eval $data_keys = $keys.rows}{include components/key-table.tpl}{/if}</section>
{if $data.dialogs}{eval $data_wsdialogs = $data.dialogs}{include components/ws-key-dialogs.tpl}{/if}
<dialog class="confirm workspace-export" id="export-dialog" aria-labelledby="export-title">
	<form class="form" data-export="/api/workspaces/{data.export}/export">
		<h3 id="export-title">Export {data.export}</h3>
		<p>Downloads <span class="mono-break">{data.export}.tui.age</span> for Import on another task-ui's Workspaces page: the workspace, its projects, env files and grants, its secrets and the SSH keys its runs get, encrypted to the passphrase. Anyone with the file and the passphrase has the secrets and the keys. No instance key and no encryption key goes; schedules, workflows, runs and artifacts stay here.</p>
		<div class="field">
			<label for="wx-passphrase">Passphrase <span class="faint">at least 12 characters</span></label>
			<div class="passphrase">
				<input type="password" id="wx-passphrase" name="passphrase" minlength="12" autocomplete="new-password" spellcheck="false" required>
				<button type="button" class="btn-icon" data-copy="" data-passphrase-copy title="Copy passphrase" aria-label="Copy passphrase" hidden><svg class="ic"><use href="#i-copy"/></svg></button>
				<button type="button" class="btn" data-passphrase-generate>Generate</button>
			</div>
		</div>
		<div class="field" data-passphrase-confirm>
			<label for="wx-confirm">Confirm passphrase</label>
			<input type="password" id="wx-confirm" name="confirm" autocomplete="new-password" spellcheck="false">
		</div>
		<p class="passphrase-note" data-passphrase-note hidden>Store this passphrase before you export: task-ui keeps no copy, and nothing recovers it.</p>
		<div class="form-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		</div>
	</form>
</dialog>
{/if}{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
