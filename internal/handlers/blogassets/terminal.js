// Vanilla port of frontend/src/components/terminal-bar.tsx for the
// server-rendered blog pages, so the header matches the SPA's. Keep the
// COMMANDS table and easter eggs in sync with the React component.
// Output is always written with textContent — never innerHTML.
(function () {
  var COMMANDS = {
    help:    { description: 'List available commands', action: 'response', value: '' },
    home:    { description: 'Go home', action: 'navigate', value: '/' },
    about:   { description: 'About me', action: 'navigate', value: '/about' },
    blog:    { description: 'Read the blog', action: 'navigate', value: '/blog' },
    contact: { description: 'Get in touch', action: 'response', value: 'Email me: dj.thompson715@gmail.com' },
    whoami:  { description: '???', action: 'response', value: 'Dillon Thompson — Senior Full Stack Engineer & Technical Consultant' },
    clear:   { description: 'Clear output', action: 'response', value: '' }
  };

  var bar = document.getElementById('term-bar');
  var input = document.getElementById('term-input');
  var out = document.getElementById('term-out');
  if (!bar || !input || !out) return;

  var history = [];      // {type, text}
  var cmdHistory = [];   // most recent first
  var historyIdx = -1;
  var helpShown = false;

  function helpText() {
    var lines = Object.keys(COMMANDS).map(function (cmd) {
      return '  ' + (cmd + '            ').slice(0, 12) + ' ' + COMMANDS[cmd].description;
    });
    return 'Available commands (type one and hit enter):\n' + lines.join('\n');
  }

  function render() {
    out.textContent = '';
    history.forEach(function (line) {
      var pre = document.createElement('pre');
      pre.textContent = line.text;
      if (line.type === 'input') pre.className = 'term-line-input';
      if (line.type === 'error') pre.className = 'term-line-error';
      out.appendChild(pre);
    });
    out.hidden = history.length === 0;
    out.scrollTop = out.scrollHeight;
  }

  function push(type, text) { history.push({ type: type, text: text }); }

  function execute(raw) {
    var trimmed = raw.trim().toLowerCase();
    if (!trimmed) return;
    push('input', '> ' + trimmed);
    cmdHistory.unshift(trimmed);
    historyIdx = -1;

    if (trimmed === 'clear') { history = []; render(); return; }
    if (trimmed === 'help') { push('output', helpText()); render(); return; }
    if (trimmed === 'sudo hire dillon') {
      push('output', "Great choice. Let's talk → dj.thompson715@gmail.com"); render(); return;
    }
    if (trimmed === 'neofetch') {
      push('output', [
        '  dillon@portfolio',
        '  ─────────────────',
        '  OS:     macOS / Linux',
        '  Stack:  Go, React, Postgres',
        '  Editor: VS Code + Claude Code',
        '  Shell:  fish',
        '  Theme:  dark'
      ].join('\n'));
      render(); return;
    }

    var cmd = Object.prototype.hasOwnProperty.call(COMMANDS, trimmed) ? COMMANDS[trimmed] : null;
    if (!cmd) {
      push('error', "command not found: " + trimmed + ". Type 'help' for available commands.");
      render(); return;
    }
    if (cmd.action === 'navigate') {
      push('output', '→ navigating to ' + cmd.value);
      render();
      // Full page load: these pages are served by Go, the others by the SPA.
      setTimeout(function () { window.location.assign(cmd.value); }, 300);
      return;
    }
    if (cmd.value) { push('output', cmd.value); }
    render();
  }

  function open() {
    if (!helpShown) { helpShown = true; push('output', helpText()); }
    render();
  }

  bar.addEventListener('click', function () { input.focus(); });
  input.addEventListener('focus', open);

  input.addEventListener('keydown', function (e) {
    if (e.key === 'Enter') {
      execute(input.value);
      input.value = '';
    } else if (e.key === 'Escape') {
      out.hidden = true;
      input.blur();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (cmdHistory.length > 0) {
        historyIdx = Math.min(historyIdx + 1, cmdHistory.length - 1);
        input.value = cmdHistory[historyIdx];
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIdx > 0) {
        historyIdx -= 1;
        input.value = cmdHistory[historyIdx];
      } else {
        historyIdx = -1;
        input.value = '';
      }
    } else if (e.key === 'Tab') {
      e.preventDefault();
      var prefix = input.value.toLowerCase();
      var match = Object.keys(COMMANDS).filter(function (c) { return c.indexOf(prefix) === 0; })[0];
      if (match) input.value = match;
    }
  });

  window.addEventListener('keydown', function (e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      input.focus();
    }
  });
})();
