<?php $_v=&$this->vars;?>
<div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
<?php if($_v['data_space']['none']){?>
		<option value="" data-workspace="" data-rel="."<?php if($_v['data_space']['noneSelected']){?>
 selected<?php }?>>none (the workspace root)</option>
<?php }?><?php if(!empty($_v['data_space']['spaces']))foreach($_v['data_space']['spaces'] as $_v['one']){?>		<option value="<?php echo htmlspecialchars($_v['one']['id'], ENT_QUOTES);?>
"<?php if($_v['one']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['one']['title'], ENT_QUOTES);?>
</option>
<?php }?><?php if($_v['data_space']['flows']){?>
		<option value="workflow:"<?php if($_v['data_space']['flowsSelected']){?>
 selected<?php }?>>Workflows</option>
<?php }?>	</select>
</div>
