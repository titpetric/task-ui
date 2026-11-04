{* input scope: data_user_keys. See user-keys.md. *}
<section class="block" id="keys">
{eval $data_block = $data_user_keys.block}{include components/block-head.tpl}
{eval $data_keys = $data_user_keys.keys}{include components/key-table.tpl}
</section>
