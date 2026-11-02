<?php $_v=&$this->vars;if($_v['data']['crumbs']){?>

			<div class="crumbs"><a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES);?>
/">Dashboard</a><?php if(!empty($_v['data']['crumbs']))foreach($_v['data']['crumbs'] as $_v['crumb']){?><span><?php if($_v['crumb']['href']){?>
<a href="<?php echo htmlspecialchars($_v['data']['base'], ENT_QUOTES); echo htmlspecialchars($_v['crumb']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['crumb']['text'], ENT_QUOTES);?>
</a><?php }else{ echo htmlspecialchars($_v['crumb']['text'], ENT_QUOTES); }?></span><?php }?></div>
<?php }?>
