Wire up a browser console relay so errors appear in the terminal instead of DevTools.

Usage: Run /browser-relay in any project doing browser work. Patches the dev server to POST console.error/warn + unhandled errors to /__log, which prints them to stdout. Never ask the user to open DevTools again.

# Browser Console Relay Setup

## What this does
Injects a small script into every HTML response that POSTs `console.error`, `console.warn`, and unhandled errors/rejections to `/__log` on the dev server. The server prints them to stdout with color prefixes. You watch the terminal; the user never touches DevTools.

## Step 1: Find the dev server

Look for the project's dev server file:
```bash
ls *.mjs *.js server* serve* | head -20
```

Common locations: `serve.mjs`, `server.mjs`, `dev.mjs`, `index.js`, `server.js`

## Step 2: Check if relay already exists

```bash
grep -l "__log\|RELAY_SCRIPT\|browser-relay" *.mjs *.js 2>/dev/null
```

If found, skip to Step 4.

## Step 3: Patch the server

Add the relay script constant and /__log endpoint. The exact insertion point depends on the framework, but the pattern is always:

**Relay script (inject into every HTML response after `<head>`):**
```js
const RELAY_SCRIPT = `<script>
(function(){
  const _post = (lvl, args) => {
    try { fetch('/__log', {method:'POST', headers:{'content-type':'application/json'},
      body: JSON.stringify({lvl, msg: args.map(a => typeof a === 'object' ? JSON.stringify(a) : String(a)).join(' ')})
    }); } catch(e){}
  };
  ['error','warn'].forEach(lvl => {
    const orig = console[lvl].bind(console);
    console[lvl] = (...a) => { orig(...a); _post(lvl, a); };
  });
  window.addEventListener('error', e => _post('error', [e.message, e.filename+':'+e.lineno]));
  window.addEventListener('unhandledrejection', e => _post('error', ['unhandledrejection', String(e.reason)]));
})();
</script>`;
```

**/__log endpoint (add to request handler before file serving):**
```js
if (rawPath === '/__log' && req.method === 'POST') {
  let body = '';
  req.on('data', c => body += c);
  req.on('end', () => {
    try {
      const { lvl, msg } = JSON.parse(body);
      const prefix = lvl === 'error' ? '\x1b[31m[browser error]\x1b[0m' : '\x1b[33m[browser warn]\x1b[0m';
      console.log(prefix, msg);
    } catch {}
    res.writeHead(204); res.end();
  });
  return;
}
```

**HTML injection (wrap the HTML file-serving block):**
```js
if (ext === '.html') {
  fs.readFile(fp, 'utf8', (err2, html) => {
    if (err2) { res.writeHead(500); return res.end('read error'); }
    const injected = html.replace('<head>', '<head>' + RELAY_SCRIPT);
    res.writeHead(200, { 'content-type': MIME[ext], 'cache-control': 'no-store' });
    res.end(injected);
  });
  return;
}
```

## Step 4: Restart the server

Kill the old server process and restart:
```bash
pkill -f "node.*serve" 2>/dev/null
pkill -f "python.*http" 2>/dev/null
sleep 0.5
node serve.mjs &
```

Verify it's up:
```bash
sleep 0.5 && curl -s -o /dev/null -w "%{http_code}" http://localhost:$(grep -o 'PORT.*||.*[0-9]*' serve.mjs | grep -o '[0-9]*$' || echo 8080)/
```

## Step 5: Confirm relay is active

```bash
curl -s http://localhost:8080/index.html | grep -c "__log"
```

Should return `1`. If 0, the injection missed — check that the HTML files contain `<head>` (lowercase).

## Rules going forward

- Python http.server: kill it, never use it for this project
- `console.error` in terminal = fix it; don't ask user to check DevTools
- Shader compile errors appear as `[browser error] THREE.WebGLProgram: Shader Error...`
- Module load failures appear as `[browser error] SyntaxError: ...`
