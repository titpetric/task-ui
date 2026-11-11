// tap posts every block of microphone samples to voice.js, which keeps them
// for the recording. It is a file of its own because the page's CSP is
// script-src 'self', and an AudioWorklet module is a script: a blob: URL is
// refused.
class Tap extends AudioWorkletProcessor {
  process(inputs) {
    const ch = inputs[0] && inputs[0][0];
    if (ch) this.port.postMessage(new Float32Array(ch));
    return true;
  }
}

registerProcessor("tap", Tap);
