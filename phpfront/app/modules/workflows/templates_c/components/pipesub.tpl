<?php $_v=&$this->vars;?>
<li class="<?php echo htmlspecialchars($_v['data_sub']['className'], ENT_QUOTES);?>
" data-step="<?php echo htmlspecialchars($_v['data_sub']['pos'], ENT_QUOTES);?>
">
	<span class="rail"><i class="dot<?php if($_v['data_sub']['state']){?>
 <?php echo htmlspecialchars($_v['data_sub']['state'], ENT_QUOTES); }?>"></i></span>
	<div class="what">
		<div class="step-head">
			<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_sub']['href'], ENT_QUOTES);?>
"<?php if($_v['data_sub']['runId']){?>
 data-run-link="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['data_sub']['workflow']){?>
<i class="ic ic-flow lead" aria-hidden="true"></i><?php }elseif($_v['data_sub']['summary']){?>
<i class="ic ic-summary lead" aria-hidden="true"></i><?php }elseif($_v['data_sub']['check']){?>
<i class="ic ic-rotate lead" aria-hidden="true"></i><?php }elseif($_v['data_sub']['prompt']){?>
<i class="ic ic-prompt lead" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['data_sub']['text'], ENT_QUOTES);?>
</a>
<?php if($_v['data_sub']['expand']){?>
			<button type="button" class="btn btn-sm" data-expand="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
" aria-expanded="false"><i class="ic ic-expand" aria-hidden="true"></i><span>Expand<?php if($_v['data_sub']['lines']){?>
 <?php echo htmlspecialchars($_v['data_sub']['lines'], ENT_QUOTES);?>
 lines<?php }?></span></button>
<?php }?>		</div>
		<small><?php echo htmlspecialchars($_v['data_sub']['small'], ENT_QUOTES);?>
</small>
<?php if($_v['data_sub']['prompt']){?>
		<p class="step-prompt"><?php echo htmlspecialchars($_v['data_sub']['prompt'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_sub']['runId']){?>
		<a class="frame" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php if($_v['data_sub']['live']){?>
 data-frame="<?php echo htmlspecialchars($_v['data_sub']['runId'], ENT_QUOTES);?>
"<?php }?> role="img" aria-label="Terminal of <?php echo htmlspecialchars($_v['data_sub']['text'], ENT_QUOTES);?>
"><?php echo $_v['data_sub']['frame'];?>
</a>
<?php }?><?php if($_v['data_sub']['subId']){?>
		<ol class="pipeline sub" data-sub="<?php echo htmlspecialchars($_v['data_sub']['subId'], ENT_QUOTES);?>
" data-workflow="<?php echo htmlspecialchars($_v['data_sub']['workflowId'], ENT_QUOTES);?>
">
		</ol>
<?php }?>	</div>
</li>
