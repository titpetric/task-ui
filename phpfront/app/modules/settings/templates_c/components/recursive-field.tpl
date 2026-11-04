<?php $_v=&$this->vars;?>

	<div class="field">
		<label class="check"><input type="checkbox" id="<?php echo htmlspecialchars($_v['data_recursive']['id'], ENT_QUOTES);?>
" name="recursive"<?php if($_v['data_recursive']['checked']){?>
 checked<?php }?>> Recursive</label>
		<small class="faint">Lists every folder inside; off lists only the folder and its top-level folders.</small>
	</div>
