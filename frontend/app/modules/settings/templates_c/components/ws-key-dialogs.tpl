<?php $_v=&$this->vars;?>
<dialog class="confirm" id="key-generate-dialog" aria-labelledby="key-generate-title">
	<?php $_v['data_wsgen'] = $_v['data_wsdialogs']['generate'];?>
<form class="form" data-api="<?php echo htmlspecialchars($_v['data_wsgen']['api'], ENT_QUOTES);?>
" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data_wsgen']['next'], ENT_QUOTES);?>
">
<?php if($_v['data_wsgen']['dialog']){?>
	<h3 id="key-generate-title">Generate key</h3>
	<p>Generates an SSH key of <?php echo htmlspecialchars($_v['data_wsgen']['folder'], ENT_QUOTES);?>
 in <?php echo htmlspecialchars($_v['data_wsgen']['dir'], ENT_QUOTES);?>
, named <span class="mono">workspace_<?php echo htmlspecialchars($_v['data_wsgen']['folder'], ENT_QUOTES);?>
_&lt;name&gt;_&lt;algorithm&gt;</span>.</p>
<?php }?>	<div class="field">
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
		<input type="text" id="wkg-comment" name="comment" value="<?php echo htmlspecialchars($_v['data_wsgen']['comment'], ENT_QUOTES);?>
">
	</div>
<?php $_v['data_wsfoot'] = $_v['data_wsgen']['foot'];?>
	<div class="form-foot">
<?php if($_v['data_wsfoot']['dialog']){?>
		<button type="button" class="btn" data-dialog-close>Cancel</button>
<?php }else{?>
		<a class="btn" href="<?php echo htmlspecialchars($_v['data_wsfoot']['cancel'], ENT_QUOTES);?>
">Cancel</a>
<?php }?>		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_wsfoot']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>
</dialog>
<dialog class="confirm" id="key-upload-dialog" aria-labelledby="key-upload-title">
	<?php $_v['data_wsup'] = $_v['data_wsdialogs']['upload'];?>
<form class="form" data-api="<?php echo htmlspecialchars($_v['data_wsup']['api'], ENT_QUOTES);?>
" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data_wsup']['next'], ENT_QUOTES);?>
">
<?php if($_v['data_wsup']['dialog']){?>
	<h3 id="key-upload-title">Upload key</h3>
	<p>Saves a private key you already use as an SSH key of <?php echo htmlspecialchars($_v['data_wsup']['folder'], ENT_QUOTES);?>
. The folder refuses a key it holds, under any name.</p>
<?php }?>	<div class="field">
		<label for="wku-file">Private key file</label>
		<input type="file" id="wku-file" data-read-into="wku-content">
	</div>
	<div class="field">
		<label for="wku-content">Or paste the private key</label>
		<textarea id="wku-content" name="content" rows="<?php echo htmlspecialchars($_v['data_wsup']['rows'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	<div class="field">
		<label for="wku-name">Name</label>
		<input type="text" id="wku-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wku-comment">Comment</label>
		<input type="text" id="wku-comment" name="comment" value="<?php echo htmlspecialchars($_v['data_wsup']['comment'], ENT_QUOTES);?>
">
	</div>
<?php $_v['data_wsfoot'] = $_v['data_wsup']['foot'];?>
	<div class="form-foot">
<?php if($_v['data_wsfoot']['dialog']){?>
		<button type="button" class="btn" data-dialog-close>Cancel</button>
<?php }else{?>
		<a class="btn" href="<?php echo htmlspecialchars($_v['data_wsfoot']['cancel'], ENT_QUOTES);?>
">Cancel</a>
<?php }?>		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_wsfoot']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>
</dialog>
