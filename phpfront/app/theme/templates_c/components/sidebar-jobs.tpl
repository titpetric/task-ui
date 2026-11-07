<?php $_v=&$this->vars; if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
<div class="job<?php if($_v['job']['pinned']){?>
 pinned<?php }?><?php if($_v['job']['active']){?>
 active<?php }?>" data-job-key="<?php echo htmlspecialchars($_v['job']['key'], ENT_QUOTES);?>
">
	<i class="shebang<?php if($_v['job']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
	<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"<?php if($_v['job']['hint']){?>
 title="<?php echo htmlspecialchars($_v['job']['hint'], ENT_QUOTES);?>
"<?php }?>><?php if($_v['job']['kind']){?>
<i><?php echo htmlspecialchars($_v['job']['kind'], ENT_QUOTES);?>
:</i><?php }?><?php echo htmlspecialchars($_v['job']['label'], ENT_QUOTES);?>
</a>
<?php if($_v['job']['running']){?>

	<button type="button" class="play" data-attach="<?php echo htmlspecialchars($_v['job']['attach'], ENT_QUOTES);?>
" title="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
, running" aria-label="Attach to <?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></button>
<?php }elseif($_v['job']['can']){?>

	<button type="button" class="play" data-run data-folder="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
"<?php if($_v['job']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['job']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['job']['job'], ENT_QUOTES);?>
"<?php if($_v['job']['args']){?>
 data-run-args="<?php echo htmlspecialchars($_v['job']['args'], ENT_QUOTES);?>
"<?php }?> data-title="<?php echo htmlspecialchars($_v['job']['folder'], ENT_QUOTES);?>
 / <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" title="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['job']['said'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
</div>
<?php }?>
