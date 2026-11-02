<?php $_v=&$this->vars;?>

			<nav class="folders">
<?php if(!empty($_v['data']))foreach($_v['data'] as $_v['group']){?>
				<div class="group<?php if($_v['group']['workspace']){?>
 workspace<?php }?>">
					<a class="name" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['group']['href'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['group']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['group']['title'], ENT_QUOTES);?>
</span><?php if($_v['group']['dirty']){?>
<i class="dirty" title="<?php echo htmlspecialchars($_v['group']['dirty'], ENT_QUOTES);?>
 changed files"><?php echo htmlspecialchars($_v['group']['dirty'], ENT_QUOTES);?>
</i><?php }?></a>
<?php if(!empty($_v['group']['jobs']))foreach($_v['group']['jobs'] as $_v['job']){?>
					<a class="job" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['job']['href'], ENT_QUOTES);?>
"><i class="shebang" aria-hidden="true">#!</i><span><?php echo htmlspecialchars($_v['job']['name'], ENT_QUOTES);?>
</span></a>
<?php }?>
				</div>
<?php }else{ if($_v['chrome']['can']['project/edit']){?>
				<div class="empty">No folders yet. <a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/projects/new">Add a git repository</a>, or put a folder with jobs under the root.</div>
<?php }else{?>
				<div class="empty">No folders yet. Add a git repository in Settings, or put a folder with jobs under the root.</div>
<?php }?><?php }?>
			</nav>
