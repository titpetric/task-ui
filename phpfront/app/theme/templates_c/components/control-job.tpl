<?php $_v=&$this->vars;?>

<span class="cc-slot" data-cc="<?php echo htmlspecialchars($_v['data_cc']['key'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_cc']['folder'], ENT_QUOTES);?>
" data-job="<?php echo htmlspecialchars($_v['data_cc']['job'], ENT_QUOTES);?>
" data-name="<?php echo htmlspecialchars($_v['data_cc']['name'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
" data-last="<?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['can']){?>
 data-can<?php }?><?php if($_v['data_cc']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_cc']['dir'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['desc']){?>
 data-desc="<?php echo htmlspecialchars($_v['data_cc']['desc'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_cc']['args'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['confirm']){?>
 data-confirm="<?php echo htmlspecialchars($_v['data_cc']['confirm'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['path']){?>
 data-path="<?php echo htmlspecialchars($_v['data_cc']['path'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['terminal']){?>
 data-interactive<?php }?>><?php if($_v['data_cc']['run']){?>
<a class="cc named running" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_cc']['run'], ENT_QUOTES);?>
" data-attach="<?php echo htmlspecialchars($_v['data_cc']['run'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></a><?php }elseif($_v['data_cc']['can']){?>
<button type="button" class="cc named <?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
" data-run data-folder="<?php echo htmlspecialchars($_v['data_cc']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_cc']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_cc']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cc']['what'], ENT_QUOTES);?>
"<?php if($_v['data_cc']['confirm']){?>
 data-confirm="<?php echo htmlspecialchars($_v['data_cc']['confirm'], ENT_QUOTES);?>
" data-confirm-ok="Run <?php echo htmlspecialchars($_v['data_cc']['name'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_cc']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_cc']['args'], ENT_QUOTES);?>
"<?php }?> title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-<?php if($_v['data_cc']['terminal']){?>
terminal<?php }else{?>
play<?php }?>"/></svg><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></button><?php }else{?>
<a class="cc named idle <?php echo htmlspecialchars($_v['data_cc']['last'], ENT_QUOTES);?>
" href="<?php echo htmlspecialchars($_v['data_cc']['href'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cc']['title'], ENT_QUOTES);?>
"><i class="shebang" aria-hidden="true">#!</i><span><?php echo htmlspecialchars($_v['data_cc']['label'], ENT_QUOTES);?>
</span></a><?php }?></span>
