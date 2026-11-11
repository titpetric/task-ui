<?php $_v=&$this->vars;?>
	<div class="field" data-folder-pick<?php if($_v['data_pick']['label']){?>
 role="group" aria-labelledby="folders-label"<?php }?>>
<?php if($_v['data_pick']['label']){?>
		<span class="label" id="folders-label">Folders</span>
<?php }?>		<label class="check"><input type="radio" name="scope" value="all"<?php if($_v['data_pick']['all']){?>
 checked<?php }?>> All folders</label>
		<label class="check"><input type="radio" name="scope" value="custom"<?php if(!$_v['data_pick']['all']){?>
 checked<?php }?>> Custom</label>
		<div class="folder-choices">
<?php if(!empty($_v['data_pick']['choices']))foreach($_v['data_pick']['choices'] as $_v['choice']){?>			<label class="check"><input type="checkbox" name="folders" value="<?php echo htmlspecialchars($_v['choice']['id'], ENT_QUOTES);?>
" data-type="set"<?php if($_v['choice']['checked']){?>
 checked<?php }?><?php if($_v['data_pick']['all']){?>
 disabled<?php }?>> <span class="mono"><?php echo htmlspecialchars($_v['choice']['id'], ENT_QUOTES);?>
</span></label>
<?php }?><?php if(!$_v['data_pick']['choices']){?>
			<span class="faint">No folders yet.</span>
<?php }?>		</div>
	</div>
