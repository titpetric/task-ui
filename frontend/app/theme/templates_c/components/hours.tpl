<?php $_v=&$this->vars; if($_v['data_hours']['text']){ echo htmlspecialchars($_v['data_hours']['line'], ENT_QUOTES);?>
 <?php if($_v['data_hours']['open']){?>
<span class="badge ok">open now</span><?php }else{?>
<span class="badge bad">out of hours now</span><?php }?><?php }else{?>
<span class="faint">always</span><?php }?>
