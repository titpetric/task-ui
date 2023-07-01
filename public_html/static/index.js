class Runner {
	exec(terminal_id, name, interactive) {
		var terminal_dom = document.getElementById(terminal_id)
		if (!terminal_dom) {
			return;
		}
		terminal_dom.innerHTML = '';
		var term;
		var isHTTPS = window.location.protocol === 'https:'
		var websocketProto = isHTTPS ? 'wss:' : 'ws:';
		var websocketHost = isHTTPS ? window.location.hostname : window.location.host;
		var websocket = new WebSocket(websocketProto + "//" + websocketHost + "/ws/" + name);
		websocket.onopen = function(evt) {
			if (this.term) {
				this.term.destroy();
			}
			this.term = term = new Terminal({
				cols: 163,
				rows: 20,
				screenKeys: true,
				useStyle: true,
				cursorBlink: true,
			});
			term.on('data', function(data) {
				if (interactive) {
					websocket.send(data);
				}
			});
			term.on('title', function(title) {
				document.title = title;
			});
			term.open(terminal_dom);

			websocket.onmessage = function(evt) {
				term.write(evt.data);
			}
			websocket.onclose = function(evt) {
				term.write("Session terminated");
			}
			websocket.onerror = function(evt) {
				if (typeof console.log == "function") {
					console.log(evt)
				}
			}
		}
	}

	history(terminal_id, history_id, replay) {
		var interactive = false
		var terminal_dom = document.getElementById(terminal_id)
		if (!terminal_dom) {
			return;
		}
		terminal_dom.innerHTML = '';
		var term;
		var isHTTPS = window.location.protocol === 'https:'
		var websocketProto = isHTTPS ? 'wss:' : 'ws:';
		var websocketHost = isHTTPS ? window.location.hostname : window.location.host;
		var websocketURL = websocketProto + "//" + websocketHost + "/ws/history/" + history_id
		if (replay) {
			websocketURL = websocketURL + "/replay"
		}
		var websocket = new WebSocket(websocketURL);
		websocket.onopen = function(evt) {
			if (this.term) {
				this.term.destroy();
			}
			this.term = term = new Terminal({
				cols: 163,
				rows: 20,
				screenKeys: true,
				useStyle: true,
				cursorBlink: true,
			});
			term.on('data', function(data) {
				if (interactive) {
					websocket.send(data);
				}
			});
			term.on('title', function(title) {
				document.title = title;
			});
			term.open(terminal_dom);

			websocket.onmessage = function(evt) {
				term.write(evt.data);
			}
			websocket.onclose = function(evt) {
				term.write("Session terminated");
			}
			websocket.onerror = function(evt) {
				if (typeof console.log == "function") {
					console.log(evt)
				}
			}
		}
	}
}

var runner = new Runner();
