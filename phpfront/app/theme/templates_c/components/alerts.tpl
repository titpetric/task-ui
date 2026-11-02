<?php $_v=&$this->vars;if($_v['data']['notice']){?>

			<div class="alert ok" role="status"><?php echo htmlspecialchars($_v['data']['notice'], ENT_QUOTES);?>
</div>
<?php } if($_v['data']['error']){?>

			<div class="alert bad" role="alert"><?php echo htmlspecialchars($_v['data']['error'], ENT_QUOTES);?>
</div>
<?php }?>
