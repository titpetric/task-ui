<?php $_v=&$this->vars;?>

			<nav class="menu">
<?php if(!empty($_v['data']))foreach($_v['data'] as $_v['item']){?>
				<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['item']['href'], ENT_QUOTES);?>
" class="<?php if($_v['item']['active']){?>
active<?php }?>"><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['title'], ENT_QUOTES);?>
</span><?php if($_v['item']['pulse']){?>
<i class="dot" data-any-running></i><?php }?></a>
<?php }?>
			</nav>
