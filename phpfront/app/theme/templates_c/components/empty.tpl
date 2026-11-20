<?php $_v=&$this->vars; if($_v['data_empty']['tag'] == 'li'){?>
<li class="empty-state" role="status"><?php echo htmlspecialchars($_v['data_empty']['text'], ENT_QUOTES);?>
</li><?php }else{?>
<p class="empty-state" role="status"><?php echo htmlspecialchars($_v['data_empty']['text'], ENT_QUOTES); if($_v['data_empty']['href']){?>
 <a href="<?php echo htmlspecialchars($_v['data_empty']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_empty']['label'], ENT_QUOTES);?>
</a><?php }?></p><?php }?>
