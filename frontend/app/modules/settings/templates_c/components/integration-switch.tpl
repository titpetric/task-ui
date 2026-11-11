<?php $_v=&$this->vars;?>
<label class="switch env-switch" title="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
 in a run">
	<input type="checkbox" name="<?php echo htmlspecialchars($_v['data_switch']['name'], ENT_QUOTES);?>
" value="1" data-type="toggle"<?php if($_v['data_switch']['on']){?>
 checked<?php }?> aria-label="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
">
	<i aria-hidden="true"></i>
</label>
