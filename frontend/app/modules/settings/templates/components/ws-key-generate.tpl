<form class="form" data-api="{data_wsgen.api}" data-method="POST" data-next="{data_wsgen.next}">
{if $data_wsgen.dialog}	<h3 id="key-generate-title">Generate key</h3>
	<p>Generates an SSH key of {data_wsgen.folder} in {data_wsgen.dir}, named <span class="mono">workspace_{data_wsgen.folder}_&lt;name&gt;_&lt;algorithm&gt;</span>.</p>
{/if}	<div class="field">
		<label for="wkg-name">Name</label>
		<input type="text" id="wkg-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wkg-algo">Algorithm</label>
		<select id="wkg-algo" name="algo" data-key-algo>
			<option value="ed25519" selected>ed25519</option>
			<option value="ecdsa" data-bits="256">ECDSA P-256</option>
			<option value="ecdsa" data-bits="384">ECDSA P-384</option>
			<option value="ecdsa" data-bits="521">ECDSA P-521</option>
			<option value="rsa" data-bits="2048">RSA 2048</option>
			<option value="rsa" data-bits="3072">RSA 3072</option>
			<option value="rsa" data-bits="4096">RSA 4096</option>
		</select>
		<input type="hidden" name="bits" value="0" data-type="int">
	</div>
	<div class="field">
		<label for="wkg-comment">Comment</label>
		<input type="text" id="wkg-comment" name="comment" value="{data_wsgen.comment}">
	</div>
{eval $data_wsfoot = $data_wsgen.foot}{include components/ws-key-foot.tpl}</form>
