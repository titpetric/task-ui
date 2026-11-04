<?php $_v=&$this->vars;?>

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
