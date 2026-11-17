<?php $_v=&$this->vars;?>
	<footer>
		<span class="credit">
			<a class="wordmark" href="https://github.com/titpetric/task-ui">task-ui</a>
			<span class="sep" aria-hidden="true">|</span>
			<span>pipelines by <a href="https://github.com/titpetric">@titpetric</a></span>
			<span class="sep" aria-hidden="true">|</span>
			<a href="https://blacksky.digital/products/task-ui">blacksky.digital</a>
		</span>
	</footer>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/ext/xterm.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/ext/addon-fit.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/mdedit/mdedit.global.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/task-ui.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/voice/orbs.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
	<script src="<?php echo htmlspecialchars($_v['data']['assets'], ENT_QUOTES);?>
/theme/voice/voice.js?v=<?php echo htmlspecialchars($_v['data']['assetVersion'], ENT_QUOTES);?>
" defer></script>
<?php if(!empty($_v['data']['scripts']))foreach($_v['data']['scripts'] as $_v['script']){?>	<script src="<?php echo htmlspecialchars($_v['script'], ENT_QUOTES);?>
" defer></script>
<?php }?>
