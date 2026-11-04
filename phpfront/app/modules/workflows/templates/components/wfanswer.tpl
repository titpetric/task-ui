{* input scope: data_wfanswer. See wfanswer.md. *}{if $data_wfanswer.body}			<tr class="wf-answer-row" id="wf-answer-{data_wfanswer.id}" data-answer-of="{data_wfanswer.id}"{if !$data_wfanswer.open} hidden{/if}><td colspan="{data_wfanswer.cols}"><div class="note-body">{data_wfanswer.body|raw}</div></td></tr>
{/if}
