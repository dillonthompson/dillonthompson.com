// Theme handling for server-rendered blog pages. Loaded as an external,
// blocking script in <head> (the CSP forbids inline scripts) so the stored
// theme is applied before first paint. Mirrors frontend/src/hooks/use-theme.ts
// and frontend/src/lib/theme-transition.ts: same localStorage key, `dark`
// (the sunset theme) by default, `dark` class on <html>; the toggle animates a
// sunrise/sunset View Transition when the browser supports it.
(function () {
  var root = document.documentElement;
  function apply(theme) {
    var dark = theme === 'dark';
    root.classList.toggle('dark', dark);
    // Keep the browser toolbar color and the toggle's label in step.
    var meta = document.querySelector('meta[name="theme-color"]');
    if (meta) meta.setAttribute('content', dark ? '#07162c' : '#d9efff');
    var btn = document.getElementById('theme-toggle');
    if (btn) btn.setAttribute('aria-label', dark ? 'Switch to sunrise theme' : 'Switch to sunset theme');
  }
  var stored = null;
  try { stored = localStorage.getItem('theme'); } catch (e) {}
  apply(stored === 'light' ? 'light' : 'dark');

  var DURATION_MS = 1300;
  // Sample the eased arc about every 8ms; keyframes interpolate linearly, so
  // coarser sampling makes the circle step as it follows the sun.
  var SAMPLES = Math.ceil(DURATION_MS / 8);

  // Where the toggle is headed (the class only flips a frame into the
  // transition, so quick double clicks must not read it), and the transition
  // that owns the root classes (a superseded one must not strip them).
  var target = root.classList.contains('dark') ? 'dark' : 'light';
  var active = null;

  function easeInOut(t) { return t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2; }

  // Mirrors runThemeTransition in frontend/src/lib/theme-transition.ts: the sun
  // arcs east to west (rise on the left, set on the right) and the theme wash is
  // a circle centred on it.
  function transitionTo(next) {
    var toSunset = next === 'dark';
    var reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (reduce || typeof document.startViewTransition !== 'function') {
      apply(next);
      return;
    }
    var token = {};
    active = token;

    var sun;
    var vt = document.startViewTransition(function () {
      // Scope the transition CSS to this animation (see blog.css).
      root.classList.add('theme-anim');
      root.classList.toggle('theme-to-sunset', toSunset);
      apply(next);
      sun = document.createElement('div');
      sun.id = 'theme-sun';
      sun.setAttribute('aria-hidden', 'true');
      document.body.appendChild(sun);
    });
    // fill: 'forwards' animations on <html> outlive the transition and would
    // stack onto the next one (choppier each toggle), so cancel them.
    var animations = [];
    function cleanup() {
      animations.forEach(function (a) { a.cancel(); });
      if (sun) sun.remove();
      if (active === token) {
        active = null;
        root.classList.remove('theme-anim', 'theme-to-sunset');
      }
    }
    vt.ready.then(function () {
      var w = window.innerWidth, h = window.innerHeight;
      var size = sun ? sun.offsetWidth : 224;
      var rx = w * 0.4, ry = h * 0.7;
      function sunAt(p) {
        var phi = ((toSunset ? 0.5 : 0) + p * 0.5) * Math.PI;
        return { x: w / 2 - rx * Math.cos(phi), y: h - ry * Math.sin(phi) + size * 0.2 };
      }
      function reach(x, y) {
        return Math.ceil(Math.hypot(Math.max(x, w - x), Math.max(y, h - y)));
      }
      var sunFrames = [], clipFrames = [];
      for (var i = 0; i <= SAMPLES; i++) {
        var p = easeInOut(i / SAMPLES);
        var c = sunAt(p);
        sunFrames.push({ transform: 'translate(' + (c.x - size / 2) + 'px, ' + (c.y - size / 2) + 'px)' });
        var r = reach(c.x, c.y) * (toSunset ? 1 - p : p);
        clipFrames.push({ clipPath: 'circle(' + r + 'px at ' + c.x + 'px ' + c.y + 'px)' });
      }
      // fill: 'forwards' holds the final state until teardown (no snap-back flicker).
      function opts(pseudo) {
        return { duration: DURATION_MS, easing: 'linear', fill: 'forwards', pseudoElement: pseudo };
      }
      animations.push(root.animate(clipFrames, opts(toSunset ? '::view-transition-old(root)' : '::view-transition-new(root)')));
      animations.push(root.animate(sunFrames, opts('::view-transition-group(theme-sun)')));
      animations.push(root.animate({ opacity: [0, 1, 1, 0], offset: [0, 0.15, 0.6, 1] }, opts('::view-transition-image-pair(theme-sun)')));
    }).catch(cleanup);
    vt.finished.then(cleanup, cleanup);
  }

  document.addEventListener('DOMContentLoaded', function () {
    var btn = document.getElementById('theme-toggle');
    if (!btn) return;
    apply(target);
    btn.addEventListener('click', function () {
      var next = target === 'dark' ? 'light' : 'dark';
      target = next;
      try { localStorage.setItem('theme', next); } catch (e) {}
      transitionTo(next);
    });
  });
})();
