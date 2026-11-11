<?php $_v=&$this->vars;?>
<section class="block" data-live>
	<h3>Terminal <small>as it looks now, <?php if($_v['data_terminal']['watching']){ echo htmlspecialchars($_v['data_terminal']['watching'], ENT_QUOTES);?>
 watching<?php }else{?>
nobody attached<?php }?></small></h3>
	<a class="tile wide" href="#" data-attach="<?php echo htmlspecialchars($_v['data_terminal']['id'], ENT_QUOTES);?>
"><div class="frame" data-frame="<?php echo htmlspecialchars($_v['data_terminal']['id'], ENT_QUOTES);?>
" role="img" aria-label="Terminal of <?php echo htmlspecialchars($_v['data_terminal']['name'], ENT_QUOTES);?>
"><?php echo $_v['data_terminal']['frame'];?>
</div></a>
</section>
