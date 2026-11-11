{eval $data_head = $data.head}
{include components/page-head.tpl}

<form method="get" action="{base}/runs" class="filters">
	<label for="runs-folder">Folder
		<select id="runs-folder" name="folder">
			<option value="">all folders</option>
{foreach $data.folders as $folder}<option value="{folder}"{if $data.form.folder === $folder} selected{/if}>{folder}</option>
{/foreach}		</select>
	</label>
	<label>Job <input type="text" name="job" value="{data.form.job}" spellcheck="false" autocomplete="off"></label>
	<label for="runs-state">State
		<select id="runs-state" name="state">
			<option value="">any state</option>
{foreach $data.states as $state}<option value="{state}"{if $data.form.state === $state} selected{/if}>{state}</option>
{/foreach}		</select>
	</label>
	<label for="runs-user">By
		<select id="runs-user" name="user">
			<option value="">anyone</option>
{foreach $data.users as $user}<option value="{user}"{if $data.form.user === $user} selected{/if}>{user}</option>
{/foreach}		</select>
	</label>
	<label class="page-size" for="page-size">Page size
		<select id="page-size" name="n">
{foreach $data.sizes as $size}<option value="{size}"{if $data.form.n === $size} selected{/if}>{size}</option>
{/foreach}		</select>
	</label>
	<button class="btn" type="submit">Filter</button>
</form>

{if $data.activity}<section class="block" data-fold="runs-activity">
	<h3>{eval $data_fold = $data.fold}{include components/fold.tpl} <small>in the server's time zone</small></h3>
	<div class="fold-body" id="runs-activity">
{eval $data_chart = $data.activity}
{include components/chart.tpl}
	</div>
</section>
{/if}
{if $data.tooling}<section class="block" data-fold="runs-tooling">
	<h3>{eval $data_fold = $data.toolFold}{include components/fold.tpl} <small>what the commands inside the runs spent their time on</small></h3>
	<div class="fold-body" id="runs-tooling">
{eval $data_tooling = $data.tooling}
{include components/tooling.tpl}
	</div>
</section>
{/if}
<section class="block" data-sse="runs">
	<h3>Runs <small>newest first</small></h3>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Folder</th><th>Job</th><th>By</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody data-runs-table{if $data.summarize} data-summarize{/if}>
{foreach $data.rows as $row}			<tr data-run-id="{row.id}">
				<td class="nowrap" data-label="Started" title="{row.stamp}">{row.ago}</td>
				<td data-label="Folder">{if $row.folderHref}<a href="{base}{row.folderHref}">{row.place}</a>{else}{row.place}{/if}</td>
				<td class="key">{if $row.jobHref}<a href="{base}{row.jobHref}">{row.job}</a>{else}{row.job}{/if}</td>
				<td data-label="By"><span class="who-ran">{eval $data_user = $row.who.user}{include components/user.tpl}{if $row.who.scheduled} <span class="badge">scheduled</span>{/if}{if $row.who.reason} <span class="badge bad" title="{row.who.reason}">out of hours</span>{elseif $row.who.forced} <span class="badge signal">forced</span>{/if}</span></td>
				<td class="nowrap" data-label="Took" data-took data-started="{row.started}">{row.duration}</td>
				<td data-label="Result" data-result>{eval $data_state = $row.state}{include components/state.tpl}{if $row.message} <span class="faint">{row.message}</span>{/if}{if $row.title} <span class="term-title" title="{row.title}">{row.title}</span>{/if}{if $row.watching} <span class="faint">{row.watching} watching</span>{/if}</td>
				<td class="actions" data-actions>
					<a class="btn btn-sm" href="{base}/runs/{row.id}" data-attach="{row.id}">{row.attach}</a>
{eval $data_menu = $row.menu}
{include components/menu.tpl}
				</td>
			</tr>
{else}			<tr class="empty"><td colspan="7">No runs match. Run a job from its folder and it appears here.</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
{eval $data_pager = $data.pager}
{include components/pager.tpl}
</section>
