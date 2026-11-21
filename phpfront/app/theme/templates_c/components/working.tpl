<?php $_v=&$this->vars;?>
<div class="said-live" data-status data-state="<?php echo htmlspecialchars($_v['data_working']['state'], ENT_QUOTES);?>
" data-started="<?php echo htmlspecialchars($_v['data_working']['started'], ENT_QUOTES);?>
">
	<p class="said-state" aria-live="polite"><i class="dot working"></i> <span data-status-text><?php echo htmlspecialchars($_v['data_working']['label'], ENT_QUOTES);?>
</span> <span class="faint" data-status-elapsed></span><?php if($_v['data_working']['canEnd']){?>
<span class="actions"><button class="btn-icon" type="button" data-stop="<?php echo htmlspecialchars($_v['data_working']['run'], ENT_QUOTES);?>
" title="Stop this run and let it settle" aria-label="Stop this run"><i class="ic ic-stop" aria-hidden="true"></i></button><button class="btn-icon danger" type="button" data-kill="<?php echo htmlspecialchars($_v['data_working']['run'], ENT_QUOTES);?>
" title="Kill this run and everything it started" aria-label="Kill this run and everything it started"><i class="ic ic-trash" aria-hidden="true"></i></button></span><?php }?></p>
	<div class="<?php echo htmlspecialchars($_v['data_working']['progressClass'], ENT_QUOTES);?>
" data-work-progress><?php if($_v['data_working']['progress']){ echo $_v['data_working']['progress']; }?></div>
<?php if($_v['data_working']['agents']){?>
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
