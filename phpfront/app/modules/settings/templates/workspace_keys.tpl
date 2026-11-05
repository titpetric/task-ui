{eval $data_head = $data.head}
{include components/page-head.tpl}
{if $data.dir}<section class="block workspace-keys" id="keys">
{eval $data_block = $data.block}{include components/block-head.tpl}{eval $data_keys = $data.keys}{include components/key-table.tpl}</section>
{if $data.dialogs}{eval $data_wsdialogs = $data.dialogs}{include components/ws-key-dialogs.tpl}{/if}
{else}<div class="alert">The instance has no key folder, so it keeps no SSH keys.</div>
{/if}{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
