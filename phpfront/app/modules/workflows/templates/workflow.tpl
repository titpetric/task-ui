{* One workflow, its pipeline and its runs. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}{include components/page-head.tpl}

<section class="block" data-sse="workflow:{data.id}" data-user="{data.user}">
	<h3>Pipeline</h3>
{if $data.latest}	<p class="faint" data-latest>Latest run <span class="state"><i class="dot {data.latest.state}"></i>{data.latest.state}</span>{if $data.latest.message} {data.latest.message},{/if} <span title="{data.latest.stamp}">{data.latest.ago}</span> by {data.latest.userName}</p>
{else}	<p class="faint" data-latest>Not run yet.</p>
{/if}	<ol class="pipeline" data-pipeline>
{foreach $data.pipeline as $data_node}{include components/pipenode.tpl}{/foreach}	</ol>
</section>

<section class="block">
	<h3>Runs</h3>
{if $data.activity}{eval $data_chart = $data.activity}
{include components/chart.tpl}
{/if}	<form method="get" action="{base}/workflows/{data.id}" class="filters">
		<label class="page-size" for="page-size">Page size
			<select id="page-size" name="n">
{foreach $data.sizes as $size}<option value="{size}"{if $data.form.n === $size} selected{/if}>{size}</option>
{/foreach}			</select>
		</label>
		<button class="btn" type="submit">Filter</button>
	</form>
	<div class="table-wrap">
	<table class="table cards wf-runs">
		<thead><tr><th>Run</th><th>By</th><th>Steps</th><th></th></tr></thead>
		<tbody data-runs{if $data.summarize} data-summarize{/if}>
{foreach $data.rows as $row}			<tr data-run-id="{row.id}">
				<td class="key" data-result>{eval $data_wffold = $row.fold}{include components/wffold.tpl}<a href="{base}{row.href}" title="{row.stamp}"><i class="dot {row.state}" title="{row.state}"></i>{row.ago}</a>{if $row.message}<small class="faint wf-desc">{row.message}</small>{/if}</td>
				<td data-label="By">{eval $data_user = $row.user}{include components/user.tpl}</td>
				<td data-label="Steps"><span class="chain">{foreach $row.chain as $link}{if $link.then}<span class="faint then" aria-hidden="true"></span>{/if}<a class="nowrap" href="{base}{link.href}" data-pos="{link.pos}"><i class="dot {link.state}"></i>{if $link.workflow}<svg class="ic lead"><use href="#i-flow"/></svg>{elseif $link.summary}<svg class="ic lead"><use href="#i-summary"/></svg>{elseif $link.check}<svg class="ic lead"><use href="#i-rotate"/></svg>{elseif $link.prompt}<svg class="ic lead"><use href="#i-prompt"/></svg>{/if}{link.text}</a>{/foreach}</span></td>
				<td class="actions"><a class="btn btn-sm" href="{base}{row.href}" data-workflow-view>View</a>{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
			</tr>
{eval $data_wfanswer = $row.answer}
{include components/wfanswer.tpl}
{else}			<tr class="empty"><td colspan="4">Not run yet.</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
{eval $data_pager = $data.pager}
{include components/pager.tpl}
</section>
