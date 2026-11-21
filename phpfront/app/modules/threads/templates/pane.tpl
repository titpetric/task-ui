<div class="conversation">
<a class="conversation-back" href="{base}/threads"><i class="ic ic-back" aria-hidden="true"></i> All threads</a>
{eval $data_head = $data_pane.head_view}{include components/conversation-head.tpl}
{eval $data_status = $data_pane.status}{include components/status.tpl}

	<ol class="talk" data-talk data-thread="{data_pane.id}" data-last-event="{data_pane.last_event}"{if $data_pane.chat_thread} data-chat-thread="{data_pane.id}"{/if}>
{foreach $data_pane.messages as $message}		{eval $data_message = $message}{include components/message.tpl}
{/foreach}	</ol>

{if $data_pane.replies}{eval $data_composer = $data_pane.composer}{include components/composer.tpl}
{/if}<button type="button" class="scroll-latest" data-scroll-latest title="Scroll to the latest" aria-label="Scroll to the latest" hidden><i class="ic ic-down" aria-hidden="true"></i></button></div>
