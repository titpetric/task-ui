<?php $_v=&$this->vars;?>
<li><a href="<?php echo htmlspecialchars($_v['data_finding']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_finding']['folder'], ENT_QUOTES);?>
</a> <code><?php echo htmlspecialchars($_v['data_finding']['path'], ENT_QUOTES);?>
</code><?php if($_v['data_finding']['job']){?>
 <span class="dim">job</span> <b><?php echo htmlspecialchars($_v['data_finding']['job'], ENT_QUOTES);?>
</b><?php }?>: <span<?php if($_v['data_finding']['issue']){?>
 title="<?php echo htmlspecialchars($_v['data_finding']['issue'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_finding']['detail'], ENT_QUOTES);?>
</span></li>
