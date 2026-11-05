<section class="block" data-pipeline-view>
	<div class="page-head">
		<div>
			<h2>{data.flow.name}</h2>
			<p class="faint"><span class="state"><i class="dot {data.run.state}"></i>{data.run.state}</span>{if $data.run.message}: {data.run.message};{/if} started <span title="{data.run.stamp}">{data.run.ago}</span> by {data.run.userName}{if $data.run.finished}, took {data.run.took}{/if}</p>
		</div>
	</div>
{eval $data_thread = $data.thread}
{include components/wfthread.tpl}
	<h3 class="wf-steps-title">Steps</h3>
	<ol class="pipeline">
{foreach $data.pipeline as $data_node}{include components/pipenode.tpl}{/foreach}	</ol>
{if $data.prompt}{eval $data_prompt = $data.prompt}
{include components/prompt.tpl}
{/if}</section>
