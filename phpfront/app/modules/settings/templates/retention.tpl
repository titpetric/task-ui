{* The data retention page. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
<form class="form retention" data-api="/api/retention" data-method="PUT" data-next="{data.next}">
<section class="block">
	<h3>What the instance keeps <small>a row per area, the oldest items go first</small></h3>
	<p class="dim">Applied by <a href="{data.runs}" title="The runs of @cleanup">@cleanup</a>, which {if $data.last.ran}last ran <a href="{data.last.href}" title="{data.last.stamp}">{data.last.ago}</a> {eval $data_state = $data.last.state}{include components/state.tpl}{else}<span class="faint">has not run yet</span>{/if}.</p>
	<div class="table-wrap">
	<table class="table retention">
		<thead><tr><th>Area</th><th>What goes</th><th>Keep</th><th>Size limit</th><th>Holds now</th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr>
				<td class="key">{row.area.title}</td>
				<td class="desc">{row.area.what}</td>
{if $row.area.fixed}				<td colspan="2" class="dim">{row.area.fixed}</td>
{else}				<td><select name="{row.area.kind}.age" aria-label="Keep {row.area.title} for"{if !$data.admin} disabled{/if}>{foreach $row.ages as $age}<option value="{age.value}"{if $age.selected} selected{/if}>{age.label}</option>{/foreach}</select></td>
				<td>{if $row.limited}<select name="{row.area.kind}.limit" aria-label="Size limit of {row.area.title}"{if !$data.admin} disabled{/if}>{foreach $row.limits as $limit}<option value="{limit.value}"{if $limit.selected} selected{/if}>{limit.label}</option>{/foreach}</select>{else}<span class="faint">none</span>{/if}</td>
{/if}				<td class="nowrap">{row.holds}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{if $data.admin}	<div class="form-foot">
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
{/if}</section>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
