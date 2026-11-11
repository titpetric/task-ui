<figure class="chart">
	<figcaption class="chart-title">{data_line_chart.title}</figcaption>
{if $data_line_chart.empty}	<p class="faint chart-empty">No records in the last 30 days.</p>
{else}	{data_line_chart.svg|raw}
	<table class="chart-legend">
		<thead><tr><th></th><th>min</th><th>avg</th><th>max</th></tr></thead>
		<tbody>
{foreach $data_line_chart.legend as $row}			<tr><td><i class="swatch s{row.color}"></i>{row.name}</td><td>{row.min}</td><td>{row.avg}</td><td>{row.max}</td></tr>
{/foreach}{if $data_line_chart.total}			<tr class="total"><td>{data_line_chart.total.name}</td><td>{data_line_chart.total.min}</td><td>{data_line_chart.total.avg}</td><td>{data_line_chart.total.max}</td></tr>
{/if}		</tbody>
	</table>
{/if}</figure>
