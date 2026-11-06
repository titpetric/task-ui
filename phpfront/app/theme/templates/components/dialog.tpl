	<dialog class="run" id="run-dialog" aria-labelledby="run-title">
		<div class="run-head">
			<span class="label" id="run-kind">Run</span>
			<h3 id="run-title"></h3>
			<span class="run-subtitle" id="run-subtitle" hidden></span>
			<span class="tag rec" id="run-rec" hidden>Recording</span>
			<span class="tag input-on" id="run-input" hidden>Input on</span>
			<span class="spacer"></span>
			<span class="dim" id="run-who"></span>
			<span class="elapsed" id="run-elapsed"></span>
			<span class="state" id="run-state"></span>
			<button type="button" class="btn btn-sm" id="run-replay" title="Replay the recording" aria-label="Replay the recording" hidden>|&gt;</button>
			<a class="btn btn-sm" id="run-page" href="{base}/runs" hidden>Open run page</a>
		</div>
		<div class="run-body">
			<div class="terminal-host" id="run-terminal"></div>
		</div>
		<p class="run-notice" id="run-notice" role="status"></p>
		<div class="run-reason" id="run-reason" hidden>
			<label for="run-reason-text">Why does this need to run outside its hours? task-ui saves your reason with the run.</label>
			<input type="text" id="run-reason-text" placeholder="customer is waiting on the hotfix" autocomplete="off" maxlength="200">
		</div>
		<div class="run-foot">
			<nav class="run-chain" id="run-chain" aria-label="The runs of the chain" hidden>
				<button type="button" class="btn btn-sm" id="run-prev" title="The run before this one">‹ Previous</button>
				<span class="run-pos" id="run-pos"></span>
				<button type="button" class="btn btn-sm" id="run-next" title="The run after this one">Next ›</button>
			</nav>
			<a class="btn btn-sm" id="run-report" href="#" title="Claude's answer or summary, as a document" hidden>Report</a>
			<button type="button" class="btn" id="run-close" aria-label="Close">Close</button>
			<span class="hint" id="run-hint"></span>
			<button type="button" class="btn" id="run-attach" hidden>Attach</button>
			<button type="button" class="btn" id="run-again" hidden>Run again</button>
			<button type="button" class="btn" id="run-take" aria-pressed="false" title="Send your keystrokes to the run, to answer a prompt it waits on" hidden>Take input</button>
			<button type="button" class="btn-icon" id="run-mic" aria-haspopup="dialog" title="Dictate into the run" aria-label="Dictate into the run" hidden><svg class="ic"><use href="#i-mic"/></svg></button>
			<button type="button" class="btn" id="run-stop" hidden>Stop</button>
			<button type="button" class="btn btn-danger" id="run-kill" hidden>Kill</button>
			<button type="button" class="btn" id="run-pause" hidden>Pause</button>
			<button type="button" class="btn" id="run-restart" hidden>Restart</button>
			<button type="button" class="btn btn-primary" id="run-force" hidden>Run force</button>
		</div>
	</dialog>
	<dialog class="run preview" id="preview-dialog" aria-labelledby="preview-title">
		<div class="run-head">
			<span class="label" id="preview-label">Artifact</span>
			<h3 id="preview-title"></h3>
		</div>
		<div class="run-body preview-body"><img id="preview-image" alt=""><div id="preview-doc" class="preview-doc" hidden></div><iframe id="preview-frame" class="preview-frame" title="The page the run wrote" sandbox referrerpolicy="no-referrer" hidden></iframe></div>
		<div class="run-foot">
			<span class="hint"><select id="preview-version" aria-label="Version" hidden></select></span>
			<a class="btn" id="preview-download" href="#" download>Download</a>
			<span class="preview-actions" id="preview-actions"></span>
			<button type="button" class="btn" id="preview-edit" hidden>Edit</button>
			<button type="button" class="btn btn-primary" id="preview-save" hidden>Save</button>
			<button type="button" class="btn" id="preview-close" aria-label="Close">Close</button>
		</div>
	</dialog>
	<dialog class="run pipeline-window" id="pipeline-dialog" aria-labelledby="pipeline-title">
		<div class="run-head">
			<span class="label">Workflow run</span>
			<h3 id="pipeline-title"></h3>
		</div>
		<div class="run-body pipeline-body" id="pipeline-body"></div>
		<div class="run-foot">
			<span class="hint"></span>
			<a class="btn" id="pipeline-page" href="#">Open page</a>
			<button type="button" class="btn" id="pipeline-close" aria-label="Close">Close</button>
		</div>
	</dialog>
	<dialog class="run preview run-summary" id="summary-dialog" aria-labelledby="summary-title">
		<div class="run-head">
			<span class="label">Summary</span>
			<h3 id="summary-title"></h3>
		</div>
		<div class="run-body preview-body"><div id="summary-text" class="preview-doc markdown"></div></div>
		<div class="run-foot">
			<span class="hint" id="summary-meta"></span>
			<a class="btn" id="summary-run" href="#">View run</a>
			<button type="button" class="btn" id="summary-close" aria-label="Close">Close</button>
		</div>
	</dialog>
	<dialog class="run source" id="source-dialog" aria-labelledby="source-title">
		<div class="run-head">
			<span class="label">Source</span>
			<h3 id="source-title"></h3>
		</div>
		<div class="run-body"><pre class="code" id="source-code"></pre></div>
		<div class="run-foot">
			<span class="hint"></span>
			<button type="button" class="btn" id="source-close" aria-label="Close">Close</button>
		</div>
	</dialog>
