<?php $_v=&$this->vars;?>
				<div class="project<?php if($_v['data_project']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
">
					<a class="section<?php if($_v['data_project']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_project']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_project']['id'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['data_project']['title'], ENT_QUOTES);?>
">
<?php $_v['data_icon'] = $_v['data_project'];?>
<svg class="ic<?php if($_v['data_icon']['dirty']){?>
 dirty<?php }?>"><?php if($_v['data_icon']['dirty']){?>
<title><?php echo htmlspecialchars($_v['data_icon']['changed'], ENT_QUOTES);?>
</title><?php }?><use href="#i-<?php echo htmlspecialchars($_v['data_icon']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['data_project']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_project']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_project']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_project']['badge'], ENT_QUOTES);?>
</small><?php }elseif($_v['data_project']['counted']){?>
<small><?php echo htmlspecialchars($_v['data_project']['count'], ENT_QUOTES);?>
</small><?php }?>
					</a>
				</div>
