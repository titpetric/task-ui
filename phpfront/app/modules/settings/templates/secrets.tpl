{* The secrets settings page. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
{if $data.workspace}
<p class="dim secret-filter">Showing what reaches the workspace <b>{data.workspace}</b>. <a href="{data.every}">Every workspace</a></p>
{/if}
<section class="block">
{eval $data_block = $data.block}{include components/block-head.tpl}{if $data.error}	<div class="alert bad">{data.error}</div>
{/if}{if $data.rows}	<div class="table-wrap">
	<table class="table secrets">
		<thead><tr><th>Name</th><th>Value</th><th>Applies to</th><th>Reads</th><th class="actions" aria-label="Actions"></th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr>
				<td class="key">{row.name}{if $row.secret} <span class="badge" title="The value is encrypted and never shown">secret</span>{else} <span class="badge" title="The value is shown on this page">env</span>{/if}</td>
				<td>{if $row.secret}<span class="faint" title="The value is never shown">{row.mask}</span>{elseif $row.value}<code>{row.value}</code>{else}<span class="faint">empty</span>{/if}</td>
				<td class="scope faint">{row.applies}</td>
				<td class="nowrap">{if $row.reads}<a href="{row.readsHref}" title="The runs that got this variable">{row.reads}</a>{else}<span class="faint">0</span>{/if} {if $row.read.run}<a class="faint" href="{row.read.href}" title="{row.read.title}">{row.read.ago}</a>{else}<span class="faint" title="No run has got this variable">never</span>{/if}</td>
				<td class="actions">{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}</section>
{if $data.env}
<section class="block" id="integration-env">
{eval $data_block = $data.envBlock}{include components/block-head.tpl}	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Name</th><th>Value</th></tr></thead>
		<tbody>
{foreach $data.env as $var}			<tr><td class="key">{var.name}</td><td>{if $var.redacted}<span class="faint mono">REDACTED</span>{elseif $var.value}<code>{var.value}</code>{else}<span class="faint">empty</span>{/if}</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
</section>
{/if}{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
