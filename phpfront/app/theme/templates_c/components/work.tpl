<?php $_v=&$this->vars; if($_v['data_work']['stats'] || $_v['data_work']['trace']['states']){?>
<div class="work-run" data-run-id="<?php echo htmlspecialchars($_v['data_work']['id'], ENT_QUOTES);?>
">
<?php if($_v['data_work']['trace']['states']){?>
	<button type="button" class="work-trace" data-trace data-mode="duration" data-ms="<?php echo htmlspecialchars($_v['data_work']['trace']['ms'], ENT_QUOTES);?>
" data-count="<?php echo htmlspecialchars($_v['data_work']['trace']['count'], ENT_QUOTES);?>
" title="Switch the bar between duration, count and time">
		<span class="work-trace-mode" data-trace-mode>Duration</span>
		<span class="work-trace-bar">
<?php if(!empty($_v['data_work']['trace']['states']))foreach($_v['data_work']['trace']['states'] as $_v['state']){?><span class="work-trace-group <?php echo htmlspecialchars($_v['state']['key'], ENT_QUOTES);?>
" style="width:<?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
%" data-share="<?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
" data-count="<?php echo htmlspecialchars($_v['state']['count'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['state']['label'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['state']['label'], ENT_QUOTES);?>
: <?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
%"></span>
<?php }?><?php if(!empty($_v['data_work']['trace']['segments']))foreach($_v['data_work']['trace']['segments'] as $_v['segment']){?><span class="work-trace-event <?php echo htmlspecialchars($_v['segment']['key'], ENT_QUOTES);?>
" data-ms="<?php echo htmlspecialchars($_v['segment']['ms'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['segment']['label'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['segment']['label'], ENT_QUOTES);?>
: <?php echo htmlspecialchars($_v['segment']['span'], ENT_QUOTES);?>
"></span>
<?php }?>		</span>
	</button>
<?php }?><?php if($_v['data_work']['stats']){?>
	<div class="work-stats"><?php echo htmlspecialchars($_v['data_work']['stats'], ENT_QUOTES);?>
</div>
<?php }?></div><?php }?>
