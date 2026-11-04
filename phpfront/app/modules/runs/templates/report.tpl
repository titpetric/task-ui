{* The report of a run, as a document. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

{if !$data.summary}<section class="block">
	<h3>Question</h3>
	<div class="markdown-doc report-question">{data.question|raw}</div>
</section>
{/if}
<section class="block">
	<h3>{if $data.summary}Summary{else}Answer{/if}</h3>
	<article class="markdown-doc report-doc">{data.body|raw}</article>
</section>
