// orbs: the avatar in the voice window.
//
// Four orbs are drawn on one canvas, each a radial greyscale that is white in
// the middle and black at the rim, summed with a composite operation of
// "lighter". Two that overlap sum past 255. The sum is read through a palette
// of 256 entries, black to colour to white, so the pair draws as one body
// with one outline.
//
// Every orb keeps the path it came along. The trail is drawn out of it, and
// in the snake each follower sits one orb's width back along the path of the
// orb in front.
//
// See docs/guide/voice.md.
(function () {
  "use strict";

  // SIZE is the field's own resolution, whatever box the canvas fills. The
  // smallest feature in the field is an orb's radius, so the browser scales
  // it up and a phone draws 256 square however large the window is.
  var SIZE = 256;
  var MID = SIZE / 2;
  var TWO = Math.PI * 2;

  // DEPTH is the camera distance that turns an orb's z into its size.
  var DEPTH = 340;

  // TRAIL is how many times behind now each orb is drawn again and STEP the
  // seconds between them. SPAN is how much path is kept: the trail reaches
  // TRAIL * STEP back, and the last of the chain reaches however long the
  // head takes to travel the chain's length.
  var TRAIL = 14;
  var STEP = 0.03;
  var SPAN = 4;

  // WAKE is the field a trail adds over the orb that laid it, as a fraction
  // of the orb. The samples fall on the orb itself wherever it is slow, so
  // the weights are normalised to WAKE whatever TRAIL is and whatever an
  // orb's own tail is: an orb with a long trail would otherwise draw a third
  // brighter in the middle than one without.
  var WAKE = 0.18;

  // SPRITE is the radius the one greyscale orb is drawn at. Every orb is that
  // image scaled, so a frame costs one drawImage per orb per trail sample.
  var SPRITE = 64;

  // PEAK is the field one orb puts in at its middle and CORE the fraction of
  // its radius that is all of it. One orb alone is white in the middle; two
  // that overlap run past the top of the palette.
  var PEAK = 255;
  var CORE = 0.16;

  // FLOOR is the field below which there is no blob, and RISE how far above
  // it the blob has faded in. Both are soft: a hard cut draws a ring at the
  // edge of every orb.
  var FLOOR = 30;
  var RISE = 42;

  // MOODS are the palettes the summed field is read through, one per state of
  // the window. Each climbs evenly from black to white through its colour, so
  // a blob draws as a solid shape with a lit middle. White is held to the top
  // of the ramp: one orb's field is still 165 at half its radius, and a ramp
  // pale by 165 draws a white disc with a coloured rim and no body.
  var MOODS = {
    idle: [
      [0, 0, 0, 0], [40, 22, 5, 42], [90, 66, 16, 112], [140, 116, 34, 176],
      [190, 166, 70, 216], [228, 206, 122, 238], [248, 245, 205, 250], [255, 255, 255, 255],
    ],
    recording: [
      [0, 0, 0, 0], [40, 40, 4, 6], [90, 108, 12, 14], [140, 170, 28, 22],
      [190, 222, 70, 40], [228, 246, 140, 96], [248, 253, 215, 190], [255, 255, 255, 255],
    ],
    answered: [
      [0, 0, 0, 0], [40, 4, 36, 16], [90, 12, 98, 44], [140, 28, 158, 78],
      [190, 80, 206, 124], [228, 150, 234, 180], [248, 215, 250, 230], [255, 255, 255, 255],
    ],
    paused: [
      [0, 0, 0, 0], [40, 44, 18, 2], [90, 114, 48, 6], [140, 178, 84, 12],
      [190, 226, 128, 36], [228, 244, 176, 96], [248, 252, 224, 190], [255, 255, 255, 255],
    ],
    transcribing: [
      [0, 0, 0, 0], [40, 4, 36, 34], [90, 10, 92, 86], [140, 20, 148, 138],
      [190, 56, 196, 184], [228, 130, 228, 218], [248, 208, 248, 244], [255, 255, 255, 255],
    ],
    thinking: [
      [0, 0, 0, 0], [40, 6, 14, 44], [90, 14, 38, 110], [140, 28, 70, 172],
      [190, 62, 116, 214], [228, 128, 170, 238], [248, 206, 224, 250], [255, 255, 255, 255],
    ],
    speaking: [
      [0, 0, 0, 0], [40, 40, 4, 34], [90, 104, 12, 90], [140, 164, 26, 142],
      [190, 212, 70, 190], [228, 238, 134, 222], [248, 250, 208, 244], [255, 255, 255, 255],
    ],
    // A run that answered with nothing drops the hue: grey with a trace of
    // blue, which separates it from a canvas that failed to paint.
    failed: [
      [0, 0, 0, 0], [40, 36, 38, 40], [90, 88, 90, 94], [140, 138, 140, 144],
      [190, 186, 188, 192], [228, 222, 224, 226], [248, 244, 245, 246], [255, 255, 255, 255],
    ],
  };

  // BLEND is how long one palette takes to become the next, MORPH how long
  // one formation takes.
  var BLEND = 0.4;
  var MORPH = 1.6;

  // Each orb walks a plane turned by turn degrees, once every secs, starting
  // at the fraction at of the turn. A negative secs walks it the other way.
  // Its distance from the middle swings between low of R and R every swing
  // seconds, starting at the fraction swingAt of that, and sway rocks the
  // angle by that many radians every swaySecs, so the path draws a rosette
  // whose lobes move. None of the periods share a factor, and the three
  // swingAt hold the three orbs off one radius at a time. tilt and size are
  // solved below.
  var ORBS = [
    // An R of nothing holds this one at the centre. A tail of two is the
    // shortest there is: it holds still in two of the formations, where
    // fourteen samples of a trail would land on it and nowhere else.
    { R: 0, turn: 0, secs: 1, size: 0, at: 0, tail: 2, pulse: 0.12, pulseSecs: 4.5, meter: 0.4 },
    { R: 72, turn: 15, secs: 6.3, size: 34, at: 0, low: 0.55, swing: 9.7, swingAt: 0, sway: 0.35, swaySecs: 5.1 },
    { R: 66, turn: 70, secs: 8.8, size: 26, at: 0.11, low: 0.45, swing: 7.3, swingAt: 0.38, sway: 0.5, swaySecs: 6.7 },
    // dive takes this one through 0,0,0 and back out past the middle orb, so
    // the pair merges at the inner end of every swing and parts again. Its R
    // is the shortest of the three, about four fifths of the longest, and its
    // tilt is written down: no tilt clears a path that reaches the centre.
    { R: 58, turn: 120, secs: -6.9, size: 32, at: 0.72, low: 0, dive: true, tilt: 34, swing: 8.6, swingAt: 0.71, sway: 0.28, swaySecs: 4.3 },
    // only names the one formation these two are drawn in. Everywhere else
    // their radius is nothing, so they grow out of the middle as the ring
    // forms and shrink back into it when it breaks.
    { R: 64, turn: 200, secs: 7.7, size: 24, at: 0.3, low: 0.5, swing: 9.1, swingAt: 0.2, sway: 0.3, swaySecs: 5.9, only: "answered" },
    { R: 70, turn: 300, secs: -8.2, size: 22, at: 0.6, low: 0.5, swing: 8.3, swingAt: 0.55, sway: 0.4, swaySecs: 6.1, only: "answered" },
  ];

  // MOVERS are the orbs every formation draws, and RINGERS how many the ring
  // has: the two counts are what space a formation's orbs around it.
  var MOVERS = ORBS.filter(function (orb) { return orb.R && !orb.only; });
  var RINGERS = ORBS.length - 1;

  // NECK is the field wanted at the thinnest point between two orbs at their
  // closest. bridge solves the distance that leaves it, and lean tilts each
  // orbit so it comes no nearer than that to the middle.
  var NECK = 96;

  // LEADX and LEADY are the head's sweep, twice across the canvas for once
  // down. A circular path turns tighter than the chain is long, and the
  // followers then spread around the circle. DRIFT shares no factor with
  // LEAD, so the path takes minutes to close on itself.
  var LEAD = 4.3;
  var LEADX = 76;
  var LEADY = 68;
  var DRIFT = 11.3;

  // The ring is flat and behind the middle: every z on it is negative, so its
  // orbs draw smaller than they do anywhere else and their screen y moves
  // with the perspective. RINGY is solved below, short enough that the ring
  // runs over the middle orb at the front and the back of the turn.
  var RING = 6;
  var RINGR = 79;
  var RINGZ = -86;
  var RINGY = 0;

  // One period each for the formations the three take while the instance
  // works, and the reach of the two that leave the middle.
  var COIL = 7.4;
  var SCAN = 3.1;
  var SCANY = 44;
  var SWARM = 3.2;
  var SWARMT = 1.1;
  var BEAT = 1.9;
  var BEATR = 96;
  var LOST = 16;
  var FAR = 84;

  // near is how much closer to the camera a z puts something.
  function near(z) {
    return DEPTH / (DEPTH - z);
  }

  // waist returns the least field along the line between two orbs this far
  // apart.
  function waist(d, r1, r2) {
    var least = Infinity;
    for (var i = 0; i <= 40; i++) {
      var a = (d * i / 40) / r1;
      var b = (d - d * i / 40) / r2;
      var v = PEAK * ((a < 1 ? kernel(a) : 0) + (b < 1 ? kernel(b) : 0));
      if (v < least) least = v;
    }
    return least;
  }

  // bridge returns how far apart two orbs leave NECK between them. The waist
  // only falls as they part, so halving the span they can be apart finds it.
  function bridge(r1, r2) {
    var lo = 0;
    var hi = r1 + r2;
    for (var n = 0; n < 24; n++) {
      var d = (lo + hi) / 2;
      if (waist(d, r1, r2) > NECK) lo = d;
      else hi = d;
    }
    return (lo + hi) / 2;
  }

  // nearest returns the smallest screen distance from the middle an orbit of
  // this tilt reaches. It falls at the far side of the turn, where the
  // perspective divisor is smallest.
  function nearest(R, tilt) {
    var least = Infinity;
    for (var i = 0; i < 180; i++) {
      var th = Math.PI * i / 90;
      var x = R * Math.cos(th);
      var y = R * Math.sin(th) * Math.cos(tilt);
      var m = Math.sqrt(x * x + y * y) * near(R * Math.sin(th) * Math.sin(tilt));
      if (m < least) least = m;
    }
    return least;
  }

  // lean returns the steepest tilt in degrees whose orbit still clears the
  // middle by want. A plane near edge on crosses the middle on screen, and
  // the two orbs draw as one body for the whole turn.
  function lean(R, want) {
    if (nearest(R, 0) <= want) return 0;
    var lo = 0;
    var hi = Math.PI / 2;
    for (var n = 0; n < 20; n++) {
      var t = (lo + hi) / 2;
      if (nearest(R, t) > want) lo = t;
      else hi = t;
    }
    return (lo + hi) / 2 * 180 / Math.PI;
  }

  // The orb in the middle is the mean of the three that travel. Each of those
  // is held off it by bridge's distance, by raising low until the inner end
  // of the swing clears it and tilting the plane until the far side of the
  // turn does too. lift is the perspective at the near end of the orbit,
  // which every sample of the orb is drawn at, so depth moves an orb and
  // holds its size. Size and distance each read the other, so two passes
  // settle them.
  ORBS.forEach(function (orb) {
    orb.lift = 1;
    orb.path = [];
  });
  for (var pass = 0; pass < 2; pass++) {
    ORBS[0].size = MOVERS.reduce(function (sum, orb) {
      return sum + orb.size * orb.lift;
    }, 0) / MOVERS.length;
    ORBS.slice(1).forEach(function (orb) {
      orb.clear = bridge(ORBS[0].size, orb.size);
      if (!orb.dive) {
        orb.low = Math.max(orb.low, Math.min(1, orb.clear / orb.R));
        orb.tilt = lean(orb.R * orb.low, orb.clear);
      }
      orb.lift = near(orb.R * Math.sin(orb.tilt * Math.PI / 180));
    });
  }

  // The ring hangs a tenth over the middle orb's radius below the middle,
  // near enough that the two merge at the front and the back of every turn
  // and far enough that the ring still reads as the lower plane.
  RINGY = ORBS[0].size * 1.1;

  // ramp returns a palette as 256 entries of red, green, blue and alpha,
  // interpolated between the stops, with the alpha fading in over RISE.
  function ramp(stops) {
    var map = new Uint8ClampedArray(256 * 4);
    for (var v = 0; v < 256; v++) {
      var i = 0;
      while (i < stops.length - 2 && v > stops[i + 1][0]) i++;
      var lo = stops[i];
      var hi = stops[i + 1];
      var t = (v - lo[0]) / (hi[0] - lo[0]);
      map[v * 4] = lo[1] + (hi[1] - lo[1]) * t;
      map[v * 4 + 1] = lo[2] + (hi[2] - lo[2]) * t;
      map[v * 4 + 2] = lo[3] + (hi[3] - lo[3]) * t;
      map[v * 4 + 3] = v <= FLOOR ? 0 : 255 * Math.min(1, (v - FLOOR) / RISE);
    }
    return map;
  }

  // kernel is the field one orb puts in at a fraction of its radius:
  // Wyvill's polynomial, 1 at the middle and 0 at the rim, with no corner at
  // either. A straight gradient sums to a waist where two orbs meet; this
  // bulges towards the other orb and rounds the pair into one blob.
  function kernel(q) {
    if (q <= CORE) return 1;
    var d = (q - CORE) / (1 - CORE);
    var d2 = d * d;
    return 1 - (22 / 9) * d2 + (17 / 9) * d2 * d2 - (4 / 9) * d2 * d2 * d2;
  }

  // sprite returns the one orb as a canvas of that field in grey, built once
  // and scaled for every orb.
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

  // wide returns an orb's radius now: its size, its pulse, the meter and the
  // press, and nothing at all for an orb outside the one formation it is in.
  function wide(orb, now) {
    if (orb.only && drawing !== FORMS[orb.only]) return 0;
    var r = orb.size * orb.lift;
    if (orb.pulse) r *= 1 + orb.pulse * Math.sin(TWO * now / orb.pulseSecs);
    return r * swell * (orb.meter ? 1 + vu * orb.meter : 1);
  }

  // spin is the formation of an idle window: the middle orb holds at the
  // centre and each of the others walks its own tilted circle.
  function spin(orb, i, now) {
    if (!orb.R) return { x: MID, y: MID, r: wide(orb, now) };
    var th = TWO * (now / orb.secs + orb.at) + orb.sway * Math.sin(TWO * now / orb.swaySecs);
    var tilt = orb.tilt * Math.PI / 180;
    var turn = orb.turn * Math.PI / 180;
    var reach = 0.5 + 0.5 * Math.sin(TWO * (now / orb.swing + orb.swingAt));
    var rad = orb.R * (orb.low + (1 - orb.low) * reach);
    var x = rad * Math.cos(th);
    var y = rad * Math.sin(th) * Math.cos(tilt);
    var look = near(rad * Math.sin(th) * Math.sin(tilt));
    return {
      x: MID + (x * Math.cos(turn) - y * Math.sin(turn)) * look,
      y: MID + (x * Math.sin(turn) + y * Math.cos(turn)) * look,
      r: wide(orb, now),
    };
  }

  // snake is the formation from the first word to the reply: the middle orb
  // leads and each of the others sits a gap back along the path it left.
  function snake(orb, i, now) {
    var r = wide(orb, now);
    if (i === 0) {
      var drift = Math.sin(TWO * now / DRIFT);
      return {
        x: MID + LEADX * Math.sin(TWO * now / LEAD) * (0.8 + 0.2 * drift),
        y: MID + LEADY * Math.sin(Math.PI * now / LEAD + drift),
        r: r,
      };
    }
    var gap = 0;
    for (var k = 1; k <= i; k++) {
      gap += ORBS[k - 1].size * ORBS[k - 1].lift + ORBS[k].size * ORBS[k].lift;
    }
    var spot = behind(ORBS[0], gap);
    return { x: spot ? spot.x : MID, y: spot ? spot.y : MID, r: r };
  }

  // ring is the formation once a reply is in: the middle orb returns to the
  // centre and the others turn on one flat circle below it.
  function ring(orb, i, now) {
    var r = wide(orb, now);
    if (i === 0) return { x: MID, y: MID, r: r };
    var th = TWO * now / RING + (i - 1) * TWO / RINGERS;
    var look = near(RINGZ + RINGR * Math.sin(th));
    return {
      x: MID + RINGR * Math.cos(th) * look,
      y: MID + RINGY * look,
      r: r * look,
    };
  }

  // coil is the formation of a held recording: the chain curls onto one
  // upright circle around the middle orb and turns slowly on it.
  function coil(orb, i, now) {
    var r = wide(orb, now);
    if (i === 0) return { x: MID, y: MID, r: r };
    var th = TWO * now / COIL + (i - 1) * TWO / MOVERS.length;
    return { x: MID + orb.clear * Math.cos(th), y: MID + orb.clear * Math.sin(th), r: r };
  }

  // scan is the formation while the recording is transcribed: the three stack
  // into a column and sweep across the canvas, through the middle orb.
  function scan(orb, i, now) {
    var r = wide(orb, now);
    if (i === 0) return { x: MID, y: MID, r: r };
    var walk = (now / SCAN) % 1;
    return { x: -SPRITE + walk * (SIZE + SPRITE * 2), y: MID + (i - 2) * SCANY, r: r };
  }

  // swarm is the formation while Claude holds the question: the three turn
  // fast and close around the middle orb, each on its own tilted plane.
  function swarm(orb, i, now) {
    var r = wide(orb, now);
    if (i === 0) return { x: MID, y: MID, r: r };
    // Each takes its own period, its own tilt, a plane that turns and a
    // radius that breathes, all off the orb's own numbers. Three on one
    // period at one tilt draw a mechanism turning.
    var th = TWO * now / (SWARM * (0.8 + 0.25 * i)) + (i - 1) * TWO / MOVERS.length
      + orb.sway * Math.sin(TWO * now / orb.swaySecs);
    var turn = (i - 1) * Math.PI / 3 + 0.4 * Math.sin(TWO * now / orb.swing);
    var tilt = SWARMT * (0.65 + 0.22 * i);
    var rad = orb.clear * (1 + 0.18 * Math.sin(TWO * (now / orb.swing + orb.swingAt)));
    var x = rad * Math.cos(th);
    var y = rad * Math.sin(th) * Math.cos(tilt);
    var look = near(rad * Math.sin(th) * Math.sin(tilt));
    return {
      x: MID + (x * Math.cos(turn) - y * Math.sin(turn)) * look,
      y: MID + (x * Math.sin(turn) + y * Math.cos(turn)) * look,
      r: r * look,
    };
  }

  // beat is the formation while the instance synthesises speech: each of the
  // three runs out from the middle orb along its own spoke, swelling from
  // nothing and back to nothing, one after the other.
  function beat(orb, i, now) {
    var r = wide(orb, now);
    if (i === 0) return { x: MID, y: MID, r: r };
    var walk = (now / BEAT + (i - 1) / MOVERS.length) % 1;
    var th = (i - 1) * TWO / MOVERS.length + TWO * now / LOST;
    return {
      x: MID + walk * BEATR * Math.cos(th),
      y: MID + walk * BEATR * Math.sin(th),
      r: r * Math.sin(Math.PI * walk),
    };
  }

  // scatter is the formation of a run that answered with nothing: the three
  // stand off at the rim and turn there, and the middle orb is left alone.
  function scatter(orb, i, now) {
    var r = wide(orb, now);
    if (i === 0) return { x: MID, y: MID, r: r };
    var th = TWO * now / LOST + (i - 1) * TWO / MOVERS.length;
    return { x: MID + FAR * Math.cos(th), y: MID + FAR * Math.sin(th), r: r };
  }

  // One formation per state of the window. voice.js names the state on every
  // step of the run, so the three change flight path as the work moves on.
  var FORMS = {
    idle: spin,
    recording: snake,
    paused: coil,
    transcribing: scan,
    thinking: swarm,
    speaking: beat,
    answered: ring,
    failed: scatter,
  };

  // remember keeps where an orb is, and drops what is older than SPAN.
  function remember(orb, secs, spot) {
    orb.path.push({ t: secs, x: spot.x, y: spot.y, r: spot.r });
    while (orb.path.length > 2 && orb.path[0].t < secs - SPAN) orb.path.shift();
  }

  // at returns an orb's place at a time, read out of its path.
  function at(orb, secs) {
    var p = orb.path;
    if (!p.length) return null;
    if (secs >= p[p.length - 1].t) return p[p.length - 1];
    for (var i = p.length - 1; i > 0; i--) {
      if (p[i - 1].t > secs) continue;
      var span = p[i].t - p[i - 1].t;
      var f = span > 0 ? (secs - p[i - 1].t) / span : 0;
      return {
        x: p[i - 1].x + (p[i].x - p[i - 1].x) * f,
        y: p[i - 1].y + (p[i].y - p[i - 1].y) * f,
        r: p[i - 1].r + (p[i].r - p[i - 1].r) * f,
      };
    }
    return p[0];
  }

  // behind returns the point a distance back along an orb's path. Walking it
  // by length holds the gap as the head's speed changes.
  function behind(orb, want) {
    var p = orb.path;
    if (p.length < 2) return p[0] || null;
    var gone = 0;
    for (var i = p.length - 1; i > 0; i--) {
      var dx = p[i].x - p[i - 1].x;
      var dy = p[i].y - p[i - 1].y;
      var step = Math.sqrt(dx * dx + dy * dy);
      if (gone + step >= want) {
        var f = step > 0 ? (want - gone) / step : 0;
        return { x: p[i].x - dx * f, y: p[i].y - dy * f, r: p[i].r };
      }
      gone += step;
    }
    return p[0];
  }

  // clock is the time every part of this reads, in seconds.
  function clock() {
    return window.performance.now() / 1000;
  }

  // One palette per state, so a step of the run changes the colour and the
  // flight path together.
  var PALETTE = {};
  Object.keys(MOODS).forEach(function (name) {
    PALETTE[name] = ramp(MOODS[name]);
  });
  var mix = new Uint8ClampedArray(256 * 4);
  var from = PALETTE.idle;
  var into = PALETTE.idle;
  var blendAt = -1;
  var tapAt = 0;
  var form = FORMS.idle;   // the formation the orbs are taking
  var fore = FORMS.idle;   // the one they are leaving
  var drawing = FORMS.idle;  // the one being asked where an orb goes
  var morphAt = -1;
  var ORB = null;    // the greyscale sprite, made on the first frame
  var show = null;   // the canvas on the page
  var paint = null;  // the field the orbs are summed into, offscreen
  var out = null;    // the page canvas's context
  var raf = 0;
  var level = 0;
  var vu = 0;
  var swell = 1;

  // place asks a formation where an orb goes, and names it for wide.
  // An orb of one formation alone waits at the centre of every other, so it
  // grows out of the middle rather than flying in from off the canvas.
  function place(shape, orb, i, secs) {
    drawing = shape;
    if (orb.only && FORMS[orb.only] !== shape) return { x: MID, y: MID, r: 0 };
    return shape(orb, i, secs);
  }

  // draw sums the orbs and their trails into the field, reads the sum
  // through the palette and puts the colour on the page's canvas.
  function draw(secs) {
    // The meter takes a rise at once and falls 6% a frame, so a word swells
    // the middle orb and silence lets it down over about a third of a second.
    vu = level > vu ? level : vu + (level - vu) * 0.06;
    // A press swells every orb by a quarter, decaying over 0.18s.
    swell = tapAt ? 1 + 0.25 * Math.exp(-(secs - tapAt) / 0.18) : 1;
    if (swell < 1.002) {
      tapAt = 0;
      swell = 1;
    }
    // One formation becomes the next over MORPH seconds, eased at both ends.
    // Both are evaluated every frame while w is under 1.
    var w = 1;
    if (morphAt >= 0) {
      w = (secs - morphAt) / MORPH;
      if (w < 0) w = 0;
      if (w >= 1) {
        morphAt = -1;
        w = 1;
      } else {
        w = w * w * (3 - 2 * w);
      }
    }
    var i, orb;
    for (i = 0; i < ORBS.length; i++) {
      orb = ORBS[i];
      var goes = place(form, orb, i, secs);
      if (w < 1) {
        var came = place(fore, orb, i, secs);
        goes = {
          x: came.x + (goes.x - came.x) * w,
          y: came.y + (goes.y - came.y) * w,
          r: came.r + (goes.r - came.r) * w,
        };
      }
      remember(orb, secs, goes);
    }
    paint.globalCompositeOperation = "source-over";
    paint.globalAlpha = 1;
    paint.fillStyle = "#000000";
    paint.fillRect(0, 0, SIZE, SIZE);
    paint.globalCompositeOperation = "lighter";
    for (i = 0; i < ORBS.length; i++) {
      orb = ORBS[i];
      var tail = orb.tail || TRAIL;
      var sum = 0;
      for (var s = 1; s < tail; s++) sum += Math.pow(1 - s / tail, 3);
      for (var j = tail - 1; j >= 0; j--) {
        var back = 1 - j / tail;
        paint.globalAlpha = j === 0 ? 1 : WAKE * back * back * back / sum;
        var spot = at(orb, secs - j * STEP);
        if (!spot) continue;
        var r = spot.r * (0.45 + 0.55 * back);
        paint.drawImage(ORB, spot.x - r, spot.y - r, r * 2, r * 2);
      }
    }
    paint.globalAlpha = 1;
    var lut = into;
    if (blendAt >= 0) {
      var t = (secs - blendAt) / BLEND;
      if (t >= 1) {
        blendAt = -1;
      } else {
        for (var k = 0; k < mix.length; k++) mix[k] = from[k] + (into[k] - from[k]) * t;
        lut = mix;
      }
    }
    var frame = paint.getImageData(0, 0, SIZE, SIZE);
    var data = frame.data;
    for (var p = 0; p < data.length; p += 4) {
      var v = data[p] * 4;
      data[p] = lut[v];
      data[p + 1] = lut[v + 1];
      data[p + 2] = lut[v + 2];
      data[p + 3] = lut[v + 3];
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
      // A window that opens again starts from nothing: a kept path draws a
      // trail from wherever the last one ended, and a kept meter draws the
      // centre orb at the size the last loud word left it.
      ORBS.forEach(function (orb) { orb.path = []; });
      level = vu = 0;
      swell = 1;
      tapAt = 0;
      morphAt = -1;
      form = fore = drawing = FORMS.idle;
      from = into = PALETTE.idle;
      blendAt = -1;
      paint.globalCompositeOperation = "source-over";
      paint.fillStyle = "#000000";
      paint.fillRect(0, 0, SIZE, SIZE);
      if (still) {
        draw(0);
        return;
      }
      raf = window.requestAnimationFrame(tick);
    },
    // stop ends the loop, for a window that has closed.
    stop: function () {
      if (raf) window.cancelAnimationFrame(raf);
      raf = 0;
    },
    // level is how loud the microphone is, 0 to 1. It sets the size of the
    // orb at the centre.
    level: function (loud) {
      level = loud;
    },
    // tap swells every orb for a moment and settles it back.
    tap: function () {
      tapAt = clock();
      if (still) draw(tapAt);
    },
    // mood is what the window is doing, one of the keys of FORMS. It picks
    // the palette and the formation, and fades over BLEND and MORPH from the
    // pair in use.
    mood: function (name) {
      var next = PALETTE[name] || PALETTE.idle;
      var shape = FORMS[name] || FORMS.idle;
      if (next !== into) {
        from = blendAt >= 0 ? mix.slice(0) : into;
        into = next;
        blendAt = clock();
      }
      if (shape !== form) {
        fore = form;
        form = shape;
        morphAt = clock();
      }
      if (still) draw(clock() + MORPH);
    },
  };
})();
