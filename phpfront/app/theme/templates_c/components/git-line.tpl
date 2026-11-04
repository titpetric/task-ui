<?php $_v=&$this->vars;?>

<div class="actionsbar">
	<div class="gitline" title="<?php echo htmlspecialchars($_v['data_git']['title'], ENT_QUOTES);?>
">
		<span class="branch"><span class="glyph" aria-hidden="true">⎇</span> <?php if($_v['data_git']['branch']){ echo htmlspecialchars($_v['data_git']['branch'], ENT_QUOTES); }else{?>
<span class="faint">detached</span> <?php echo htmlspecialchars($_v['data_git']['detached'], ENT_QUOTES); }?></span>
<?php if($_v['data_git']['ahead']){?>
		<span class="ahead" title="<?php echo htmlspecialchars($_v['data_git']['ahead'], ENT_QUOTES);?>
 commits to push">↑<?php echo htmlspecialchars($_v['data_git']['ahead'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_git']['behind']){?>
		<span class="behind" title="<?php echo htmlspecialchars($_v['data_git']['behind'], ENT_QUOTES);?>
 commits to pull">↓<?php echo htmlspecialchars($_v['data_git']['behind'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_git']['dirty']){?>
		<button type="button" class="dirty" data-git-status aria-expanded="false" aria-controls="git-status" title="Show the changed files"><span class="glyph" aria-hidden="true">●</span> <?php echo htmlspecialchars($_v['data_git']['files'], ENT_QUOTES);?>
<svg class="ic chev" aria-hidden="true"><use href="#i-chev"/></svg></button>
<?php }?><?php if($_v['data_git']['clean']){?>
		<span class="clean">up to date</span>
<?php }?>		<span class="where">· <?php echo htmlspecialchars($_v['data_git']['where'], ENT_QUOTES);?>
</span>
	</div>
<?php if($_v['data_git']['dirty']){?>
	<div class="gitstatus" id="git-status" hidden>
		<div class="table-wrap">
			<table class="table gitstatus-table">
				<thead><tr><th>Status</th><th>File</th></tr></thead>
				<tbody>
<?php if(!empty($_v['data_git']['changes']))foreach($_v['data_git']['changes'] as $_v['change']){?>					<tr><td class="gitstate"><code><?php echo htmlspecialchars($_v['change']['code'], ENT_QUOTES);?>
</code> <span class="faint"><?php echo htmlspecialchars($_v['change']['state'], ENT_QUOTES);?>
</span></td><td class="gitpath"><?php echo htmlspecialchars($_v['change']['path'], ENT_QUOTES); if($_v['change']['from']){?>
 <span class="faint">from <?php echo htmlspecialchars($_v['change']['from'], ENT_QUOTES);?>
</span><?php }?></td></tr>
<?php }?>				</tbody>
			</table>
		</div>
	</div>
<?php }?></div>
