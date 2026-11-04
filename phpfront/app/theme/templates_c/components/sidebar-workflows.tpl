<?php $_v=&$this->vars;if(!empty($_v['data_workflows']))foreach($_v['data_workflows'] as $_v['flow']){ if(!$_v['flow']['hidden']){?>
					<div class="flow<?php if($_v['flow']['active']){?>
 active<?php }?>">
						<i class="shebang<?php if($_v['flow']['running']){?>
 running<?php }?>" aria-hidden="true">#!</i>
						<a class="name" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/workflows/<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['flow']['title'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
</a>
<?php if($_v['flow']['running']){?>
						<a class="play" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/workflows/<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
, running" aria-label="Watch <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-eye"/></svg></a>
<?php }elseif($_v['flow']['canRun']){?>
						<button type="button" class="play" data-workflow-run="<?php echo htmlspecialchars($_v['flow']['id'], ENT_QUOTES);?>
"<?php if($_v['flow']['prompted']){?>
 data-workflow-prompt="<?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"<?php }?> title="Run <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
" aria-label="Run <?php echo htmlspecialchars($_v['flow']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-play"/></svg></button>
<?php }?>
					</div>
<?php } }?>
