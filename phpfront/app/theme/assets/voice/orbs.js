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

  // PEAK is how much field one orb puts in at its middle. Short of 255 on
  // purpose: a lone orb then reads most of the palette, and a pair that
  // overlaps runs past the top of it and goes white.
  var PEAK = 180;

  // FLOOR is the field below which there is no blob: the palette starts
  // there, so the band at the edge is the blob's outline.
  var FLOOR = 56;

  // BANDS are the palette, as flat steps: a step draws the line where the
  // summed field crosses it, so two orbs that overlap show one outline
  // around the pair.
  // They alternate dark and light as they climb, because a palette that only
  // climbs reads as one smooth gradient and draws no line at all.
  var BANDS = [
    [34, 8, 66], [116, 30, 176], [52, 12, 96], [150, 52, 206],
    [70, 18, 126], [182, 84, 226], [92, 26, 156], [212, 134, 240],
    [118, 36, 188], [236, 180, 248], [150, 60, 214], [255, 255, 255],
  ];

  // Each orb walks a circle of radius R in a plane tilted by tilt and turned
  // by turn, once every secs, starting at a different point of it. A
  // negative secs walks it the other way.
  var ORBS = [
    { R: 72, tilt: 72, turn: 15, secs: 19, size: 27, at: 0 },
    { R: 62, tilt: 52, turn: -48, secs: -27, size: 23, at: 0.37 },
    { R: 76, tilt: 81, turn: 70, secs: 37, size: 20, at: 0.11 },
    { R: 52, tilt: 38, turn: 120, secs: -23, size: 29, at: 0.72 },
    { R: 66, tilt: 64, turn: -105, secs: 31, size: 22, at: 0.55 },
  ];

  // ramp returns the palette as 256 entries of red, green, blue and alpha:
  // nothing below FLOOR, and a flat band for each step above it.
  function ramp() {
    var map = new Uint8ClampedArray(256 * 4);
    var step = (256 - FLOOR) / BANDS.length;
    for (var v = FLOOR; v < 256; v++) {
      var band = BANDS[Math.min(BANDS.length - 1, Math.floor((v - FLOOR) / step))];
      map[v * 4] = band[0];
      map[v * 4 + 1] = band[1];
      map[v * 4 + 2] = band[2];
      map[v * 4 + 3] = 255;
    }
    return map;
  }

  // kernel is the field one orb puts into the world at a fraction of its
  // radius: Wyvill's polynomial, which is 1 at the middle and reaches 0 at
  // the rim with no corner. Summing these is what rounds two orbs into one
  // blob; a straight gradient would leave a waist between them.
  function kernel(q) {
    var q2 = q * q;
    return 1 - (22 / 9) * q2 + (17 / 9) * q2 * q2 - (4 / 9) * q2 * q2 * q2;
  }

  // sprite returns the one orb as a canvas of that field in grey, drawn once
  // and scaled for each.
  function sprite() {
    var pad = document.createElement("canvas");
    pad.width = pad.height = SPRITE * 2;
    var ctx = pad.getContext("2d");
    var fill = ctx.createRadialGradient(SPRITE, SPRITE, 0, SPRITE, SPRITE, SPRITE);
    for (var i = 0; i <= 20; i++) {
      var q = i / 20;
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
        paint.globalAlpha = j === 0 ? 1 : 0.1 * back * back * back;
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
