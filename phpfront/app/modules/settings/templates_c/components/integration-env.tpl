<?php $_v=&$this->vars;?>
						<p class="faint integration-env-note">A set API key sends every variable below; an empty value takes its default.</p>
						<div class="table-wrap integration-env">
							<table class="table">
								<thead><tr><th>Variable</th><th>Value</th></tr></thead>
								<tbody>
<?php if(!empty($_v['data_env']))foreach($_v['data_env'] as $_v['row']){?>								<tr>
									<td><b class="mono"><?php echo htmlspecialchars($_v['row']['env'], ENT_QUOTES);?>
</b><div class="faint"><?php echo htmlspecialchars($_v['row']['help'], ENT_QUOTES);?>
</div></td>
<?php if($_v['row']['source']){?>
									<td><span class="faint">the API key</span>
<?php }elseif($_v['row']['empty']){?>
									<td><span class="faint">empty</span>
<?php }else{?>
									<td>
									<input type="text" id="<?php echo htmlspecialchars($_v['row']['id'], ENT_QUOTES);?>
" name="<?php echo htmlspecialchars($_v['row']['name'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['row']['value'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['row']['env'], ENT_QUOTES);?>
" autocomplete="off" spellcheck="false">
<?php }?>									</td>
								</tr>
<?php }?>								</tbody>
							</table>
						</div>
