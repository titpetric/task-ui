<?php $_v=&$this->vars;?>
<div class="voice" id="voice-overlay" hidden>
	<svg class="voice-avatar" viewBox="0 0 320 320" aria-hidden="true" focusable="false">
		<defs>
			<filter id="voice-fog" x="-30%" y="-30%" width="160%" height="160%" color-interpolation-filters="sRGB">
				<feTurbulence type="fractalNoise" baseFrequency="0.012 0.02" numOctaves="4" seed="7" result="noise"/>
				<feDisplacementMap in="SourceGraphic" in2="noise" scale="12" xChannelSelector="R" yChannelSelector="G" result="bent"/>
				<feGaussianBlur in="bent" stdDeviation="2.4"/>
			</filter>
			<filter id="voice-haze" x="-40%" y="-40%" width="180%" height="180%" color-interpolation-filters="sRGB">
				<feTurbulence type="fractalNoise" baseFrequency="0.008 0.014" numOctaves="3" seed="19" result="n2"/>
				<feDisplacementMap in="SourceGraphic" in2="n2" scale="26" xChannelSelector="R" yChannelSelector="B" result="b2"/>
				<feGaussianBlur in="b2" stdDeviation="9"/>
			</filter>
			<radialGradient id="voice-core">
				<stop offset="0%" stop-color="var(--voice-hot)" stop-opacity="0.95"/>
				<stop offset="45%" stop-color="var(--voice-warm)" stop-opacity="0.45"/>
				<stop offset="100%" stop-color="var(--voice-cool)" stop-opacity="0"/>
			</radialGradient>
			<linearGradient id="voice-arm-a" x1="0" y1="0" x2="1" y2="1">
				<stop offset="0%" stop-color="var(--voice-warm)" stop-opacity="0"/>
				<stop offset="35%" stop-color="var(--voice-warm)" stop-opacity="0.75"/>
				<stop offset="100%" stop-color="var(--voice-cool)" stop-opacity="0"/>
			</linearGradient>
			<linearGradient id="voice-arm-b" x1="1" y1="0" x2="0" y2="1">
				<stop offset="0%" stop-color="var(--voice-cool)" stop-opacity="0"/>
				<stop offset="40%" stop-color="var(--voice-cool)" stop-opacity="0.7"/>
				<stop offset="100%" stop-color="var(--voice-violet)" stop-opacity="0"/>
			</linearGradient>
			<linearGradient id="voice-arm-c" x1="0" y1="1" x2="1" y2="0">
				<stop offset="0%" stop-color="var(--voice-violet)" stop-opacity="0"/>
				<stop offset="45%" stop-color="var(--voice-violet)" stop-opacity="0.6"/>
				<stop offset="100%" stop-color="var(--voice-hot)" stop-opacity="0"/>
			</linearGradient>
		</defs>
		<g class="voice-haze" filter="url(#voice-haze)">
			<ellipse cx="146" cy="150" rx="104" ry="86"/>
			<ellipse cx="176" cy="168" rx="88" ry="104"/>
			<ellipse cx="160" cy="142" rx="118" ry="70"/>
		</g>
		<circle class="voice-core" cx="160" cy="160" r="86" fill="url(#voice-core)"/>
		<g class="voice-arms" filter="url(#voice-fog)">
			<g class="voice-whorl voice-whorl-1">
				<path d="M166 160C166 161 167 163 166 164C166 166 165 167 163 168C162 169 159 170 157 170C155 170 152 169 150 168C148 166 146 164 145 161C144 158 144 154 145 151C146 147 148 143 152 140C155 138 161 135 166 135C171 135 178 136 183 140C188 143 194 149 196 156C199 163 200 173 198 181C195 190 190 200 182 207C174 213 161 220 149 221C136 221 119 219 107 211C94 203 80 189 73 173C66 156 68 122 67 112" stroke="url(#voice-arm-a)" stroke-width="9"/>
				<path d="M155 163C154 163 153 161 153 160C152 158 152 156 153 155C154 153 155 151 157 150C159 149 162 148 164 148C167 148 170 149 172 151C175 153 177 157 178 160C179 164 179 168 177 172C176 176 173 181 168 184C164 187 157 189 151 189C145 189 137 187 131 182C126 178 119 170 117 162C114 153 113 141 117 132C120 122 128 110 138 102C148 95 163 89 178 89C193 89 213 93 227 103C242 114 258 144 264 152" stroke="url(#voice-arm-c)" stroke-width="6"/>
				<path d="M163 155C164 155 166 156 167 157C168 158 169 160 169 162C169 163 168 166 167 168C166 169 164 171 162 172C160 173 156 174 153 173C150 173 147 171 144 168C142 166 139 161 139 157C138 153 139 147 141 142C144 138 149 133 154 130C160 128 168 126 175 127C182 128 191 132 197 138C204 145 210 155 211 165C213 175 212 189 207 200C201 212 190 225 177 232C164 238 145 244 127 241C109 239 87 231 72 216C56 202 41 164 35 153" stroke="url(#voice-arm-b)" stroke-width="4"/>
			</g>
			<g class="voice-whorl voice-whorl-2">
				<path d="M163 165C162 166 161 167 159 167C158 168 156 167 154 167C152 166 151 164 150 162C149 160 148 158 148 155C149 153 150 149 152 147C154 145 158 143 161 142C165 142 170 142 174 144C178 145 182 149 185 154C187 158 189 165 188 171C188 177 185 185 180 190C175 196 167 201 159 203C150 205 138 205 129 201C119 197 107 189 101 178C95 167 89 152 90 137C91 122 97 103 108 89C120 75 151 61 159 55" stroke="url(#voice-arm-b)" stroke-width="8"/>
				<path d="M155 157C155 156 156 154 157 153C158 152 160 151 161 151C163 151 166 151 167 153C169 154 171 156 172 158C173 160 174 164 173 167C173 170 171 173 168 176C166 178 162 181 157 181C153 182 147 181 143 179C138 176 133 172 130 166C128 161 126 153 127 145C128 138 132 129 138 123C144 117 154 110 164 109C175 107 189 107 200 113C211 119 224 129 231 142C238 155 244 175 242 192C239 210 232 232 217 248C202 263 165 279 154 285" stroke="url(#voice-arm-a)" stroke-width="5"/>
			</g>
			<g class="voice-whorl voice-whorl-3">
				<path d="M159 166C158 166 156 166 154 165C153 164 152 163 151 161C151 159 150 157 151 155C151 153 153 150 155 149C157 147 160 145 163 145C166 145 170 146 173 148C176 150 180 153 181 157C183 161 184 167 183 172C181 177 179 183 174 187C169 191 162 195 155 196C147 197 138 196 130 191C122 187 114 179 109 170C105 160 102 146 104 134C106 122 113 106 124 96C134 86 151 75 169 73C187 70 219 80 229 81" stroke="url(#voice-arm-c)" stroke-width="7"/>
				<path d="M158 154C158 154 160 153 162 153C163 153 165 153 167 154C168 155 170 157 170 159C171 161 172 164 171 166C170 169 169 172 166 174C164 176 160 177 157 178C153 178 148 177 145 175C141 172 137 168 135 164C133 159 132 152 133 146C135 140 138 132 144 128C149 123 158 118 166 117C175 116 187 118 196 123C205 128 216 138 221 149C226 160 230 177 227 191C224 206 216 224 203 237C190 249 158 260 149 264" stroke="url(#voice-arm-a)" stroke-width="4"/>
			</g>
		</g>
	</svg>
	<p class="voice-said" id="voice-said" role="status" aria-live="polite"></p>
	<div class="voice-acts">
		<button type="button" class="voice-act discard" id="voice-discard" title="Discard" aria-label="Discard"><svg class="ic"><use href="#i-trash"/></svg></button>
		<button type="button" class="voice-act send" id="voice-send" title="Send" aria-label="Send"><svg class="ic"><use href="#i-send"/></svg></button>
	</div>
</div>
