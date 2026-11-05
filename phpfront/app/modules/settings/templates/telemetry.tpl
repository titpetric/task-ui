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
		<label for="telemetry-debug-min">Shortest debug span <span class="faint">a span marked as detail is dropped under this, and counted; <code>0</code> keeps every one</span></label>
		<input type="text" id="telemetry-debug-min" name="debug_span_min_duration" value="{data.opts.debug_span_min_duration}"{if !$data.admin} disabled{/if}>
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
	</div>

	<div class="field">
		<label for="telemetry-refresh">Refresh interval <span class="faint">seconds, where the browser cannot stream; 0 never reloads</span></label>
		<input type="number" id="telemetry-refresh" name="refresh_interval" min="0" step="1" value="{data.opts.refresh_interval}" data-type="int"{if !$data.admin} disabled{/if}>
	</div>

	<div class="field">
		<label for="telemetry-networks">Allowed networks <span class="faint">CIDR ranges, one per line; empty is every network</span></label>
		<textarea id="telemetry-networks" name="allowed_networks" rows="3"{if !$data.admin} disabled{/if}>{data.networks}</textarea>
	</div>

	<div class="field">
		<label for="telemetry-path">Path <span class="faint">where the dashboard is mounted</span></label>
		<input type="text" id="telemetry-path" name="path" value="{data.opts.option_path}"{if !$data.admin} disabled{/if}>
		<p class="faint"><code>{data.path}</code></p>
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
	</div>

	<div class="field">
		<label class="switch-row" for="telemetry-memory">
			<span class="switch">
				<input type="checkbox" id="telemetry-memory" name="track_memory_use" value="1" data-type="toggle"{if $data.opts.track_memory_use} checked{/if}{if !$data.admin} disabled{/if}>
				<i aria-hidden="true"></i>
			</span>
			<span>Track memory use</span>
		</label>
	</div>

	<div class="field">
		<label class="switch-row" for="telemetry-request-id">
			<span class="switch">
				<input type="checkbox" id="telemetry-request-id" name="trust_request_id" value="1" data-type="toggle"{if $data.opts.trust_request_id} checked{/if}{if !$data.admin} disabled{/if}>
				<i aria-hidden="true"></i>
			</span>
			<span>Trust Request-Id</span>
		</label>
	</div>
{if $data.admin}	<div class="form-foot">
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
{/if}</section>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
