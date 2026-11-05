{eval $data_head = $data.head}
{include components/page-head.tpl}

{if $data.space}<section class="block">
	<h3>Space</h3>
{eval $data_space = $data.space}
{include components/space.tpl}
</section>
{/if}
<section class="block">
{eval $data_block = $data.block}{include components/block-head.tpl}
	<div class="table-wrap">
	<table class="table file-table">
		<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
		<tbody>
{foreach $data.files as $row}			<tr>
				<td class="key"><svg class="ic{if $row.dir} dir{/if}"><use href="#i-{if $row.dir}folder{else}file{/if}"/></svg>{if $row.dir} <a href="{row.href}">{row.name}/</a>{else} <a href="{row.href}" data-artifact="{row.path}">{row.name}</a>{/if}</td>
				<td class="num nowrap">{row.size}</td>
				<td class="hide-sm nowrap" title="{row.stamp}">{row.ago}</td>
				<td class="actions">{eval $data_menu = $row.menu}
{include components/menu.tpl}
				</td>
			</tr>
{else}			<tr class="empty"><td colspan="4">This folder is empty.</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>History</h3>
	<form method="get" action="{base}/artifacts" class="filters">
		<label for="art-folder">Folder
			<select id="art-folder" name="folder">
				<option value="">all folders</option>
{foreach $data.folders as $folder}<option value="{folder}"{if $data.form.folder === $folder} selected{/if}>{folder}</option>
{/foreach}			</select>
		</label>
		<label>Job <input type="text" name="job" value="{data.form.job}" spellcheck="false" autocomplete="off"></label>
		<button class="btn" type="submit">Filter</button>
	</form>

	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Created</th><th>Folder</th><th>Job</th><th>Run</th><th>Name</th><th class="hide-sm">Version</th><th>Size</th><th></th></tr></thead>
		<tbody>
{foreach $data.history as $row}			<tr>
				<td class="nowrap" title="{row.stamp}">{row.ago}</td>
{if $row.run}				<td>{if $row.folderHref}<a href="{row.folderHref}">{row.folder}</a>{else}{row.folder}{/if}</td>
				<td>{if $row.jobHref}<a href="{row.jobHref}">{row.job}</a>{else}{row.job}{/if}</td>
				<td><a class="faint mono" href="{row.runHref}" title="{row.run}">{row.short}</a></td>
{else}				<td colspan="3" class="dim">written over MCP or the API by {eval $data_user = $row.user}{include components/user.tpl}</td>
{/if}				<td class="key">{if $row.commit}<a href="{row.nameHref}" data-artifact="{row.path}" data-version="{row.commit}" title="View this version">{row.name}</a>{else}<a href="{row.nameHref}">{row.name}</a>{/if}</td>
				<td class="hide-sm">{if $row.commit}<span class="faint mono" title="{row.commit}">{row.version}</span>{else}<span class="faint">kept apart</span>{/if}</td>
				<td class="nowrap">{row.size}</td>
				<td class="actions">{eval $data_menu = $row.menu}
{include components/menu.tpl}
				</td>
			</tr>
{else}			<tr class="empty"><td colspan="8">No artifacts yet. Files a job writes to <code>$TASK_UI_ARTIFACTS</code> appear here.</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
</section>
