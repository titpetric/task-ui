<?php $_v=&$this->vars;?>
			<nav class="menu">
<?php if(!empty($_v['data']))foreach($_v['data'] as $_v['item']){?>
				<a href="<?php echo htmlspecialchars($_v['item']['href'], ENT_QUOTES);?>
" class="<?php if($_v['item']['active']){?>
active<?php }?>"><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['title'], ENT_QUOTES);?>
</span><?php if($_v['item']['pulse']){?>
<i class="dot<?php if($_v['chrome']['running']){?>
 running<?php }?>" data-any-running></i><?php }?></a>
<?php if($_v['item']['flows'] && $_v['chrome']['workflows']){ $_v['data_workflows'] = $_v['chrome']['workflows'];?>
				<div class="flows"><?php if(!empty($_v['data_workflows']))foreach($_v['data_workflows'] as $_v['flow']){ if(!$_v['flow']['hidden']){?>
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
				</div>
<?php } }?>
			</nav>
