	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
{foreach $data_kv.rows as $row}			<tr><th>{row.label}</th><td{if $row.class} class="{row.class}"{/if}>{if $row.who}{eval $data_who = $row.who}{include components/who.tpl}{/if}{if $row.state}{eval $data_state = $row.state}{include components/state.tpl}{/if}{if $row.link.href}<a href="{row.link.href}">{if $row.link.icon}<svg class="ic"><use href="#i-{row.link.icon}"/></svg> {/if}{row.link.text}</a>{/if}{foreach $row.code as $value}<code>{value}</code> {/foreach}{row.text}{if $row.faint}<span class="faint">{row.faint}</span>{/if}{if $row.after} <span class="dim">{row.after}</span>{/if}{if $row.badge.text} <span class="badge {row.badge.tone}">{row.badge.text}</span>{/if}</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
