<?php $_v=&$this->vars;?>

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
