<?php $_v=&$this->vars;?>

<details class="folder<?php if($_v['data_folder']['active']){?>
 active<?php }?>" data-folder="<?php echo htmlspecialchars($_v['data_folder']['id'], ENT_QUOTES);?>
"<?php if($_v['data_folder']['open']){?>
 open<?php }?>>
	<summary title="<?php echo htmlspecialchars($_v['data_folder']['title'], ENT_QUOTES);?>
">
		<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_folder']['href'], ENT_QUOTES);?>
"><?php $_v['data_icon'] = $_v['data_folder'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php if($_v['data_folder']['dir']){?>
<i><?php echo htmlspecialchars($_v['data_folder']['dir'], ENT_QUOTES);?>
/</i><?php }?><b><?php echo htmlspecialchars($_v['data_folder']['name'], ENT_QUOTES);?>
</b></span></a>
		<i class="dot"></i>
<?php if($_v['data_folder']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_folder']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_folder']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_folder']['count'], ENT_QUOTES);?>
</small><?php }?>
		<svg class="ic chev"><use href="#i-chev"/></svg>
	</summary>
	<div class="jobs">
<?php $_v['data_jobs'] = $_v['data_folder'];if(!empty($_v['data_jobs']['jobs']))foreach($_v['data_jobs']['jobs'] as $_v['job']){?>
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
<?php }else{?>

<div class="empty"><?php echo htmlspecialchars($_v['data_jobs']['empty'], ENT_QUOTES);?>
</div>
<?php }?>

	</div>
</details>
