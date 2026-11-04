<?php $_v=&$this->vars; if($_v['data_wffold']['answer']){?>
<button type="button" class="btn-icon wf-fold" data-answer-toggle aria-expanded="<?php if($_v['data_wffold']['open']){?>
true<?php }else{?>
false<?php }?>" aria-controls="wf-answer-<?php echo htmlspecialchars($_v['data_wffold']['id'], ENT_QUOTES);?>
" title="<?php if($_v['data_wffold']['open']){?>
Hide<?php }else{?>
Show<?php }?> the answer" aria-label="<?php if($_v['data_wffold']['open']){?>
Hide<?php }else{?>
Show<?php }?> the answer"><svg class="ic chev"><use href="#i-chev"/></svg></button><?php }else{?>
<span class="wf-fold" aria-hidden="true"></span><?php }?>
