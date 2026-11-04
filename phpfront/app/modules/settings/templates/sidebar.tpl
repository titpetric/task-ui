{* The sidebar settings page. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
<form class="form sidebar-form" {if $data.admin}data-api="/api/sidebar" data-method="PUT" data-next="reload" data-notice="Sidebar saved" data-sidebar{/if} novalidate>
<section class="block">
	<h3>The items <small>a row per link, the order the sidebar reads</small></h3>
	<div class="table-wrap">
	<table class="table sidebar-rows"{if $data.admin} data-sort-rows{/if}>
		<thead><tr><th class="num">Order</th><th>Item</th><th>Link</th><th class="num">Shown</th><th></th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr data-row data-href="{row.href}">
				<td class="num">
{if $data.admin}					<button type="button" class="grip" data-row-grip title="Drag to move the item, or focus it and press the up and down arrows" aria-label="Move {row.title}"><svg class="ic"><use href="#i-grip"/></svg></button>
{/if}					<span class="mono" data-row-num>{row.number}</span>
				</td>
				<td class="key">
					<svg class="ic"><use href="#i-{row.icon}"/></svg>
{if $row.note}					{row.title}<small class="faint">{row.note}</small>
				</td>
{else}					{row.title}
				</td>
{/if}				<td class="mono"><a href="{row.link}">{row.href}</a></td>
				<td class="num">
					<label class="switch" title="{if $row.hidden}Left out of the sidebar{else}Listed in the sidebar{/if}">
						<input type="checkbox" data-row-shown="{if $row.hidden}false{else}true{/if}"{if !$row.hidden} checked{/if}{if !$data.admin} disabled{/if} aria-label="Show {row.title} in the sidebar">
						<i aria-hidden="true"></i>
					</label>
				</td>
				<td class="row-actions">
{if $data.admin}					{eval $data_menu = $row.menu}{include components/menu.tpl}
{/if}				</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{if $data.admin}	<div class="form-foot">
		<span class="faint" data-sidebar-changed>No changes yet.</span>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
	<p class="wf-live" aria-live="polite" data-rows-live></p>
{/if}</section>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
