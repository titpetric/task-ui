{* The form that generates a workspace's SSH key. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
{eval $data_wsgen = $data.form}{include components/ws-key-generate.tpl}
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
