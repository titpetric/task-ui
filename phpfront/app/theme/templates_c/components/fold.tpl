<?php $_v=&$this->vars;?>

<button type="button" class="fold-toggle" aria-expanded="true" aria-controls="<?php echo htmlspecialchars($_v['data_fold']['id'], ENT_QUOTES);?>
"><svg class="ic chev"><use href="#i-chev"/></svg><?php echo htmlspecialchars($_v['data_fold']['label'], ENT_QUOTES);?>
</button>
