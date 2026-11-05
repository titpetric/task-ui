<?php $_v=&$this->vars;?>
<span class="meter<?php if($_v['data_meter']['share'] >= 90){?>
 high<?php }elseif($_v['data_meter']['share'] >= 75){?>
 warm<?php }?>"><span class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_meter']['width'], ENT_QUOTES);?>
%"></i></span><span class="nowrap"><?php echo htmlspecialchars($_v['data_meter']['pct'], ENT_QUOTES);?>
%</span></span>
