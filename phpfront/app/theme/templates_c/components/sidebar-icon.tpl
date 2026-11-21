<?php $_v=&$this->vars;?>
<i class="ic ic-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES); if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"<?php if($_v['data_icon']['dirty']){?>
 title="<?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
"<?php }?>></i>