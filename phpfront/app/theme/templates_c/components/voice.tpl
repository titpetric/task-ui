<?php $_v=&$this->vars;?>
	<button type="button" class="voice-orb" id="voice-orb" aria-haspopup="dialog" title="Dictate" aria-label="Dictate" hidden></button>
	<dialog class="voice" id="voice-overlay" aria-label="Dictate">
		<button type="button" class="voice-stage" id="voice-stage" aria-pressed="false" title="Record" aria-label="Record" autofocus>
			<span class="voice-well" aria-hidden="true"></span>
			<svg class="voice-avatar" viewBox="0 0 320 320" aria-hidden="true" focusable="false">
				<defs>
					<filter id="voice-gravity" x="-20%" y="-20%" width="140%" height="140%" color-interpolation-filters="sRGB">
						<feImage id="voice-well-map" x="0" y="0" width="100%" height="100%" preserveAspectRatio="none" result="field"/>
						<feDisplacementMap in="SourceGraphic" in2="field" scale="54" xChannelSelector="R" yChannelSelector="G"/>
					</filter>
					<filter id="voice-fog" x="-30%" y="-30%" width="160%" height="160%" color-interpolation-filters="sRGB">
						<feTurbulence type="fractalNoise" baseFrequency="0.01 0.016" numOctaves="3" seed="7" result="noise"/>
						<feDisplacementMap in="SourceGraphic" in2="noise" scale="7" xChannelSelector="R" yChannelSelector="G" result="bent"/>
						<feGaussianBlur in="bent" stdDeviation="1.1"/>
					</filter>
					<filter id="voice-haze" x="-40%" y="-40%" width="180%" height="180%" color-interpolation-filters="sRGB">
						<feTurbulence type="fractalNoise" baseFrequency="0.008 0.014" numOctaves="3" seed="19" result="n2"/>
						<feDisplacementMap in="SourceGraphic" in2="n2" scale="22" xChannelSelector="R" yChannelSelector="B" result="b2"/>
						<feGaussianBlur in="b2" stdDeviation="10"/>
					</filter>
					<radialGradient id="voice-core">
						<stop offset="0%" stop-color="#fffaf0" stop-opacity="1"/>
						<stop offset="18%" stop-color="var(--voice-hot)" stop-opacity="0.9"/>
						<stop offset="52%" stop-color="var(--voice-warm)" stop-opacity="0.34"/>
						<stop offset="100%" stop-color="var(--voice-cool)" stop-opacity="0"/>
					</radialGradient>
					<radialGradient id="voice-shadow">
						<stop offset="0%" stop-color="#01040a" stop-opacity="0.92"/>
						<stop offset="62%" stop-color="#01040a" stop-opacity="0.55"/>
						<stop offset="100%" stop-color="#01040a" stop-opacity="0"/>
					</radialGradient>
					<linearGradient id="voice-ring-a" x1="0" y1="0" x2="1" y2="1">
						<stop offset="0%" stop-color="var(--voice-cool)" stop-opacity="0"/>
						<stop offset="30%" stop-color="var(--voice-cool)" stop-opacity="0.5"/>
						<stop offset="70%" stop-color="var(--voice-warm)" stop-opacity="0.5"/>
						<stop offset="100%" stop-color="var(--voice-warm)" stop-opacity="0"/>
					</linearGradient>
					<linearGradient id="voice-ring-b" x1="1" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="var(--voice-violet)" stop-opacity="0"/>
						<stop offset="35%" stop-color="var(--voice-violet)" stop-opacity="0.45"/>
						<stop offset="75%" stop-color="var(--voice-cool)" stop-opacity="0.45"/>
						<stop offset="100%" stop-color="var(--voice-cool)" stop-opacity="0"/>
					</linearGradient>
					<linearGradient id="voice-ring-c" x1="0" y1="1" x2="1" y2="0">
						<stop offset="0%" stop-color="var(--voice-hot)" stop-opacity="0"/>
						<stop offset="40%" stop-color="var(--voice-hot)" stop-opacity="0.4"/>
						<stop offset="80%" stop-color="var(--voice-violet)" stop-opacity="0.4"/>
						<stop offset="100%" stop-color="var(--voice-violet)" stop-opacity="0"/>
					</linearGradient>
					<radialGradient id="voice-grey" fx="38%" fy="32%">
						<stop offset="0%" stop-color="#ffffff" stop-opacity="1"/>
						<stop offset="20%" stop-color="#e0e0e0" stop-opacity="0.98"/>
						<stop offset="45%" stop-color="#9c9c9c" stop-opacity="0.92"/>
						<stop offset="68%" stop-color="#5a5a5a" stop-opacity="0.75"/>
						<stop offset="85%" stop-color="#262626" stop-opacity="0.42"/>
						<stop offset="100%" stop-color="#000000" stop-opacity="0"/>
					</radialGradient>
					<filter id="voice-phong" x="-15%" y="-15%" width="130%" height="130%" color-interpolation-filters="sRGB">
						<feComponentTransfer>
							<feFuncR type="table" tableValues="0 0.09 0.2 0.5 0.9 1 1"/>
							<feFuncG type="table" tableValues="0 0.12 0.46 0.78 0.94 1 1"/>
							<feFuncB type="table" tableValues="0 0.26 0.6 0.64 0.62 0.92 1"/>
						</feComponentTransfer>
					</filter>
					<filter id="voice-wake" x="-25%" y="-25%" width="150%" height="150%" color-interpolation-filters="sRGB">
						<feGaussianBlur stdDeviation="1.3"/>
					</filter>
					<path id="voice-path-a" pathLength="1000" d="M232 178 L231 185 L230 192 L228 198 L224 205 L220 211 L215 217 L209 223 L203 227 L195 232 L187 235 L178 238 L169 240 L160 240 L150 240 L140 239 L130 237 L121 235 L111 231 L102 226 L93 220 L86 214 L78 206 L72 199 L67 190 L63 181 L60 172 L58 163 L57 153 L57 144 L59 135 L62 126 L66 117 L71 109 L77 102 L84 95 L92 90 L101 85 L110 81 L120 78 L130 76 L140 75 L150 75 L160 76 L170 78 L180 81 L189 84 L198 89 L206 94 L213 100 L219 106 L225 113 L229 120 L233 128 L236 135 L237 143 L238 151 L238 158 L237 165 L235 172 L232 178 L228 184 L224 189 L220 194 L214 198 L209 202 L203 204 L197 207 L191 208 L184 209 L178 210 L172 210 L166 209 L160 208 L154 206 L149 204 L144 202 L139 199 L135 196 L131 192 L128 189 L125 185 L123 181 L120 177 L119 173 L117 169 L117 165 L116 161 L116 157 L117 153 L117 149 L118 146 L120 142 L122 138 L124 135 L127 132 L130 129 L133 126 L137 124 L141 121 L145 119 L150 118 L155 117 L160 116 L165 116 L171 116 L177 117 L182 118 L188 119 L194 122 L199 124 L205 128 L210 132 L214 136 L219 141 L223 146 L226 152 L228 158 L230 164 L231 171 Z"/>
					<path id="voice-path-b" pathLength="1000" d="M217 69 L227 76 L235 85 L243 94 L250 103 L256 114 L260 124 L263 135 L265 146 L266 157 L266 167 L264 178 L261 188 L257 197 L252 206 L247 214 L240 221 L233 228 L225 233 L216 238 L207 241 L198 244 L188 246 L179 246 L169 246 L160 244 L151 242 L142 239 L134 235 L126 231 L119 225 L113 220 L107 213 L101 207 L97 200 L93 192 L90 185 L88 177 L87 170 L86 162 L86 155 L87 148 L89 141 L92 134 L95 127 L98 121 L103 116 L107 111 L113 106 L118 102 L125 99 L131 96 L138 94 L145 92 L152 92 L160 91 L167 92 L175 93 L182 95 L190 98 L197 101 L204 105 L210 110 L216 115 L222 121 L227 128 L231 134 L235 142 L238 149 L241 158 L242 166 L243 174 L243 183 L241 191 L239 200 L236 208 L232 215 L227 223 L221 230 L215 236 L207 242 L199 246 L190 250 L180 253 L171 255 L160 256 L150 256 L139 255 L128 252 L118 249 L107 244 L97 239 L88 232 L79 224 L71 216 L64 206 L58 196 L53 186 L50 175 L47 163 L46 152 L46 140 L47 129 L50 118 L54 107 L59 97 L65 87 L73 79 L81 71 L91 64 L101 58 L112 54 L124 50 L135 48 L148 47 L160 48 L172 50 L184 53 L195 57 L206 62 Z"/>
					<path id="voice-path-c" pathLength="1000" d="M170 191 L167 190 L164 190 L161 189 L159 187 L156 186 L154 184 L152 183 L150 181 L148 179 L147 177 L146 175 L144 173 L143 171 L142 168 L141 166 L140 164 L140 161 L139 159 L139 156 L138 154 L138 151 L138 148 L139 145 L139 142 L140 139 L141 136 L142 133 L144 131 L146 128 L149 125 L152 123 L155 121 L158 119 L162 118 L166 117 L170 117 L175 117 L180 118 L185 120 L190 122 L194 125 L199 128 L204 133 L208 138 L212 143 L215 149 L218 156 L221 163 L222 171 L223 178 L223 186 L223 194 L221 202 L219 210 L216 217 L212 224 L207 231 L201 237 L195 242 L188 246 L180 249 L172 251 L164 253 L156 253 L147 252 L138 249 L130 246 L121 241 L114 236 L106 230 L99 222 L93 214 L88 205 L83 196 L80 186 L77 176 L75 166 L75 156 L75 146 L77 136 L79 126 L83 118 L87 109 L92 102 L98 96 L104 90 L111 85 L118 82 L126 79 L133 78 L141 77 L149 78 L156 80 L164 82 L171 85 L177 90 L183 94 L189 100 L193 106 L197 112 L201 118 L203 125 L205 132 L207 138 L207 145 L207 151 L207 157 L206 162 L204 167 L202 172 L199 176 L197 180 L193 183 L190 186 L187 188 L183 189 L180 190 L177 191 L173 191 Z"/>
				</defs>
				<circle class="voice-shadow" cx="160" cy="160" r="150" fill="url(#voice-shadow)"/>
				<g class="voice-haze" filter="url(#voice-haze)">
					<ellipse cx="150" cy="154" rx="92" ry="78"/>
					<ellipse cx="172" cy="166" rx="78" ry="92"/>
				</g>
				<g class="voice-orbits" filter="url(#voice-fog)">
					<use class="voice-ring" href="#voice-path-a" stroke="url(#voice-ring-a)" stroke-width="1.6"/>
					<use class="voice-ring" href="#voice-path-b" stroke="url(#voice-ring-b)" stroke-width="1.4"/>
					<use class="voice-ring" href="#voice-path-c" stroke="url(#voice-ring-c)" stroke-width="1.2"/>
				</g>
				<circle class="voice-core" cx="160" cy="160" r="72" fill="url(#voice-core)"/>
				<g class="voice-wakes" filter="url(#voice-wake)">
					<g class="wake wake-a">
						<use class="w1" href="#voice-path-a"><animate attributeName="stroke-dashoffset" dur="19s" begin="-3s" repeatCount="indefinite" from="90" to="-910"/></use>
						<use class="w2" href="#voice-path-a"><animate attributeName="stroke-dashoffset" dur="19s" begin="-3s" repeatCount="indefinite" from="200" to="-800"/></use>
						<use class="w3" href="#voice-path-a"><animate attributeName="stroke-dashoffset" dur="19s" begin="-3s" repeatCount="indefinite" from="380" to="-620"/></use>
						<use class="w4" href="#voice-path-a"><animate attributeName="stroke-dashoffset" dur="19s" begin="-3s" repeatCount="indefinite" from="600" to="-400"/></use>
					</g>
					<g class="wake wake-b">
						<use class="w1" href="#voice-path-b"><animate attributeName="stroke-dashoffset" dur="31s" begin="-19s" repeatCount="indefinite" from="-1000" to="0"/></use>
						<use class="w2" href="#voice-path-b"><animate attributeName="stroke-dashoffset" dur="31s" begin="-19s" repeatCount="indefinite" from="-1000" to="0"/></use>
						<use class="w3" href="#voice-path-b"><animate attributeName="stroke-dashoffset" dur="31s" begin="-19s" repeatCount="indefinite" from="-1000" to="0"/></use>
						<use class="w4" href="#voice-path-b"><animate attributeName="stroke-dashoffset" dur="31s" begin="-19s" repeatCount="indefinite" from="-1000" to="0"/></use>
					</g>
					<g class="wake wake-c">
						<use class="w1" href="#voice-path-c"><animate attributeName="stroke-dashoffset" dur="47s" begin="-34s" repeatCount="indefinite" from="90" to="-910"/></use>
						<use class="w2" href="#voice-path-c"><animate attributeName="stroke-dashoffset" dur="47s" begin="-34s" repeatCount="indefinite" from="200" to="-800"/></use>
						<use class="w3" href="#voice-path-c"><animate attributeName="stroke-dashoffset" dur="47s" begin="-34s" repeatCount="indefinite" from="380" to="-620"/></use>
						<use class="w4" href="#voice-path-c"><animate attributeName="stroke-dashoffset" dur="47s" begin="-34s" repeatCount="indefinite" from="600" to="-400"/></use>
					</g>
				</g>
				<g class="voice-balls" filter="url(#voice-phong)">
					<circle class="ball nucleus" cx="160" cy="160" r="42" fill="url(#voice-grey)"/>
					<circle class="ball ball-a" r="40" fill="url(#voice-grey)">
						<animateMotion dur="19s" begin="-3s" repeatCount="indefinite"><mpath href="#voice-path-a"/></animateMotion>
					</circle>
					<circle class="ball ball-b" r="34" fill="url(#voice-grey)">
						<animateMotion dur="31s" begin="-19s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#voice-path-b"/></animateMotion>
					</circle>
					<circle class="ball ball-c" r="28" fill="url(#voice-grey)">
						<animateMotion dur="47s" begin="-34s" repeatCount="indefinite"><mpath href="#voice-path-c"/></animateMotion>
					</circle>
				</g>
			</svg>
		</button>
		<p class="voice-said" id="voice-said" role="status" aria-live="polite"></p>
		<p class="voice-where" id="voice-where"></p>
		<div class="voice-acts">
			<button type="button" class="voice-act discard" id="voice-discard" title="Discard" aria-label="Discard"><svg class="ic"><use href="#i-trash"/></svg></button>
			<button type="button" class="voice-act replay" id="voice-replay" title="Play again" aria-label="Play again" hidden><svg class="ic"><use href="#i-play"/></svg></button>
			<button type="button" class="voice-act send" id="voice-send" title="Send" aria-label="Send" disabled><svg class="ic"><use href="#i-send"/></svg></button>
		</div>
	</dialog>
