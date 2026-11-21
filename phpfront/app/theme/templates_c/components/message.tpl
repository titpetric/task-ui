<?php $_v=&$this->vars;?>
<li class="<?php echo htmlspecialchars($_v['data_message']['className'], ENT_QUOTES);?>
" id="note-<?php echo htmlspecialchars($_v['data_message']['id'], ENT_QUOTES);?>
" data-run-id="<?php echo htmlspecialchars($_v['data_message']['run'], ENT_QUOTES);?>
">
	<header>
		<span class="author"><?php $_v['data_user'] = $_v['data_message']['author'];?>
<span class="user-ref"><?php if($_v['data_user']['href']){?>
<a class="user-link" href="<?php echo htmlspecialchars($_v['data_user']['href'], ENT_QUOTES);?>
"><?php $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span></a><?php }else{ $_v['data_avatar'] = $_v['data_user'];?>
<span class="avatar sm<?php if($_v['data_avatar']['role']){?>
 role-<?php echo htmlspecialchars($_v['data_avatar']['role'], ENT_QUOTES); }?>" title="<?php echo htmlspecialchars($_v['data_avatar']['name'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_avatar']['initials'], ENT_QUOTES);?>
</span> <span class="user-name"><?php echo htmlspecialchars($_v['data_user']['name'], ENT_QUOTES);?>
</span><?php }?></span>
</span>
		<span class="faint" title="<?php echo htmlspecialchars($_v['data_message']['stamp'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_message']['ago'], ENT_QUOTES);?>
</span>
		<?php if($_v['data_message']['agent']){?>
			<span class="badge agent" title="The agent's answer">agent</span>
<?php }?><?php if($_v['data_message']['state']){?>
			<span class="badge note-state <?php echo htmlspecialchars($_v['data_message']['state'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_message']['state'], ENT_QUOTES);?>
</span>
<?php }?><?php if($_v['data_message']['run']){?>
			<a class="faint mono" href="<?php echo htmlspecialchars($_v['base'], ENT_QUOTES);?>
/runs/<?php echo htmlspecialchars($_v['data_message']['run'], ENT_QUOTES);?>
" title="The run this message comes from">run <?php echo htmlspecialchars($_v['data_message']['short'], ENT_QUOTES);?>
</a>
<?php }?>	</header>
<?php if($_v['data_message']['body']){?>
	<div class="said-body prose"><?php echo $_v['data_message']['body'];?>
</div>
<?php }?><?php if($_v['data_message']['working']){ $_v['data_working'] = $_v['data_message']['working_view'];?>
<div class="said-live" data-status data-state="<?php echo htmlspecialchars($_v['data_working']['state'], ENT_QUOTES);?>
" data-started="<?php echo htmlspecialchars($_v['data_working']['started'], ENT_QUOTES);?>
">
	<div class="<?php echo htmlspecialchars($_v['data_working']['progressClass'], ENT_QUOTES);?>
" data-work-progress><?php if($_v['data_working']['progressHtml']){ echo $_v['data_working']['progressHtml']; }?></div>
	<p class="said-state" aria-live="polite"><i class="dot working"></i> <span data-status-text><?php echo htmlspecialchars($_v['data_working']['label'], ENT_QUOTES);?>
</span> <span class="faint" data-status-elapsed></span><?php if($_v['data_working']['canEnd']){?>
<span class="actions"><button class="btn-icon" type="button" data-stop="<?php echo htmlspecialchars($_v['data_working']['run'], ENT_QUOTES);?>
" title="Stop this run and let it settle" aria-label="Stop this run"><i class="ic ic-stop" aria-hidden="true"></i></button><button class="btn-icon danger" type="button" data-kill="<?php echo htmlspecialchars($_v['data_working']['run'], ENT_QUOTES);?>
" title="Kill this run and everything it started" aria-label="Kill this run and everything it started"><i class="ic ic-trash" aria-hidden="true"></i></button></span><?php }?></p>
<?php if($_v['data_working']['todos']){?>
	<ul class="status-todos" data-status-todos>
<?php if(!empty($_v['data_working']['todos']))foreach($_v['data_working']['todos'] as $_v['todo']){?>		<li class="todo <?php echo htmlspecialchars($_v['todo']['state'], ENT_QUOTES);?>
"><i class="<?php echo htmlspecialchars($_v['todo']['className'], ENT_QUOTES);?>
" aria-hidden="true"></i> <span><?php echo htmlspecialchars($_v['todo']['text'], ENT_QUOTES);?>
</span></li>
<?php }?>	</ul>
<?php }?><?php if($_v['data_working']['agents']){?>
	<ul class="status-agents" data-status-agents>
<?php if(!empty($_v['data_working']['agents']))foreach($_v['data_working']['agents'] as $_v['agent']){?>		<li><details class="<?php echo htmlspecialchars($_v['agent']['className'], ENT_QUOTES);?>
" data-agent="<?php echo htmlspecialchars($_v['agent']['id'], ENT_QUOTES);?>
">
			<summary><i class="ic ic-users" aria-hidden="true"></i> <span><?php echo htmlspecialchars($_v['agent']['task'], ENT_QUOTES);?>
</span><?php if($_v['agent']['kind']){?>
 <small class="faint"><?php echo htmlspecialchars($_v['agent']['kind'], ENT_QUOTES);?>
</small><?php }?> <span class="<?php echo htmlspecialchars($_v['agent']['badge'], ENT_QUOTES);?>
" data-status-agent-state><?php echo htmlspecialchars($_v['agent']['state'], ENT_QUOTES);?>
</span></summary>
<?php if($_v['agent']['prompt']){?>
			<pre class="status-agent-prompt"><?php echo htmlspecialchars($_v['agent']['prompt'], ENT_QUOTES);?>
</pre>
<?php }?><?php if($_v['agent']['result']){?>
			<pre class="status-agent-result"><?php echo htmlspecialchars($_v['agent']['result'], ENT_QUOTES);?>
</pre>
<?php }?>		</details></li>
<?php }?>	</ul>
<?php }?>	<p class="said-live-foot"><span class="work-stats" data-work-stats data-tools="<?php echo htmlspecialchars($_v['data_working']['tools'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_working']['stats'], ENT_QUOTES);?>
</span></p>
</div>

<?php }else{ if($_v['data_message']['agent']){ if(!empty($_v['data_message']['work']))foreach($_v['data_message']['work'] as $_v['run']){?><?php $_v['data_work'] = $_v['run']; if($_v['data_work']['stats'] || $_v['data_work']['trace']['states']){?>
<div class="work-run" data-run-id="<?php echo htmlspecialchars($_v['data_work']['id'], ENT_QUOTES);?>
">
<?php if($_v['data_work']['trace']['states']){?>
	<button type="button" class="work-trace" data-trace data-mode="duration" data-ms="<?php echo htmlspecialchars($_v['data_work']['trace']['ms'], ENT_QUOTES);?>
" data-count="<?php echo htmlspecialchars($_v['data_work']['trace']['count'], ENT_QUOTES);?>
" title="Switch the bar between duration, count and time">
		<span class="work-trace-mode" data-trace-mode>Duration</span>
		<span class="work-trace-bar">
<?php if(!empty($_v['data_work']['trace']['states']))foreach($_v['data_work']['trace']['states'] as $_v['state']){?><span class="work-trace-group <?php echo htmlspecialchars($_v['state']['key'], ENT_QUOTES);?>
" style="width:<?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
%" data-share="<?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
" data-count="<?php echo htmlspecialchars($_v['state']['count'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['state']['label'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['state']['label'], ENT_QUOTES);?>
: <?php echo htmlspecialchars($_v['state']['share'], ENT_QUOTES);?>
%"></span>
<?php }?><?php if(!empty($_v['data_work']['trace']['segments']))foreach($_v['data_work']['trace']['segments'] as $_v['segment']){?><span class="work-trace-event <?php echo htmlspecialchars($_v['segment']['key'], ENT_QUOTES);?>
" data-ms="<?php echo htmlspecialchars($_v['segment']['ms'], ENT_QUOTES);?>
" data-label="<?php echo htmlspecialchars($_v['segment']['label'], ENT_QUOTES);?>
" data-tip="<?php echo htmlspecialchars($_v['segment']['label'], ENT_QUOTES);?>
: <?php echo htmlspecialchars($_v['segment']['span'], ENT_QUOTES);?>
"></span>
<?php }?>		</span>
	</button>
<?php }?><?php if($_v['data_work']['stats']){?>
	<div class="work-stats"><?php echo htmlspecialchars($_v['data_work']['stats'], ENT_QUOTES);?>
</div>
<?php }?></div><?php }}?><?php }?><?php }?><?php if($_v['data_message']['audio']){?>
	<audio class="said-audio" controls preload="none" src="<?php echo htmlspecialchars($_v['data_message']['audio'], ENT_QUOTES);?>
"></audio>
<?php }?></li>
