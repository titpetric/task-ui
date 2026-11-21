<?php $_v=&$this->vars; if($_v['data_work']['stats']){?>
<div class="work-run" data-run-id="<?php echo htmlspecialchars($_v['data_work']['id'], ENT_QUOTES);?>
">
	<div class="work-stats"><?php echo htmlspecialchars($_v['data_work']['stats'], ENT_QUOTES);?>
</div>
</div><?php }?>