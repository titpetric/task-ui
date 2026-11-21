<?php $_v=&$this->vars;?>
<button type="button" class="btn btn-sm show-source" data-source="<?php echo htmlspecialchars($_v['data_source']['folder'], ENT_QUOTES);?>
" data-file="<?php echo htmlspecialchars($_v['data_source']['file'], ENT_QUOTES);?>
" aria-haspopup="dialog"><i class="ic ic-file" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['data_source']['label'], ENT_QUOTES);?>
</span></button>
