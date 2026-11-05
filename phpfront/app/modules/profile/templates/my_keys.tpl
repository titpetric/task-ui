{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="block">
{eval $data_block = $data.block}{include components/block-head.tpl}
{eval $data_keys = $data.keys}{include components/key-table.tpl}
</section>
