{* The Add workflow page. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

{eval $data_editor = $data.editor}
{include components/stepeditor.tpl}
