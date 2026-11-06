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
					<radialGradient id="voice-nucleus">
						<stop offset="0%" stop-color="#fffdf7" stop-opacity="1"/>
						<stop offset="40%" stop-color="var(--voice-hot)" stop-opacity="0.72"/>
						<stop offset="100%" stop-color="var(--voice-warm)" stop-opacity="0"/>
					</radialGradient>
					<filter id="voice-wake" x="-25%" y="-25%" width="150%" height="150%" color-interpolation-filters="sRGB">
						<feGaussianBlur stdDeviation="1.3"/>
					</filter>
					<path id="voice-path-a" pathLength="1000" d="M160 46a62 114 0 1 1-0.1 0z" transform="rotate(14 160 160)"/>
					<path id="voice-path-b" pathLength="1000" d="M160 58a104 102 0 1 1-0.1 0z" transform="rotate(-62 160 160)"/>
					<path id="voice-path-c" pathLength="1000" d="M160 70a132 90 0 1 1-0.1 0z" transform="rotate(68 160 160)"/>
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
				<circle class="voice-rim" cx="160" cy="160" r="54" fill="none" stroke="url(#voice-ring-a)" stroke-width="1.1"/>
				<circle class="voice-nucleus" cx="160" cy="160" r="27" fill="url(#voice-nucleus)"/>
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
				<g class="voice-comets">
					<g class="voice-comet comet-a">
						<circle class="coma" r="9" fill="#fff6e2"/>
						<circle class="head" r="3.2" fill="#fffdf4"/>
						<animateMotion dur="19s" begin="-3s" repeatCount="indefinite"><mpath href="#voice-path-a"/></animateMotion>
					</g>
					<g class="voice-comet comet-b">
						<circle class="coma" r="7.5" fill="#f3e9ff"/>
						<circle class="head" r="2.7" fill="#fdfaff"/>
						<animateMotion dur="31s" begin="-19s" repeatCount="indefinite" keyPoints="1;0" keyTimes="0;1" calcMode="linear"><mpath href="#voice-path-b"/></animateMotion>
					</g>
					<g class="voice-comet comet-c">
						<circle class="coma" r="6" fill="#fff3d6"/>
						<circle class="head" r="2.2" fill="#fffaf0"/>
						<animateMotion dur="47s" begin="-34s" repeatCount="indefinite"><mpath href="#voice-path-c"/></animateMotion>
					</g>
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
