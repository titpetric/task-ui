<?php $_v=&$this->vars;?>
<span class="tri-cell"><button type="button" class="tri" data-tri="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_cell']['title'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_cell']['label'], ENT_QUOTES);?>
"<?php if($_v['data_cell']['locked']){?>
 disabled<?php }?>><?php echo htmlspecialchars($_v['data_cell']['mark'], ENT_QUOTES);?>
</button><?php if($_v['data_cell']['field']){?>
<input type="hidden" name="<?php echo htmlspecialchars($_v['data_cell']['path'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" data-was="<?php echo htmlspecialchars($_v['data_cell']['effect'], ENT_QUOTES);?>
" disabled><?php }?><?php if($_v['data_cell']['shown']){?>
<span class="eff <?php echo htmlspecialchars($_v['data_cell']['tone'], ENT_QUOTES);?>
" title="Result: <?php echo htmlspecialchars($_v['data_cell']['says'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_cell']['result'], ENT_QUOTES);?>
</span><?php }?></span>
