// orbs: the avatar in the voice window.
//
// Five orbs circle one centre in tilted planes. Each is drawn as a plain
// radial greyscale, white in the middle and black at the rim, and the five
// are added together, so where two overlap the sum runs past white. That
// summed field is read through a palette of black to purple to white, which
// is what makes a pair read as one body with one outline.
//
// Each orb is drawn again at where it was a moment ago, dimmer each time, so
// it drags the path it took behind it and the trail dims down the palette.
//
// See docs/guide/voice.md.
(function () {
  "use strict";

  // SIZE is what the field is drawn at, whatever box it is shown in: the
  // field is smooth, so scaling it up costs nothing to look at and holds the
  // per-frame work still on a phone.
  var SIZE = 256;
  var MID = SIZE / 2;

  // DEPTH is the camera distance that turns an orb's z into its size.
  var DEPTH = 340;

  // TRAIL is how many times behind now each orb is drawn again, STEP the
  // seconds between them. The trail is read back out of the orbit, so it
  // costs no memory: a field kept and decayed climbs to its own ceiling
  // wherever an orb is slow, and every blob there goes white.
  var TRAIL = 14;
  var STEP = 0.03;

  // SPRITE is the radius the one greyscale orb is drawn at, and every orb is
  // that image scaled: a gradient is smooth, so one is enough.
  var SPRITE = 64;

  // PEAK is how much field one orb puts in at its middle, and CORE the
  // fraction of its radius that is all of it: an orb on its own is white in
  // the middle, and a pair that overlaps runs past the top of the palette.
  var PEAK = 255;
  var CORE = 0.16;

  // FLOOR is the field below which there is no blob, and RISE how far above
  // it the blob has faded in. Both are soft: a hard cut draws a ring at the
  // edge of every orb.
  var FLOOR = 30;
  var RISE = 42;

  // STOPS are the palette the summed field is read through, black to purple
  // to white, and the steps between them are interpolated: one even climb,
  // so a blob is a solid shape with a lit middle and not a set of rings.
  var STOPS = [
    [0, 0, 0, 0],
    [48, 26, 6, 48],
    [96, 74, 18, 120],
    [144, 130, 40, 190],
    [192, 180, 95, 225],
    [224, 218, 158, 242],
    [255, 255, 255, 255],
  ];

  // Each orb walks a circle of radius R in a plane tilted by tilt and turned
  // by turn, once every secs, starting at a different point of it. A
  // negative secs walks it the other way. Half of secs is the time it takes
  // to cross the window, which reads best between three and four seconds.
  var ORBS = [
    // A radius of nothing leaves this one at the centre, which the others
    // pass through and merge with. It puts in less field than they do, or
    // the middle of the window is white whatever is happening there.
    { R: 0, tilt: 0, turn: 0, secs: 1, size: 38, at: 0, peak: 0.72 },
    { R: 72, tilt: 72, turn: 15, secs: 6.3, size: 27, at: 0 },
    { R: 76, tilt: 81, turn: 70, secs: 8.8, size: 20, at: 0.11 },
    { R: 52, tilt: 38, turn: 120, secs: -6.9, size: 29, at: 0.72 },
  ];

  // ramp returns the palette as 256 entries of red, green, blue and alpha,
  // interpolated between the stops, with the alpha fading in over RISE.
  function ramp() {
    var map = new Uint8ClampedArray(256 * 4);
    for (var v = 0; v < 256; v++) {
      var i = 0;
      while (i < STOPS.length - 2 && v > STOPS[i + 1][0]) i++;
      var lo = STOPS[i];
      var hi = STOPS[i + 1];
      var t = (v - lo[0]) / (hi[0] - lo[0]);
      map[v * 4] = lo[1] + (hi[1] - lo[1]) * t;
      map[v * 4 + 1] = lo[2] + (hi[2] - lo[2]) * t;
      map[v * 4 + 2] = lo[3] + (hi[3] - lo[3]) * t;
      map[v * 4 + 3] = v <= FLOOR ? 0 : 255 * Math.min(1, (v - FLOOR) / RISE);
    }
    return map;
  }

  // kernel is the field one orb puts into the world at a fraction of its
  // radius: Wyvill's polynomial, which is 1 at the middle and reaches 0 at
  // the rim with no corner. Summing these is what rounds two orbs into one
  // blob; a straight gradient would leave a waist between them.
  function kernel(q) {
    if (q <= CORE) return 1;
    var d = (q - CORE) / (1 - CORE);
    var d2 = d * d;
    return 1 - (22 / 9) * d2 + (17 / 9) * d2 * d2 - (4 / 9) * d2 * d2 * d2;
  }

  // sprite returns the one orb as a canvas of that field in grey, drawn once
  // and scaled for each.
  function sprite() {
    var pad = document.createElement("canvas");
    pad.width = pad.height = SPRITE * 2;
    var ctx = pad.getContext("2d");
    var fill = ctx.createRadialGradient(SPRITE, SPRITE, 0, SPRITE, SPRITE, SPRITE);
    for (var i = 0; i <= 32; i++) {
      var q = i / 32;
      var v = Math.round(PEAK * Math.max(0, kernel(q)));
      fill.addColorStop(q, "rgb(" + v + "," + v + "," + v + ")");
    }
    ctx.fillStyle = fill;
    ctx.beginPath();
    ctx.arc(SPRITE, SPRITE, SPRITE, 0, Math.PI * 2);
    ctx.fill();
    return pad;
  }

  // place returns where an orb is now and how big: the circle in its tilted
  // plane, turned into the view, with z read as distance from the camera.
  function place(orb, now) {
    var th = Math.PI * 2 * (now / orb.secs + orb.at);
    var tilt = orb.tilt * Math.PI / 180;
    var turn = orb.turn * Math.PI / 180;
    var x = orb.R * Math.cos(th);
    var y = orb.R * Math.sin(th) * Math.cos(tilt);
    var z = orb.R * Math.sin(th) * Math.sin(tilt);
    var near = DEPTH / (DEPTH - z);
    return {
      x: MID + (x * Math.cos(turn) - y * Math.sin(turn)) * near,
      y: MID + (x * Math.sin(turn) + y * Math.cos(turn)) * near,
      r: orb.size * near,
    };
  }

  var PALETTE = ramp();
  var ORB = null;    // the greyscale sprite, made on the first frame
  var show = null;   // the canvas on the page
  var paint = null;  // the field the orbs are summed into, offscreen
  var out = null;    // the page canvas's context
  var raf = 0;
  var level = 0;

  // draw sums the orbs and their trails into the field, reads the sum
  // through the palette and puts the colour on the page's canvas.
  function draw(secs) {
    paint.globalCompositeOperation = "source-over";
    paint.globalAlpha = 1;
    paint.fillStyle = "#000000";
    paint.fillRect(0, 0, SIZE, SIZE);
    paint.globalCompositeOperation = "lighter";
    var swell = 1 + level * 0.35;
    for (var i = 0; i < ORBS.length; i++) {
      for (var j = TRAIL - 1; j >= 0; j--) {
        var back = 1 - j / TRAIL;
        // A trail sample is a tenth of the orb: a slow orb's trail sits on
        // top of the orb itself, and at full weight the sum of them all
        // would run the head past white wherever it lingers.
        paint.globalAlpha = (j === 0 ? 1 : 0.1 * back * back * back) * (ORBS[i].peak || 1);
        var spot = place(ORBS[i], secs - j * STEP);
        var r = spot.r * swell * (0.45 + 0.55 * back);
        paint.drawImage(ORB, spot.x - r, spot.y - r, r * 2, r * 2);
      }
    }
    paint.globalAlpha = 1;
    var frame = paint.getImageData(0, 0, SIZE, SIZE);
    var data = frame.data;
    for (var p = 0; p < data.length; p += 4) {
      var v = data[p] * 4;
      data[p] = PALETTE[v];
      data[p + 1] = PALETTE[v + 1];
      data[p + 2] = PALETTE[v + 2];
      data[p + 3] = PALETTE[v + 3];
    }
    out.putImageData(frame, 0, 0);
  }

  function tick(now) {
    draw(now / 1000);
    raf = window.requestAnimationFrame(tick);
  }

  var still = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  window.voiceOrbs = {
    // start draws into the canvas and keeps drawing. Reduced motion gets one
    // frame of it and no loop.
    start: function (canvas) {
      if (!canvas || !canvas.getContext || raf) return;
      show = canvas;
      show.width = show.height = SIZE;
      out = show.getContext("2d");
      if (!ORB) ORB = sprite();
      if (!paint) {
        var field = document.createElement("canvas");
        field.width = field.height = SIZE;
        paint = field.getContext("2d", { willReadFrequently: true });
      }
      paint.globalCompositeOperation = "source-over";
      paint.fillStyle = "#000000";
      paint.fillRect(0, 0, SIZE, SIZE);
      if (still) {
        draw(0);
        return;
      }
      raf = window.requestAnimationFrame(tick);
    },
    // stop ends the loop: the window is closed and nothing of it is seen.
    stop: function () {
      if (raf) window.cancelAnimationFrame(raf);
      raf = 0;
    },
    // level is how loud the microphone is, 0 to 1, which swells the orbs.
    level: function (loud) {
      level = loud;
    },
  };
})();
