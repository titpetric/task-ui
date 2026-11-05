{eval $data_head = $data.head}
{include components/page-head.tpl}

{if $data.terminal}{eval $data_terminal = $data.terminal}{include components/terminal.tpl}
{/if}{if $data.output.expired}<section class="block">
	<h3>Output <small>the recording expired</small></h3>
	<p class="dim">{data.output.text} <a href="{base}/settings/retention">Data retention</a></p>
</section>
{elseif $data.output.recording}<section class="block">
	<h3>Output <small>every line the run printed</small></h3>
	<pre class="output" data-output="{data.output.id}" data-cols="{data.output.cols}" data-rows="{data.output.rows}"><span class="faint">Reading the recording</span></pre>
</section>
{/if}{if $data.prompt}{eval $data_prompt = $data.prompt}{include components/prompt.tpl}
{/if}
<div class="grid-2">
<section class="block">
	<h3>Run</h3>
{eval $data_kv = $data.kv}{include components/kv.tpl}
</section>

<section class="block">
{eval $data_block = $data.artifacts.block}{include components/block-head.tpl}
{if $data.artifacts.rows}	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Name</th><th>Size</th><th>Type</th><th></th></tr></thead>
		<tbody>
{foreach $data.artifacts.rows as $file}			<tr>
				<td class="key"><a href="{file.href}">{file.name}</a></td>
				<td class="nowrap">{file.size}</td>
				<td class="dim">{file.type}</td>
				<td class="actions">{eval $data_menu = $file.menu}{include components/menu.tpl}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}</section>
</div>
