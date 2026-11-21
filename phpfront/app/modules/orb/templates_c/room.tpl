<?php $_v=&$this->vars; $_v['data_head'] = $_v['data']['head'];?>

<div class="page-head">
	<div>
		<h2><?php if($_v['data_head']['icon']){?>
<i class="ic ic-<?php echo htmlspecialchars($_v['data_head']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i> <?php }?><?php echo htmlspecialchars($_v['data_head']['title'], ENT_QUOTES); if($_v['data_head']['badge']){?>
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
"<?php }?>><i class="ic ic-play" aria-hidden="true"></i>Run<?php if($_v['data_run']['takesArgs']){?>
...<?php }?></button>
<?php if($_v['data_run']['menu']){?>
	<details class="more-menu run-menu">
		<summary class="btn btn-primary btn-sm run-caret" aria-haspopup="menu" aria-expanded="false" title="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
" aria-label="More ways to run <?php echo htmlspecialchars($_v['data_run']['name'], ENT_QUOTES);?>
"><i class="ic ic-down" aria-hidden="true"></i></summary>
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
"<?php }?> title="Run and summarize"><i class="ic ic-summary" aria-hidden="true"></i><span>Run and summarize</span></button>
<?php }?><?php if($_v['data_run']['force']){?>
				<button type="button" class="more-item danger" role="menuitem" data-run data-force data-folder="<?php echo htmlspecialchars($_v['data_run']['folder'], ENT_QUOTES);?>
"<?php if($_v['data_run']['dir']){?>
 data-dir="<?php echo htmlspecialchars($_v['data_run']['dir'], ENT_QUOTES);?>
"<?php }?> data-job="<?php echo htmlspecialchars($_v['data_run']['job'], ENT_QUOTES);?>
" data-title="<?php echo htmlspecialchars($_v['data_run']['place'], ENT_QUOTES);?>
" title="Run out of hours, with a reason"><i class="ic ic-play" aria-hidden="true"></i><span>Run force</span></button>
<?php }?>			</div>
		</div>
	</details>
<?php }?></div>

<?php }?><?php if($_v['data_head']['buttons']['items']){?>
		<?php $_v['data_buttons'] = $_v['data_head']['buttons']; if(!empty($_v['data_buttons']['items']))foreach($_v['data_buttons']['items'] as $_v['button']){?><?php if($_v['button']['bare']){?>
	<button type="button" class="btn-icon"<?php if($_v['button']['attrs']){?>
 <?php echo $_v['button']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i></button>
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
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
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
<i class="ic ic-<?php echo htmlspecialchars($_v['button']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><?php }?><?php echo htmlspecialchars($_v['button']['label'], ENT_QUOTES);?>
</button>
<?php }?><?php }}?><?php if($_v['data_head']['actions']){?>
		<?php $_v['data_menu'] = $_v['data_head']['actions'];?>

		<?php if(!$_v['data_menu']['empty']){?>
<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
" aria-label="<?php echo htmlspecialchars($_v['data_menu']['label'], ENT_QUOTES);?>
"><i class="ic ic-more" aria-hidden="true"></i></summary>
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
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></a>
<?php }else{?>
			<button type="button" class="more-item<?php if($_v['item']['danger']){?>
 danger<?php }?>" role="menuitem"<?php if($_v['item']['disabled']){?>
 disabled<?php }?><?php if($_v['item']['attrs']){?>
 <?php echo $_v['item']['attrs']; }?>><i class="ic ic-<?php echo htmlspecialchars($_v['item']['icon'], ENT_QUOTES);?>
" aria-hidden="true"></i><span><?php echo htmlspecialchars($_v['item']['label'], ENT_QUOTES);?>
</span><?php if($_v['item']['note']){?>
<small class="faint"><?php echo htmlspecialchars($_v['item']['note'], ENT_QUOTES);?>
</small><?php }?></button>
<?php }?><?php }?><?php }?>
		</div>
	</div>
</details><?php }}?>	</div>
<?php }?></div>


<section class="orb-room" data-orb-room
	data-meeting="<?php echo htmlspecialchars($_v['data']['room']['id'], ENT_QUOTES);?>
" data-state="<?php echo htmlspecialchars($_v['data']['room']['state'], ENT_QUOTES);?>
"
	data-self="<?php echo htmlspecialchars($_v['data']['room']['self'], ENT_QUOTES);?>
" data-self-name="<?php echo htmlspecialchars($_v['data']['room']['self_name'], ENT_QUOTES);?>
"
	data-host="<?php if($_v['data']['room']['host']){?>
1<?php }else{?>
0<?php }?>"
	data-can-edit="<?php if($_v['data']['room']['can_edit']){?>
1<?php }else{?>
0<?php }?>"
	data-can-delete="<?php if($_v['data']['room']['can_delete']){?>
1<?php }else{?>
0<?php }?>"
	data-started="<?php echo htmlspecialchars($_v['data']['room']['started'], ENT_QUOTES);?>
" data-limit="<?php echo htmlspecialchars($_v['data']['room']['limit'], ENT_QUOTES);?>
"
	data-api="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/api" data-base="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
">
	<div class="orb-people" data-people>
<?php if(!empty($_v['data']['people']))foreach($_v['data']['people'] as $_v['person']){?>		<span class="orb-person<?php if($_v['person']['self']){?>
 self<?php }?><?php if($_v['person']['here']){?>
 here<?php }else{?>
 gone<?php }?>" data-person="<?php echo htmlspecialchars($_v['person']['id'], ENT_QUOTES);?>
" data-here="<?php if($_v['person']['here']){?>
1<?php }else{?>
0<?php }?>">
			<span class="orb-dot"></span><span class="orb-name"><?php echo htmlspecialchars($_v['person']['name'], ENT_QUOTES);?>
</span><?php if($_v['person']['joins'] > 1){?>
 <span class="orb-joins" title="Times joined"><?php echo htmlspecialchars($_v['person']['joins'], ENT_QUOTES);?>
</span><?php }?>
		</span>
<?php }?>	</div>
<?php if($_v['data']['room']['open']){?>
	<div class="orb-controls">
		<div class="orb-action">
<?php if($_v['data']['room']['can_edit']){?>
			<button class="btn btn-primary btn-lg" type="button" data-orb-join><i class="ic ic-users" aria-hidden="true"></i>Join meeting</button>
			<div class="orb-speak" data-orb-speak hidden>
				<button class="orb-mic" type="button" data-orb-mic aria-label="Start speaking" aria-pressed="false" title="Start speaking"><i class="ic ic-mic" aria-hidden="true"></i></button>
				<span class="orb-miclabel faint" data-orb-miclabel>Microphone off</span>
			</div>
<?php }else{?>
			<span class="faint">You can read this room but not speak in it.</span>
<?php }?>		</div>
		<span class="orb-timer" data-orb-timer title="Time left"></span>
	</div>
<?php if($_v['data']['room']['pick_agent']){?>
	<form class="orb-agent" data-orb-agent>
		<label class="faint" for="orb-agent">Agent</label>
		<select id="orb-agent" name="agent">
			<option value="">No agent</option>
<?php if(!empty($_v['data']['room']['agent_names']))foreach($_v['data']['room']['agent_names'] as $_v['name']){?>			<option value="<?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
"<?php if($_v['name'] == $_v['data']['room']['agent']){?>
 selected<?php }?>><?php echo htmlspecialchars($_v['name'], ENT_QUOTES);?>
</option>
<?php }?>		</select>
		<button class="btn" type="submit">Save</button>
	</form><?php }?>
	<p class="orb-status faint" data-orb-status aria-live="polite"><?php if($_v['data']['room']['can_edit']){?>
Join to hear the room and speak.<?php }else{?>
You need <code>meeting/edit</code> to join.<?php }?></p>
	<p class="faint orb-hint">Headphones keep the room's own voices out of your microphone.</p>
<?php }else{?>
	<p class="orb-status faint" data-orb-status>This meeting ended <?php echo htmlspecialchars($_v['data']['room']['ended'], ENT_QUOTES);?>
.</p>
<?php }if(!$_v['data']['room']['open']){?>
<section class="block orb-summary" data-orb-summary data-state="<?php echo htmlspecialchars($_v['data']['room']['summary_state'], ENT_QUOTES);?>
">
	<h3>Summary<?php if($_v['data']['room']['summary_state'] == "working"){?>
 <small class="faint">Claude is reading the transcript...</small><?php }?></h3>
	<div class="prose" data-summary-body><?php echo $_v['data']['room']['summary'];?>
</div>
	<p class="faint" data-summary-message><?php echo htmlspecialchars($_v['data']['room']['summary_message'], ENT_QUOTES);?>
</p>
<?php if($_v['data']['room']['can_edit']){?>
	<button class="btn" type="button" data-summary-retry<?php if($_v['data']['room']['summary_state'] != "failed"){?>
 hidden<?php }?>>Ask Claude again</button>
<?php }?></section>
<?php }?>

<details class="block orb-log-block"<?php if($_v['data']['room']['open']){?>
 open<?php }?>>
	<summary>Transcript<?php if(!$_v['data']['room']['open']){?>
, with joins and parts<?php }?></summary>
	<div class="orb-log" data-orb-log aria-live="polite">
<?php if(!empty($_v['data']['lines']))foreach($_v['data']['lines'] as $_v['line']){?><?php if($_v['line']['said']){?>
		<div class="orb-line" data-line-id="<?php echo htmlspecialchars($_v['line']['id'], ENT_QUOTES);?>
">
			<span class="orb-who"><?php echo htmlspecialchars($_v['line']['name'], ENT_QUOTES);?>
</span>
			<span class="orb-said"><?php echo htmlspecialchars($_v['line']['body'], ENT_QUOTES);?>
</span>
			<span class="orb-at" data-at="<?php echo htmlspecialchars($_v['line']['created'], ENT_QUOTES);?>
" title="<?php echo htmlspecialchars($_v['line']['at'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['line']['at'], ENT_QUOTES);?>
</span>
			<button class="btn-icon" type="button" data-play data-user="<?php echo htmlspecialchars($_v['line']['user'], ENT_QUOTES);?>
" data-start="<?php echo htmlspecialchars($_v['line']['start'], ENT_QUOTES);?>
" data-end="<?php echo htmlspecialchars($_v['line']['end'], ENT_QUOTES);?>
" title="Play this line"><i class="ic ic-play" aria-hidden="true"></i></button>
		</div>
<?php }else{?>
		<div class="orb-line system" data-line-id="<?php echo htmlspecialchars($_v['line']['id'], ENT_QUOTES);?>
"><span><?php echo htmlspecialchars($_v['line']['body'], ENT_QUOTES);?>
</span><span class="orb-at" data-at="<?php echo htmlspecialchars($_v['line']['created'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['line']['at'], ENT_QUOTES);?>
</span></div>
<?php }?><?php }?>		<p class="faint orb-empty"<?php if($_v['data']['lines']){?>
 hidden<?php }?> data-orb-empty>Nothing said yet.</p>
	</div>
</details>
</section>
