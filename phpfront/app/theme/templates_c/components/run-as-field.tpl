<?php $_v=&$this->vars; if($_v['data_runas']['options']['users']){?>
	<div class="field">
		<label for="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
">Runs as</label>
		<select id="<?php echo htmlspecialchars($_v['data_runas']['id'], ENT_QUOTES);?>
" name="user">
<?php $_v['data_options'] = $_v['data_runas']['options'];if(!empty($_v['data_options']['users']))foreach($_v['data_options']['users'] as $_v['option']){?>			<option value="<?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES);?>
"<?php if($_v['option']['name'] == $_v['data_options']['selected']){?>
 selected<?php }?> data-keys="<?php echo htmlspecialchars($_v['option']['keys'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['option']['name'], ENT_QUOTES); if(!$_v['data_options']['short']){?>
 (<?php echo htmlspecialchars($_v['option']['role'], ENT_QUOTES);?>
)<?php }?> · <?php echo htmlspecialchars($_v['option']['said'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
		<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>
	</div>
<?php }?>