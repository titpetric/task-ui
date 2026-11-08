<?php $_v=&$this->vars;?>
			<nav class="folders">
<?php if(!empty($_v['data']))foreach($_v['data'] as $_v['group']){ if($_v['group']['workspace']){ $_v['data_workspace'] = $_v['group']; if($_v['data_workspace']['top']){?>

<details class="workspace<?php if($_v['data_workspace']['active']){?>
 active<?php }?>" data-project="<?php echo htmlspecialchars($_v['data_workspace']['id'], ENT_QUOTES);?>
" open>
	<summary<?php if($_v['data_workspace']['here']){?>
 class="active"<?php }?> title="<?php echo htmlspecialchars($_v['data_workspace']['title'], ENT_QUOTES);?>
">
		<svg class="ic chev"><use href="#i-chev"/></svg>
		<a class="section<?php if($_v['data_workspace']['here']){?>
 active<?php }?>" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES); echo htmlspecialchars($_v['data_workspace']['href'], ENT_QUOTES);?>
" data-folder="<?php echo htmlspecialchars($_v['data_workspace']['id'], ENT_QUOTES);?>
">
			<span><?php echo htmlspecialchars($_v['data_workspace']['name'], ENT_QUOTES);?>
</span>
<?php if($_v['data_workspace']['bad']){?>
<small class="bad" title="<?php echo htmlspecialchars($_v['data_workspace']['badTitle'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_workspace']['badge'], ENT_QUOTES);?>
</small><?php }else{?>
<small><?php echo htmlspecialchars($_v['data_workspace']['count'], ENT_QUOTES);?>
</small><?php }?>
		</a>
	</summary>
<?php if(!empty($_v['data_workspace']['projects']))foreach($_v['data_workspace']['projects'] as $_v['project']){?><?php $_v['data_project'] = $_v['project'];?>
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
<?php }?>
</details>
<?php }else{if(!empty($_v['data_workspace']['projects']))foreach($_v['data_workspace']['projects'] as $_v['project']){?><?php $_v['data_project'] = $_v['project'];?>
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
<?php } }}else{ $_v['data_project'] = $_v['group'];?>
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

<?php } }else{?>

				<div class="empty">No folders yet. <?php if($_v['chrome']['can']['project/edit']){?>
<a href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/settings/projects/new">Add a git repository</a><?php }else{?>
Add a git repository in Settings<?php }?>, or put a folder with jobs under the root.</div>
<?php }?>
			</nav>
