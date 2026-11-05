<form class="form" data-api="{data_wsup.api}" data-method="POST" data-next="{data_wsup.next}">
{if $data_wsup.dialog}	<h3 id="key-upload-title">Upload key</h3>
	<p>Saves a private key you already use as an SSH key of {data_wsup.folder}. The folder refuses a key it holds, under any name.</p>
{/if}	<div class="field">
		<label for="wku-file">Private key file</label>
		<input type="file" id="wku-file" data-read-into="wku-content">
	</div>
	<div class="field">
		<label for="wku-content">Or paste the private key</label>
		<textarea id="wku-content" name="content" rows="{data_wsup.rows}" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	<div class="field">
		<label for="wku-name">Name</label>
		<input type="text" id="wku-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wku-comment">Comment</label>
		<input type="text" id="wku-comment" name="comment" value="{data_wsup.comment}">
	</div>
{eval $data_wsfoot = $data_wsup.foot}{include components/ws-key-foot.tpl}</form>
