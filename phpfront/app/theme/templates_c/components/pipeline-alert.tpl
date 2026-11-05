<?php $_v=&$this->vars;?>
<div class="alert bad pipeline-alert" role="alert">
	<b><?php echo htmlspecialchars($_v['data_findings']['pipelines'], ENT_QUOTES);?>
 <?php echo htmlspecialchars($_v['data_findings']['says'], ENT_QUOTES);?>
 errors.</b> atkins lists no jobs of a pipeline its linter refuses; fix the file and they return.
	<ul class="pipeline-errors">
<?php if(!empty($_v['data_findings']['shown']))foreach($_v['data_findings']['shown'] as $_v['finding']){?><?php $_v['data_finding'] = $_v['finding'];?>
<li><a href="<?php echo htmlspecialchars($_v['data_finding']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_finding']['folder'], ENT_QUOTES);?>
</a> <code><?php echo htmlspecialchars($_v['data_finding']['path'], ENT_QUOTES);?>
</code><?php if($_v['data_finding']['job']){?>
 <span class="dim">job</span> <b><?php echo htmlspecialchars($_v['data_finding']['job'], ENT_QUOTES);?>
</b><?php }?>: <span<?php if($_v['data_finding']['issue']){?>
 title="<?php echo htmlspecialchars($_v['data_finding']['issue'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_finding']['detail'], ENT_QUOTES);?>
</span></li>

<?php }?>	</ul>
<?php if($_v['data_findings']['rest']){?>
	<details class="pipeline-more">
		<summary><svg class="ic chev"><use href="#i-chev"/></svg>Show <?php echo htmlspecialchars($_v['data_findings']['rest'], ENT_QUOTES);?>
 more</summary>
		<ul class="pipeline-errors">
<?php if(!empty($_v['data_findings']['more']))foreach($_v['data_findings']['more'] as $_v['finding']){?><?php $_v['data_finding'] = $_v['finding'];?>
<li><a href="<?php echo htmlspecialchars($_v['data_finding']['href'], ENT_QUOTES);?>
"><?php echo htmlspecialchars($_v['data_finding']['folder'], ENT_QUOTES);?>
</a> <code><?php echo htmlspecialchars($_v['data_finding']['path'], ENT_QUOTES);?>
</code><?php if($_v['data_finding']['job']){?>
 <span class="dim">job</span> <b><?php echo htmlspecialchars($_v['data_finding']['job'], ENT_QUOTES);?>
</b><?php }?>: <span<?php if($_v['data_finding']['issue']){?>
 title="<?php echo htmlspecialchars($_v['data_finding']['issue'], ENT_QUOTES);?>
"<?php }?>><?php echo htmlspecialchars($_v['data_finding']['detail'], ENT_QUOTES);?>
</span></li>

<?php }?>		</ul>
	</details>
<?php }?></div>
