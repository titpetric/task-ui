<?php $_v=&$this->vars;?>
		<li class="wf-step integration-step kind-<?php echo htmlspecialchars($_v['data_card']['id'], ENT_QUOTES);?>
" id="<?php echo htmlspecialchars($_v['data_card']['id'], ENT_QUOTES);?>
">
			<span class="wf-node" aria-hidden="true"><?php echo htmlspecialchars($_v['data_card']['number'], ENT_QUOTES);?>
</span>
			<div class="wf-card">
				<div class="wf-card-head">
					<button type="button" class="wf-toggle" data-card-toggle aria-expanded="false">
						<span class="wf-kind"><svg class="ic" aria-hidden="true"><use href="#i-plug"/></svg>Integration</span>
						<span class="wf-line"><?php echo htmlspecialchars($_v['data_card']['title'], ENT_QUOTES);?>
</span><?php if($_v['data_card']['configured']){?>
<span class="badge ok">set up</span><?php }else{?>
<span class="badge">not set up</span><?php }?>
						<svg class="ic wf-chev" aria-hidden="true"><use href="#i-chev"/></svg>
					</button>
					<?php $_v['data_menu'] = $_v['data_card']['menu']; if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-more"/></svg></summary>
	<div class="more-list">
<?php if($_v['data_menu']['details']){?>
		<dl class="more-details">
<?php if(!empty($_v['data_menu']['details']))foreach($_v['data_menu']['details'] as $_v['detail']){?>			<dt><?php echo htmlspecialchars($_v['detail']['label'], ENT_QUOTES);?>
</dt><dd<?php if($_v['detail']['mono']){?>
 class="mono"<?php }?><?php if($_v['detail']['title']){?>
 title="<?php echo htmlspecialchars($_v['detail']['title'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['detail']['value'], ENT_QUOTES);?>
</dd>
<?php }?>		</dl>
<?php }?>		<div class="more-items" role="menu" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_menu']['groups']))foreach($_v['data_menu']['groups'] as $_v['index'] => $_v['group']){?><?php if($_v['index']){?>
			<hr class="more-sep">
<?php }?><?php if(!empty($_v['group']))foreach($_v['group'] as $_v['item']){?><?php if($_v['item']['href']){?>
			<a class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem" href="<?php echo htmlspecialchars($_v['item']['href'], ENT_QUOTES);?>
"<?php if($_v['item']['download']){?>
 download<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
"/></svg><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }?>

				</div>
				<div class="wf-card-body">
					<div class="dim desc integration-desc"><?php echo $_v['data_card']['desc'];?>
</div>
					<form class="form integration-form" data-api="<?php echo htmlspecialchars($_v['data_card']['api'], ENT_QUOTES);?>
" data-method="PUT" data-next="<?php echo htmlspecialchars($_v['data_card']['next'], ENT_QUOTES);?>
">
						<div class="wf-fields">
<?php if(!empty($_v['data_card']['fields']))foreach($_v['data_card']['fields'] as $_v['field']){?><?php $_v['data_field'] = $_v['field']; if($_v['data_field']['kind'] === "who"){?>
							<div class="field wf-wide">
								<span class="label"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES);?>
</span><?php $_v['data_whofield'] = $_v['data_field']['who'];?>
<div class="whofield<?php if($_v['data_whofield']['compact']){?>
 compact<?php }?>" data-who-empty="<?php echo htmlspecialchars($_v['data_whofield']['empty'], ENT_QUOTES);?>
">
	<div class="checks who-roles">
		<span class="faint">Roles</span><?php if(!empty($_v['data_whofield']['roles']))foreach($_v['data_whofield']['roles'] as $_v['role']){?>
		<label class="check"><input type="checkbox" name="<?php echo htmlspecialchars($_v['data_whofield']['name'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['role']['name'], ENT_QUOTES);?>
" data-type="<?php echo htmlspecialchars($_v['data_whofield']['type'], ENT_QUOTES);?>
"<?php if($_v['role']['checked']){?>
 checked<?php }?>> <?php echo htmlspecialchars($_v['role']['name'], ENT_QUOTES);?>
 <small class="count">· <?php echo htmlspecialchars($_v['role']['count'], ENT_QUOTES);?>
</small></label><?php }?>
	</div>
	<div class="checks who-users">
		<span class="faint">Users</span>
		<ul class="who-list"><?php if(!empty($_v['data_whofield']['users']))foreach($_v['data_whofield']['users'] as $_v['user']){?>
		<li><label class="check who-user" data-role="<?php echo htmlspecialchars($_v['user']['role'], ENT_QUOTES);?>
"><input type="checkbox" name="<?php echo htmlspecialchars($_v['data_whofield']['userName'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['user']['name'], ENT_QUOTES);?>
" data-type="<?php echo htmlspecialchars($_v['data_whofield']['type'], ENT_QUOTES);?>
"<?php if($_v['user']['checked']){?>
 checked<?php }?>> <?php echo htmlspecialchars($_v['user']['name'], ENT_QUOTES);?>
 <small class="role"><?php echo htmlspecialchars($_v['user']['role'], ENT_QUOTES);?>
</small> <small class="who-state"></small></label></li><?php }?><?php if(!empty($_v['data_whofield']['extra']))foreach($_v['data_whofield']['extra'] as $_v['other']){?>
		<li><label class="check who-user" data-role=""><input type="checkbox" name="<?php echo htmlspecialchars($_v['data_whofield']['userName'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['other'], ENT_QUOTES);?>
" data-type="<?php echo htmlspecialchars($_v['data_whofield']['type'], ENT_QUOTES);?>
" checked> <?php echo htmlspecialchars($_v['other'], ENT_QUOTES);?>
 <small class="role">no such user</small> <small class="who-state"></small></label></li><?php }?>
		</ul>
	</div><?php if(!$_v['data_whofield']['compact']){?>

	<p class="dim who-summary"></p><?php }?>
</div><?php if($_v['data_field']['help']){?>
<small class="faint"><?php echo htmlspecialchars($_v['data_field']['help'], ENT_QUOTES);?>
</small><?php }?>
							</div>
<?php }elseif($_v['data_field']['kind'] === "env"){?>
							<div class="field wf-wide env-flag">
								<div class="field-head">
									<span class="label"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES);?>
</span><?php $_v['data_switch'] = $_v['data_field']['switch'];?>
<label class="switch env-switch" title="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
 in a run">
	<input type="checkbox" name="<?php echo htmlspecialchars($_v['data_switch']['name'], ENT_QUOTES);?>
" value="1" data-type="toggle"<?php if($_v['data_switch']['on']){?>
 checked<?php }?> aria-label="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
">
	<i aria-hidden="true"></i>
</label>
								</div>
								<small class="faint"><?php if($_v['data_field']['help']){ echo htmlspecialchars($_v['data_field']['help'], ENT_QUOTES); if($_v['data_field']['spaced']){?>
 <?php }?><?php }?><code><?php echo htmlspecialchars($_v['data_field']['env'], ENT_QUOTES);?>
</code></small>
							</div>
<?php }else{?>
							<div class="field<?php if($_v['data_field']['wide']){?>
 wf-wide<?php }?>">
<?php if($_v['data_field']['toggle']){?>
								<div class="field-head">
									<label for="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES); if($_v['data_field']['badge']){?>
 <span class="badge<?php if($_v['data_field']['tone']){?>
 <?php echo htmlspecialchars($_v['data_field']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_field']['badge'], ENT_QUOTES);?>
</span><?php }?></label><?php $_v['data_switch'] = $_v['data_field']['switch'];?>
<label class="switch env-switch" title="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
 in a run">
	<input type="checkbox" name="<?php echo htmlspecialchars($_v['data_switch']['name'], ENT_QUOTES);?>
" value="1" data-type="toggle"<?php if($_v['data_switch']['on']){?>
 checked<?php }?> aria-label="Send <?php echo htmlspecialchars($_v['data_switch']['env'], ENT_QUOTES);?>
">
	<i aria-hidden="true"></i>
</label>
								</div>
<?php }else{?>
								<label for="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_field']['label'], ENT_QUOTES); if($_v['data_field']['badge']){?>
 <span class="badge<?php if($_v['data_field']['tone']){?>
 <?php echo htmlspecialchars($_v['data_field']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_field']['badge'], ENT_QUOTES);?>
</span><?php }?></label>
<?php }?><?php if($_v['data_field']['kind'] === "secret"){?>
								<input type="password" id="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
" name="<?php echo htmlspecialchars($_v['data_field']['name'], ENT_QUOTES);?>
" placeholder="<?php echo htmlspecialchars($_v['data_field']['placeholder'], ENT_QUOTES);?>
" autocomplete="new-password" spellcheck="false">
<?php }elseif($_v['data_field']['kind'] === "select"){?>
								<select id="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
" name="<?php echo htmlspecialchars($_v['data_field']['name'], ENT_QUOTES);?>
">
<?php if(!empty($_v['data_field']['options']))foreach($_v['data_field']['options'] as $_v['option']){?>									<option value="<?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
"<?php if($_v['option']['selected']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['option']['value'], ENT_QUOTES);?>
</option>
<?php }?>								</select>
<?php }else{?>
								<input type="text" id="<?php echo htmlspecialchars($_v['data_field']['id'], ENT_QUOTES);?>
" name="<?php echo htmlspecialchars($_v['data_field']['name'], ENT_QUOTES);?>
" value="<?php echo htmlspecialchars($_v['data_field']['value'], ENT_QUOTES);?>
" placeholder="<?php echo htmlspecialchars($_v['data_field']['placeholder'], ENT_QUOTES);?>
" autocomplete="off" spellcheck="false">
<?php }?>								<small class="faint"><?php if($_v['data_field']['help']){ echo htmlspecialchars($_v['data_field']['help'], ENT_QUOTES); if($_v['data_field']['spaced']){?>
 <?php }?><?php }?><?php if($_v['data_field']['env']){?>
<code><?php echo htmlspecialchars($_v['data_field']['env'], ENT_QUOTES);?>
</code><?php }?></small>
							</div>
<?php }?><?php }?>						</div>
<?php if($_v['data_card']['env']){ $_v['data_env'] = $_v['data_card']['env'];?>
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
<?php }?>						<details class="integration-key">
							<summary>Get the <?php echo htmlspecialchars($_v['data_card']['credential'], ENT_QUOTES);?>
</summary>
							<ol class="steps">
<?php if(!empty($_v['data_card']['steps']))foreach($_v['data_card']['steps'] as $_v['step']){?>								<li><?php echo $_v['step'];?>
</li>
<?php }?>							</ol>
<?php if($_v['data_card']['hook']){?>
							<p class="dim"><?php echo $_v['data_card']['hook'];?>

							</p>
<?php }?>						</details>
						<div class="form-foot">
<?php if($_v['data_card']['runs']){?>
							<button type="button" class="btn" data-run data-folder="internal" data-job="@forgejo-sync" data-run-args='["--dry-run"]' data-title="internal / forgejo-sync --dry-run" title="Run @forgejo-sync --dry-run: the report alone, nothing changes">Dry run</button>
							<button type="button" class="btn" data-run data-folder="internal" data-job="@forgejo-sync" data-title="internal / forgejo-sync" title="Run @forgejo-sync now: workspaces, homes, projects and hooks">Sync now</button>
<?php }?><?php if($_v['data_card']['links']){?>
							<button type="button" class="btn" data-spotify-link title="Link your own Spotify account to this user">Connect</button>
<?php }?>							<button class="btn btn-primary" type="submit" disabled>Save</button>
						</div>
					</form>
<?php if($_v['data_card']['sync']){?>
					<p class="dim integration-sync"><?php if($_v['data_card']['last']){?>
Last sync <?php $_v['data_state'] = $_v['data_card']['last']['state']; if($_v['data_state']['id']){?>
<span class="state" data-run-id="<?php echo htmlspecialchars($_v['data_state']['id'], ENT_QUOTES);?>
"<?php if($_v['data_state']['ended']){?>
 data-finished="true"<?php }?>><i class="dot <?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES);?>
"></i><?php echo htmlspecialchars($_v['data_state']['state'], ENT_QUOTES); if($_v['data_state']['code'] !== ""){?>
 <span class="faint">(<?php echo htmlspecialchars($_v['data_state']['code'], ENT_QUOTES);?>
)</span><?php }?></span><?php }else{?>
<span class="faint">never run</span><?php }?>
 <?php echo htmlspecialchars($_v['data_card']['last']['ago'], ENT_QUOTES);?>
 by <?php echo htmlspecialchars($_v['data_card']['last']['who'], ENT_QUOTES);?>
: <a href="<?php echo htmlspecialchars($_v['data_card']['last']['href'], ENT_QUOTES);?>
">the report</a>.
<?php }else{?>
No sync ran yet: <b>Dry run</b> shows what one would do.
<?php }?>					</p>
<?php }?>				</div>
			</div>
		</li>
