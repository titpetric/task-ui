<?php $_v=&$this->vars; if($_v['data_wfanswer']['body']){?>
			<tr class="wf-answer-row" id="wf-answer-<?php echo htmlspecialchars($_v['data_wfanswer']['id'], ENT_QUOTES);?>
" data-answer-of="<?php echo htmlspecialchars($_v['data_wfanswer']['id'], ENT_QUOTES);?>
"<?php if(!$_v['data_wfanswer']['open']){?>
 hidden<?php }?>><td colspan="<?php echo htmlspecialchars($_v['data_wfanswer']['cols'], ENT_QUOTES);?>
"><div class="note-body"><?php echo $_v['data_wfanswer']['body'];?>
</div></td></tr>
<?php }?>
