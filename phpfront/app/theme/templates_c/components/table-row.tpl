<?php $_v=&$this->vars;?>
<tr<?php if($_v['data_row']['id']){?>
 id="<?php echo htmlspecialchars($_v['data_row']['id'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['run']){?>
 data-run-id="<?php echo htmlspecialchars($_v['data_row']['run'], ENT_QUOTES);?>
"<?php }?><?php if($_v['data_row']['class']){?>
 class="<?php echo htmlspecialchars($_v['data_row']['class'], ENT_QUOTES);?>
"<?php }?>><?php if(!empty($_v['data_row']['cells']))foreach($_v['data_row']['cells'] as $_v['cell']){?><td<?php if($_v['cell']['class']){?>
 class="<?php echo htmlspecialchars($_v['cell']['class'], ENT_QUOTES);?>
"<?php }?><?php if($_v['cell']['span']){?>
 colspan="<?php echo htmlspecialchars($_v['cell']['span'], ENT_QUOTES);?>
"<?php }?>><?php echo $_v['cell']['html'];?>
</td><?php }?></tr>
