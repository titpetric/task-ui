<?php $_v=&$this->vars;?>
<div class="field">
	<label for="key-name">Name</label>
	<input type="text" id="key-name" name="name" value="<?php echo htmlspecialchars($_v['data_keyname']['value'], ENT_QUOTES);?>
" required pattern="<?php echo htmlspecialchars($_v['data_keyname']['pattern'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off">
	<small class="faint">Lower case letters, digits, dashes and underscores; the algorithm follows, as deploy_ed25519.</small>
</div>