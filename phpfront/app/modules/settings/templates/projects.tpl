{* The projects settings page. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
{foreach $data.groups as $group}<section class="block">
{eval $data_block = $group.block}{include components/block-head.tpl}{if $group.table.rows}{eval $data_projects = $group.table}{include components/project-table.tpl}
{/if}</section>
{/foreach}{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
