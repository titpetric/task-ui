	<footer>
		<span class="credit">
			<a class="wordmark" href="https://github.com/titpetric/task-ui">task-ui</a>
			<span class="sep" aria-hidden="true">|</span>
			<span>pipelines by <a href="https://github.com/titpetric">@titpetric</a></span>
			<span class="sep" aria-hidden="true">|</span>
			<a href="https://blacksky.digital/products/task-ui">blacksky.digital</a>
		</span>
	</footer>
	<script src="{data.assets}/theme/vendor/xterm.js?v={data.assetVersion}" defer></script>
	<script src="{data.assets}/theme/vendor/addon-fit.js?v={data.assetVersion}" defer></script>
	<script src="{data.assets}/theme/task-ui.js?v={data.assetVersion}" defer></script>
	<script src="{data.assets}/theme/voice/orbs.js?v={data.assetVersion}" defer></script>
	<script src="{data.assets}/theme/voice/voice.js?v={data.assetVersion}" defer></script>
{foreach $data.scripts as $script}	<script src="{script}" defer></script>
{/foreach}
