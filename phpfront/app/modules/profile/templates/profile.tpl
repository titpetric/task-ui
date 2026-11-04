{* One user's page. The scope is {data}; the chrome is {chrome}. *}
<div class="page-head profile-head">
	<div class="profile-who">
		<span class="avatar lg role-{data.who.role}" aria-hidden="true">{data.who.initials}</span>
		<div>
			<h2>{data.who.name}</h2>
			<p>{data.who.lead}</p>
		</div>
	</div>
	<div class="actions">
{eval $data_buttons = $data.buttons}{include components/buttons.tpl}{eval $data_menu = $data.menu}{include components/menu.tpl}	</div>
</div>

<section class="block" id="runs">
{eval $data_block = $data.runs.block}{include components/block-head.tpl}
{if $data.runs.rows}	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Job</th><th>Duration</th><th>Result</th><th class="actions" aria-label="Actions"></th></tr></thead>
		<tbody>
{foreach $data.runs.rows as $row}			<tr>
				<td class="nowrap" data-label="Started" title="{row.stamp}">{row.ago}</td>
				<td class="key"><a href="{base}{row.href}">{row.name}</a></td>
				<td class="nowrap" data-label="Took">{row.duration}</td>
				<td data-label="Result">{eval $data_state = $row.state}{include components/state.tpl}</td>
				<td class="actions"><a class="btn btn-sm" href="{base}/runs/{row.id}" data-attach="{row.id}">{row.attach}</a>{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
	<p class="more-link"><a href="{data.runs.more.href}">{data.runs.more.label}</a></p>
{/if}</section>

<section class="block" id="workflow-runs">
{eval $data_block = $data.flows.block}{include components/block-head.tpl}
{if $data.flows.rows}	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Workflow</th><th>Result</th><th class="actions" aria-label="Actions"></th></tr></thead>
		<tbody>
{foreach $data.flows.rows as $row}			<tr>
				<td class="nowrap" data-label="Started" title="{row.stamp}">{row.ago}</td>
				<td class="key"><a href="{base}{row.href}"><svg class="ic lead"><use href="#i-flow"/></svg>{row.name}</a></td>
				<td data-label="Result"><span class="state"><i class="dot {row.state}"></i>{row.state}</span></td>
				<td class="actions"><a class="btn btn-sm" href="{base}{row.view}">View</a></td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}</section>

<section class="block" id="notes">
{eval $data_block = $data.notes.block}{include components/block-head.tpl}
{if $data.notes.rows}	<ul class="profile-list">
{foreach $data.notes.rows as $row}		<li><svg class="ic"><use href="#i-note"/></svg><a class="excerpt" href="{row.href}">{row.excerpt}</a>{if $row.state} <span class="badge note-state {row.state}">{row.state}</span>{/if}<span class="faint when" title="{row.stamp}">{row.ago}</span></li>
{/foreach}	</ul>
{/if}</section>

<section class="block" id="prompts">
{eval $data_block = $data.prompts.block}{include components/block-head.tpl}
{if $data.prompts.rows}	<ul class="profile-list">
{foreach $data.prompts.rows as $row}		<li><svg class="ic"><use href="#i-terminal"/></svg><a class="excerpt" href="{row.href}">{row.excerpt}</a> <span class="faint mono">{row.dir}/</span>{if $row.state} <span class="badge note-state {row.state}">{row.state}</span>{/if}<span class="faint when" title="{row.stamp}">{row.ago}</span></li>
{/foreach}	</ul>
{/if}</section>
{if $data.keys}{eval $data_user_keys = $data.keys}{include components/user-keys.tpl}{/if}
