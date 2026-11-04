{* The settings index. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
<section class="block">
	<h3>Steps</h3>
	<div class="table-wrap">
	<table class="table settings">
		<tbody>
{foreach $data.steps as $step}			<tr>
				<td class="num dim">{step.number}</td>
				<td class="key nowrap"><a href="{step.href}"><svg class="ic"><use href="#i-{step.icon}"/></svg> {step.name}</a></td>
				<td class="desc">{step.desc}</td>
				<td class="count dim"><span class="nowrap">{step.count}</span></td>
				<td class="actions"><a class="btn btn-sm" href="{step.href}">Open</a></td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
</section>
