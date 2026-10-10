// Theme handling for server-rendered blog pages. Loaded as an external,
// blocking script in <head> (the CSP forbids inline scripts) so the stored
// theme is applied before first paint. Mirrors frontend/src/hooks/use-theme.ts:
// same localStorage key, dark by default, `dark` class on <html>.
(function () {
  var root = document.documentElement;
  function apply(theme) {
    root.classList.toggle('dark', theme === 'dark');
  }
  var stored = null;
  try { stored = localStorage.getItem('theme'); } catch (e) {}
  apply(stored === 'light' ? 'light' : 'dark');

  document.addEventListener('DOMContentLoaded', function () {
    var btn = document.getElementById('theme-toggle');
    if (!btn) return;
    btn.addEventListener('click', function () {
      var next = root.classList.contains('dark') ? 'light' : 'dark';
      apply(next);
      try { localStorage.setItem('theme', next); } catch (e) {}
    });
  });
})();
