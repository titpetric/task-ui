<?php $_v=&$this->vars;?>
	<button type="button" class="voice-orb" id="voice-orb" aria-haspopup="dialog" title="Dictate" aria-label="Dictate" hidden></button>
	<dialog class="voice on-dark" id="voice-overlay" aria-label="Dictate">
		<button type="button" class="voice-stage" id="voice-stage" aria-pressed="false" title="Record" aria-label="Record" autofocus>
			<span class="voice-well" aria-hidden="true"></span>
			<svg class="voice-field" viewBox="0 0 320 320" aria-hidden="true" focusable="false">
				<defs>
					<filter id="voice-gravity" x="-20%" y="-20%" width="140%" height="140%" color-interpolation-filters="sRGB">
						<feImage id="voice-well-map" x="0" y="0" width="100%" height="100%" preserveAspectRatio="none" result="field"/>
						<feDisplacementMap in="SourceGraphic" in2="field" scale="54" xChannelSelector="R" yChannelSelector="G"/>
					</filter>
				</defs>
			</svg>
			<canvas class="voice-avatar" id="voice-avatar" width="256" height="256" aria-hidden="true"></canvas>
			<span class="voice-rec" aria-hidden="true"><i></i>Recording</span>
		</button>
		<p class="voice-said" id="voice-said" role="status" aria-live="polite"></p>
		<p class="voice-where" id="voice-where"></p>
		<div class="voice-quiz" id="voice-quiz" hidden>
			<p class="voice-quiz-count" id="voice-quiz-count"></p>
			<p class="voice-quiz-ask" id="voice-quiz-ask" role="status" aria-live="polite"></p>
			<div class="voice-quiz-picks" id="voice-quiz-picks" role="group" aria-labelledby="voice-quiz-ask"></div>
			<p class="voice-quiz-help" id="voice-quiz-help"></p>
			<div class="voice-quiz-acts">
				<button type="button" class="btn" id="voice-quiz-cancel">Cancel</button>
				<button type="button" class="btn btn-primary" id="voice-quiz-submit">Submit</button>
			</div>
		</div>
		<div class="voice-acts">
			<button type="button" class="voice-act discard" id="voice-discard" title="Discard" aria-label="Discard"><svg class="ic"><use href="#i-trash"/></svg></button>
			<button type="button" class="voice-act replay" id="voice-replay" title="Play again" aria-label="Play again" hidden><svg class="ic"><use href="#i-play"/></svg></button>
			<button type="button" class="voice-act send" id="voice-send" title="Send" aria-label="Send" disabled><svg class="ic"><use href="#i-send"/></svg></button>
		</div>
	</dialog>
