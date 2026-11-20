<?php $_v=&$this->vars; if($_v['data_status']['working']){?>
<div class="conversation-status" data-status data-state="<?php echo htmlspecialchars($_v['data_status']['state'], ENT_QUOTES);?>
" data-started="<?php echo htmlspecialchars($_v['data_status']['started'], ENT_QUOTES);?>
">
	<p aria-live="polite"><i class="dot working"></i> <span data-status-text><?php echo htmlspecialchars($_v['data_status']['label'], ENT_QUOTES);?>
</span> <span class="faint" data-status-elapsed></span></p>
<?php if($_v['data_status']['can_end']){?>
	<div class="actions">
		<button class="btn btn-sm" type="button" data-stop="<?php echo htmlspecialchars($_v['data_status']['run'], ENT_QUOTES);?>
" title="Stop this run and let it settle">Stop</button>
		<button class="btn btn-sm btn-danger" type="button" data-kill="<?php echo htmlspecialchars($_v['data_status']['run'], ENT_QUOTES);?>
" title="Kill this run and everything it started">Kill</button>
	</div>
<?php }?></div>
<?php }?>
