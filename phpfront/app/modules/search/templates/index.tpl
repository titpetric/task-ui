{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="block">
	<form class="form search-form" method="post" action="{base}/search/ask" role="search">
		<label for="search-q">What to look for</label>
		<input type="search" id="search-q" name="q" autocomplete="off" spellcheck="false" placeholder="How do I deploy? ci:daily?" autofocus>
		<div class="form-foot">
			<span class="faint">The answer cites the documents and job targets it read.</span>
			<button class="btn btn-primary" type="submit">Search</button>
		</div>
	</form>
</section>
