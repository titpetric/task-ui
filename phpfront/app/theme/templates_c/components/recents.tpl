<?php $_v=&$this->vars;?>

		<aside class="rail" aria-label="Recents">
			<div class="rail-inner">
				<div class="label side-label"><a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data']['href'], ENT_QUOTES);?>
">Recents</a></div>
				<nav class="recents" id="recents">
<?php if(!empty($_v['data']['groups']))foreach($_v['data']['groups'] as $_v['group']){?>
					<div class="recent-group"><?php echo htmlspecialchars($_v['group']['label'], ENT_QUOTES);?>
</div>
<?php if(!empty($_v['group']['entries']))foreach($_v['group']['entries'] as $_v['entry']){?>
					<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['entry']['href'], ENT_QUOTES);?>
" class="recent<?php if($_v['entry']['flow']){?>
 flow<?php }?>" title="<?php echo htmlspecialchars($_v['entry']['title'], ENT_QUOTES);?>
">
						<i class="dot <?php echo htmlspecialchars($_v['entry']['state'], ENT_QUOTES);?>
"></i>
						<span class="what"><b><?php echo htmlspecialchars($_v['entry']['name'], ENT_QUOTES);?>
</b><small><?php if($_v['entry']['flow']){?>
<svg class="ic lead"><use href="#i-flow"/></svg><?php }?><?php echo htmlspecialchars($_v['entry']['place'], ENT_QUOTES);?>
</small></span>
<?php if(!$_v['entry']['ended']){?>
						<svg class="ic act"><use href="#i-<?php echo htmlspecialchars($_v['entry']['act'], ENT_QUOTES);?>
"/></svg><?php }?>
						<span class="avatar sm role-<?php echo htmlspecialchars($_v['entry']['role'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['entry']['user'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['entry']['initials'], ENT_QUOTES);?>
</span>
					</a>
<?php } } if(!$_v['data']['groups']){?>

					<div class="empty">Runs started by anyone show up here.</div>
<?php }?>
					<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data']['href'], ENT_QUOTES);?>
" class="recent more" data-recents-more hidden><svg class="ic"><use href="#i-list"/></svg><span class="what"><b>All runs</b><small></small></span></a>
				</nav>
			</div>
		</aside>
