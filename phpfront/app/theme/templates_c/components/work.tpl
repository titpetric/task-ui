<?php $_v=&$this->vars;?>
<div class="work-run" data-run="<?php echo htmlspecialchars($_v['data_work']['id'], ENT_QUOTES);?>
">
<?php if($_v['data_work']['stats']){?>
	<div class="work-stats"><?php echo htmlspecialchars($_v['data_work']['stats'], ENT_QUOTES);?>
</div>
<?php }?><?php if(!empty($_v['data_work']['events']))foreach($_v['data_work']['events'] as $_v['event']){?><?php if($_v['event']['kind'] == 'text' || $_v['event']['kind'] == 'result'){?>
	<div class="work-text prose" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><?php echo $_v['event']['prose'];?>
</div>
<?php }elseif($_v['event']['kind'] == 'thinking'){?>
	<details class="work-thinking" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary>Thinking</summary><pre><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'tool_use'){?>
	<details class="work-tool" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary><code><?php echo htmlspecialchars($_v['event']['tool'], ENT_QUOTES);?>
</code><?php if($_v['event']['body']){?>
 <span class="faint"><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</span><?php }?></summary><pre><?php echo htmlspecialchars($_v['event']['input'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'tool_result'){?>
	<details class="work-result<?php if($_v['event']['error']){?>
 error<?php }?>" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary><?php if($_v['event']['error']){?>
Error<?php }else{?>
Result<?php }?></summary><pre><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'system'){?>
	<details class="work-system" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><summary>Session</summary><pre><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</pre></details>
<?php }elseif($_v['event']['kind'] == 'error'){?>
	<div class="work-error" data-seq="<?php echo htmlspecialchars($_v['event']['seq'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['event']['body'], ENT_QUOTES);?>
</div>
<?php }?><?php }?></div>
