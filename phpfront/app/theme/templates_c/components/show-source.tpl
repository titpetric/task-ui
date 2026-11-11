<?php $_v=&$this->vars;?>
<button type="button" class="btn btn-sm show-source" data-source="<?php echo htmlspecialchars($_v['data_source']['folder'], ENT_QUOTES);?>
" data-file="<?php echo htmlspecialchars($_v['data_source']['file'], ENT_QUOTES);?>
" aria-haspopup="dialog"><svg class="ic"><use href="#i-file"/></svg><span><?php echo htmlspecialchars($_v['data_source']['label'], ENT_QUOTES);?>
</span></button>
