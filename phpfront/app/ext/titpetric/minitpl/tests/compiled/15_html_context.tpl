<?php $_v=&$this->vars;?>
<!-- comment context: <?php echo htmlspecialchars($_v['comment'], ENT_QUOTES);?>
 -->
<p class="<?php echo htmlspecialchars($_v['class'], ENT_QUOTES);?>
" data-id='<?php echo htmlspecialchars($_v['id'], ENT_QUOTES);?>
' title=<?php echo htmlspecialchars($_v['bare'], ENT_QUOTES);?>
><?php echo htmlspecialchars($_v['text'], ENT_QUOTES);?>
</p>
<a href="<?php echo htmlspecialchars($_v['url'], ENT_QUOTES);?>
"><?php echo $_v['label'];?>
</a>
<script type="text/javascript">
var config = <?php echo $_v['json'];?>
;
</script>
<style>
.theme { color: <?php echo $_v['color'];?>
; }
</style>
<textarea><?php echo htmlspecialchars($_v['text'], ENT_QUOTES);?>
</textarea>
