<?php $_v=&$this->vars;?>
<div class="spacebar" role="img" aria-label="<?php echo htmlspecialchars($_v['data_space']['total'], ENT_QUOTES);?>
 in all: <?php echo htmlspecialchars($_v['data_space']['files'], ENT_QUOTES);?>
 of files, <?php echo htmlspecialchars($_v['data_space']['history'], ENT_QUOTES);?>
 of history">
	<div class="spacebar-track"><i class="files" style="width: <?php echo htmlspecialchars($_v['data_space']['filesPercent'], ENT_QUOTES);?>
%"></i><i class="history" style="width: <?php echo htmlspecialchars($_v['data_space']['historyPercent'], ENT_QUOTES);?>
%"></i></div>
	<div class="spacebar-legend">
		<span><i class="swatch files"></i>Files <b><?php echo htmlspecialchars($_v['data_space']['files'], ENT_QUOTES);?>
</b></span>
		<span><i class="swatch history"></i>History <b><?php echo htmlspecialchars($_v['data_space']['history'], ENT_QUOTES);?>
</b></span>
		<span class="faint"><?php echo htmlspecialchars($_v['data_space']['total'], ENT_QUOTES);?>
 in all<?php if($_v['data_space']['ignored']){?>
, and <?php echo htmlspecialchars($_v['data_space']['ignored'], ENT_QUOTES);?>
 git ignores<?php }?></span>
	</div>
</div>
