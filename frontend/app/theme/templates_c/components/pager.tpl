<?php $_v=&$this->vars; if($_v['data_pager']['prev'] || $_v['data_pager']['next']){?>
	<nav class="pager" aria-label="Pages">
<?php if($_v['data_pager']['prev']){?>
		<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_pager']['prev'], ENT_QUOTES);?>
" rel="prev">Previous</a>
<?php }?><?php if($_v['data_pager']['next']){?>
		<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_pager']['next'], ENT_QUOTES);?>
" rel="next">Next</a>
<?php }?>	</nav>
<?php }?>
