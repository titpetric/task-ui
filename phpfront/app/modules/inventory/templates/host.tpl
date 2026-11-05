{eval $data_head = $data.head}
{include components/page-head.tpl}
<div class="host-detail" data-host-detail="{data.name}">
	<p class="host-facts faint">
		<span><code>{data.host}</code></span>{if $data.login}<span>ssh <code>{data.login}</code></span>{else}<span>no ssh login</span>{/if}{if $data.cpus}<span>{data.cpus} cores</span>{/if}{if $data.up}<span>up {data.uptime}</span>{else}<span class="bad">{data.status}</span>{/if}{if $data.agent}<span>Beszel agent {data.agent}</span>{/if}{if $data.read}<span>read over ssh {data.read}</span>{/if}
	</p>

	<h4 class="section-title">Containers <small>{data.count}{if $data.measured}, measures from Beszel{/if}{if $data.sized}, sizes from docker ps{/if}</small></h4>
{if $data.containers}	<div class="table-wrap">
	<table class="table host-containers" data-sortable>
		<thead><tr>
			<th data-sort="text">Container</th>
			<th data-sort="text">Status</th>
			<th class="num" data-sort="num" aria-sort="descending">CPU</th>
			<th class="num" data-sort="num">Memory</th>
			<th class="num" data-sort="num">Size</th>
			<th class="num" data-sort="num">Net</th>
		</tr></thead>
		<tbody>
{foreach $data.containers as $row}			<tr>
				<td data-v="{row.name}"><span class="key">{row.name}</span>{if $row.image}<span class="faint under">{row.image}</span>{/if}</td>
				<td data-v="{row.status}" class="nowrap{if $row.faint} faint{/if}">{if $row.status}{row.status}{else}<span class="faint">-</span>{/if}</td>
{if $row.measured}				<td class="num" data-v="{row.cpuV}">{row.cpu}%</td>
				<td class="num nowrap" data-v="{row.memoryV}">{row.memory}{if $row.shares}<span class="faint under">{row.share}% of RAM</span>{/if}</td>
{else}				<td class="num faint" data-v="-1" title="{data.measureNote}">-</td>
				<td class="num faint" data-v="-1" title="{data.measureNote}">-</td>
{/if}{if $row.sized}				<td class="num nowrap" data-v="{row.sizeV}">{row.size}<span class="faint under">of {row.virtual}</span></td>
{else}				<td class="num faint" data-v="-1" title="{data.sizeNote}">-</td>
{/if}{if $row.traffic}				<td class="num nowrap" data-v="{row.netV}">{row.net}</td>
{else}				<td class="num faint" data-v="-1" title="{if $row.measured}Beszel reports no traffic of the container{else}{data.measureNote}{/if}">-</td>
{/if}			</tr>
{/foreach}		</tbody>
	</table>
	</div>{if !$data.sized}<p class="faint host-note">No sizes: {data.unsized}.</p>{/if}
{else}	<p class="empty-state">{if $data.known}The host runs no containers.{else}Neither Beszel nor docker ps lists containers on the host: {data.unsized}.{/if}</p>
{/if}
	<h4 class="section-title">Filesystems <small>durable storage{if $data.fromDF}, from df over ssh{elseif $data.filesystems}, the ones Beszel watches{/if}</small></h4>
{if $data.filesystems}	<div class="table-wrap">
	<table class="table host-filesystems">
		<thead><tr><th>Mount</th>{if $data.fromDF}<th>Device</th><th>Type</th>{/if}<th class="num">Size</th><th class="num">Used</th><th class="num">Free</th><th>Use</th></tr></thead>
		<tbody>
{foreach $data.filesystems as $row}			<tr>
				<td class="key"><code>{row.mount}</code></td>{if $data.fromDF}<td class="faint"><code>{row.device}</code></td><td class="faint">{row.type}</td>{/if}
				<td class="num nowrap">{row.size}</td>
				<td class="num nowrap">{row.used}</td>
				<td class="num nowrap">{row.free}</td>
				<td>{eval $data_meter = $row.use}{include components/meter.tpl}			</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{else}	<p class="empty-state">Neither df over ssh nor Beszel reports the host's filesystems{if !$data.login}: task-ui has no ssh login to it, and Beszel no measures{/if}.</p>
{/if}{if $data.charts}	<h4 class="section-title">The last 30 days <small>Beszel's records</small></h4>
	<div class="charts">{foreach $data.charts as $chart}{eval $data_line_chart = $chart}{include components/line-chart.tpl}{/foreach}	</div>
{else}	<p class="empty-state">Beszel holds no records of the host in the last 30 days.</p>
{/if}</div>
