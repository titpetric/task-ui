<?php $_v=&$this->vars; $_v['data_head'] = $_v['data']['head'];?>

<div class="page-head">
	<div>
		<h2><?php if($_v['data_head']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['data_head']['icon'], ENT_QUOTES);?>
"/></svg> <?php }?><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES); if($_v['data_head']['badge']){?>
 <span class="badge<?php if($_v['data_head']['tone']){?>
 <?php echo htmlspecialchars($_v['data_head']['tone'], ENT_QUOTES); }?>"><?php echo htmlspecialchars($_v['data_head']['badge'], ENT_QUOTES);?>
</span><?php }?><?php if(!empty($_v['data_head']['tags']))foreach($_v['data_head']['tags'] as $_v['tag']){?> <span class="badge"><?php echo htmlspecialchars($_v['tag'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_head']['note']){?>
 <span class="term-title" title="<?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_head']['note'], ENT_QUOTES);?>
</span><?php }?></h2>
<?php if($_v['data_head']['lead'] || $_v['data_head']['leads']){?>
		<p><?php echo $_v['data_head']['lead'];?>
</p>
<?php }?><?php if($_v['data_head']['faint']){?>
		<p class="faint"><?php echo htmlspecialchars($_v['data_head']['faint'], ENT_QUOTES);?>
</p>
<?php }?>	</div>
<?php if($_v['data_head']['show']){?>
	<div class="actions"<?php if($_v['data_head']['live']){?>
 data-actions<?php }?>>
<?php if($_v['data_head']['run']){?>
		<?php $_v['data_run'] = $_v['data_head']['run'];?>
<div class="run-split">
	<button type="button" class="btn btn-primary btn-sm" data-run data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
" title="Run with arguments after <?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?>><svg class="ic"><use href="#i-play"/></svg>Run<?php if($_v['data_run']['takesArgs']){?>
...<?php }?></button>
<?php if($_v['data_run']['menu']){?>
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
" aria-label="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-down"/></svg></summary>
		<div class="more-list">
			<div class="more-items" role="menu" aria-label="Run options">
<?php if($_v['data_run']['summarize']){?>
				<button type="button" class="more-item" role="menuitem" data-run data-summarize data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['what'], ENT_QUOTES);?>
"<?php if($_v['data_run']['args']){?>
 data-args="<?php echo htmlspecialchars($_v['data_run']['args'], ENT_QUOTES);?>
"<?php }?> title="Run and summarize"><svg class="ic"><use href="#i-summary"/></svg><span>Run and summarize</span></button>
<?php }?><?php if($_v['data_run']['force']){?>
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['place'], ENT_QUOTES);?>
" title="Run out of hours, with a reason"><svg class="ic"><use href="#i-play"/></svg><span>Run force</span></button>
<?php }?>			</div>
		</div>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data_head']['buttons']['items']){?>
		<?php $_v['data_buttons'] = $_v['data_head']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }}?><?php if($_v['data_head']['actions']){?>
		<?php $_v['data_menu'] = $_v['data_head']['actions'];?>

		<?php if(!$_v['data_menu']['empty']){?>
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
</details><?php }}?>	</div>
<?php }?></div>

<?php if($_v['data']['dir']){?>
<section class="block workspace-keys" id="keys">
<?php $_v['data_block'] = $_v['data']['block'];?>
<div class="block-head<?php if($_v['data_block']['empty']){?>
 is-empty<?php }?>">
	<h3><?php echo htmlspecialchars($_v['data_block']['title'], ENT_QUOTES); if($_v['data_block']['note']){?>
 <small><?php echo htmlspecialchars($_v['data_block']['note'], ENT_QUOTES);?>
</small><?php }?></h3>
<?php if($_v['data_block']['empty']){?>
	<p class="empty-line"><?php echo htmlspecialchars($_v['data_block']['empty'], ENT_QUOTES);?>
</p>
<?php }?><?php if($_v['data_block']['buttons']['items']){?>
	<div class="actions"><?php $_v['data_buttons'] = $_v['data_block']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg></button>
<?php }elseif($_v['button']['faint']){?>
	<span class="faint run-denied"><?php echo htmlspecialchars($_v['button']['faint'], ENT_QUOTES);?>
</span>
<?php }elseif($_v['button']['href']){?>
	<a class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>" href="<?php echo htmlspecialchars($_v['button']['href'], ENT_QUOTES);?>
"<?php if($_v['button']['download']){?>
 download<?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</a>
<?php }else{?>
	<button type="button" class="btn<?php if($_v['data_buttons']['small'] || $_v['button']['small']){?>
 btn-sm<?php }?><?php if($_v['button']['primary']){?>
 btn-primary<?php }?><?php if($_v['button']['danger']){?>
 btn-danger<?php }?>"<?php if($_v['button']['disabled']){?>
 disabled<?php if($_v['button']['note']){?>
 title="<?php echo htmlspecialchars($_v['button']['note'], ENT_QUOTES);?>
"<?php }?><?php }?><?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><?php if($_v['button']['icon']){?>
<svg class="ic"><use href="#i-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
"/></svg><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }?>
	</div>
<?php }?></div>
<?php $_v['data_keys'] = $_v['data']['keys']; if($_v['data_keys']){?>
	<div class="table-wrap fit">
	<table class="table keys key-table cards">
		<colgroup><col class="c-name"><col class="c-algo"><col class="c-size"><col class="c-pub"><col class="c-actions"></colgroup>
		<thead><tr><th>Name</th><th class="hide-narrow">Algorithm</th><th class="hide-narrow">Size</th><th>Public key</th><th class="actions" aria-label="Actions"></th></tr></thead>

		<tbody>
<?php if(!empty($_v['data_keys']))foreach($_v['data_keys'] as $_v['row']){?><?php $_v['data_key'] = $_v['row'];?>
			<tr id="key-<?php echo htmlspecialchars($_v['data_key']['name'], ENT_QUOTES);?>
">
				<td class="key"><?php if($_v['data_key']['href']){?>
<a class="mono" href="<?php echo htmlspecialchars($_v['data_key']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_key']['name'], ENT_QUOTES);?>
</a><?php }else{?>
<span class="mono"><?php echo htmlspecialchars($_v['data_key']['name'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_key']['encrypts']){?>
 <span class="badge signal">encryption</span><?php }?><?php if($_v['data_key']['backup']){?>
<span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: <?php echo htmlspecialchars($_v['data_key']['backup'], ENT_QUOTES);?>
</span><?php }?><?php if($_v['data_key']['comment']){?>
<div class="faint"><?php echo htmlspecialchars($_v['data_key']['comment'], ENT_QUOTES);?>
</div><?php }?></td>
				<td class="mono hide-narrow" data-label="Algorithm"><?php echo htmlspecialchars($_v['data_key']['algo'], ENT_QUOTES);?>
</td>
				<td class="mono nowrap hide-narrow" data-label="Size"><?php echo htmlspecialchars($_v['data_key']['size'], ENT_QUOTES);?>
 bits</td>
				<td class="pubkey" data-label="Public key"><?php $_v['data_copy'] = $_v['data_key']['copy'];?>
<span class="copyable<?php if($_v['data_copy']['fill']){?>
 fill<?php }?>"><code<?php if($_v['data_copy']['fill'] || $_v['data_copy']['shown'] !== $_v['data_copy']['value']){?>
 title="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_copy']['shown'], ENT_QUOTES);?>
</code><button type="button" class="btn-icon" data-copy="<?php echo htmlspecialchars($_v['data_copy']['value'], ENT_QUOTES);?>
" title="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
" aria-label="Copy <?php echo htmlspecialchars($_v['data_copy']['label'], ENT_QUOTES);?>
"><svg class="ic"><use href="#i-copy"/></svg></button></span>
</td>
				<td class="actions"><?php $_v['data_menu'] = $_v['data_key']['menu']; if(!$_v['data_menu']['empty']){?>
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
</td>
			</tr>

<?php }?>		</tbody>
	</table>
	</div>
<?php }?>
</section>
<?php if($_v['data']['dialogs']){ $_v['data_wsdialogs'] = $_v['data']['dialogs'];?>
<dialog class="confirm" id="key-generate-dialog" aria-labelledby="key-generate-title">
	<?php $_v['data_wsgen'] = $_v['data_wsdialogs']['generate'];?>
<form class="form" data-api="<?php echo htmlspecialchars($_v['data_wsgen']['api'], ENT_QUOTES);?>
" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data_wsgen']['next'], ENT_QUOTES);?>
">
<?php if($_v['data_wsgen']['dialog']){?>
	<h3 id="key-generate-title">Generate key</h3>
	<p>Generates an SSH key of <?php echo htmlspecialchars($_v['data_wsgen']['folder'], ENT_QUOTES);?>
 in <?php echo htmlspecialchars($_v['data_wsgen']['dir'], ENT_QUOTES);?>
, named <span class="mono">workspace_<?php echo htmlspecialchars($_v['data_wsgen']['folder'], ENT_QUOTES);?>
_&lt;name&gt;_&lt;algorithm&gt;</span>.</p>
<?php }?>	<div class="field">
		<label for="wkg-name">Name</label>
		<input type="text" id="wkg-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wkg-algo">Algorithm</label>
		<select id="wkg-algo" name="algo" data-key-algo>
			<option value="ed25519" selected>ed25519</option>
			<option value="ecdsa" data-bits="256">ECDSA P-256</option>
			<option value="ecdsa" data-bits="384">ECDSA P-384</option>
			<option value="ecdsa" data-bits="521">ECDSA P-521</option>
			<option value="rsa" data-bits="2048">RSA 2048</option>
			<option value="rsa" data-bits="3072">RSA 3072</option>
			<option value="rsa" data-bits="4096">RSA 4096</option>
		</select>
		<input type="hidden" name="bits" value="0" data-type="int">
	</div>
	<div class="field">
		<label for="wkg-comment">Comment</label>
		<input type="text" id="wkg-comment" name="comment" value="<?php echo htmlspecialchars($_v['data_wsgen']['comment'], ENT_QUOTES);?>
">
	</div>
<?php $_v['data_wsfoot'] = $_v['data_wsgen']['foot'];?>
	<div class="form-foot">
<?php if($_v['data_wsfoot']['dialog']){?>
		<button type="button" class="btn" data-dialog-close>Cancel</button>
<?php }else{?>
		<a class="btn" href="<?php echo htmlspecialchars($_v['data_wsfoot']['cancel'], ENT_QUOTES);?>
">Cancel</a>
<?php }?>		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_wsfoot']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>
</dialog>
<dialog class="confirm" id="key-upload-dialog" aria-labelledby="key-upload-title">
	<?php $_v['data_wsup'] = $_v['data_wsdialogs']['upload'];?>
<form class="form" data-api="<?php echo htmlspecialchars($_v['data_wsup']['api'], ENT_QUOTES);?>
" data-method="POST" data-next="<?php echo htmlspecialchars($_v['data_wsup']['next'], ENT_QUOTES);?>
">
<?php if($_v['data_wsup']['dialog']){?>
	<h3 id="key-upload-title">Upload key</h3>
	<p>Saves a private key you already use as an SSH key of <?php echo htmlspecialchars($_v['data_wsup']['folder'], ENT_QUOTES);?>
. The folder refuses a key it holds, under any name.</p>
<?php }?>	<div class="field">
		<label for="wku-file">Private key file</label>
		<input type="file" id="wku-file" data-read-into="wku-content">
	</div>
	<div class="field">
		<label for="wku-content">Or paste the private key</label>
		<textarea id="wku-content" name="content" rows="<?php echo htmlspecialchars($_v['data_wsup']['rows'], ENT_QUOTES);?>
" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	<div class="field">
		<label for="wku-name">Name</label>
		<input type="text" id="wku-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wku-comment">Comment</label>
		<input type="text" id="wku-comment" name="comment" value="<?php echo htmlspecialchars($_v['data_wsup']['comment'], ENT_QUOTES);?>
">
	</div>
<?php $_v['data_wsfoot'] = $_v['data_wsup']['foot'];?>
	<div class="form-foot">
<?php if($_v['data_wsfoot']['dialog']){?>
		<button type="button" class="btn" data-dialog-close>Cancel</button>
<?php }else{?>
		<a class="btn" href="<?php echo htmlspecialchars($_v['data_wsfoot']['cancel'], ENT_QUOTES);?>
">Cancel</a>
<?php }?>		<button class="btn btn-primary" type="submit"><?php echo htmlspecialchars($_v['data_wsfoot']['label'], ENT_QUOTES);?>
</button>
	</div>
</form>
</dialog>
<?php } }else{?>
<div class="alert">The instance has no key folder, so it keeps no SSH keys.</div>
<?php }?><?php if($_v['data']['trail']){ $_v['data_trail'] = $_v['data']['trail'];?>
<nav class="setup-trail" aria-label="Setup steps">
<?php if($_v['data_trail']['previous']){?>
	<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data_trail']['previous']['href'], ENT_QUOTES);?>
" rel="prev"><svg class="ic back"><use href="#i-chev"/></svg>Previous: <?php echo htmlspecialchars($_v['data_trail']['previous']['name'], ENT_QUOTES);?>
</a>
<?php }?>	<ol class="steps">
<?php if(!empty($_v['data_trail']['steps']))foreach($_v['data_trail']['steps'] as $_v['step']){?>		<li><?php if($_v['step']['current']){?>
<b aria-current="step"><?php echo htmlspecialchars($_v['step']['name'], ENT_QUOTES);?>
</b><?php }else{?>
<a href="<?php echo htmlspecialchars($_v['step']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['step']['name'], ENT_QUOTES);?>
</a><?php }?></li>
<?php }?>	</ol>
<?php if($_v['data_trail']['next']){?>
	<a class="btn btn-sm" href="<?php echo htmlspecialchars($_v['data_trail']['next']['href'], ENT_QUOTES);?>
" rel="next">Next: <?php echo htmlspecialchars($_v['data_trail']['next']['name'], ENT_QUOTES);?>
<svg class="ic"><use href="#i-chev"/></svg></a>
<?php }?></nav>
<?php }?>
