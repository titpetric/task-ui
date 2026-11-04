{* The telemetry settings page. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
<form class="form telemetry" {if $data.admin}data-api="/api/telemetry" data-method="PUT" data-next="{data.next}"{/if} novalidate>
<section class="block">
	<h3>Recording <small>what a request leaves behind</small></h3>

	<div class="field">
		<label class="switch-row" for="telemetry-enabled">
			<span class="switch">
				<input type="checkbox" id="telemetry-enabled" name="enabled" value="1" data-type="toggle"{if $data.opts.enabled} checked{/if}{if !$data.admin} disabled{/if}>
				<i aria-hidden="true"></i>
			</span>
			<span>Enabled</span>
		</label>
		<p class="faint">Off, a request passes through and records nothing. This is the one setting that takes effect on its own: everything else below rebuilds the recorder, which empties the traces it holds.</p>
	</div>

	<div class="field">
		<label for="telemetry-service">Service name <span class="faint">on every trace, and in the dashboard's heading</span></label>
		<input type="text" id="telemetry-service" name="service_name" value="{data.opts.service_name}"{if !$data.admin} disabled{/if}>
	</div>

	<div class="row">
		<div class="field">
			<label for="telemetry-rate">Sample rate <span class="faint">percent of requests traced</span></label>
			<input type="number" id="telemetry-rate" name="sample_rate" min="0" max="100" step="1" value="{data.rate}" data-type="int"{if !$data.admin} disabled{/if}>
		</div>
		<div class="field">
			<label for="telemetry-ring">Retained traces <span class="faint">the oldest go first</span></label>
			<input type="number" id="telemetry-ring" name="ring_buffer_size" min="1" step="1" value="{data.opts.ring_buffer_size}" data-type="int"{if !$data.admin} disabled{/if}>
		</div>
	</div>

	<div class="row">
		<div class="field">
			<label for="telemetry-top">Top requests <span class="faint">groups the statistics keep</span></label>
			<input type="number" id="telemetry-top" name="top_requests" min="1" step="1" value="{data.opts.top_requests}" data-type="int"{if !$data.admin} disabled{/if}>
		</div>
		<div class="field">
			<label for="telemetry-spans">Spans per trace <span class="faint">the rest are counted, not kept; 0 is no cap</span></label>
			<input type="number" id="telemetry-spans" name="max_spans_per_trace" min="0" step="1" value="{data.opts.max_spans_per_trace}" data-type="int"{if !$data.admin} disabled{/if}>
		</div>
	</div>

	<div class="field">
		<label for="telemetry-ignore">Ignored paths <span class="faint">one per line; an entry ending in <code>/*</code> matches by prefix</span></label>
		<textarea id="telemetry-ignore" name="ignore_paths" rows="4"{if !$data.admin} disabled{/if}>{data.ignore}</textarea>
	</div>
</section>

<section class="block">
	<h3>The dashboard <small>how the live view reads, and who may open it</small></h3>

	<div class="field">
		<label class="switch-row" for="telemetry-stream">
			<span class="switch">
				<input type="checkbox" id="telemetry-stream" name="live_stream" value="1" data-type="toggle"{if $data.opts.live_stream} checked{/if}{if !$data.admin} disabled{/if}>
				<i aria-hidden="true"></i>
			</span>
			<span>Live stream</span>
		</label>
		<p class="faint">Serves the live view over server sent events, so a trace appears as it happens instead of on a timer.</p>
	</div>

	<div class="field">
		<label for="telemetry-refresh">Refresh interval <span class="faint">seconds, where the browser cannot stream; 0 never reloads</span></label>
		<input type="number" id="telemetry-refresh" name="refresh_interval" min="0" step="1" value="{data.opts.refresh_interval}" data-type="int"{if !$data.admin} disabled{/if}>
	</div>

	<div class="field">
		<label for="telemetry-networks">Allowed networks <span class="faint">CIDR ranges, one per line; empty is every network</span></label>
		<textarea id="telemetry-networks" name="allowed_networks" rows="3"{if !$data.admin} disabled{/if}>{data.networks}</textarea>
		<p class="faint">The dashboard is an admin's either way: a trace holds request paths, query strings and the errors a handler recorded, and a request from outside these ranges gets a 404 rather than a refusal.</p>
	</div>

	<div class="field">
		<label for="telemetry-path">Path <span class="faint">where the dashboard is mounted</span></label>
		<input type="text" id="telemetry-path" name="path" value="{data.opts.option_path}"{if !$data.admin} disabled{/if}>
		<p class="faint">The router reads this once, at startup, so a path saved here takes a restart. It is <code>{data.path}</code> now.</p>
	</div>
</section>

<section class="block">
	<h3>What a trace holds <small>beside the spans</small></h3>

	<div class="field">
		<label class="switch-row" for="telemetry-logs">
			<span class="switch">
				<input type="checkbox" id="telemetry-logs" name="capture_logs" value="1" data-type="toggle"{if $data.opts.capture_logs} checked{/if}{if !$data.admin} disabled{/if}>
				<i aria-hidden="true"></i>
			</span>
			<span>Capture logs</span>
		</label>
		<p class="faint">Records the log lines a trace writes. Off, an error still lands on its span.</p>
	</div>

	<div class="field">
		<label class="switch-row" for="telemetry-memory">
			<span class="switch">
				<input type="checkbox" id="telemetry-memory" name="track_memory_use" value="1" data-type="toggle"{if $data.opts.track_memory_use} checked{/if}{if !$data.admin} disabled{/if}>
				<i aria-hidden="true"></i>
			</span>
			<span>Track memory use</span>
		</label>
		<p class="faint">Records the process's allocation change per trace.</p>
	</div>

	<div class="field">
		<label class="switch-row" for="telemetry-request-id">
			<span class="switch">
				<input type="checkbox" id="telemetry-request-id" name="trust_request_id" value="1" data-type="toggle"{if $data.opts.trust_request_id} checked{/if}{if !$data.admin} disabled{/if}>
				<i aria-hidden="true"></i>
			</span>
			<span>Trust Request-Id</span>
		</label>
		<p class="faint">Reuses a client's <code>Request-Id</code> header as the trace identifier, so a request crossing two services reads as one trace. Behind a trusted proxy alone: a client that picks its own identifier can collide with another's.</p>
	</div>
{if $data.admin}	<div class="form-foot">
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
{/if}</section>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
