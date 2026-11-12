import { chromium } from 'playwright';
import assert from 'node:assert/strict';
import { mkdirSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';

const base = process.env.AIFT_BROWSER_BASE || 'http://127.0.0.1:8765';
const prefix = process.env.AIFT_UXI_PREFIX || '';
const out = process.env.AIFT_SCREENSHOT_DIR || 'uxi-browser-evidence';
mkdirSync(out, { recursive: true });
const browser = await chromium.launch({ headless: true });
const findings = [];
const pages = [['observatory', 'living-universe-2.html'], ['expedition', 'living-universe-expedition.html']];
try {
  for (const width of [375, 1280]) {
    const context = await browser.newContext({
      viewport: { width, height: 900 },
      reducedMotion: 'reduce',
      colorScheme: 'dark'
    });
    for (const [name, path] of pages) {
      const page = await context.newPage();
      const errors = [];
      page.on('pageerror', e => errors.push(e.message));
      page.on('console', m => { if (m.type() === 'error') errors.push('console: ' + m.text()); });
      const response = await page.goto(new URL(prefix + path, base + '/').href, { waitUntil: 'load', timeout: 30000 });
      assert.equal(response?.status(), 200, name + ' HTTP status');
      const body = await page.locator('body').innerText();
      assert.ok(body.length > 300, name + ' substantive content');
      assert.match(body, /root/i, name + ' root vocabulary');
      assert.match(body, /evidence|provenance/i, name + ' evidence vocabulary');
      assert.deepEqual(await page.locator('a[href]').evaluateAll(nodes =>
        nodes.filter(a => a.href.startsWith('https:') && a.target === '_blank' &&
          (!a.rel.includes('noopener') || !a.rel.includes('noreferrer'))).map(a => a.href)
      ), [], name + ' safe external links');
      const buttons = page.locator('button:visible');
      if (await buttons.count()) {
        await buttons.first().focus();
        assert.ok(await buttons.first().evaluate(e => document.activeElement === e), name + ' keyboard focus');
      }
      const before = await page.evaluate(() => ({
        documentWidth: document.documentElement.scrollWidth,
        viewportWidth: innerWidth,
        motion: matchMedia('(prefers-reduced-motion: reduce)').matches
      }));
      assert.equal(before.motion, true, name + ' reduced-motion preference');
      // Horizontal overflow is recorded, not blindly failed: the constellation intentionally scrolls.
      await page.screenshot({ path: join(out, name + '-' + width + '-baseline.png'), fullPage: true, animations: 'disabled' });
      if (name === 'expedition') {
        assert.match(await page.locator('#status').innerText(), /no evidence|empty|import/i, 'fail-closed empty state');
        await page.locator('#search').fill('nonexistent-record');
        await page.locator('#zoom').fill('140');
        assert.match(await page.locator('#zoom-value').innerText(), /140/, 'zoom feedback');
        await page.locator('#river-search').fill('nonexistent-record');
        await page.locator('#river-filter').selectOption('gaps');
        await page.locator('#quest-search').fill('nonexistent-record');
        await page.screenshot({ path: join(out, name + '-' + width + '-interaction.png'), fullPage: true, animations: 'disabled' });
        await page.locator('#river-reset').click();
        assert.equal(await page.locator('#river-search').inputValue(), '', 'river reset clears search');
        await page.locator('#clear').click();
        assert.match(await page.locator('#status').innerText(), /no evidence|empty|clear|import/i, 'clear stays fail-closed');
      } else {
        await page.locator('#search').fill('nonexistent-record');
        await page.screenshot({ path: join(out, name + '-' + width + '-interaction.png'), fullPage: true, animations: 'disabled' });
      }
      assert.deepEqual(errors, [], name + ' browser errors');
      findings.push({ name, width, ...before, errors, screenshots: 2 });
      await page.close();
    }
    await context.close();
  }
  writeFileSync(join(out, 'acceptance-results.json'), JSON.stringify({
    scope: 'local checked-out pages, empty evidence state; not a live deployment or verified CI receipt',
    findings
  }, null, 2));
  console.log('Browser acceptance passed: two surfaces, mobile/desktop, interactive states, fail-closed behavior, keyboard focus, reduced motion, screenshots');
} finally {
  await browser.close();
}
