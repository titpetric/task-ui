<div class="spacebar" role="img" aria-label="{data_space.total} in all: {data_space.files} of files, {data_space.history} of history">
	<div class="spacebar-track"><i class="files" style="width: {data_space.filesPercent}%"></i><i class="history" style="width: {data_space.historyPercent}%"></i></div>
	<div class="spacebar-legend">
		<span><i class="swatch files"></i>Files <b>{data_space.files}</b></span>
		<span><i class="swatch history"></i>History <b>{data_space.history}</b></span>
		<span class="faint">{data_space.total} in all{if $data_space.ignored}, and {data_space.ignored} git ignores{/if}</span>
	</div>
</div>
