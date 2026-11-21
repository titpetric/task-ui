<?php $_v=&$this->vars; if($_v['data_status']['working']){?>
<div class="conversation-status" data-status data-state="<?php echo htmlspecialchars($_v['data_status']['state'], ENT_QUOTES);?>
" data-started="<?php echo htmlspecialchars($_v['data_status']['started'], ENT_QUOTES);?>
">
	<p aria-live="polite"><i class="dot working"></i> <span data-status-text><?php echo htmlspecialchars($_v['data_status']['label'], ENT_QUOTES);?>
</span> <span class="faint" data-status-elapsed></span></p>
<?php if($_v['data_status']['agents']){?>
	<ul class="status-agents" data-status-agents>
<?php if(!empty($_v['data_status']['agents']))foreach($_v['data_status']['agents'] as $_v['agent']){?>		<li><details class="status-agent <?php echo htmlspecialchars($_v['agent']['state'], ENT_QUOTES);?>
" data-agent="<?php echo htmlspecialchars($_v['agent']['id'], ENT_QUOTES);?>
">
			<summary><i class="ic ic-users" aria-hidden="true"></i> <span><?php echo htmlspecialchars($_v['agent']['task'], ENT_QUOTES);?>
</span><?php if($_v['agent']['kind']){?>
 <small class="faint"><?php echo htmlspecialchars($_v['agent']['kind'], ENT_QUOTES);?>
</small><?php }?> <span class="badge <?php if($_v['agent']['state'] == 'working'){?>
signal<?php }elseif($_v['agent']['state'] == 'failed'){?>
bad<?php }else{?>
ok<?php }?>" data-status-agent-state><?php echo htmlspecialchars($_v['agent']['state'], ENT_QUOTES);?>
</span></summary>
<?php if($_v['agent']['prompt']){?>
			<pre class="status-agent-prompt"><?php echo htmlspecialchars($_v['agent']['prompt'], ENT_QUOTES);?>
</pre>
<?php }?><?php if($_v['agent']['result']){?>
			<pre class="status-agent-result"><?php echo htmlspecialchars($_v['agent']['result'], ENT_QUOTES);?>
</pre>
<?php }?>		</details></li>
<?php }?>	</ul>
<?php }?><?php if($_v['data_status']['can_end']){?>
	<div class="actions">
		<button class="btn-icon" type="button" data-stop="<?php echo htmlspecialchars($_v['data_status']['run'], ENT_QUOTES);?>
" title="Stop this run and let it settle" aria-label="Stop this run"><i class="ic ic-stop" aria-hidden="true"></i></button>
		<button class="btn-icon danger" type="button" data-kill="<?php echo htmlspecialchars($_v['data_status']['run'], ENT_QUOTES);?>
" title="Kill this run and everything it started" aria-label="Kill this run and everything it started"><i class="ic ic-trash" aria-hidden="true"></i></button>
	</div>
<?php }?></div>
<?php }?>