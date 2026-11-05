{eval $data_head = $data.head}
{include components/page-head.tpl}
<section class="block integration-cards">
	<div class="wf-steps-head">
		<h3 class="section-title">The services</h3>
		<button type="button" class="btn btn-sm" data-cards-fold aria-pressed="false">Expand all</button>
	</div>
	<ol class="wf-rail">
{foreach $data.cards as $card}{eval $data_card = $card}{include components/integration-card.tpl}
{/foreach}	</ol>
</section>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
