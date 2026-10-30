<?php $_v=&$this->vars; for($_v['i']=99; $_v['i']>=1; $_v['i']--){echo htmlspecialchars($_v['i'], ENT_QUOTES);?>
 bottles of beer on the wall, <?php echo htmlspecialchars($_v['i'], ENT_QUOTES);?>
 bottles of beer. Take one down and pass it around, <?php echo $_v['i']-1;?>
 bottles of beer on the wall.
<?php }if($_v['i'] > 10){?>

	We have some beer left.
<?php }elseif($_v['i'] > 3){?>

	Our beer is going to run out.
<?php }else{?>

	Critical! Only <?php echo htmlspecialchars($_v['i'], ENT_QUOTES);?>
 beer left.
<?php }?>