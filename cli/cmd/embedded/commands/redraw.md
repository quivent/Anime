# /redraw — Take a WebGL screenshot and describe it

Takes a Metal-backed WebGL screenshot of the current scene and describes what you see.

## Usage
`/redraw` — screenshot `/virgo/port` and describe
`/redraw virgo` — screenshot `/virgo` (polygon reference)
`/redraw both` — screenshot both and compare

## Protocol

```js
// Node.js puppeteer snippet — run from /tmp
const puppeteer = require('puppeteer');
const browser = await puppeteer.launch({
  headless: 'new',
  args: ['--no-sandbox', '--use-angle=metal', '--ignore-gpu-blocklist', '--enable-gpu']
});
const page = await browser.newPage();
await page.setViewport({ width: 1280, height: 720 });
await page.setCacheEnabled(false);
await page.goto('http://localhost:7180/virgo/port', { waitUntil: 'networkidle0', timeout: 20000 });
await new Promise(r => setTimeout(r, 4000));
await page.screenshot({ path: '/tmp/redraw.png' });
await browser.close();
```

Then read `/tmp/redraw.png` and describe the scene precisely in English before proposing changes.
